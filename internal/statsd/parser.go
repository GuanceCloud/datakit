// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf/plugins/parsers/graphite"
)

// parser monitors the col.opts.in channel, if there is a packet ready, it parses the
// packet into statsd strings and then calls parseStatsdLine, which parses a
// single statsd metric into a struct.
func (col *Collector) parser(idx int) {
	for {
		select {
		case <-col.done:
			return

		case in := <-col.in:

			lines := strings.Split(in.Buffer.String(), "\n")

			col.bufPool.Put(in.Buffer)

			for _, line := range lines {
				line = strings.TrimSpace(line)
				col.opts.l.Debugf("[%d] statsd line: %s", idx, line)

				switch {
				case line == "": // pass
				case col.opts.dataDogExtensions && strings.HasPrefix(line, "_e"):
					if err := col.parseEventMessage(in.Time, line, in.Addr); err != nil {
						col.opts.l.Warnf("[%d] parseEventMessage: %s, ignored", idx, err.Error())
					}
				default:
					if err := col.parseStatsdLine(line); err != nil {
						col.opts.l.Warnf("[%d] parseEventMessage: %s, ignored", idx, err.Error())
					}
				}
			}
		}
	}
}

// parseStatsdLine will parse the given statsd line, validating it as it goes.
// If the line is valid, it will be cached for the next call to Gather().
func (col *Collector) parseStatsdLine(line string) error {
	lineTags := make(map[string]string)
	if col.opts.dataDogExtensions {
		recombinedSegments := make([]string, 0)
		// datadog tags look like this:
		//
		// users.online:1|c|@0.5|#country:china,environment:production
		// users.online:1|c|#sometagwithnovalue
		// namespace.test_gauge:21|g|#globalTags,globalTags2,tag1,tag2|T1658997712
		//
		// we will split on the pipe and remove any elements that are datadog
		// tags, parse them, and rebuild the line sans the datadog tags
		pipesplit := strings.Split(line, "|")
		for _, segment := range pipesplit {
			if len(segment) > 0 && segment[0] == '#' {
				// we have ourselves a tag; they are comma separated
				parseDataDogTags(lineTags, segment[1:])
			} else {
				recombinedSegments = append(recombinedSegments, segment)
			}
		}
		line = strings.Join(recombinedSegments, "|")
	}

	// Validate splitting the line on ":"
	bits := strings.Split(line, ":")
	if len(bits) < 2 {
		col.opts.l.Debugf("Splitting ':', unable to parse metric: %s", line)
		return nil
	}

	// Extract bucket name from individual metric bits
	bucketName, bits := bits[0], bits[1:]

	// Add a metric for each bit available
	for _, bit := range bits {
		m := metric{}

		m.bucket = bucketName

		// Validate splitting the bit on "|"
		pipesplit := strings.Split(bit, "|")
		if len(pipesplit) < 2 {
			col.opts.l.Debugf("splitting '|', unable to parse metric: %s, ignored", line)
			return nil
		} else if len(pipesplit) > 2 { // with sample rate
			sr := pipesplit[2]

			if strings.Contains(sr, "@") && len(sr) > 1 {
				samplerate, err := strconv.ParseFloat(sr[1:], 64)
				if err != nil {
					col.opts.l.Errorf("parsing sample rate error: %s, line: %s", err.Error(), line)
				} else {
					// sample rate successfully parsed
					m.samplerate = samplerate
				}
			} else {
				col.opts.l.Debugf("sample rate must be in format like: "+
					"@0.1, @0.5, etc. Ignoring sample rate for line: %s", line)
			}
		}

		// Validate metric type
		switch pipesplit[1] {
		case "g", "c", "s", "ms", "h", "d":
			m.mtype = pipesplit[1]
		default:
			col.opts.l.Debugf("metric type %q unsupported, line: %s", pipesplit[1], line)
			return nil
		}

		// Parse the value
		if strings.HasPrefix(pipesplit[0], "-") || strings.HasPrefix(pipesplit[0], "+") {
			if m.mtype != "g" && m.mtype != "c" {
				col.opts.l.Errorf("+- values are only supported for gauges and counters, unable to parse metric: %s", line)
				return errors.New("error parsing statsd line")
			}
			m.additive = true
		}

		switch m.mtype {
		case "g", "ms", "h", "d", "c":
			v, err := strconv.ParseFloat(pipesplit[0], 64)
			if err != nil {
				col.opts.l.Errorf("Parsing value to float64, unable to parse metric: %s", line)
				return errors.New("error parsing statsd line")
			}
			// If a sample rate is given with a counter, divide value by the rate
			if m.samplerate != 0 && m.mtype == "c" {
				v /= m.samplerate
			}
			m.floatvalue = v
		case "s":
			m.strvalue = pipesplit[0]
		}

		// Parse the name & tags from bucket
		m.name, m.field, m.tags = col.parseName(m.bucket)
		switch m.mtype {
		case "c":
			m.tags["metric_type"] = "counter"
		case "g":
			m.tags["metric_type"] = "gauge"
		case "s":
			m.tags["metric_type"] = "set"
		case "ms":
			m.tags["metric_type"] = "timing"
		case "h":
			m.tags["metric_type"] = "histogram"
		case "d":
			m.tags["metric_type"] = "distribution"
		}

		if len(lineTags) > 0 {
			for k, v := range lineTags {
				m.tags[k] = v
			}
		}

		// Make a unique key for the measurement name/tags
		var tg []string
		for k, v := range m.tags {
			tg = append(tg, k+"="+v)
		}
		sort.Strings(tg)
		tg = append(tg, m.name)
		m.hash = strings.Join(tg, "")

		col.aggregate(m)
	}

	return nil
}

// parseName parses the given bucket name with the list of bucket maps in the
// config file. If there is a match, it will parse the name of the metric and
// map of tags.
// Return values are (<name>, <field>, <tags>).
func (col *Collector) parseName(bucket string) (string, string, map[string]string) {
	col.Lock()
	defer col.Unlock()
	tags := make(map[string]string)

	bucketparts := strings.Split(bucket, ",")
	// Parse out any tags in the bucket
	if len(bucketparts) > 1 {
		for _, btag := range bucketparts[1:] {
			k, v := parseKeyValue(btag)
			if k != "" {
				tags[k] = v
			}
		}
	}

	var field string
	name := bucketparts[0]

	p := col.graphiteParser
	var err error

	if p == nil || col.graphiteParser.Separator != col.opts.metricSeparator {
		p, err = graphite.NewGraphiteParser(col.opts.metricSeparator, col.Templates, nil)
		col.graphiteParser = p
	}

	if err == nil {
		p.DefaultTags = tags
		name, tags, field, _ = p.ApplyTemplate(name)
	}

	if col.opts.convertNames {
		name = strings.ReplaceAll(name, ".", "_")
		name = strings.ReplaceAll(name, "-", "__")
	}
	if field == "" {
		field = defaultFieldName
	}

	return name, field, tags
}

// Parse the key,value out of a string that looks like "key=value".
func parseKeyValue(keyvalue string) (string, string) {
	var key, val string

	split := strings.Split(keyvalue, "=")
	// Must be exactly 2 to get anything meaningful out of them
	if len(split) == 2 {
		key = split[0]
		val = split[1]
	} else if len(split) == 1 {
		val = split[0]
	}

	return key, val
}

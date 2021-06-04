package statsd

import (
	"bytes"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type job struct {
	*bytes.Buffer
	time.Time
	Addr string
}

type metric struct {
	name, field, bucket, hash, mtype string

	fields map[string]interface{}

	additive   bool
	samplerate float64

	intval   int64
	floatval float64
	strval   string
	ts       time.Time

	tags map[string]string
}

func (m *metric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "以 statsd 中实际数据而定",
		Fields: map[string]interface{}{},
		Tags:   map[string]interface{}{},
	}
}

func (m *metric) getVal() interface{} {

	switch m.mtype {
	case "g", "ms", "h", "d":
		return m.floatval
	case "c":
		return m.intval
	case "s":
		return m.strval
	default:
		return nil
	}
}

func (m *metric) LineProto() (*io.Point, error) {
	if m.fields == nil {
		m.fields = map[string]interface{}{
			m.field: m.getVal(),
		}
	} else {
		// TODO: merge m.fields & m.filed
	}

	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

type parser struct {
	ref   *input
	cache []*metric
}

func (x *input) run(idx int) {

	p := &parser{
		ref: x,
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Infof("statsd parser worker %d exit.", idx)
			return
		case j := <-x.in:
			_ = p.doParse(j)
			p.feedIO()
			p.reset()
		}
	}
}

func (p *parser) feedIO() {
	// TODO
}

func (p *parser) reset() {
	// TODO
}

func (p *parser) doParse(j *job) error {
	lines := strings.Split(j.Buffer.String(), "\n")
	p.ref.bufpool.Put(j.Buffer)
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		switch {
		case ln == "":
		case strings.HasPrefix(ln, "_e"):
			if err := p.parseEventMsg(j.Time, ln, j.Addr); err != nil {
				return err
			}

		default:
			if err := p.parseStatsdLine(ln); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *parser) parseStatsdLine(line string) error {
	var lineTags map[string]string

	if p.ref.DataDogExtensions {
		recombinedSegments := make([]string, 0)
		// datadog tags look like this:
		// users.online:1|c|@0.5|#country:china,environment:production
		// users.online:1|c|#sometagwithnovalue
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
		l.Errorf("Splitting ':', unable to parse metric: %s", line)
		return errors.New("error Parsing statsd line")
	}

	// Extract bucket name from individual metric bits
	bucketName, bits := bits[0], bits[1:]

	// Add a metric for each bit available
	for _, bit := range bits {
		m := metric{
			bucket: bucketName,
		}

		// Validate splitting the bit on "|"
		pipesplit := strings.Split(bit, "|")
		if len(pipesplit) < 2 {
			l.Errorf("Splitting '|', unable to parse metric: %s", line)
			return errors.New("error parsing statsd line")
		} else if len(pipesplit) > 2 {
			sr := pipesplit[2]

			if strings.Contains(sr, "@") && len(sr) > 1 {
				samplerate, err := strconv.ParseFloat(sr[1:], 64)
				if err != nil {
					l.Errorf("Parsing sample rate: %s", err.Error())
				} else {
					// sample rate successfully parsed
					m.samplerate = samplerate
				}
			} else {
				l.Debugf("Sample rate must be in format like: "+
					"@0.1, @0.5, etc. Ignoring sample rate for line: %s", line)
			}
		}

		// Validate metric type
		switch pipesplit[1] {
		case "g", "c", "s", "ms", "h", "d":
			m.mtype = pipesplit[1]
		default:
			l.Errorf("Metric type %q unsupported", pipesplit[1])
			return errors.New("error parsing statsd line")
		}

		// Parse the value
		if strings.HasPrefix(pipesplit[0], "-") || strings.HasPrefix(pipesplit[0], "+") {
			if m.mtype != "g" && m.mtype != "c" {
				l.Errorf("+- values are only supported for gauges & counters, unable to parse metric: %s", line)
				return errors.New("error parsing statsd line")
			}
			m.additive = true
		}

		switch m.mtype {
		case "g", "ms", "h", "d":
			v, err := strconv.ParseFloat(pipesplit[0], 64)
			if err != nil {
				l.Errorf("Parsing value to float64, unable to parse metric: %s", line)
				return errors.New("error parsing statsd line")
			}
			m.floatval = v
		case "c":
			var v int64
			v, err := strconv.ParseInt(pipesplit[0], 10, 64)
			if err != nil {
				v2, err2 := strconv.ParseFloat(pipesplit[0], 64)
				if err2 != nil {
					l.Errorf("Parsing value to int64, unable to parse metric: %s", line)
					return errors.New("error parsing statsd line")
				}
				v = int64(v2)
			}
			// If a sample rate is given with a counter, divide value by the rate
			if m.samplerate != 0 && m.mtype == "c" {
				v = int64(float64(v) / m.samplerate)
			}
			m.intval = v
		case "s":
			m.strval = pipesplit[0]
		}

		// Parse the name & tags from bucket
		m.name, m.field, m.tags = p.parseName(m.bucket)
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

		p.aggregate(&m)
	}

	return nil
}

func (p *parser) parseName(bucket string) (string, string, map[string]string) {
	tags := make(map[string]string)

	bucketparts := strings.Split(bucket, ",")
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

	gp := s.graphiteParser
	var err error

	if gp == nil || p.ref.graphiteParser.Separator != p.ref.MetricSeparator {
		gp, err = graphite.NewGraphiteParser(p.ref.MetricSeparator, p.ref.Templates, nil)
		p.ref.graphiteParser = gp
	}

	if err == nil {
		p.ref.DefaultTags = tags
		name, tags, field, _ = p.ref.ApplyTemplate(name)
	}

	if s.ConvertNames {
		name = strings.Replace(name, ".", "_", -1)
		name = strings.Replace(name, "-", "__", -1)
	}
	if field == "" {
		field = defaultFieldName
	}

	return name, field, tags
}

func (p *parser) aggregate(m *metric) {
	// TODO
}

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

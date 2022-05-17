// Package socket collect socket metrics
package socket

import (
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func (i *Input) SampleConfig() string {
	return sample
}

func (i *Input) appendMeasurement(name string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	tmp := &Measurement{name: name, tags: tags, fields: fields, ts: ts}
	i.collectCache = append(i.collectCache, tmp)
}

func (i *Input) Catalog() string {
	return "socket"
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("socket input started")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)
	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()

	for {
		i.collectCache = make([]inputs.Measurement, 0)

		start := time.Now()
		if err := i.Collect(); err != nil {
			l.Errorf("Collect: %s", err)
			io.FeedLastError(inputName, err.Error())
		}

		if len(i.collectCache) > 0 {
			if err := inputs.FeedMeasurement(metricName,
				datakit.Metric,
				i.collectCache,
				&io.Option{CollectCost: time.Since((start))}); err != nil {
				l.Errorf("FeedMeasurement: %s", err)
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("hostdir input exit")
			return

		case <-i.semStop.Wait():
			l.Infof("hostdir input return")
			return
		}
	}
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
	}
}

func (i *Input) Collect() error {
	for _, proto := range i.SocketProto {
		out, err := i.lister(i.cmdName, proto)
		if err != nil {
			return err
		}
		i.CollectMeasurement(out, proto)
	}
	return nil
}

func (i *Input) CollectMeasurement(data *bytes.Buffer, proto string) {
	TimeNow := time.Now()
	scanner := bufio.NewScanner(data)
	tags := map[string]string{}
	fields := make(map[string]interface{})

	scanner.Scan()

	flushData := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		words := strings.Fields(line)

		if i.isNewConnection.MatchString(line) {
			for _, word := range words {
				if !i.validValues.MatchString(word) {
					continue
				}
				// kv will have 2 fields because it matched the regexp
				kv := strings.Split(word, ":")
				v, err := strconv.ParseUint(kv[1], 10, 64)
				if err != nil {
					l.Infof("socket couldn't parse metric %q: %v", word, err)
					continue
				}
				fields[kv[0]] = v
			}
			if !flushData {
				l.Warnf("socket input found orphaned metrics: %s", words)
				l.Warn("socket input added them to the last known connection.")
			}
			i.appendMeasurement(inputName, tags, fields, TimeNow)
			flushData = false
			continue
		}
		// A line with no starting whitespace means we're going to parse a new connection.
		// Flush what we gathered about the previous one, if any.
		if flushData {
			i.appendMeasurement(inputName, tags, fields, TimeNow)
		}

		// Delegate the real parsing to getTagsAndState, which manages various
		// formats depending on the protocol.
		tags, fields = getTagsAndState(proto, words)

		// This line containted metrics, so record that.
		flushData = true
	}
	if flushData {
		i.appendMeasurement(inputName, tags, fields, TimeNow)
	}
}

func socketList(cmdName string, proto string) (*bytes.Buffer, error) {
	// Run ss for the given protocol, return the output as bytes.Buffer
	args := []string{"-in", "--" + proto}
	cmd := exec.Command(cmdName, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return &out, err
	}
	return &out, nil
}

func getTagsAndState(proto string, words []string) (map[string]string, map[string]interface{}) {
	tags := map[string]string{
		"proto": proto,
	}
	fields := make(map[string]interface{})
	switch proto {
	case "udp", "raw":
		words = append([]string{"dummy"}, words...)
	case "tcp", "dccp", "sctp":
		fields["state"] = words[0]
	}
	switch proto {
	case "tcp", "udp", "raw", "dccp", "sctp":
		// Local and remote addresses are fields 3 and 4
		// Separate addresses and ports with the last ':'
		localIndex := strings.LastIndex(words[3], ":")
		remoteIndex := strings.LastIndex(words[4], ":")
		tags["local_addr"] = words[3][:localIndex]
		tags["local_port"] = words[3][localIndex+1:]
		tags["remote_addr"] = words[4][:remoteIndex]
		tags["remote_port"] = words[4][remoteIndex+1:]
	case "unix", "packet":
		fields["netid"] = words[0]
		tags["local_addr"] = words[4]
		tags["local_port"] = words[5]
		tags["remote_addr"] = words[6]
		tags["remote_port"] = words[7]
	}
	tags["proto"] = proto
	v, err := strconv.ParseUint(words[1], 10, 64)
	if err != nil {
		l.Warnf("Couldn't read recv_q in %q: %v", words, err)
	} else {
		fields["recv_q"] = v
	}
	v, err = strconv.ParseUint(words[2], 10, 64)
	if err != nil {
		l.Warnf("Couldn't read send_q in %q: %v", words, err)
	} else {
		fields["send_q"] = v
	}
	return tags, fields
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 5},
			semStop:  cliutils.NewSem(),
		}
		if len(s.SocketProto) == 0 {
			s.SocketProto = []string{"tcp", "udp"}
		}

		// Initialize regexps to validate input data
		validFields := "(bytes_acked|bytes_received|segs_out|segs_in|data_segs_in|data_segs_out)"
		s.validValues = regexp.MustCompile("^" + validFields + ":[0-9]+$")
		s.isNewConnection = regexp.MustCompile(`^\s+.*$`)

		s.lister = socketList
		ssPath, err := exec.LookPath("ss")
		if err != nil {
			io.FeedLastError(inputName, "socket input init error:"+err.Error())
		}
		s.cmdName = ssPath

		return s
	})
}

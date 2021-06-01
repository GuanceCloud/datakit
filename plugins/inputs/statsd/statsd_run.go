package statsd

import (
	"bufio"
	"errors"
	"net"
	"regexp"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var (
	ConnectionReset = errors.New("ConnectionReset")
)

const (
	MaxGatherInterval = 30 * time.Minute
	MinGatherInterval = 1 * time.Second
)

func (p *StatsdParams) gather() {
	var err error
	var d time.Duration

	switch p.input.Interval.(type) {
	case int64:
		d = time.Duration(p.input.Interval.(int64)) * time.Second
	case string:
		d, err = time.ParseDuration(p.input.Interval.(string))
		if err != nil {
			p.log.Errorf("parse interval err: %s", err.Error())
			return
		}
	default:
		p.log.Errorf("interval type unsupported")
		return
	}

	d = config.ProtectedInterval(MinGatherInterval, MaxGatherInterval, d)
	tick := time.NewTicker(d)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			_, err = p.getMetrics(false)
			if err != nil {
				io.FeedLastError(inputName, err.Error())
				p.reportNotUp()
			}

		case <-datakit.Exit.Wait():
			p.log.Info("input statsd exit")
			return
		}
	}
}

func (p *StatsdParams) getMetrics(isTest bool) ([]byte, error) {
	var pt []byte
	var err error

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["host"] = p.input.Host
	for tag, tagV := range p.input.Tags {
		tags[tag] = tagV
	}
	fields["is_up"] = true

	conn, err := net.Dial("tcp", p.input.Host)
	if err != nil {
		if !isTest {
			p.reportNotUp()
		}
		return nil, err
	}
	defer conn.Close()

	err = getMetric(conn, "counters", fields)
	if err != nil {
		goto ERR
	}

	err = getMetric(conn, "gauges", fields)
	if err != nil {
		goto ERR
	}

	err = getMetric(conn, "timers", fields)
	if err != nil {
		goto ERR
	}

	fields["can_connect"] = true
	pt, err = io.MakeMetric(p.input.MetricsName, tags, fields, time.Now())
	if err != nil {
		return nil, err
	}

	if !isTest {
		err = p.output.IoFeed(pt, datakit.Metric, inputName)
	}

	return pt, err

ERR:
	fields["can_connect"] = false
	pt, _ = io.MakeMetric(p.input.MetricsName, tags, fields, time.Now())

	if !isTest {
		err = p.output.IoFeed(pt, datakit.Metric, inputName)
	}
	return pt, err
}

func getMetric(conn net.Conn, msg string, fields map[string]interface{}) error {
	//buf := make([]byte, 0, 1024)
	_, err := conn.Write([]byte(msg))
	if err != nil {
		return err
	}
	bio := bufio.NewReader(conn)
	s, err := bio.ReadString('}')
	if err != nil {
		return err
	}

	exp := `(?s:\{(.*)\})`
	r := regexp.MustCompile(exp)
	matchs := r.FindStringSubmatch(s)
	if len(matchs) < 2 {
		return nil
	}

	cnt := strings.Count(matchs[1], ":")
	fields[msg+"_count"] = cnt

	return nil
}

func (p *StatsdParams) reportNotUp() error {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["host"] = p.input.Host
	for tag, tagV := range p.input.Tags {
		tags[tag] = tagV
	}
	fields["is_up"] = false
	fields["can_connect"] = false

	pt, err := io.MakeMetric(p.input.MetricsName, tags, fields, time.Now())
	if err != nil {
		return err
	}
	err = p.output.IoFeed(pt, datakit.Metric, inputName)
	return err
}

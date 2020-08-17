package statsd

import (
	"bufio"
	"errors"
	"net"
	"regexp"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var (
	ConnectionReset = errors.New("ConnectionReset")
)

func (p *StatsdParams) gather() {
	var connectFail bool = true
	var conn net.Conn
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
	tick := time.NewTicker(d)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			if connectFail {
				conn, err = net.Dial("tcp", p.input.Host)
				if err != nil {
					connectFail = true
				} else {
					connectFail = false
				}
			}

			if connectFail == false && conn != nil {
				err = p.getMetrics(conn)
				if err != nil && err != ConnectionReset {
					p.log.Errorf("getMetrics err: %s", err.Error())
				}

				if err == ConnectionReset {
					connectFail = true
					conn.Close()
				}
			} else {
				err = p.reportNotUp()
				if err != nil {
					p.log.Errorf("reportNotUp err: %s", err.Error())
				}
			}

		case <-datakit.Exit.Wait():
			p.log.Info("input statsd exit")
			return
		}
	}
}

func (p *StatsdParams) getMetrics(conn net.Conn) error {
	var pt []byte
	var err error

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["host"] = p.input.Host
	for tag, tagV := range p.input.Tags {
		tags[tag] = tagV
	}
	fields["is_up"] = true

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
		return err
	}
	err = p.output.IoFeed(pt, io.Metric, inputName)
	return err

ERR:
	fields["can_connect"] = false
	pt, _ = io.MakeMetric(p.input.MetricsName, tags, fields, time.Now())
	err = p.output.IoFeed(pt, io.Metric, inputName)
	return ConnectionReset
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
	err = p.output.IoFeed(pt, io.Metric, inputName)
	return err
}

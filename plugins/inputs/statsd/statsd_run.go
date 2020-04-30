package statsd

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

var (
	ConnectionReset = errors.New("ConnectionReset")
)

func (p *StatsdParams) gather(ctx context.Context) {
	var connectFail bool = true
	var conn net.Conn
	var err error

	for {
		select {
		case <-stopChan:
			return
		case <-ctx.Done():
			return
		default:
		}

		if connectFail {
			conn, err = net.Dial("tcp", "127.0.0.1:8126")
			if err != nil {
				connectFail = true
			} else {
				connectFail = false
			}
		}

		if connectFail == false && conn != nil {
			err = p.getMetrics(conn)
			if err != nil {
				log.Printf("W! [statsd] %s", err.Error())
			}
			if err == ConnectionReset {
				connectFail = true
				conn.Close()
			}
		} else {
			p.reportNotUp()
		}

		err = internal.SleepContext(ctx, time.Duration(p.input.Interval)*time.Second)
		if err != nil {
			log.Printf("W! [statsd] %s", err.Error())
		}
	}
}

func (p *StatsdParams) getMetrics(conn net.Conn) error {
	var metrics telegraf.Metric
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["host"] = p.input.Host
	fields["is_up"] = true

	err := getMetric(conn, "counters", fields)
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
	metrics, err = metric.New(p.input.MetricName, tags, fields, time.Now())
	if err != nil {
		return err
	}
	p.output.acc.AddMetric(metrics)
	return nil

ERR:
	fields["can_connect"] = false
	metrics, _ = metric.New(p.input.MetricName, tags, fields, time.Now())
	p.output.acc.AddMetric(metrics)
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
	r:= regexp.MustCompile(exp)
	matchs := r.FindStringSubmatch(s)
	if len(matchs) < 2 {
		return nil
	}

	cnt := strings.Count(matchs[1], ":")
	fields[msg+"_count"] = cnt

	return nil
}

func (p *StatsdParams) reportNotUp() {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["host"] = p.input.Host
	fields["is_up"] = false
	fields["can_connect"] = false

	pointMetric, _ := metric.New(p.input.MetricName, tags, fields, time.Now())
	p.output.acc.AddMetric(pointMetric)
}

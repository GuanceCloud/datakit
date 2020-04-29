package traefik

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"encoding/json"

	traefikLog "github.com/siddontang/go-log/log"

	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type TraefikServStats struct {
	Pid      int    `json:"pid"`
	Hostname string `json:"hostname"`

	Uptime       float64 `json:"uptime_sec"`
	TotalCount   int64    `json:"total_count"`
	TotalRepTime float64 `json:"total_response_time_sec"`
	TotalRepSize int64    `json:"total_response_size"`
	AvergRepTime float64 `json:"average_response_time_sec"`
	AvergRepSize int64    `json:"average_response_size"`
	TotalStatCodeCnt  map[string]int `json:"total_status_code_count"`
}
type TraefikTarget struct {
	Interval int
	Active   bool
	Url      string
}

type Traefik struct {
	MetricName string `toml:"metric_name"`
	Targets    []TraefikTarget
	acc        telegraf.Accumulator
}

type TraefikInput struct {
	TraefikTarget
	MetricName string
}

type TraefikOutput struct {
	acc telegraf.Accumulator
}

type TraefikParam struct {
	input  TraefikInput
	output TraefikOutput
}

type TraefikLogWriter struct {
	io.Writer
}

const traefikConfigSample = `### metric_name: the name of metric, default is "traefik"
### You need to configure an [[targets]] for each traefik to be monitored.
### interval: monitor interval second, unit is second. The default value is 60.
### active: whether to monitor traefik.
### url: traefik service WebUI url.

#metric_name="traefik"
#[[targets]]
#	interval = 60
#	active   = true
#	url     = "http://127.0.0.1:8080/health"

#[[targets]]
#	interval = 60
#	active   = true
#	url     = "http://127.0.0.1:8080/health"
`

var (
	activeTargets     = 0
	stopChan          chan bool
	ctx               context.Context
	cfun              context.CancelFunc
	defaultMetricName = "traefik"
	defaultInterval   = 60
)

func (t *Traefik) SampleConfig() string {
	return traefikConfigSample
}

func (t *Traefik) Description() string {
	return "Monitor Traefik Service Status"
}

func (t *Traefik) Gather(telegraf.Accumulator) error {
	return nil
}

func (t *Traefik) Start(acc telegraf.Accumulator) error {
	log.Printf("I! [traefik] start")
	ctx, cfun = context.WithCancel(context.Background())
	setupLogger()

	metricName := defaultMetricName
	if t.MetricName != "" {
		metricName = t.MetricName
	}

	targetCnt := 0
	for _, target := range t.Targets {
		if target.Active && target.Url != "" {
			if target.Interval == 0 {
				target.Interval = defaultInterval
			}

			input := TraefikInput{target, metricName}
			output := TraefikOutput{acc}

			p := &TraefikParam{input, output}
			go p.gather(ctx)
			targetCnt += 1
		}
	}
	activeTargets = targetCnt
	stopChan = make(chan bool, targetCnt)
	return nil
}

func (p *Traefik) Stop() {
	for i := 0; i < activeTargets; i++ {
		stopChan <- true
	}
	cfun()
}

func (p *TraefikParam) gather(ctx context.Context) {
	for {
		select {
		case <-stopChan:
			return
		case <-ctx.Done():
			return
		default:
		}

		err := p.getMetrics()
		if err != nil {
			log.Printf("W! [traefik] %s", err.Error())
		}

		err = internal.SleepContext(ctx, time.Duration(p.input.Interval)*time.Second)
		if err != nil {
			log.Printf("W! [traefik] %s", err.Error())
		}
	}
}

func (p *TraefikParam) getMetrics() error {
	var s TraefikServStats
	s.TotalStatCodeCnt = make(map[string]int)

	tags := make(map[string]string)
	fields := make(map[string]interface{})
	tags["url"] = p.input.Url

	resp, err := http.Get(p.input.Url)
	if err != nil || resp.StatusCode != 200 {
		fields["can_connect"] = false
		p.output.acc.AddFields(p.input.MetricName, fields, tags)
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return fmt.Errorf("decode json err: %s", err.Error())
	}

	tags["pid"] = fmt.Sprintf("%d", s.Pid)
	tags["hostname"] = s.Hostname

	fields["can_connect"] = true
	fields["uptime"] = s.Uptime
	fields["total_count"] = s.TotalCount
	fields["total_time"] = s.TotalRepTime
	fields["total_size"] = s.TotalRepSize
	fields["average_time"] = s.AvergRepTime
	fields["average_size"] = s.AvergRepSize

	for k,v := range s.TotalStatCodeCnt {
		fields["http_" + k + "_count"] = v
	}
	p.output.acc.AddFields(p.input.MetricName, fields, tags)
	return nil
}

func getReadableTimeStr(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d)
	} else if d < time.Millisecond {
		return fmt.Sprintf("%fus", float64(d)/float64(time.Microsecond))
	} else if d < time.Second {
		return fmt.Sprintf("%fms", float64(d)/float64(time.Millisecond))
	} else {
		return fmt.Sprintf("%fs", float64(d)/float64(time.Second))
	}
}
func setupLogger() {
	loghandler, _ := traefikLog.NewStreamHandler(&TraefikLogWriter{})
	traefikLogger := traefikLog.New(loghandler, 0)
	traefikLog.SetLevel(traefikLog.LevelDebug)
	traefikLog.SetDefaultLogger(traefikLogger)
}

func init() {
	inputs.Add("traefik", func() telegraf.Input {
		p := &Traefik{}
		return p
	})
}

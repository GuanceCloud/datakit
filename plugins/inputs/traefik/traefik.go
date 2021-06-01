package traefik

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type IoFeed func(data []byte, category, name string) error

type TraefikServStats struct {
	Pid      int    `json:"pid"`
	Hostname string `json:"hostname"`

	Uptime           float64        `json:"uptime_sec"`
	TotalCount       int64          `json:"total_count"`
	TotalRepTime     float64        `json:"total_response_time_sec"`
	TotalRepSize     int64          `json:"total_response_size"`
	AvergRepTime     float64        `json:"average_response_time_sec"`
	AvergRepSize     int64          `json:"average_response_size"`
	TotalStatCodeCnt map[string]int `json:"total_status_code_count"`
}

type Traefik struct {
	Interval    interface{}
	Active      bool
	Url         string
	MetricsName string
	Tags        map[string]string
}

type TraefikInput struct {
	Traefik
}

type TraefikOutput struct {
	IoFeed
}

type TraefikParam struct {
	input  TraefikInput
	output TraefikOutput
	log    *logger.Logger
}

var (
	inputName = "traefik"

	defaultMetricName   = inputName
	defaultInterval     = "60s"
	traefikConfigSample = `
### You need to configure an [[inputs.traefik]] for each traefik to be monitored.
### interval: monitor interval, the default value is "60s".
### url: traefik service WebUI url.
### metricsName: the name of metric, default is "traefik"

#[[inputs.traefik]]
#	interval    = "60s"
#	url         = "http://127.0.0.1:8080/health"
#	metricsName = "traefik"
#	[inputs.traefik.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"`
)

const (
	MaxGatherInterval = 30 * time.Minute
	MinGatherInterval = 30 * time.Second
)

func (t *Traefik) SampleConfig() string {
	return traefikConfigSample
}

func (t *Traefik) Catalog() string {
	return "traefik"
}

func (t *Traefik) Run() {
	if t.Url == "" {
		return
	}

	p := t.genParam()
	p.log.Infof("traefik input started...")
	p.gather()
}

func (t *Traefik) genParam() *TraefikParam {
	if t.MetricsName == "" {
		t.MetricsName = defaultMetricName
	}

	if t.Interval == nil {
		t.Interval = defaultInterval
	}

	input := TraefikInput{*t}
	output := TraefikOutput{io.NamedFeed}

	p := &TraefikParam{input, output, logger.SLogger("traefik")}
	return p
}
func (p *TraefikParam) gather() {
	var d time.Duration
	var err error

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
				p.log.Errorf("getMetrics err: %s", err.Error())
			}

		case <-datakit.Exit.Wait():
			p.log.Info("input traefik exit")
			return
		}
	}
}

func (p *TraefikParam) getMetrics(isTest bool) ([]byte, error) {
	var s TraefikServStats
	s.TotalStatCodeCnt = make(map[string]int)

	tags := make(map[string]string)
	fields := make(map[string]interface{})
	tags["url"] = p.input.Url
	for tag, tagV := range p.input.Tags {
		tags[tag] = tagV
	}

	resp, err := http.Get(p.input.Url)
	if err != nil || resp.StatusCode != 200 {
		fields["can_connect"] = false
		pt, _ := io.MakeMetric(p.input.MetricsName, tags, fields, time.Now())
		p.output.IoFeed(pt, datakit.Metric, inputName)
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, fmt.Errorf("decode json err: %s", err.Error())
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

	for k, v := range s.TotalStatCodeCnt {
		fields["http_"+k+"_count"] = v
	}

	pt, err := io.MakeMetric(p.input.MetricsName, tags, fields, time.Now())
	if err != nil {
		return pt, err
	}

	if !isTest {
		err = p.output.IoFeed(pt, datakit.Metric, inputName)
		return pt, err
	}

	return pt, nil
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

func init() {
	inputs.Add(inputName, func() inputs.Input {
		p := &Traefik{}
		return p
	})
}

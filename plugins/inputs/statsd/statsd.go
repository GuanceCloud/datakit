package statsd

import (
	"regexp"

	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type IoFeed func(data []byte, category string) error

type StatsD struct {
	Interval    int
	Active      bool
	Host        string
	MetricsName string
	Tags        map[string]string
}

type StatsdInput struct {
	StatsD
}

type StatsdOutput struct {
	IoFeed
}

type StatsdParams struct {
	input  StatsdInput
	output StatsdOutput
	log    *zap.SugaredLogger
}

var (
	statsdConfigSample = `### You need to configure an [[inputs.statsd]] for each statsd service to be monitored.
### active: whether to monitor statsd.
### interval: monitor interval second, unit is second. The default value is 60.
### host: statsd service ip:port, if "127.0.0.1", default port is 8126.
### metricsName: the name of metric, default is "statsd"

#[[inputs.statsd]]
#	active      = true
#	interval    = 60
#	host        = "127.0.0.1:8126"
#	metricsName = "statsd"
#	[inputs.statsd.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"

#[[inputs.statsd]]
#	active      = true
#	interval    = 60
#	host        = "127.0.0.1:8126"
#	metricsName = "statsd"
#	[inputs.statsd.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
`
	defaultMetricName = "statsd"
	defaultInterval   = 60
)

func (t *StatsD) Catalog() string {
	return "statsd"
}

func (t *StatsD) SampleConfig() string {
	return statsdConfigSample
}

func (t *StatsD) Run() {
	if !t.Active || t.Host == "" {
		return
	}

	reg, _ := regexp.Compile(`:\d{1,5}$`)

	if t.MetricsName == "" {
		t.MetricsName = defaultMetricName
	}

	if t.Interval <= 0 {
		t.Interval = defaultInterval
	}

	if !reg.MatchString(t.Host) {
		t.Host += ":8126"
	}

	input := StatsdInput{*t}
	output := StatsdOutput{io.Feed}
	p := StatsdParams{input, output, logger.SLogger("statsd")}
	p.log.Info("statsd input started...")
	p.gather()
}

func init() {
	inputs.Add("statsd", func() inputs.Input {
		p := &StatsD{}
		return p
	})
}

package statsd

import (
	"regexp"
	"sync"


	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type IoFeed func(data []byte, category string) error

type Target struct {
	Interval int
	Active   bool
	Host     string
}

type StatsD struct {
	MetricName string `toml:"metric_name"`
	Targets    []Target
}

type StatsdInput struct {
	MetricName string
	Interval   int
	Host       string
}

type StatsdOutput struct {
	IoFeed
}

type StatsdParams struct {
	input  StatsdInput
	output StatsdOutput
}

var (
	statsdConfigSample = `### metric_name: the name of metric, default is "statsd".
### You need to configure an [[targets]] for each statsd service to be monitored.
### interval: monitor interval second, unit is second. The default value is 60.
### active: whether to monitor statsd.
### host: statsd service ip:port, if "127.0.0.1", default port is 8126.

#metric_name="statsd"
#[[targets]]
#	interval = 60
#	active   = true
#	host     = "127.0.0.1:8126"

#[[targets]]
#	interval = 60
#	active   = true
#	host     = "127.0.0.1:8126"
`
	defaultMetricName = "statsd"
	defaultInterval   = 60
	Log *zap.SugaredLogger
)


func (t *StatsD) Catalog() string {
	return "statsd"
}

func (t *StatsD) SampleConfig() string {
	return statsdConfigSample
}

func (t *StatsD) Run() {
	isActive := false
	wg := sync.WaitGroup{}
	Log = logger.SLogger("statsd")
	reg, _ := regexp.Compile(`:\d{1,5}$`)


	if t.MetricName == "" {
		t.MetricName = defaultMetricName
	}

	for _, target := range t.Targets {
		if target.Active == false || target.Host == "" {
			continue
		}

		if !isActive {
			Log.Info("statsd input started...")
			isActive = true
		}

		input := StatsdInput{
			t.MetricName,
			target.Interval,
			target.Host,
		}

		if input.Interval <= 0 {
			input.Interval = defaultInterval
		}

		if !reg.MatchString(target.Host) {
			target.Host += ":8126"
		}

		output := StatsdOutput{io.Feed}

		p := StatsdParams{input, output}
		wg.Add(1)
		go p.gather(&wg)
	}
	wg.Wait()
}

func init() {
	inputs.Add("statsd", func() inputs.Input {
		p := &StatsD{}
		return p
	})
}

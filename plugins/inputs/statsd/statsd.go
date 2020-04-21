package statsd

import (
	"context"
	"io"
	"log"
	"regexp"

	"github.com/influxdata/telegraf"
	sdlog "github.com/siddontang/go-log/log"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

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
	acc telegraf.Accumulator
}

type StatsdParams struct {
	input  StatsdInput
	output StatsdOutput
}

type sdLogWriter struct {
	io.Writer
}

const statsdConfigSample = `### metric_name: the name of metric, default is "statsd".
### You need to configure an [[targets]] for each statsd service to be monitored.
### interval: monitor interval second, unit is second. The default value is 60.
### active: whether to monitor statsd.
### host: statsd service ip:port, if "127.0.0.1", default port is 8126.

#metric_name="statsd"
#[[targets]]
#	interval = 60
#	active   = false
#	host     = "127.0.0.1:8126"

#[[targets]]
#	interval = 60
#	active   = false
#	host     = "127.0.0.1:8126"
`

var (
	ctx           context.Context
	cfun          context.CancelFunc
	activeTargets = 0
	stopChan      chan bool
)

const (
	defaultMetricName = "statsd"
	defaultInterval   = 60
)

func (t *StatsD) SampleConfig() string {
	return statsdConfigSample
}

func (t *StatsD) Description() string {
	return "Monitor StatsD Service"
}

func (t *StatsD) Gather(telegraf.Accumulator) error {
	return nil
}

func (t *StatsD) Start(acc telegraf.Accumulator) error {
	setupLogger()
	reg, _ := regexp.Compile(`:\d{1,5}$`)
	log.Printf("I! [statsd] start")
	ctx, cfun = context.WithCancel(context.Background())

	if t.MetricName == "" {
		t.MetricName = defaultMetricName
	}

	activeCnt := 0
	for _, target := range t.Targets {
		if target.Active == false || target.Host == "" {
			continue
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

		output := StatsdOutput{acc}

		p := StatsdParams{input, output}
		go p.gather(ctx)

		activeCnt += 1
	}

	activeTargets = activeCnt
	stopChan = make(chan bool, activeTargets)
	return nil
}

func (t *StatsD) Stop() {
	for i := 0; i < activeTargets; i++ {
		stopChan <- true
	}
	cfun()
}

func setupLogger() {
	loghandler, _ := sdlog.NewStreamHandler(&sdLogWriter{})
	sdlogger := sdlog.New(loghandler, 0)
	sdlog.SetLevel(sdlog.LevelDebug)
	sdlog.SetDefaultLogger(sdlogger)
}

func init() {
	inputs.Add("statsd", func() telegraf.Input {
		p := &StatsD{}
		return p
	})
}

package prometheus

import (
	"context"
	"io"
	"log"
	"strings"

	"github.com/influxdata/telegraf"
	plog "github.com/siddontang/go-log/log"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	pluginName = "prometheus"
	promeths   *Prometheus
	stopChan   chan bool
	activeHost int
	defaultInterval = 60
)

type prometheusLogWriter struct {
	io.Writer
}

type Target struct {
	Host     string
	Interval int
	Active   bool
}

type Prometheus struct {
	Targets []Target

	ctx  context.Context
	cfun context.CancelFunc
	acc  telegraf.Accumulator
}

type PrometheusInput struct {
	host     string
	interval int
}

type PrometheusOutput struct {
	ctx  context.Context
	cfun context.CancelFunc
	acc  telegraf.Accumulator
}

type PrometheusParam struct {
	input PrometheusInput
	output PrometheusOutput
}
func (p *Prometheus) SampleConfig() string {
	return prometheusConfigSample
}

func (p *Prometheus) Description() string {
	return "Gather Prometheus Exporter Data to Dataway"
}

func (p *Prometheus) Gather(telegraf.Accumulator) error {
	return nil
}

func (p *Prometheus) globalInit() {
	promeths = p
}

func (p *Prometheus) Start(acc telegraf.Accumulator) error {
	var targets []string

	log.Printf("I! [Prometheus] start")
	setupLogger()

	for _, target := range p.Targets {
		if target.Active && target.Host != ""{
			targets = append(targets, target.Host)

			interval := defaultInterval
			if target.Interval != 0 {
				interval = target.Interval
			}

			input := PrometheusInput{target.Host, interval}
			output := PrometheusOutput{p.ctx, p.cfun, acc}

			p := &PrometheusParam{input, output}
			go p.gather()
		}
	}

	log.Printf("I! [Prometheus] active host: %s", strings.Join(targets, ","))

	activeHost = len(targets)
	stopChan = make(chan bool, activeHost)

	return nil
}

func (p *Prometheus) Stop() {
	for i :=0; i < activeHost; i++ {
		stopChan <- true
	}
}

func setupLogger() {
	loghandler, _ := plog.NewStreamHandler(&prometheusLogWriter{})
	plogger := plog.New(loghandler, 0)
	plog.SetLevel(plog.LevelDebug)
	plog.SetDefaultLogger(plogger)
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input {
		p := &Prometheus{}
		p.ctx, p.cfun = context.WithCancel(context.Background())
		return p
	})
}
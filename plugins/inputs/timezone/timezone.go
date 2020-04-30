package timezone

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	tzlog "github.com/siddontang/go-log/log"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Timezone struct {
	Metricname string `toml:"metric_name"`
	Active     bool
	Interval   int
	Hostname   string
}

type TzInput struct {
	Timezone
}

type TzOutput struct {
	acc telegraf.Accumulator
}

type TzParams struct {
	input  TzInput
	output TzOutput
}

type TzLogWriter struct {
	io.Writer
}

const (
	defaultMetricName = "timezone"
	defaultInterval   = 60
)

var (
	timeZoneConfigSample = `### metric_name: the name of metric, default is "timezone".
### active: whether to monitor timezone changes.
### interval: monitor interval second, unit is second. The default value is 60.
### hostname: If not specified, the environment variable will be used.

#metric_name="timezone"
#active   = true
#interval = 60
#hostname = ""`
)

var (
	ctx          context.Context
	cfun         context.CancelFunc
)

func (t *Timezone) SampleConfig() string {
	return timeZoneConfigSample
}

func (t *Timezone) Description() string {
	return "Monitor timezone changes"
}

func (t *Timezone) Gather(telegraf.Accumulator) error {
	return nil
}

func (t *Timezone) Start(acc telegraf.Accumulator) error {
	setupLogger()

	log.Printf("I! [timezone] start")
	ctx, cfun = context.WithCancel(context.Background())

	input := TzInput{*t}
	if input.Active == false {
		return nil
	}
	if input.Interval <= 0 {
		input.Interval = defaultInterval
	}
	if input.Metricname == "" {
		input.Metricname = defaultMetricName
	}

	output := TzOutput{acc}

	p := TzParams{input, output}
	go p.gather(ctx)

	return nil
}

func (t *Timezone) Stop() {
	cfun()
}

func (p *TzParams) gather(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := p.getMetrics()
		if err != nil {
			log.Printf("W! [timezone] %s", err.Error())
		}

		err = internal.SleepContext(ctx, time.Duration(p.input.Interval)*time.Second)
		if err != nil {
			log.Printf("W! [timezone] %s", err.Error())
		}
	}
}

func (p *TzParams) getMetrics() error {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	timezone, err := getOsTimezone()
	if err != nil {
		return err
	}

	if p.input.Hostname != "" {
		tags["host"] = p.input.Hostname
	}
	fields["tz"] = timezone

	pointMetric, err := metric.New(p.input.Metricname, tags, fields, time.Now())
	if err != nil {
		return err
	}

	p.output.acc.AddMetric(pointMetric)
	return nil
}

func getOsTimezone() (string, error) {
	var outInfo bytes.Buffer
	os := runtime.GOOS

	if os == "linux" || os == "darwin" {
		cmd := exec.Command("date", `+%Z`)
		cmd.Stdout = &outInfo
		cmd.Run()
		return strings.Trim(outInfo.String(), "\n"), nil
	} else if os == "windows" {
		cmd := exec.Command("tzutil.exe", "/g")
		cmd.Stdout = &outInfo
		cmd.Run()
		return outInfo.String(), nil
	} else {
		return "", fmt.Errorf("Os: %s unsuport get timezone", os)
	}
}

func setupLogger() {
	loghandler, _ := tzlog.NewStreamHandler(&TzLogWriter{})
	logger := tzlog.New(loghandler, 0)
	tzlog.SetLevel(tzlog.LevelDebug)
	tzlog.SetDefaultLogger(logger)
}

func init() {
	inputs.Add("timezone", func() telegraf.Input {
		tz := &Timezone{}
		return tz
	})
}

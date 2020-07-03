package timezone

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type TimeIoFeed func(data []byte, category string) error

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
	ioFeed TimeIoFeed
}

type TzParams struct {
	input  TzInput
	output TzOutput
}

const (
	defaultMetricName = "timezone"
	defaultInterval   = 60
)

var (
	tzlog                *zap.SugaredLogger
	timeZoneConfigSample = `### metric_name: the name of metric, default is "timezone".
### active: whether to monitor timezone changes.
### interval: monitor interval second, unit is second. The default value is 60.
### hostname: If not specified, the environment variable will be used.

#metric_name="timezone"
#active   = true
#interval = 60
#hostname = ""`
)

func (t *Timezone) SampleConfig() string {
	return timeZoneConfigSample
}

func (t *Timezone) Catalog() string {
	return "timezone"
}

func (t *Timezone) Description() string {
	return "Monitor timezone changes"
}

func (t *Timezone) Run() {
	tzlog = logger.SLogger("timezone")

	input := TzInput{*t}
	if input.Active == false {
		return
	}
	if input.Interval <= 0 {
		input.Interval = defaultInterval
	}
	if input.Metricname == "" {
		input.Metricname = defaultMetricName
	}

	output := TzOutput{io.Feed}
	p := TzParams{input, output}

	tzlog.Info("timezone input started...")
	p.gather()
}

func (p *TzParams) gather() {
	tick := time.NewTicker(time.Duration(p.input.Interval) * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			err := p.getMetrics()
			if err != nil {
				tzlog.Errorf("getMetrics err: %s", err.Error())
			}

		case <-datakit.Exit.Wait():
			tzlog.Info("input timezone exit")
			return
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

	pt, err := influxdb.NewPoint(p.input.Metricname, tags, fields, time.Now())
	if err != nil {
		return err
	}

	if err := p.output.ioFeed([]byte(pt.String()), io.Metric); err != nil {
		return err
	}

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

func init() {
	inputs.Add("timezone", func() inputs.Input {
		tz := &Timezone{}
		return tz
	})
}

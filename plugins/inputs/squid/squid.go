package squid

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type IoFeed func(data []byte, category string) error

type Squid struct {
	MetricName string `toml:"metric_name"`
	Active     bool
	Interval   int
	Port       int
}

type SquidInput struct {
	Squid
}

type SquidOutput struct {
	IoFeed
}

type SquidParam struct {
	input  SquidInput
	output SquidOutput
}

var (
	defaultMetricName = "squid"
	defaultInterval   = 60
	defaultPort       = 3218
	sqlog             *zap.SugaredLogger
	squidConfigSample = `### metric_name: the name of metric, default is "squid"
### interval: monitor interval second, unit is second. The default value is 60.
### active: whether to monitor squid.

#metric_name = "squid"
#active   = true
#interval = 60
#port     = 3128`
)

func (s *Squid) Catalog() string {
	return "squid"
}

func (s *Squid) SampleConfig() string {
	return squidConfigSample
}

func (s *Squid) Description() string {
	return "Monitor Squid Service Status"
}

func (s *Squid) Run() {
	sqlog = logger.SLogger("squid")
	input := SquidInput{*s}
	if input.Active == false {
		return
	}
	if input.MetricName == "" {
		input.MetricName = defaultMetricName
	}
	if input.Interval == 0 {
		input.Interval = defaultInterval
	}
	if input.Port == 0 {
		input.Port = defaultPort
	}
	output := SquidOutput{io.Feed}
	p := &SquidParam{input, output}

	sqlog.Info("squid input started...")
	p.gather()
}

func (p *SquidParam) gather() {
	tick := time.NewTicker(time.Duration(p.input.Interval) * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			err := p.getMetrics()
			if err != nil {
				sqlog.Errorf("getMetrics err: %s", err.Error())
			}
		case <-datakit.Exit.Wait():
			sqlog.Info("input squid exit")
			return
		}
	}
}

func (p *SquidParam) getMetrics() (err error) {
	var outInfo bytes.Buffer

	tags := make(map[string]string)
	fields := make(map[string]interface{})
	fields["can_connect"] = true

	reg := regexp.MustCompile(" = \\d{1,}\\.{0,1}\\d{0,}$")
	portStr := fmt.Sprintf("%d", p.input.Port)

	cmd := exec.Command("squidclient", "-p", portStr, "mgr:counters")
	cmd.Stdout = &outInfo
	err = cmd.Run()
	if err != nil {
		fields["can_connect"] = false
		pt, _ := influxdb.NewPoint(p.input.Squid.MetricName, tags, fields, time.Now())
		p.output.IoFeed([]byte(pt.String()), io.Metric)
		return
	}

	s := bufio.NewScanner(strings.NewReader(outInfo.String()))
	for s.Scan() {
		str := s.Text()
		if reg.MatchString(str) == false {
			continue
		}
		strs := strings.Split(str, "=")
		if len(strs) != 2 {
			continue
		}

		keys := strings.ReplaceAll(strs[0], ".", "_")
		keys = strings.Trim(keys, " ")
		vals := strings.Trim(strs[1], " ")
		if strings.Contains(strs[1], ".") {
			val, _ := strconv.ParseFloat(vals, 64)
			fields[keys] = val
		} else {
			val, _ := strconv.ParseInt(vals, 10, 64)
			fields[keys] = val
		}
	}

	pt, err := influxdb.NewPoint(p.input.Squid.MetricName, tags, fields, time.Now())
	if err != nil {
		return
	}

	err = p.output.IoFeed([]byte(pt.String()), io.Metric)
	return
}

func init() {
	inputs.Add("squid", func() inputs.Input {
		s := &Squid{}
		return s
	})
}

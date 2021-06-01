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

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	MaxGatherInterval = 30 * time.Minute
	MinGatherInterval = 1 * time.Second
)

type IoFeed func(data []byte, category, name string) error

type Squid struct {
	Active      bool
	Interval    interface{}
	Port        int
	MetricsName string
	Tags        map[string]string
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
	log    *logger.Logger
}

var (
	inputName         = "squid"
	defaultMetricName = inputName
	defaultInterval   = "60s"
	defaultPort       = 3218
	squidConfigSample = `#[inputs.squid]
#	interval = "60s"
#	port     = 3128
#	metricsName = "squid"
#	[inputs.squid.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
`
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
	p := s.genParam()
	p.log.Info("squid input started...")
	p.gather()
}

func (s *Squid) genParam() *SquidParam {
	if s.MetricsName == "" {
		s.MetricsName = defaultMetricName
	}
	if s.Interval == nil {
		s.Interval = defaultInterval
	}
	if s.Port == 0 {
		s.Port = defaultPort
	}

	input := SquidInput{*s}
	output := SquidOutput{io.NamedFeed}
	p := &SquidParam{input, output, logger.SLogger("squid")}
	return p
}

func (p *SquidParam) gather() {
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
			p.log.Info("input squid exit")
			return
		}
	}
}

func (p *SquidParam) getMetrics(isTest bool) ([]byte, error) {
	var outInfo bytes.Buffer

	tags := make(map[string]string)
	for tag, tagV := range p.input.Tags {
		tags[tag] = tagV
	}

	fields := make(map[string]interface{})
	fields["can_connect"] = true

	reg := regexp.MustCompile(" = \\d{1,}\\.{0,1}\\d{0,}$")
	portStr := fmt.Sprintf("%d", p.input.Port)

	cmd := exec.Command("squidclient", "-p", portStr, "mgr:counters")
	cmd.Stdout = &outInfo
	err := cmd.Run()
	if err != nil {
		fields["can_connect"] = false
		pt, _ := io.MakeMetric(p.input.Squid.MetricsName, tags, fields, time.Now())
		if !isTest {
			p.output.IoFeed(pt, datakit.Metric, inputName)
		}
		return pt, err
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

	pt, err := io.MakeMetric(p.input.Squid.MetricsName, tags, fields, time.Now())
	if err != nil {
		return nil, err
	}

	if !isTest {
		err = p.output.IoFeed(pt, datakit.Metric, inputName)
	}

	return pt, err
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		s := &Squid{}
		return s
	})
}

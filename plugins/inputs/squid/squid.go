package squid

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"fmt"

	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

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
	acc telegraf.Accumulator
}

type SquidParam struct {
	input  SquidInput
	output SquidOutput
}

type SquidLogWriter struct {
	io.Writer
}

const squidConfigSample = `### metric_name: the name of metric, default is "squid"
### interval: monitor interval second, unit is second. The default value is 60.
### active: whether to monitor squid.

#metric_name = "squid"
#active   = true
#interval = 60
#port     = 3128
`

var (
	ctx               context.Context
	cfun              context.CancelFunc
	defaultMetricName = "squid"
	defaultInterval   = 60
	defaultPort       = 3218
)

func (s *Squid) SampleConfig() string {
	return squidConfigSample
}

func (s *Squid) Description() string {
	return "Monitor Squid Service Status"
}

func (s *Squid) Gather(telegraf.Accumulator) error {
	return nil
}

func (s *Squid) Start(acc telegraf.Accumulator) error {
	log.Printf("I! [squid] start")
	ctx, cfun = context.WithCancel(context.Background())

	input := SquidInput{*s}
	if input.Active == false {
		return nil
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
	output := SquidOutput{acc}
	p := &SquidParam{input, output}
	go p.gather(ctx)

	return nil
}

func (s *Squid) Stop() {
	cfun()
}

func (p *SquidParam) gather(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := p.getMetrics()
		if err != nil {
			log.Printf("W! [squid] %s", err.Error())
		}

		err = internal.SleepContext(ctx, time.Duration(p.input.Interval)*time.Second)
		if err != nil {
			log.Printf("W! [squid] %s", err.Error())
		}
	}
}

func (p *SquidParam) getMetrics() error {
	var outInfo bytes.Buffer
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	fields["can_connect"] = true

	reg := regexp.MustCompile(" = \\d{1,}\\.{0,1}\\d{0,}$")
	portStr := fmt.Sprintf("%d", p.input.Port)
	cmd := exec.Command("squidclient", "-p", portStr,"mgr:counters")
	cmd.Stdout = &outInfo
	err := cmd.Run()
	if err != nil {
		fields["can_connect"] = false
		fmt.Printf("Err: %s\n", err.Error())
		p.output.acc.AddFields(p.input.MetricName, fields, tags)
		return err
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

	p.output.acc.AddFields(p.input.MetricName, fields, tags)
	return nil
}

func init() {
	inputs.Add("squid", func() telegraf.Input {
		s := &Squid{}
		return s
	})
}

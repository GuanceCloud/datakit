// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package demo explains `what should you do' when adding new inputs into datakit.
// Except that, we still adding some new testsing features to this input, such as
// election/cgroup and so on.
package demo

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	inputName = "demo"
	l         = logger.DefaultSLogger("demo")
	g         = goroutine.NewGroup(goroutine.Option{Name: "inputs_demo"})

	_ inputs.ElectionInput = (*Input)(nil)
)

type Input struct {
	collectCache []inputs.Measurement
	Tags         map[string]string
	chpause      chan bool
	EatCPU       bool `toml:"eat_cpu"`
	Election     bool `toml:"election"`
	paused       bool

	semStop *cliutils.Sem // start stop signal
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) Collect() error {
	ipt.collectCache = []inputs.Measurement{
		&demoMetric{
			name: "demo",
			tags: map[string]string{"tag_a": "a", "tag_b": "b"},
			fields: map[string]interface{}{
				"usage":       "12.3",
				"disk_size":   5e9,
				"mem_size":    1e9,
				"some_string": "hello world",
				"ok":          true,
			},
			ts:       time.Now(),
			election: ipt.Election,
		},
	}

	// simulate long-time collect..
	time.Sleep(time.Second)

	return nil
}

func (ipt *Input) Run() {
	l = logger.SLogger("demo")
	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	n := 0

	if ipt.EatCPU {
		eatCPU(runtime.NumCPU())
	}

	for {
		n++

		select {
		case ipt.paused = <-ipt.chpause:
			l.Debugf("demo paused? %v", ipt.paused)

		case <-tick.C:

			if ipt.paused {
				l.Debugf("paused")
				continue
			}

			l.Debugf("resumed")

			l.Debugf("demo input gathering...")
			start := time.Now()
			if err := ipt.Collect(); err != nil {
				l.Error(err)
			} else {
				if err := inputs.FeedMeasurement(inputName, datakit.Metric, ipt.collectCache,
					&io.Option{
						CollectCost: time.Since(start),
					}); err != nil {
					l.Errorf("FeedMeasurement: %s", err.Error())
				}

				ipt.collectCache = ipt.collectCache[:0] // Do not forget to clean cache
				io.FeedLastError(inputName, "mocked error from demo input")
			}

		case <-datakit.Exit.Wait():
			ipt.exit()
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			return
		}
	}
}

func (ipt *Input) exit() {
	close(ipt.chpause)
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (*Input) Catalog() string { return "testing" }
func (*Input) SampleConfig() string {
	return `
[inputs.demo]
  ## 这里是一些测试配置

  # 是否开启 CPU 爆满
  eat_cpu = false

  ## Set true to enable election
  election = true

[inputs.demo.tags] # 所有采集器，都应该有 tags 配置项
  # tag_a = "val1"
  # tag_b = "val2"
`
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&demoMetric{},
		&demoMetric2{},
		&demoObj{},
		&demoLog{},
	}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case ipt.chpause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case ipt.chpause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			paused:  false,
			chpause: make(chan bool, inputs.ElectionPauseChannelLength),

			Election: true,
			semStop:  cliutils.NewSem(),
		}
	})
}

func eatCPU(n int) {
	for i := 0; i < n; i++ {
		l.Debugf("start eat_cpu: %d", i)
		g.Go(func(ctx context.Context) error {
			for { //nolint:staticcheck
			}
		})
	}
}

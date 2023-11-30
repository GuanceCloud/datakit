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
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	inputName = "demo"
	l         = logger.DefaultSLogger("demo")
	g         = goroutine.NewGroup(goroutine.Option{Name: "inputs_demo"})

	_ inputs.ElectionInput = (*input)(nil)
)

type input struct {
	collectCache []*point.Point
	Tags         map[string]string
	chpause      chan bool
	EatCPU       bool `toml:"eat_cpu"`
	Election     bool `toml:"election"`

	RandomPoints int `toml:"random_points"`

	paused bool

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
}

func (ipt *input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *input) collect() error {
	var kvs point.KVs

	kvs.Add("tag_a", "a", true, false)
	kvs.Add("tag_b", "b", true, false)

	kvs.Add("usage", "12.3", false, false)
	kvs.Add("disk_size", 5e9, false, false)
	kvs.Add("mem_size", 1e9, false, false)
	kvs.Add("some_string", "hello world", false, false)
	kvs.Add("ok", true, false, false)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(time.Now()))
	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(inputName, kvs, opts...))

	// simulate long-time collect..
	time.Sleep(time.Second)

	return nil
}

func (ipt *input) Run() {
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
			if err := ipt.collect(); err != nil {
				l.Error(err)
			} else {
				if err := ipt.feeder.Feed(inputName, point.Logging, ipt.collectCache,
					&dkio.Option{CollectCost: time.Since(start)}); err != nil {
					l.Errorf("Feed failed: %s", err.Error())

					ipt.feeder.FeedLastError("mocked error from demo input",
						dkio.WithLastErrorInput(inputName),
						dkio.WithLastErrorCategory(point.Metric),
					)
				}

				ipt.collectCache = ipt.collectCache[:0] // Do not forget to clean cache
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

func (ipt *input) exit() {
	close(ipt.chpause)
}

func (ipt *input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (*input) Catalog() string { return "testing" }
func (*input) SampleConfig() string {
	return `
[inputs.demo]
  ## 这里是一些测试配置

  # 是否开启 CPU 爆满
  eat_cpu = false

  ## Set true to enable election
  election = true

  random_points = 100

[inputs.demo.tags] # 所有采集器，都应该有 tags 配置项
  # tag_a = "val1"
  # tag_b = "val2"
`
}

func (*input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&demoMetric{},
		&demoMetric2{},
		&demoObj{},
		&demoLog{},
	}
}

func (*input) AvailableArchs() []string {
	return datakit.AllOS
}

func (ipt *input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case ipt.chpause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case ipt.chpause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
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

func defaultInput() *input {
	return &input{
		paused:  false,
		chpause: make(chan bool, inputs.ElectionPauseChannelLength),

		Election: true,
		semStop:  cliutils.NewSem(),
		feeder:   dkio.DefaultFeeder(),
		Tagger:   datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}

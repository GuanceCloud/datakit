package demo

import (
	"fmt"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "demo"
	l         = logger.DefaultSLogger("demo")
)

type Input struct {
	collectCache    []inputs.Measurement
	collectObjCache []inputs.Measurement
	Tags            map[string]string
	chpause         chan bool
	EatCPU          bool `toml:"eat_cpu"`
	paused          bool
}

func (i *Input) Collect() error {
	i.collectCache = []inputs.Measurement{
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
			ts: time.Now(),
		},
	}

	// simulate long-time collect..
	time.Sleep(time.Second)

	return nil
}

func (i *Input) Run() {
	l = logger.SLogger("demo")
	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	n := 0

	if i.EatCPU {
		eatCPU(runtime.NumCPU())
	}

	for {
		n++

		select {
		case i.paused = <-i.chpause:
			l.Debugf("demo paused? %v", i.paused)

		case <-tick.C:

			if i.paused {
				l.Debugf("paused")
				continue
			}

			l.Debugf("resumed")

			l.Debugf("demo input gathering...")
			start := time.Now()
			if err := i.Collect(); err != nil {
				l.Error(err)
			} else {
				inputs.FeedMeasurement(inputName, datakit.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start), HighFreq: (n%2 == 0)})

				i.collectCache = i.collectCache[:0] // NOTE: do not forget to clean cache

				io.FeedLastError(inputName, "mocked error from demo input")
			}

		case <-datakit.Exit.Wait():
			close(i.chpause)
			return
		}
	}
}

func (i *Input) Catalog() string { return "testing" }
func (i *Input) SampleConfig() string {
	return `
[inputs.demo]
  ## 这里是一些测试配置

  # 是否开启 CPU 爆满
  eat_cpu = false

[inputs.demo.tags] # 所有采集器，都应该有 tags 配置项
	# tag_a = "val1"
	# tag_b = "val2"
`
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&demoMetric{},
		&demoMetric2{},
		&demoObj{},
		&demoLog{},
	}
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) Pause() error {
	tick := time.NewTicker(time.Second * 5)
	select {
	case i.chpause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(time.Second * 5)
	select {
	case i.chpause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			paused:  false,
			chpause: make(chan bool),
		}
	})
}

func eatCPU(n int) {
	for i := 0; i < n; i++ {
		l.Debugf("start eat_cpu: %d", i)
		go func() {
			for {
			}
		}()
	}
}

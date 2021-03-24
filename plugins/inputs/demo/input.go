package demo

import (
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

type Demo struct{}

func (i *Demo) Run() {

	l = logger.SLogger("demo")
	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	n := 0

	for {

		n++

		select {
		case <-tick.C:
			l.Debugf("demo input gathering...")
			start := time.Now()
			pt, err := io.MakeMetric("demo",
				map[string]string{
					"tag_a": "a",
					"tag_b": "b",
				},
				map[string]interface{}{
					"f1": 123,
					"f2": 456.0,
					"f3": "abc",
					"f5": true,
				},
			)

			if err != nil {
				l.Error(err)
			} else {
				time.Sleep(time.Second)
				io.Feed("demo", io.Metric, &io.Option{CollectCost: time.Since(start), HighFreq: (n%2 == 0)}, pt)
			}

		case <-datakit.Exit.Wait():
			return
		}

	}
}

func (i *Demo) Catalog() string      { return "testing" }
func (i *Demo) SampleConfig() string { return "[inputs.demo]" }

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Demo{}
	})
}

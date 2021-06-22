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

type Input struct {
	collectCache []inputs.Measurement
	Tags         map[string]string
}

type demoMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *demoMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *demoMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "demo",
		Desc: "这是一个指标集的 demo(**务必加上每个指标集的描述**)",
		Tags: map[string]interface{}{
			"tag_a": &inputs.TagInfo{Desc: "示例 tag A"},
			"tag_b": &inputs.TagInfo{Desc: "示例 tag B"},
		},
		Fields: map[string]interface{}{
			"usage": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "this is CPU usage",
			},
			"disk_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "this is disk size",
			},
			"mem_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeIByte,
				Desc:     "this is memory size",
			},
			"some_string": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "some string field",
			},
			"ok": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "some boolean field",
			},
		},
	}
}

func (i *Input) Collect() error {

	i.collectCache = []inputs.Measurement{
		&demoMeasurement{
			name: "demo",
			tags: map[string]string{"tag_a": "a", "tag_b": "b"},
			fields: map[string]interface{}{
				"usage":       "12.3",
				"disk_size":   5e9,
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

	for {

		n++

		select {
		case <-tick.C:

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
			return
		}
	}
}

func (i *Input) Catalog() string { return "testing" }
func (i *Input) SampleConfig() string {
	return `
[inputs.demo]
# 这里无需任何配置

[inputs.demo.tags] # 所有采集器，都应该有 tags 配置项
	# tag_a = "val1"
	# tag_b = "val2"
`
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&demoMeasurement{},
	}
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}

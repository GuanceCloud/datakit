package mock

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type (
	Mock struct {
	}
)

func (m *Mock) SampleConfig() string {
	return ""
}

func (m *Mock) Description() string {
	return ""
}

func (m *Mock) makeMetric(value interface{}, name ...string) telegraf.Metric {
	if value == nil {
		panic("Cannot use a nil value")
	}
	measurement := "test1"
	if len(name) > 0 {
		measurement = name[0]
	}
	tags := map[string]string{"tag1": "value1"}
	pt, _ := metric.New(
		measurement,
		tags,
		map[string]interface{}{"value": value},
		time.Now(),
	)
	return pt
}

func (m *Mock) Gather(acc telegraf.Accumulator) error {

	// acc.AddMetric(m.makeMetric(1.0))

	// fields := map[string]interface{}{
	// 	"val1": 3.14,
	// 	"val2": 10,
	// 	"val3": true,
	// 	"val4": "",
	// }
	// tags := map[string]string{
	// 	"t1": "aa",
	// }
	// acc.AddFields("test2", fields, tags)
	return nil
}

func init() {
	inputs.Add("mock", func() telegraf.Input {
		return &Mock{}
	})
}

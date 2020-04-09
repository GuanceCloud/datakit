package mongodboplog

import (
	"context"
	"sync"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Mongodboplog struct {
	Config Config `toml:"mongodb_oplog"`

	ctx    context.Context
	cancel context.CancelFunc
	acc    telegraf.Accumulator
	wg     *sync.WaitGroup
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input {
		m := &Mongodboplog{}
		m.ctx, m.cancel = context.WithCancel(context.Background())
		return m
	})
}

func (m *Mongodboplog) Start(acc telegraf.Accumulator) error {

	m.acc = acc
	m.wg = new(sync.WaitGroup)

	for _, sub := range m.Config.Subscribes {
		m.wg.Add(1)
		stream := newStream(&sub, m)
		go stream.start(m.wg)
	}

	return nil
}

func (m *Mongodboplog) Stop() {
	m.cancel()
	m.wg.Wait()
}

func (_ *Mongodboplog) SampleConfig() string {
	return mongodboplogConfigSample
}

func (_ *Mongodboplog) Description() string {
	return "Convert MongoDB Database to Dataway"
}

func (_ *Mongodboplog) Gather(telegraf.Accumulator) error {
	return nil
}

func (m *Mongodboplog) ProcessPts(pts []*influxdb.Point) error {
	for _, pt := range pts {
		fields, err := pt.Fields()
		if err != nil {
			return err
		}
		pt_metric, err := metric.New(pt.Name(), pt.Tags(), fields, pt.Time())
		if err != nil {
			return err
		}
		m.acc.AddMetric(pt_metric)
	}
	return nil
}

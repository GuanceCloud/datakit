package etcd

import (
	"context"
	"log"
	"sync"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Etcd struct {
	Config Config `toml:"etcd"`

	ctx    context.Context
	cancel context.CancelFunc
	acc    telegraf.Accumulator
	wg     *sync.WaitGroup
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input {
		e := &Etcd{}
		e.ctx, e.cancel = context.WithCancel(context.Background())
		return e
	})
}

func (e *Etcd) Start(acc telegraf.Accumulator) error {
	e.acc = acc
	e.wg = new(sync.WaitGroup)

	log.Printf("I! [Etcd] start\n")
	log.Printf("I! [Etcd] load subscribes count: %d\n", len(e.Config.Subscribes))
	for _, sub := range e.Config.Subscribes {
		e.wg.Add(1)
		s := sub
		stream := newStream(&s, e)
		go stream.start(e.wg)
	}

	return nil
}

func (e *Etcd) Stop() {
	log.Printf("I! [Etcd] stop\n")
	e.cancel()
	e.wg.Wait()
}

func (_ *Etcd) SampleConfig() string {
	return etcdConfigSample
}

func (_ *Etcd) Description() string {
	return "Convert Etcd collection data to Dataway"
}

func (_ *Etcd) Gather(telegraf.Accumulator) error {
	return nil
}

func (e *Etcd) ProcessPts(pts []*influxdb.Point) error {
	for _, pt := range pts {
		fields, err := pt.Fields()
		if err != nil {
			return err
		}
		pt_metric, err := metric.New(pt.Name(), pt.Tags(), fields, pt.Time())
		if err != nil {
			return err
		}
		log.Printf("D! [Etcd] metric: %v\n", pt_metric)
		e.acc.AddMetric(pt_metric)
	}
	return nil
}

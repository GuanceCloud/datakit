// +build linux

package containerd

import (
	"context"
	"log"
	"sync"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Containerd struct {
	Config Config `toml:"containerd"`

	ctx    context.Context
	cancel context.CancelFunc
	acc    telegraf.Accumulator
	wg     *sync.WaitGroup
}

func init() {
	inputs.Add(pluginName, func() inputs.Input {
		e := &Containerd{}
		return e
	})
}

func (e *Containerd) Start(acc telegraf.Accumulator) error {

	e.ctx, e.cancel = context.WithCancel(context.Background())
	e.acc = acc
	e.wg = new(sync.WaitGroup)

	log.Printf("I! [Containerd] start\n")
	log.Printf("I! [Containerd] load subscribes count: %d\n", len(e.Config.Subscribes))
	for _, sub := range e.Config.Subscribes {
		e.wg.Add(1)
		s := sub
		stream := newStream(&s, e)
		go stream.start(e.wg)
	}

	return nil
}

func (e *Containerd) Stop() {
	e.cancel()
	e.wg.Wait()
	log.Printf("I! [Containerd] stop\n")
}

func (_ *Containerd) Catalog() string {
	return "containerd"
}

func (_ *Containerd) SampleConfig() string {
	return containerdConfigSample
}

func (_ *Containerd) Description() string {
	return "Convert Containerd collection metrics to Dataway"
}

func (_ *Containerd) Gather(telegraf.Accumulator) error {
	return nil
}

func (e *Containerd) ProcessPts(pts []*influxdb.Point) error {
	for _, pt := range pts {
		fields, err := pt.Fields()
		if err != nil {
			return err
		}
		pt_metric, err := metric.New(pt.Name(), pt.Tags(), fields, pt.Time())
		if err != nil {
			return err
		}
		log.Printf("D! [Containerd] metric: %v\n", pt_metric)
		e.acc.AddMetric(pt_metric)
	}
	return nil
}

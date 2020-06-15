// +build !solaris

package tailf

import (
	"context"
	"log"
	"sync"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Tailf struct {
	Config Config `toml:"tailf"`

	ctx    context.Context
	cancel context.CancelFunc
	acc    telegraf.Accumulator
	wg     *sync.WaitGroup
}

func init() {
	inputs.Add(pluginName, func() inputs.Input {
		t := &Tailf{}
		return t
	})
}

func (t *Tailf) Start(acc telegraf.Accumulator) error {

	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.acc = acc
	t.wg = new(sync.WaitGroup)

	log.Printf("I! [Tailf] start\n")
	log.Printf("I! [Tailf] load subscribes count %d\n", len(t.Config.Subscribes))
	for _, sub := range t.Config.Subscribes {
		t.wg.Add(1)
		s := sub
		stream := newStream(&s, t)
		go stream.start(t.wg)
	}

	return nil
}

func (t *Tailf) Stop() {
	t.cancel()
	t.wg.Wait()
	log.Printf("I! [Tailf] stop\n")
}

func (_ *Tailf) Catalog() string {
	return "log"
}

func (_ *Tailf) SampleConfig() string {
	return tailfConfigSample
}

func (_ *Tailf) Description() string {
	return "Convert stream a log file (like the tail -f command) to Dataway"
}

func (_ *Tailf) Gather(telegraf.Accumulator) error {
	return nil
}

func (e *Tailf) ProcessPts(pts []*influxdb.Point) error {
	for _, pt := range pts {
		fields, err := pt.Fields()
		if err != nil {
			return err
		}
		pt_metric, err := metric.New(pt.Name(), pt.Tags(), fields, pt.Time())
		if err != nil {
			return err
		}
		log.Printf("D! [Tailf] metric: %v\n", pt_metric)
		e.acc.AddMetric(pt_metric)
	}
	return nil
}

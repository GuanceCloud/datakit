package replication

import (
	"context"
	"log"
	"sync"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Replication struct {
	Config Config `toml:"replication"`

	ctx    context.Context
	cancel context.CancelFunc
	acc    telegraf.Accumulator
	wg     *sync.WaitGroup
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input {
		r := &Replication{}
		r.ctx, r.cancel = context.WithCancel(context.Background())
		return r
	})
}

func (r *Replication) Start(acc telegraf.Accumulator) error {

	r.acc = acc
	r.wg = new(sync.WaitGroup)

	log.Printf("I! [Replication] start\n")
	log.Printf("I! [Replication] load subscribes count: %d\n", len(r.Config.Subscribes))

	for _, sub := range r.Config.Subscribes {
		r.wg.Add(1)
		s := sub
		stream := newStream(&s)
		go stream.start(r, r.ctx, r.wg)
	}

	return nil
}

func (r *Replication) Stop() {
	r.cancel()
	r.wg.Wait()
	log.Printf("I! [Replication] stop\n")
}

func (_ *Replication) SampleConfig() string {
	return replicationConfigSample
}

func (_ *Replication) Description() string {
	return "Convert replication to Dataway"
}

func (_ *Replication) Gather(telegraf.Accumulator) error {
	return nil
}

func (r *Replication) ProcessPts(pts []*influxdb.Point) error {
	for _, pt := range pts {
		fields, err := pt.Fields()
		if err != nil {
			return err
		}
		pt_metric, err := metric.New(pt.Name(), pt.Tags(), fields, pt.Time())
		if err != nil {
			return err
		}
		log.Printf("D! [Replication] metric: %v\n", pt_metric)
		r.acc.AddMetric(pt_metric)
	}
	return nil
}

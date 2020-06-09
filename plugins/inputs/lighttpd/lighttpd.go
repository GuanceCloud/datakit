package lighttpd

import (
	"context"
	"log"
	"sync"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Lighttpd struct {
	Config Config `toml:"lighttpd"`

	ctx    context.Context
	cancel context.CancelFunc
	acc    telegraf.Accumulator
	wg     *sync.WaitGroup
}

func init() {
	inputs.Add(pluginName, func() inputs.Input {
		lt := &Lighttpd{}
		return lt
	})
}

func (lt *Lighttpd) Start(acc telegraf.Accumulator) error {

	lt.ctx, lt.cancel = context.WithCancel(context.Background())
	lt.acc = acc
	lt.wg = new(sync.WaitGroup)

	log.Printf("I! [Lighttpd] start\n")
	log.Printf("I! [Lighttpd] load subscribes count %d\n", len(lt.Config.Subscribes))
	for _, sub := range lt.Config.Subscribes {
		lt.wg.Add(1)
		s := sub
		stream := newStream(&s, lt)
		go stream.start(lt.wg)
	}

	return nil
}

func (lt *Lighttpd) Stop() {
	lt.cancel()
	lt.wg.Wait()
	log.Printf("I! [Lighttpd] stop\n")
}

func (_ *Lighttpd) SampleConfig() string {
	return lighttpdConfigSample
}

func (_ *Lighttpd) Catalog() string {
	return "lighttpd"
}

func (_ *Lighttpd) Description() string {
	return "Convert Lighttpd collection data to Dataway"
}

func (_ *Lighttpd) Gather(telegraf.Accumulator) error {
	return nil
}

func (lt *Lighttpd) ProcessPts(pts []*influxdb.Point) error {
	for _, pt := range pts {
		fields, err := pt.Fields()
		if err != nil {
			return err
		}
		pt_metric, err := metric.New(pt.Name(), pt.Tags(), fields, pt.Time())
		if err != nil {
			return err
		}
		log.Printf("D! [Lighttpd] metric: %v\n", pt_metric)
		lt.acc.AddMetric(pt_metric)
	}
	return nil
}

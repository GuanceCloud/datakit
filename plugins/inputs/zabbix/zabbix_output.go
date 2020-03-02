package zabbix

import (
	"context"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type ZabbixOutput struct {
	ctx  context.Context
	cfun context.CancelFunc
	acc  telegraf.Accumulator
}

func (o *ZabbixOutput) ProcessPts(pts []*influxdb.Point) error {
	for _, pt := range pts {
		fields, err := pt.Fields()
		if err != nil {
			return err
		}
		pt_metric, err := metric.New(pt.Name(), pt.Tags(), fields, pt.Time())
		if err != nil {
			return err
		}
		o.acc.AddMetric(pt_metric)
	}
	return nil
}

package zabbix

import (
	influxdb "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type IoFeed func(data []byte, category, name string) error

type ZabbixOutput struct {
	IoFeed
}

func (z *ZabbixParam) ProcessPts(pts []*influxdb.Point) error {
	for _, pt := range pts {
		fields, err := pt.Fields()
		if err != nil {
			return err
		}

		tags := pt.Tags()
		for tag, tagv := range z.input.Tags {
			tags[tag] = tagv
		}

		ps, err := io.MakeMetric(pt.Name(), tags, fields, pt.Time())
		if err != nil {
			return err
		}
		z.log.Debug(string(ps))
		err = z.output.IoFeed(ps, io.Metric, inputName)
		if err != nil {
			return err
		}
	}
	return nil
}

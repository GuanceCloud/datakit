package io

import (
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

type Point struct {
	*influxdb.Point
}

func WrapPoint(pts []*influxdb.Point) (x []*Point) {
	for _, pt := range pts {
		x = append(x, &Point{pt})
	}
	return
}

var _ sinkcommon.ISinkPoint = new(Point)

func (p *Point) ToPoint() *influxdb.Point {
	return p.Point
}

func (p *Point) ToJSON() (*sinkcommon.JSONPoint, error) {
	fields, err := p.Point.Fields()
	if err != nil {
		return nil, err
	}
	return &sinkcommon.JSONPoint{
		Measurement: p.ToPoint().Name(),
		Tags:        p.Point.Tags(),
		Fields:      fields,
		Time:        p.Point.Time(),
	}, nil
}

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

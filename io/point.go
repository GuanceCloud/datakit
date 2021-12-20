package io

import (
	influxdb "github.com/influxdata/influxdb1-client/v2"
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

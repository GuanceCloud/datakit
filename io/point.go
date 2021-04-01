package io

import (
	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type Point struct {
	*influxdb.Point
}

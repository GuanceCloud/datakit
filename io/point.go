package io

import (
	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type Point *influxdb.Point
type Points []*influxdb.Point

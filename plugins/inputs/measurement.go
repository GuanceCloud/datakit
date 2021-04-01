package inputs

import (
	"fmt"
	//"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	Int = iota
	Float
	String
	Bool

	// TODO:
	// Prometheus metric types: https://prometheus.io/docs/concepts/metric_types/
	// DataDog metricc types: https://docs.datadoghq.com/developers/metrics/types/?tab=count#metric-types
	Gauge = iota
	Count
	// add more...
)

// units
const (
	UnknownUnit = ""
	SizeByte    = "byte"  // 1000
	SizeIByte   = "ibyte" // 1024

	DurationNS     = "ns"
	DurationUS     = "us"
	DurationMS     = "ms"
	DurationSecond = "s"
	DurationMinute = "m"
	DurationHour   = "h"
	DurationDay    = "d"

	Percent = "perc"

	// TODO: add more...
)

type Measurement interface {
	LineProto() (io.Point, error)
	Info() *MeasurementInfo
}

type FieldInfo struct {
	Type     int // gauge/count/...
	DataType int // int/float/bool/...
	Unit     string
	Desc     string // markdown string
	Disabled bool
}

type MeasurementInfo struct {
	Name   string
	Fields map[string]*FieldInfo
	// tags ingored
}

func FeedMeasurement(name, category string, measurements []Measurement, opt *io.Option) error {
	if len(measurements) == 0 {
		return fmt.Errorf("no points")
	}

	var pts io.Points
	for _, m := range measurements {
		if pt, err := m.LineProto(); err != nil {
			return err
		} else {
			pts = append(pts, pt)
		}
	}
	return io.Feed(name, category, pts, opt)
}

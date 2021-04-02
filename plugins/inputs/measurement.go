package inputs

import (
	"fmt"
	"strings"
	//"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	Int    = "int"
	Float  = "float"
	String = "string"
	Bool   = "bool"

	// TODO:
	// Prometheus metric types: https://prometheus.io/docs/concepts/metric_types/
	// DataDog metricc types: https://docs.datadoghq.com/developers/metrics/types/?tab=count#metric-types
	Gauge = "gauge"
	Count = "count"
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
	LineProto() (*io.Point, error)
	Info() *MeasurementInfo
}

type FieldInfo struct {
	Type     string // gauge/count/...
	DataType string // int/float/bool/...
	Unit     string
	Desc     string // markdown string
	Disabled bool
}

type TagInfo struct {
	Desc string
}

type MeasurementInfo struct {
	Name   string
	Fields map[string]*FieldInfo
	Tags   map[string]*TagInfo
	// tags ingored
}

func (m *MeasurementInfo) FieldsMarkdownTable() string {
	tableHeader := `
| 指标 | 描述   | 类型 | 单位 |
| ---: | :----: | ---  | ---- |`

	rows := []string{tableHeader}
	for k, f := range m.Fields {
		rows = append(rows, fmt.Sprintf("|`%s`|%s|%s|%s|", k, f.Desc, f.DataType, f.Unit))
	}
	return strings.Join(rows, "\n")
}

func (m *MeasurementInfo) TagsMarkdownTable() string {
	tableHeader := `
| 标签名  | 描述    |
|---:     |---------|`

	rows := []string{tableHeader}
	for k, t := range m.Tags {
		rows = append(rows, fmt.Sprintf("|`%s`|%s|", k, t.Desc))
	}
	return strings.Join(rows, "\n")
}

func FeedMeasurement(name, category string, measurements []Measurement, opt *io.Option) error {
	if len(measurements) == 0 {
		return fmt.Errorf("no points")
	}

	var pts []*io.Point
	for _, m := range measurements {
		if pt, err := m.LineProto(); err != nil {
			return err
		} else {
			pts = append(pts, pt)
		}
	}
	return io.Feed(name, category, pts, opt)
}

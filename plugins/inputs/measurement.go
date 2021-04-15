package inputs

import (
	"fmt"
	"sort"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	TODO = "TODO" // global todo string

	Int    = "int"
	Float  = "float"
	String = "string"
	Bool   = "bool"

	// TODO:
	// Prometheus metric types: https://prometheus.io/docs/concepts/metric_types/
	// DataDog metricc types: https://docs.datadoghq.com/developers/metrics/types/?tab=count#metric-types
	Gauge = "gauge"
	Count = "count"
	Rate  = "rate"
	// add more...
)

// units
const (
	UnknownUnit = "-"

	SizeByte  = "byte(1k=1000)"  // 1000
	SizeIByte = "ibyte(1k=1024)" // 1024
	SizeMiB   = "MiB"            // 1024

	NCount = "count"

	// time units
	DurationNS     = "nsec"
	DurationUS     = "usec"
	DurationMS     = "msec"
	DurationSecond = "second"
	DurationMinute = "minute"
	DurationHour   = "hour"
	DurationDay    = "day"

	Percent = "%"

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
	Fields map[string]interface{}
	Tags   map[string]interface{}
}

func (m *MeasurementInfo) FieldsMarkdownTable() string {
	tableHeader := `
| 指标 | 描述  | 数据类型 | 单位   |
| ---- | ----  | :---:    | :----: |`

	rows := []string{tableHeader}
	keys := sortMapKey(m.Fields)
	for _, key := range keys { // XXX: f.Type not used
		f, ok := m.Fields[key].(*FieldInfo)
		if !ok {
			continue
		}

		if f.Unit == "" {
			rows = append(rows, fmt.Sprintf("|`%s`|%s|%s|-|", key, f.Desc, f.DataType))
		} else {
			rows = append(rows, fmt.Sprintf("|`%s`|%s|%s|%s|", key, f.Desc, f.DataType, f.Unit))
		}
	}
	return strings.Join(rows, "\n")
}

func (m *MeasurementInfo) TagsMarkdownTable() string {
	if len(m.Tags) == 0 {
		return "暂无"
	}

	tableHeader := `
| 标签名  | 描述    |
|----     | --------|`

	rows := []string{tableHeader}
	keys := sortMapKey(m.Tags)
	for _, key := range keys {
		t, ok := m.Tags[key].(*TagInfo)
		if !ok {
			continue
		}
		rows = append(rows, fmt.Sprintf("|`%s`|%s|", key, t.Desc))
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

func NewTagInfo(desc string) *TagInfo {
	return &TagInfo{
		Desc: desc,
	}
}

func sortMapKey(m map[string]interface{}) (res []string) {

	for k, _ := range m {
		res = append(res, k)
	}

	sort.Strings(res)
	return
}

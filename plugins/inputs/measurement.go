package inputs

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var (
	MonofontOnTagFieldName = true
	TODO                   = "TODO" // global todo string
)

const (
	Int    = "int"
	Float  = "float"
	String = "string"
	Bool   = "bool"

	// TODO:
	// Prometheus metric types: https://prometheus.io/docs/concepts/metric_types/
	// DataDog metricc types: https://docs.datadoghq.com/developers/metrics/types/?tab=count#metric-types
	Gauge       = "gauge"
	Count       = "count"
	Rate        = "rate"
	UnknownType = "unknown"
	// add more...
)

// units.
const (
	UnknownUnit = "-"

	SizeByte = "B"
	SizeMiB  = "MB"

	NCount = "count"

	// time units.
	DurationNS     = "ns"
	DurationUS     = "μs"
	DurationMS     = "ms"
	DurationSecond = "s"
	DurationMinute = "min"
	DurationHour   = "h"
	DurationDay    = "d"

	// timestamp units.
	TimestampNS  = "nsec"
	TimestampUS  = "usec"
	TimestampMS  = "msec"
	TimestampSec = "sec"

	Percent = "percent"

	// TODO: add more...
	BytesPerSec    = "B/S"
	RequestsPerSec = "reqps"
	Celsius        = "C"
)

type Measurement interface {
	LineProto() (*io.Point, error)
	Info() *MeasurementInfo
}

type FieldInfo struct {
	Type     string `json:"type"`      // gauge/count/...
	DataType string `json:"data_type"` // int/float/bool/...
	Unit     string `json:"unit"`
	Desc     string `json:"desc"` // markdown string
	Disabled bool   `json:"disabled"`
}

type TagInfo struct {
	Desc string
}

type MeasurementInfo struct {
	Name   string
	Desc   string
	Type   string
	Fields map[string]interface{}
	Tags   map[string]interface{}
}

func (m *MeasurementInfo) FieldsMarkdownTable() string {
	const tableHeader = `
| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |`
	const monoRowfmt = "|`%s`|%s|%s|%s|" // 指标/标签列等宽字体展示
	const normalRowfmt = "|%s|%s|%s|%s|"

	rowfmt := monoRowfmt
	if !MonofontOnTagFieldName {
		rowfmt = normalRowfmt
	}

	rows := []string{tableHeader}
	keys := sortMapKey(m.Fields)
	for _, key := range keys { // XXX: f.Type not used
		f, ok := m.Fields[key].(*FieldInfo)
		if !ok {
			continue
		}

		unit := f.Unit
		if unit == "" {
			unit = UnknownUnit
		}

		rows = append(rows, fmt.Sprintf(rowfmt, key, f.Desc, f.DataType, unit))
	}
	return strings.Join(rows, "\n")
}

func (m *MeasurementInfo) TagsMarkdownTable() string {
	if len(m.Tags) == 0 {
		return "暂无"
	}

	tableHeader := `
| 标签名 | 描述    |
|  ----  | --------|`

	rows := []string{tableHeader}
	keys := sortMapKey(m.Tags)
	for _, key := range keys {
		desc := ""
		switch t := m.Tags[key].(type) {
		case *TagInfo:
			desc = t.Desc
		case TagInfo:
			desc = t.Desc
		default:
		}

		if MonofontOnTagFieldName {
			rows = append(rows, fmt.Sprintf("|`%s`|%s|", key, desc))
		} else {
			rows = append(rows, fmt.Sprintf("|%s|%s|", key, desc))
		}
	}
	return strings.Join(rows, "\n")
}

func FeedMeasurement(name, category string, measurements []Measurement, opt *io.Option) error {
	if len(measurements) == 0 {
		return fmt.Errorf("no points")
	}

	pts, err := GetPointsFromMeasurement(measurements)
	if err != nil {
		return err
	}

	return io.Feed(name, category, pts, opt)
}

func GetPointsFromMeasurement(measurements []Measurement) ([]*io.Point, error) {
	var pts []*io.Point
	for _, m := range measurements {
		if pt, err := m.LineProto(); err != nil {
			return nil, err
		} else {
			pts = append(pts, pt)
		}
	}
	return pts, nil
}

func NewTagInfo(desc string) *TagInfo {
	return &TagInfo{
		Desc: desc,
	}
}

func sortMapKey(m map[string]interface{}) (res []string) {
	for k := range m {
		res = append(res, k)
	}
	sort.Strings(res)
	return
}

type ReporterMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (e ReporterMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(e.name, e.tags, e.fields, e.ts)
}

func (e ReporterMeasurement) Info() *MeasurementInfo {
	return &MeasurementInfo{}
}

func FeedReporter(reporter *io.Reporter) {
	measurement := getReporterMeasurement(reporter)
	err := FeedMeasurement("datakit", datakit.Logging, []Measurement{measurement}, nil)
	if err != nil {
		l.Errorf("send datakit logging error: %s", err.Error())
	}
}

func getReporterMeasurement(reporter *io.Reporter) ReporterMeasurement {
	now := time.Now()
	m := ReporterMeasurement{
		name: "datakit",
		ts:   now,
	}

	m.tags = reporter.Tags()
	m.fields = reporter.Fields()
	return m
}

// BuildTags used to test all measurements tags.
func BuildTags(t *testing.T, ti map[string]interface{}) map[string]string {
	t.Helper()
	x := map[string]string{}
	for k := range ti {
		x[k] = k + "-tag-val"
	}
	return x
}

// BuildFields used to test all measurements fields.
func BuildFields(t *testing.T, fi map[string]interface{}) map[string]interface{} {
	t.Helper()
	x := map[string]interface{}{}
	for k, v := range fi {
		switch _v := v.(type) {
		case *FieldInfo:
			switch _v.DataType {
			case Float:
				x[k] = 1.23
			case Int:
				x[k] = 123
			case String:
				x[k] = "abc123"
			case Bool:
				x[k] = false
			default:
				t.Errorf("invalid data field for field: %s", k)
			}

		default:
			t.Errorf("expect *FieldInfo")
		}
	}
	return x
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var TODO = "-" // global todo string

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
	Histogram   = "histogram"
	Summary     = "summary"
	Rate        = "rate"
	UnknownType = "unknown"
	EnumType    = "enum"
	// add more...

	CollectorUpMeasurement = "collector"
)

// All Units list.
//
//	See https://guanceyun.feishu.cn/wiki/HjVgwzx7iiGFO0kpNYmck8HFnef
const (
	UnknownUnit = "-"
	NoUnit      = "N/A"

	EnumValue = "enum"

	SizeByte  = "digital,B"
	SizeKB    = "digital,KB"
	SizeKBits = "digital,Kb"
	SizeMB    = "digital,MB"
	SizeMBits = "digital,Mb"
	SizeGB    = "digital,GB"
	SizeTB    = "digital,TB"
	NCount    = "count"

	// time units.
	DurationPS     = "time,ps"
	DurationNS     = "time,ns"
	DurationUS     = "time,μs"
	DurationMS     = "time,ms"
	DurationSecond = "time,s"
	DurationMinute = "time,min"
	DurationHour   = "time,h"
	DurationDay    = "time,d"

	// timestamp units.
	TimestampNS  = "timeStamp,nsec"
	TimestampUS  = "timeStamp,usec"
	TimestampMS  = "timeStamp,msec"
	TimestampSec = "timeStamp,sec"

	Percent        = "percent,percent"         // percent 0~100
	PercentDecimal = "percent,percent_decimal" // percent 0~1

	// TODO: add more...
	BytesPerSec    = "traffic,B/S"
	RequestsPerSec = "throughput,reqps"
	Celsius        = "temperature,C"
	Ampere         = "ampere"
	Watt           = "watt"
	Volt           = "volt"
	FrequencyMHz   = "frequency,MHz"
	RPMPercent     = "RPM%"
	RotationRete   = "RPM"
	PartPerMillion = "PPM"
	Millicores     = "milli-cores"
	FramePerSecond = "fps"
)

type Measurement interface {
	Info() *MeasurementInfo
}

type MeasurementV2 interface {
	Measurement
	Point() *point.Point
}

type UpMeasurement struct {
	Name     string
	Tags     map[string]string
	Fields   map[string]interface{}
	Election bool
}

func (m *UpMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	if m.Election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPoint(m.Name,
		append(point.NewTags(m.Tags), point.NewKVs(m.Fields)...),
		opts...)
}

func (m *UpMeasurement) Info() *MeasurementInfo { //nolint:funlen
	return &MeasurementInfo{
		Name:           CollectorUpMeasurement,
		Cat:            point.Metric,
		MetaDuplicated: true, // This measurement are shared among multiple collectors.
		Fields: map[string]interface{}{
			"up": &FieldInfo{
				DataType: Int,
				Type:     Gauge,
				Unit:     UnknownUnit,
				Desc:     "",
			},
		},
		Tags: map[string]interface{}{
			"job": &TagInfo{
				Desc: "Server name of the instance",
			},
			"instance": &TagInfo{
				Desc: "Server addr of the instance",
			},
		},
	}
}

// EmptyMeasurement label a collector that got no MeasurementInfo exported.
type EmptyMeasurement struct{}

var DefaultEmptyMeasurement = &EmptyMeasurement{}

func (e *EmptyMeasurement) Info() *MeasurementInfo {
	return nil
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
	Name   string                 `json:"-"`
	Desc   string                 `json:"desc"`
	DescZh string                 `json:"desc_zh"`
	Fields map[string]interface{} `json:"fields"`
	Tags   map[string]interface{} `json:"tags"`

	Cat point.Category `json:"-"`

	// do not export the measurement info
	ExportSkip bool `json:"-"`

	// This maybe a duplicated measurement info.
	// For some collector that got the same measurement(such as custome object).
	MetaDuplicated bool `json:"-"`
}

func (m *MeasurementInfo) Type() string {
	return m.Cat.String()
}

type CommonMeasurement struct {
	Name   string
	Fields map[string]interface{}
	Tags   map[string]string
}

// MarkdownTable output tags and field in single mardkdown table.
func (m *MeasurementInfo) MarkdownTable() string {
	const tableHeader = `
| Tags & Fields| Description |
| ----   |:----        |`

	const tagRowfmt = "|**%s**<br>(`tag`)|%s|"
	const fieldRowfmt = "|**%s**|%s<br>*Type: %s*<br>*Unit: %s*|"

	rows := []string{tableHeader}
	// show tags before fields
	keys := sortMapKey(m.Tags)
	for _, key := range keys {
		f, ok := m.Tags[key].(*TagInfo)
		if !ok {
			continue
		}

		rows = append(rows, fmt.Sprintf(tagRowfmt, key, f.Desc))
	}

	keys = sortMapKey(m.Fields)
	for _, key := range keys { // XXX: f.Type not used
		f, ok := m.Fields[key].(*FieldInfo)
		if !ok {
			continue
		}

		unit := f.Unit
		if unit == "" {
			unit = NoUnit
		}

		rows = append(rows, fmt.Sprintf(fieldRowfmt, key, f.Desc, f.DataType, unit))
	}
	return strings.Join(rows, "\n")
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

type I18n int

const (
	I18nZh I18n = iota
	I18nEn
)

func (x I18n) String() string {
	switch x {
	case I18nZh:
		return "zh"
	case I18nEn:
		return "en"
	default:
		panic(fmt.Sprintf("should not been here: unsupport language: %s", x.String()))
	}
}

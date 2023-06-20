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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

var (
	MonofontOnTagFieldName = true
	TODO                   = "-" // global todo string
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

	SizeByte  = "B"
	SizeKB    = "KB"
	SizeKBits = "Kb"
	SizeMB    = "MB"
	SizeMBits = "Mb"
	SizeGB    = "GB"
	NCount    = "count"

	// time units.
	DurationPS     = "ps"
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
	RequestsPerSec = "req/s"
	Celsius        = "C"
	Ampere         = "ampere"
	Watt           = "watt"
	Volt           = "volt"
	FrequencyMHz   = "MHz"
	RPMPercent     = "RPM%"
	RotationRete   = "RPM"
	PartPerMillion = "PPM"
)

type Measurement interface {
	LineProto() (*dkpt.Point, error)
	Info() *MeasurementInfo
}

type MeasurementV2 interface {
	Measurement
	Point() *point.Point
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

type CommonMeasurement struct {
	Name   string
	Fields map[string]interface{}
	Tags   map[string]string
}

func (m *MeasurementInfo) FieldsMarkdownTable() string {
	const tableHeader = `
| Metric | Description | Type | Unit |
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
		return "NA"
	}

	tableHeader := `
| Tag | Description |
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

func GetPointsFromMeasurement(measurements []Measurement) ([]*dkpt.Point, error) {
	var pts []*dkpt.Point
	for _, m := range measurements {
		if pt, err := m.LineProto(); err != nil {
			l.Warnf("make point failed: %v, ignore", err)
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

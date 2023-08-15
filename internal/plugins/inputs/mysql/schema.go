// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"time"

	gcPoint "github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type schemaMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	ts       time.Time
	election bool
}

func (m *schemaMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOptElectionV2(m.election))
}

// Point implement MeasurementV2.
func (m *schemaMeasurement) Point() *gcPoint.Point {
	opts := gcPoint.DefaultMetricOptions()

	if m.election {
		opts = append(opts, gcPoint.WithExtraTags(point.GlobalElectionTags()))
	}

	return gcPoint.NewPointV2([]byte(m.name),
		append(gcPoint.NewTags(m.tags), gcPoint.NewKVs(m.fields)...),
		opts...)
}

func (m *schemaMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mysql_schema",
		Desc: "MySQL schema information",
		Type: "metric",
		Fields: map[string]interface{}{
			"schema_size": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Size of schemas(MiB)",
			},
			"query_run_time_avg": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "Avg query response time per schema.",
			},
		},
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
			"schema_name": &inputs.TagInfo{
				Desc: "Schema name",
			},
			"host": &inputs.TagInfo{
				Desc: "The server host address",
			},
		},
	}
}

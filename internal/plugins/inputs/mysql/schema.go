// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type schemaMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *schemaMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *schemaMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameMySQLSchema,
		Desc: "MySQL schema information",
		Cat:  point.Metric,
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
				Unit:     inputs.DurationUS,
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

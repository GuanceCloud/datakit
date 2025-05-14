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

type tbMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *tbMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *tbMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc: "MySQL table information",
		Cat:  point.Metric,
		Name: metricNameMySQLTableSchema,
		Fields: map[string]interface{}{
			// status
			"data_free": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows. Some storage engines, such as MyISAM, store the exact count. For other storage engines, such as InnoDB, this value is an approximation, and may vary from the actual value by as much as 40% to 50%. In such cases, use SELECT COUNT(*) to obtain an accurate count.",
			},
			// status
			"data_length": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "For InnoDB, DATA_LENGTH is the approximate amount of space allocated for the clustered index, in bytes. Specifically, it is the clustered index size, in pages, multiplied by the InnoDB page size",
			},
			// status
			"index_length": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "For InnoDB, INDEX_LENGTH is the approximate amount of space allocated for non-clustered indexes, in bytes. Specifically, it is the sum of non-clustered index sizes, in pages, multiplied by the InnoDB page size",
			},
			// status
			"table_rows": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows. Some storage engines, such as MyISAM, store the exact count. For other storage engines, such as InnoDB, this value is an approximation, and may vary from the actual value by as much as 40% to 50%. In such cases, use SELECT COUNT(*) to obtain an accurate count.",
			},
		},
		Tags: map[string]interface{}{
			"engine": &inputs.TagInfo{
				Desc: "The storage engine for the table. See The InnoDB Storage Engine, and Alternative Storage Engines.",
			},
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
			"table_name": &inputs.TagInfo{
				Desc: "The name of the table.",
			},
			"table_schema": &inputs.TagInfo{
				Desc: "The name of the schema (database) to which the table belongs.",
			},
			"table_type": &inputs.TagInfo{
				Desc: "BASE TABLE for a table, VIEW for a view, or SYSTEM VIEW for an INFORMATION_SCHEMA table.",
			},
			"version": &inputs.TagInfo{
				Desc: "The version number of the table's .frm file.",
			},
			"host": &inputs.TagInfo{
				Desc: "The server host address",
			},
		},
	}
}

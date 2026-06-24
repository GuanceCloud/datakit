// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll,goconst,funlen // Measurement metadata contains intentionally long repeated literals.
package logging

import (
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type loggingMeasurement struct{}

//nolint:lll
func (*loggingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:           "default",
		Cat:            point.Logging,
		MetaDuplicated: true, // input `tailf' and `logging' are the same.
		Desc:           "Use the `source` of the config，if empty then use `default`",
		DescZh:         "本地日志采集器输出的日志记录，指标集名称来自配置中的 source，未配置时使用 default。",
		Tags: map[string]interface{}{
			"host":                inputs.NewTagInfo(`Host name`),
			"service":             inputs.NewTagInfo("The name of the service, if `service` is empty then use `source`."),
			"filepath":            inputs.NewTagInfo("The filepath to the log file on the host system where the log is stored."),
			"collector_source_ip": inputs.NewTagInfo("The source IP address for logs collected from sockets."),
		},
		Fields: map[string]interface{}{
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "The text body of the log entry."},
			"status":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "The log status string, defaulting to `info` when not set elsewhere."},
			"log_read_lines":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The lines of the read file."},
			"log_file_inode":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The inode of the log file, which uniquely identifies it on the file system (requires enabling the global configuration `enable_debug_fields`)."},
			"log_read_offset": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The current offset in the log file where reading has occurred, used to track progress during log collection (requires enabling the global configuration `enable_debug_fields`)."},
			"`__docid`": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Built-in extension fields added by server. The unique identifier for a log document, typically used for sorting and viewing details",
			},
			"__namespace": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Built-in extension fields added by server. The unique identifier for a log document dataType.",
			},
			"__truncated_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Built-in extension fields added by server. If the log is particularly large (usually exceeding 1M in size), the central system will split it and add three fields: `__truncated_id`, `__truncated_count`, and `__truncated_number` to define the splitting scenario. The __truncated_id field represents the unique identifier for the split log.",
			},
			"__truncated_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Built-in extension fields added by server. If the log is particularly large (usually exceeding 1M in size), the central system will split it and add three fields: `__truncated_id`, `__truncated_count`, and `__truncated_number` to define the splitting scenario. The __truncated_count field represents the total number of logs resulting from the split.",
			},
			"__truncated_number": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Built-in extension fields added by server. If the log is particularly large (usually exceeding 1M in size), the central system will split it and add three fields: `__truncated_id`, `__truncated_count`, and `__truncated_number` to define the splitting scenario. The __truncated_count field represents represents the current sequential identifier for the split logs.",
			},
			"date": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.TimestampMS,
				Desc:     "Built-in extension fields added by server. The `date` field is set to the time when the log is collected by the collector by default, but it can be overridden using a Pipeline.",
			},
			"date_ns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.DurationNS,
				Desc:     "Built-in extension fields added by server. The `date_ns` field is set to the millisecond part of the time when the log is collected by the collector by default. Its maximum value is 1.0E+6 and its unit is nanoseconds. It is typically used for sorting.",
			},
			"create_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.TimestampMS,
				Desc:     "Built-in extension fields added by server. The `create_time` field represents the time when the log is written to the storage engine.",
			},
			"df_metering_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Built-in extension fields added by server. The `df_metering_size` field records the metered log size in bytes for billing statistics.",
			},
		},
	}
}

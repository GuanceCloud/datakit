// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	measurementOracleLog     = "oracle_log"
	measurementLockedSession = "oracle_locked_session"
	measurementWaitingEvent  = "oracle_waiting_event"
)

type slowQueryMeasurement struct{}

//nolint:lll
func (x *slowQueryMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measurementOracleLog,
		Cat:    point.Logging,
		Desc:   `For full and detailed field into, see [here](https://docs.oracle.com/en/database/oracle/oracle-database/19/refrn/V-SQLAREA.html){:target="_blank"}`,
		DescZh: "Oracle 慢查询日志指标，字段详情可参考 Oracle V$SQLAREA 文档。",
		Fields: map[string]any{
			// core performance metrics
			`sql_fulltext`:   &inputs.FieldInfo{Unit: inputs.NoUnit, DataType: inputs.String, Type: inputs.String, Desc: "All characters of the SQL text for the current cursor"},
			`elapsed_time`:   &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Type: inputs.Gauge, Desc: "Elapsed time (in microseconds) used by this cursor for parsing, executing, and fetching. If the cursor uses parallel execution, then `ELAPSED_TIME` is the cumulative time..."},
			`cpu_time`:       &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Type: inputs.Gauge, Desc: "CPU time (in microseconds) used by this cursor for parsing, executing, and fetching"},
			`executions`:     &inputs.FieldInfo{Unit: inputs.NCount, DataType: inputs.Int, Type: inputs.Count, Desc: "Total number of executions, totalled over all the child cursors"},
			`disk_reads`:     &inputs.FieldInfo{Unit: inputs.NCount, DataType: inputs.Int, Type: inputs.Count, Desc: "Sum of the number of disk reads over all child cursors"},
			`buffer_gets`:    &inputs.FieldInfo{Unit: inputs.NCount, DataType: inputs.Int, Type: inputs.Count, Desc: "Sum of buffer gets over all child cursors"},
			`rows_processed`: &inputs.FieldInfo{Unit: inputs.NCount, DataType: inputs.Int, Type: inputs.Count, Desc: "Total number of rows processed on behalf of this SQL statement"},

			// wait metrics
			`user_io_wait_time`:     &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Type: inputs.Gauge, Desc: "User I/O Wait Time (in microseconds)"},
			`concurrency_wait_time`: &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Type: inputs.Gauge, Desc: "Concurrency wait time (in microseconds)"},
			`application_wait_time`: &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Type: inputs.Gauge, Desc: "Application wait time (in microseconds)"},
			`cluster_wait_time`:     &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Type: inputs.Gauge, Desc: "Cluster wait time (in microseconds)"},

			// execution plan metrics
			`plan_hash_value`: &inputs.FieldInfo{Unit: inputs.NoUnit, DataType: inputs.Int, Type: inputs.Gauge, Desc: "Numeric representation of the current SQL plan for this cursor. Comparing one `PLAN_HASH_VALUE` to another easily identifies whether or not two plans are the same (rather than comparing the two plans line by line)."},
			`parse_calls`:     &inputs.FieldInfo{Unit: inputs.NCount, DataType: inputs.Int, Type: inputs.Count, Desc: "Sum of all parse calls to all the child cursors under this parent"},
			`sorts`:           &inputs.FieldInfo{Unit: inputs.NCount, DataType: inputs.Int, Type: inputs.Count, Desc: "Sum of the number of sorts that were done for all the child cursors"},

			// context metrics
			`parsing_schema_name`: &inputs.FieldInfo{Unit: inputs.NoUnit, DataType: inputs.String, Type: inputs.String, Desc: "Schema name that was used to parse this child cursor"},
			`last_active_time`:    &inputs.FieldInfo{Unit: inputs.NoUnit, DataType: inputs.String, Type: inputs.String, Desc: "Time at which the query plan was last active"},

			// other fields
			`username`:    &inputs.FieldInfo{Unit: inputs.NoUnit, DataType: inputs.String, Type: inputs.String, Desc: "Name of the user"},
			`avg_elapsed`: &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Type: inputs.Gauge, Desc: "Average elapsed time of executions(`elapsed_time/executions`)"},
			`message`:     &inputs.FieldInfo{Unit: inputs.NoUnit, DataType: inputs.String, Type: inputs.String, Desc: "JSON dump of all queried fields of table `V$SQLAREA`"},
		},

		Tags: map[string]any{
			`sql_id`:            &inputs.TagInfo{Desc: "SQL identifier of the parent cursor in the library cache"},
			`module`:            &inputs.TagInfo{Desc: "Contains the name of the module that was executing when the SQL statement was first parsed as set by calling `DBMS_APPLICATION_INFO.SET_MODULE`"},
			`action`:            &inputs.TagInfo{Desc: "Contains the name of the action that was executing when the SQL statement was first parsed as set by calling `DBMS_APPLICATION_INFO.SET_ACTION`"},
			`command_type`:      &inputs.TagInfo{Desc: "Oracle command type definition"},
			`status`:            &inputs.TagInfo{Desc: "Log level, always `warning` here"},
			"oracle_server":     &inputs.TagInfo{Desc: "Server addr. Deprecated. Please use `server`"},
			"database_instance": &inputs.TagInfo{Desc: "Oracle instance identifier from configured tag or v$instance.host_name."},
			"server":            &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
		},
	}
}

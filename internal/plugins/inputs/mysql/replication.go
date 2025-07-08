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

type replicationMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *replicationMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
var replicationMeasurementInfo = &inputs.MeasurementInfo{
	Name: metricNameMySQLReplication,
	Cat:  point.Metric,
	Fields: map[string]interface{}{
		"Connect_Retry:": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The number of seconds between connect retries (default 60). This can be set with the CHANGE MASTER TO statement.",
		},
		"Slave_IO_Running": &inputs.FieldInfo{
			DataType: inputs.Bool,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "Whether the I/O thread is started and has connected successfully to the source. 1 if the state is Yes, 0 if the state is No.",
		},
		"Slave_SQL_Running": &inputs.FieldInfo{
			DataType: inputs.Bool,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "Whether the SQL thread is started. 1 if the state is Yes, 0 if the state is No.",
		},
		"Last_Errno": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "These columns are aliases for Last_SQL_Errno",
		},
		"Skip_Counter": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The current value of the sql_slave_skip_counter system variable.",
		},
		"Exec_Master_Log_Pos": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The position in the current source binary log file to which the SQL thread has read and executed, marking the start of the next transaction or event to be processed.",
		},
		"Relay_Log_Space": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The total combined size of all existing relay log files.",
		},
		"Seconds_Behind_Master": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The lag in seconds between the master and the slave.",
		},
		"Last_IO_Errno": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The error number of the most recent error that caused the I/O thread to stop. An error number of 0 and message of the empty string mean “no error.”",
		},
		"Last_SQL_Errno": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The error number of the most recent error that caused the SQL thread to stop. An error number of 0 and message of the empty string mean “no error.”",
		},
		"Master_Server_Id": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The server_id value from the source.",
		},
		"SQL_Delay": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The number of seconds that the replica must lag the source.",
		},
		"Auto_Position": &inputs.FieldInfo{
			DataType: inputs.Bool,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "1 if auto-positioning is in use; otherwise 0.",
		},
		"Replicas_connected": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "Number of replicas connected to a replication source.",
		},
		"count_transactions_in_queue": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The number of transactions in the queue pending conflict detection checks. Collected as group replication metric.",
		},
		"count_transactions_checked": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The number of transactions that have been checked for conflicts. Collected as group replication metric.",
		},
		"count_conflicts_detected": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The number of transactions that have not passed the conflict detection check. Collected as group replication metric.",
		},
		"count_transactions_rows_validating": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The number of transaction rows which can be used for certification, but have not been garbage collected. Collected as group replication metric.",
		},
		"count_transactions_remote_in_applier_queue": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The number of transactions that this member has received from the replication group which are waiting to be applied. Collected as group replication metric.",
		},
		"count_transactions_remote_applied": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The number of transactions this member has received from the group and applied. Collected as group replication metric.",
		},
		"count_transactions_local_proposed": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The number of transactions which originated on this member and were sent to the group. Collected as group replication metric.",
		},
		"count_transactions_local_rollback": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The number of transactions which originated on this member and were rolled back by the group. Collected as group replication metric.",
		},
	},
	Tags: map[string]interface{}{
		"server": &inputs.TagInfo{
			Desc: "Server addr",
		},
		"host": &inputs.TagInfo{
			Desc: "The server host address",
		},
	},
}

//nolint:lll,funlen
func (m *replicationMeasurement) Info() *inputs.MeasurementInfo {
	return replicationMeasurementInfo
}

type replicationLogMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *replicationLogMeasurement) Point() *point.Point {
	opts := point.DefaultLoggingOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

var replicationLogMeasurementInfo = &inputs.MeasurementInfo{
	Desc: "Record the replication string information.",
	Name: metricNameMySQLReplicationLog,
	Cat:  point.Logging,
	Fields: map[string]interface{}{
		"Master_Host": &inputs.FieldInfo{
			DataType: inputs.String,
			Type:     inputs.String,
			Unit:     inputs.NoUnit,
			Desc:     "The host name of the master.",
		},
		"Master_User": &inputs.FieldInfo{
			DataType: inputs.String,
			Type:     inputs.String,
			Unit:     inputs.NoUnit,
			Desc:     "The user name used to connect to the master.",
		},
		"Master_Port": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.NCount,
			Desc:     "The network port used to connect to the master.",
		},
		"Master_Log_File": &inputs.FieldInfo{
			DataType: inputs.String,
			Type:     inputs.String,
			Unit:     inputs.NoUnit,
			Desc:     "The name of the binary log file from which the server is reading.",
		},
		"Executed_Gtid_Set": &inputs.FieldInfo{
			DataType: inputs.String,
			Type:     inputs.String,
			Unit:     inputs.NoUnit,
			Desc:     "The set of global transaction IDs written in the binary log.",
		},
	},
	Tags: map[string]interface{}{
		"server": &inputs.TagInfo{
			Desc: "The address of the server. The value is `host:port`",
		},
		"host": &inputs.TagInfo{
			Desc: "The server host address",
		},
	},
}

func (m *replicationLogMeasurement) Info() *inputs.MeasurementInfo {
	return replicationLogMeasurementInfo
}

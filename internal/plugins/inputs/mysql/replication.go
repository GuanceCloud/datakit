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
	resData  map[string]interface{}
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

//nolint:lll,funlen
func (m *replicationMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mysql_replication",
		Type: "metric",
		Fields: map[string]interface{}{
			"Seconds_Behind_Master": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The lag in seconds between the master and the slave. Used before MySQL 8.0.22",
			},
			"Seconds_Behind_Source": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The lag in seconds between the source and the replica. Used after MySQL 8.0.22",
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
				Desc:     "The number of transactions in the queue pending conflict detection checks.",
			},
			"count_transactions_checked": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of transactions that have been checked for conflicts.",
			},
			"count_conflicts_detected": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of transactions that have not passed the conflict detection check.",
			},
			"count_transactions_rows_validating": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of transaction rows which can be used for certification, but have not been garbage collected.",
			},
			"count_transactions_remote_in_applier_queue": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of transactions that this member has received from the replication group which are waiting to be applied.",
			},
			"count_transactions_remote_applied": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of transactions this member has received from the group and applied.",
			},
			"count_transactions_local_proposed": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of transactions which originated on this member and were sent to the group.",
			},
			"count_transactions_local_rollback": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of transactions which originated on this member and were rolled back by the group.",
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
}

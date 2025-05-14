// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oceanbase

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	logName = "oceanbase_log"
)

type statMeasurement struct{}

func (m *statMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "oceanbase_stat",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"metric_value": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The value of the statistical item.",
			},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{
				Desc: "The server address or the host Name",
			},
			"cluster": &inputs.TagInfo{
				Desc: "Cluster Name",
			},
			"tenant_id": &inputs.TagInfo{
				Desc: "Tenant id",
			},
			"tenant_name": &inputs.TagInfo{
				Desc: "Tenant Name",
			},
			"metric_name": &inputs.TagInfo{
				Desc: "The name of the statistical event.",
			},
			"svr_ip": &inputs.TagInfo{
				Desc: "The IP address of the server where the information is located.",
			},
			"stat_id": &inputs.TagInfo{
				Desc: "The ID of the statistical event.",
			},
		},
	}
}

type eventMeasurement struct{}

func (m *eventMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "oceanbase_event",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"total_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The total number of waits for the event.",
			},
			"time_waited": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.DurationSecond,
				Unit:     inputs.DurationSecond,
				Desc:     "The total wait time for the event in seconds.",
			},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{
				Desc: "The server address or the host Name",
			},
			"cluster": &inputs.TagInfo{
				Desc: "Cluster Name",
			},
			"tenant_id": &inputs.TagInfo{
				Desc: "Tenant id",
			},
			"tenant_name": &inputs.TagInfo{
				Desc: "Tenant Name",
			},
			"svr_ip": &inputs.TagInfo{
				Desc: "The IP address of the server where the information is located.",
			},
			"event_group": &inputs.TagInfo{
				Desc: "The group of the event.",
			},
		},
	}
}

type sessionMeasurement struct{}

func (m *sessionMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "oceanbase_session",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"active_cnt": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The number of active sessions within a tenant.",
			},
			"all_cnt": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The total number of sessions within a tenant.",
			},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{
				Desc: "The server address or the host Name",
			},
			"cluster": &inputs.TagInfo{
				Desc: "Cluster Name",
			},
			"tenant_id": &inputs.TagInfo{
				Desc: "Tenant id",
			},
			"tenant_name": &inputs.TagInfo{
				Desc: "Tenant Name",
			},
			"svr_ip": &inputs.TagInfo{
				Desc: "The IP address of the server where the information is located.",
			},
			"svr_port": &inputs.TagInfo{
				Desc: "The port of the server where the information is located.",
			},
		},
	}
}

type clogMeasurement struct{}

func (m *clogMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "oceanbase_clog",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"max_clog_sync_delay_seconds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.DurationSecond,
				Unit:     inputs.DurationSecond,
				Desc:     "The max clog synchronization delay of an tenant.",
			},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{
				Desc: "The server address or the host Name",
			},
			"cluster": &inputs.TagInfo{
				Desc: "Cluster Name",
			},
			"tenant_id": &inputs.TagInfo{
				Desc: "Tenant id",
			},
			"tenant_name": &inputs.TagInfo{
				Desc: "Tenant Name",
			},
			"svr_ip": &inputs.TagInfo{
				Desc: "The IP address of the server where the information is located.",
			},
			"svr_port": &inputs.TagInfo{
				Desc: "The port of the server where the information is located.",
			},
			"replica_type": &inputs.TagInfo{
				Desc: "The type of the replica",
			},
		},
	}
}

type cacheBlockMeasurement struct{}

func (m *cacheBlockMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "oceanbase_cache_block",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"cache_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.SizeMB,
				Unit:     inputs.SizeMB,
				Desc:     "The block cache size in the specified statistical range.",
			},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{
				Desc: "The server address or the host Name",
			},
			"cluster": &inputs.TagInfo{
				Desc: "Cluster Name",
			},
			"tenant_id": &inputs.TagInfo{
				Desc: "Tenant id",
			},
			"tenant_name": &inputs.TagInfo{
				Desc: "Tenant Name",
			},
			"svr_ip": &inputs.TagInfo{
				Desc: "The IP address of the server where the information is located.",
			},
			"svr_port": &inputs.TagInfo{
				Desc: "The port of the server where the information is located.",
			},
			"cache_name": &inputs.TagInfo{
				Desc: "The cache name.",
			},
		},
	}
}

type cachePlanMeasurement struct{}

func (m *cachePlanMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "oceanbase_cache_plan",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"access_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The number of times that the query accesses the plan cache.",
			},
			"hit_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The number of plan cache hits.",
			},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{
				Desc: "The server address or the host Name",
			},
			"cluster": &inputs.TagInfo{
				Desc: "Cluster Name",
			},
			"tenant_id": &inputs.TagInfo{
				Desc: "Tenant id",
			},
			"tenant_name": &inputs.TagInfo{
				Desc: "Tenant Name",
			},
			"svr_ip": &inputs.TagInfo{
				Desc: "The IP address of the server where the information is located.",
			},
			"svr_port": &inputs.TagInfo{
				Desc: "The port of the server where the information is located.",
			},
		},
	}
}

type loggingMeasurement struct{}

// https://www.oceanbase.com/docs/enterprise-oceanbase-database-cn-10000000000376664
//
//nolint:lll
func (m *loggingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: logName,
		Cat:  point.Logging,
		Desc: "",
		Tags: map[string]interface{}{
			"host":                 &inputs.TagInfo{Desc: "Hostname."},
			inputName + "_server":  &inputs.TagInfo{Desc: "The address of the database instance (including port)."},
			inputName + "_service": &inputs.TagInfo{Desc: "OceanBase service name."},
			"tenant_id": &inputs.TagInfo{
				Desc: "Tenant id",
			},
			"tenant_name": &inputs.TagInfo{
				Desc: "Tenant Name",
			},
			"cluster": &inputs.TagInfo{
				Desc: "Cluster Name",
			},
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "The text of the logging."},
			"status":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "The status of the logging, only supported `info/emerg/alert/critical/error/warning/debug/OK/unknown`."},
		},
	}
}

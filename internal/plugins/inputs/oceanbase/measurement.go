// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll,goconst,funlen // Measurement metadata contains intentionally long repeated literals.
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
		Name:   "oceanbase_stat",
		Cat:    point.Metric,
		Desc:   "OceanBase system statistics from `gv$sysstat`, reported by cluster, tenant, server IP, stat ID, and metric name.",
		DescZh: "来自 `gv$sysstat` 的 OceanBase 系统统计项，按集群、租户、服务器 IP、统计项 ID 和指标名上报。",
		Fields: map[string]interface{}{
			"metric_value": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The value of the statistical item.",
				Taggedby: []string{"cluster", "tenant_id", "tenant_name", "metric_name", "svr_ip", "stat_id"},
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
		Name:   "oceanbase_event",
		Cat:    point.Metric,
		Desc:   "OceanBase non-idle system wait event totals from `gv$system_event`, grouped by cluster, tenant, server IP, and event group.",
		DescZh: "来自 `gv$system_event` 的 OceanBase 非空闲系统等待事件累计值，按集群、租户、服务器 IP 和事件分组上报。",
		Fields: map[string]interface{}{
			"total_waits": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The total number of waits for the event.",
				Taggedby: []string{"cluster", "tenant_id", "tenant_name", "svr_ip", "event_group"},
			},
			"time_waited": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.DurationSecond,
				Desc:     "The total wait time for the event in seconds.",
				Taggedby: []string{"cluster", "tenant_id", "tenant_name", "svr_ip", "event_group"},
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
		Name:   "oceanbase_session",
		Cat:    point.Metric,
		Desc:   "OceanBase current session counts from `__all_virtual_processlist`, grouped by cluster, tenant, server IP, and server port.",
		DescZh: "来自 `__all_virtual_processlist` 的 OceanBase 当前会话数量，按集群、租户、服务器 IP 和端口上报。",
		Fields: map[string]interface{}{
			"active_cnt": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of active sessions within a tenant.",
				Taggedby: []string{"cluster", "tenant_id", "tenant_name", "svr_ip", "svr_port"},
			},
			"all_cnt": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total number of sessions within a tenant.",
				Taggedby: []string{"cluster", "tenant_id", "tenant_name", "svr_ip", "svr_port"},
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
		Name:   "oceanbase_clog",
		Cat:    point.Metric,
		Desc:   "OceanBase clog synchronization delay from virtual clog statistics, grouped by cluster, tenant, server, and replica type.",
		DescZh: "来自虚拟 clog 统计表的 OceanBase clog 同步延迟，按集群、租户、服务器和副本类型上报。",
		Fields: map[string]interface{}{
			"max_clog_sync_delay_seconds": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "The max clog synchronization delay of an tenant.",
				Taggedby: []string{"cluster", "tenant_id", "tenant_name", "svr_ip", "svr_port", "replica_type"},
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
		Name:   "oceanbase_cache_block",
		Cat:    point.Metric,
		Desc:   "OceanBase block cache size from `__all_virtual_kvcache_info`, grouped by cluster, tenant, server, and cache name.",
		DescZh: "来自 `__all_virtual_kvcache_info` 的 OceanBase block cache 大小，按集群、租户、服务器和 cache 名称上报。",
		Fields: map[string]interface{}{
			"cache_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "The block cache size in bytes for the specified cache.",
				Taggedby: []string{"cluster", "tenant_id", "tenant_name", "svr_ip", "svr_port", "cache_name"},
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
		Name:   "oceanbase_cache_plan",
		Cat:    point.Metric,
		Desc:   "OceanBase plan cache access and hit counters from `gv$plan_cache_stat`, grouped by cluster, tenant, server IP, and server port.",
		DescZh: "来自 `gv$plan_cache_stat` 的 OceanBase plan cache 访问和命中计数，按集群、租户、服务器 IP 和端口上报。",
		Fields: map[string]interface{}{
			"access_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The number of times that the query accesses the plan cache.",
				Taggedby: []string{"cluster", "tenant_id", "tenant_name", "svr_ip", "svr_port"},
			},
			"hit_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The number of plan cache hits.",
				Taggedby: []string{"cluster", "tenant_id", "tenant_name", "svr_ip", "svr_port"},
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
		Name:   logName,
		Cat:    point.Logging,
		Desc:   "OceanBase slow-query log records collected from `GV$SQL_AUDIT` for the configured database service.",
		DescZh: "从配置的 OceanBase 数据库服务 `GV$SQL_AUDIT` 采集的慢查询日志。",
		Tags: map[string]interface{}{
			"host":                 &inputs.TagInfo{Desc: "Hostname."},
			inputName + "_server":  &inputs.TagInfo{Desc: "The address of the database instance (including port)."},
			inputName + "_service": &inputs.TagInfo{Desc: "OceanBase service name."},
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

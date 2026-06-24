// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package tdengine is input for TDEngine database
package tdengine

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	ts       int64
	election bool
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return tdengineMeasurementInfo()
}

func tdengineMeasurementInfo() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Cat:    point.Metric,
		Desc:   "TDengine cluster, database, data node, request, and taosAdapter metrics collected from system log tables and show commands.",
		DescZh: "从 TDengine 系统日志表和 show 命令采集的集群、数据库、数据节点、请求和 taosAdapter 指标。",
		Fields: tdengineMeasurementFields(),
		Tags:   tdengineMeasurementTags(),
	}
}

// nolint:funlen
func tdengineMeasurementFields() map[string]interface{} {
	return map[string]interface{}{
		"master_uptime": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.DurationSecond,
			Desc:     "Seconds of master's uptime",
			Taggedby: []string{"first_ep", "version"},
		},

		"expire_time": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Int,
			Unit:     inputs.DurationSecond,
			Desc:     "Time until grants expire in seconds",
		},

		"timeseries_used": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Time series used",
		},

		"timeseries_total": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total time series",
		},

		"database_count": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total number of databases",
		},

		"table_count": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total number of tables in the database",
		},

		"tables_count": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Number of tables per vgroup",
			Taggedby: []string{"database_name"},
		},

		"dnodes_total": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total number of dnodes(data nodes) in cluster",
		},
		"dnodes_alive": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total number of dnodes in ready state",
		},
		"mnodes_total": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total number of mnodes(management nodes) in cluster",
		},
		"mnodes_alive": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total number of mnodes in ready state",
		},
		"vgroups_total": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total number of vgroups in cluster",
		},
		"vgroups_alive": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total number of vgroups in ready state",
		},
		"vnodes_total": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total number of vnode in cluster",
		},
		"id": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NoUnit,
			Desc:     "TDengine node ID",
			Taggedby: []string{"end_point", "status", "offline_reason", "role", "role_time"},
		},
		"vnodes_alive": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total number of vnode in ready state",
		},
		"vnodes": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "The number of virtual node groups contained in a single data node",
			Taggedby: []string{"end_point", "status", "offline_reason"},
		},

		"req_insert_rate": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Number of insert queries received per dnode divided by monitor interval",
			Taggedby: []string{"dnode_ep"},
		},

		"req_insert_batch_rate": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Number of batch insertions divided by monitor interval",
			Taggedby: []string{"dnode_ep"},
		},

		"req_select": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Number of select queries received per dnode",
			Taggedby: []string{"dnode_ep"},
		},
		"req_select_rate": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Number of select queries received per dnode divided by monitor interval",
			Taggedby: []string{"dnode_ep"},
		},

		"req_http": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total number of requests via HTTP",
		},

		"req_http_rate": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "HTTP request rate",
		},
		"cpu_cores": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total number of CPU cores per data node",
			Taggedby: []string{"dnode_ep"},
		},

		"vnodes_num": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total number of virtual nodes per data node",
			Taggedby: []string{"dnode_ep"},
		},

		"uptime": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.DurationSecond,
			Desc:     "Data node uptime in seconds",
			Taggedby: []string{"dnode_ep"},
		},

		"cpu_engine": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.Percent,
			Desc:     "CPU usage per data node",
			Taggedby: []string{"dnode_ep"},
		},

		"disk_used": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.SizeGB,
			Desc:     "Disk usage of data nodes",
			Taggedby: []string{"dnode_ep"},
		},

		"disk_total": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.SizeGB,
			Desc:     "Total disk size of data nodes",
			Taggedby: []string{"dnode_ep"},
		},

		"disk_percent": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.Percent,
			Desc:     "Data node disk usage percentage",
			Taggedby: []string{"dnode_ep"},
		},

		"cpu_system": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "CPU system usage of data nodes",
			Taggedby: []string{"dnode_ep"},
		},

		"mem_engine": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.SizeMB,
			Desc:     "Memory usage of tdengine",
			Taggedby: []string{"dnode_ep"},
		},

		"mem_system": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.SizeMB,
			Desc:     "Available memory on the server",
			Taggedby: []string{"dnode_ep"},
		},

		"mem_total": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.SizeGB,
			Desc:     "Total memory of server",
			Taggedby: []string{"dnode_ep"},
		},

		"mem_engine_percent": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.Percent,
			Desc:     "`taosd` memory usage percentage",
			Taggedby: []string{"dnode_ep"},
		},

		"io_read_taosd": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.SizeMB,
			Desc:     "Average data size of IO reads per second",
			Taggedby: []string{"dnode_ep"},
		},

		"io_write_taosd": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.SizeMB,
			Desc:     "Average data size of IO writes per second",
			Taggedby: []string{"dnode_ep"},
		},

		"net_in": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.SizeKB,
			Desc:     "IO rate of the ingress network",
			Taggedby: []string{"dnode_ep"},
		},

		"net_out": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.SizeKB,
			Desc:     "IO rate of egress network",
			Taggedby: []string{"dnode_ep"},
		},

		"total_req_count": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Total adapter requests",
			Taggedby: []string{"endpoint", "status_code"},
		},

		"status_code": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Status code returned by the request",
		},

		"client_ip_count": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Client IP request statistics",
			Taggedby: []string{"client_ip"},
		},

		"request_in_flight": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.NCount,
			Desc:     "Number of requests being sorted",
		},

		"cpu_percent": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.Percent,
			Desc:     "Adapter occupies CPU usage",
			Taggedby: []string{"endpoint"},
		},

		"mem_percent": &inputs.FieldInfo{
			Type:     inputs.Gauge,
			DataType: inputs.Float,
			Unit:     inputs.Percent,
			Desc:     "Adapter memory usage",
			Taggedby: []string{"endpoint"},
		},
	}
}

func tdengineMeasurementTags() map[string]interface{} {
	return map[string]interface{}{
		"host":          &inputs.TagInfo{Desc: "Host name"},
		"cluster_name":  &inputs.TagInfo{Desc: "Cluster name"},
		"endpoint":      &inputs.TagInfo{Desc: "taosAdapter endpoint"},
		"end_point":     &inputs.TagInfo{Desc: "Remote address name, the general naming rule is (host:port)"},
		"dnode_ep":      &inputs.TagInfo{Desc: "Data node name, generally equivalent to `end_point`"},
		"database_name": &inputs.TagInfo{Desc: "Database name"},
		"vgroup_id":     &inputs.TagInfo{Desc: "VGroup ID"},
		"client_ip":     &inputs.TagInfo{Desc: "Client IP"},
		"status_code":   &inputs.TagInfo{Desc: "HTTP response status code"},
		"status":        &inputs.TagInfo{Desc: "Data node status"},
		"offline_reason": &inputs.TagInfo{
			Desc: "Reason the data node is offline",
		},
		"role":      &inputs.TagInfo{Desc: "Management node role"},
		"role_time": &inputs.TagInfo{Desc: "Management node role start time"},
		"version":   &inputs.TagInfo{Desc: "Version"},
		"first_ep":  &inputs.TagInfo{Desc: "First endpoint"},
	}
}

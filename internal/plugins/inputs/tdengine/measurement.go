// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package tdengine is input for TDEngine database
package tdengine

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	ts       time.Time
	election bool
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Type: "metric",
		Fields: map[string]interface{}{
			"master_uptime": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationSecond,
				Desc: "Seconds of master's uptime",
			},

			"expire_time": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.DurationSecond,
				Desc: "Time until grants expire in seconds",
			},

			"timeseries_used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Time series used",
			},

			"timeseries_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total time series",
			},

			"database_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of databases",
			},

			"table_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of tables in the database",
			},

			"tables_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Number of tables per vgroup",
			},

			"dnodes_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of dnodes(data nodes) in cluster",
			},
			"dnodes_alive": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of dnodes in ready state",
			},
			"mnodes_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of mnodes(management nodes) in cluster",
			},
			"mnodes_alive": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of mnodes in ready state",
			},
			"vgroups_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of vgroups in cluster",
			},
			"vgroups_alive": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of vgroups in ready state",
			},
			"vnodes_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of vnode in cluster",
			},
			"vnodes_alive": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of vnode in ready state",
			},
			"vnodes": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of virtual node groups contained in a single data node",
			},

			"req_insert_rate": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Number of insert queries received per dnode divided by monitor interval",
			},

			"req_insert_batch_rate": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Number of batch insertions divided by monitor interval",
			},

			"req_select": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Number of select queries received per dnode",
			},
			"req_select_rate": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Number of select queries received per dnode divided by monitor interval",
			},

			"req_http": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of requests via HTTP",
			},

			"req_http_rate": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "HTTP request rate",
			},
			"cpu_cores": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of CPU cores per data node",
			},

			"vnodes_num": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of virtual nodes per data node",
			},

			"cpu_engine": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "CPU usage per data node",
			},

			"disk_used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeGB,
				Desc: "Disk usage of data nodes",
			},

			"disk_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeGB,
				Desc: "Total disk size of data nodes",
			},

			"disk_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Data node disk usage percentage",
			},

			"cpu_system": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "CPU system usage of data nodes",
			},

			"mem_engine": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeMB,
				Desc: "Memory usage of tdengine",
			},

			"mem_system": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeMB,
				Desc: "Available memory on the server",
			},

			"mem_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeGB,
				Desc: "Total memory of server",
			},

			"mem_engine_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "`taosd` memory usage percentage",
			},

			"io_read_taosd": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeMB,
				Desc: "Average data size of IO reads per second",
			},

			"io_write_taosd": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeMB,
				Desc: "Average data size of IO writes per second",
			},

			"net_in": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeKB,
				Desc: "IO rate of the ingress network",
			},

			"net_out": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeKB,
				Desc: "IO rate of egress network",
			},

			"total_req_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total adapter requests",
			},

			"status_code": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Status code returned by the request",
			},

			"client_ip_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Client IP request statistics",
			},

			"request_in_flight": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Number of requests being sorted",
			},

			"cpu_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Adapter occupies CPU usage",
			},

			"mem_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Adapter memory usage",
			},
		},
		Tags: map[string]interface{}{
			"host":          &inputs.TagInfo{Desc: "Host name"},
			"cluster_name":  &inputs.TagInfo{Desc: "Cluster name"},
			"end_point":     &inputs.TagInfo{Desc: "Remote address name, the general naming rule is (host:port)"},
			"dnode_ep":      &inputs.TagInfo{Desc: "Data node name, generally equivalent to `end_point`"},
			"database_name": &inputs.TagInfo{Desc: "Database name"},
			"vgroup_id":     &inputs.TagInfo{Desc: "VGroup ID"},
			"client_ip":     &inputs.TagInfo{Desc: "Client IP"},
			"version":       &inputs.TagInfo{Desc: "Version"},
			"first_ep":      &inputs.TagInfo{Desc: "First endpoint"},
		},
	}
}

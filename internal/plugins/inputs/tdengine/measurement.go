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
		Fields: map[string]interface{}{
			"master_uptime": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationSecond,
				Desc: "从 dnode 当选为 master 的时间",
			},

			"expire_time": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.DurationSecond,
				Desc: "企业版到期时间",
			},

			"timeseries_used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "企业版已使用测点数",
			},

			"timeseries_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "企业版总测点数",
			},

			"database_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "数据库总个数",
			},

			"table_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "数据库中的表总数",
			},

			"tables_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "数据库中每个 database 中表数量的指标",
			},

			"dnodes_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "集群中数据节点(dnode) 的总个数",
			},
			"dnodes_alive": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "集群中数据节点存活个数",
			},
			"mnodes_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "数据库管理节点(`mnode`)个数",
			},
			"mnodes_alive": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "数据库管理节点存活个数",
			},
			"vgroups_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "数据库中虚拟节点组总数",
			},
			"vgroups_alive": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "数据库中虚拟节点组总存活数",
			},
			"vnodes_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "数据库中虚拟节点总数",
			},
			"vnodes_alive": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "数据库中虚拟节点总存活数",
			},
			"vnodes": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "单个数据节点中包括虚拟节点组的数量",
			},

			"req_insert_rate": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "请求插入数据的速率",
			},

			"req_insert_batch_rate": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "请求插入数据批次速率",
			},

			"req_select": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "查询数量",
			},
			"req_select_rate": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "查询速率",
			},

			"req_http": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "通过 http 请求的总数",
			},

			"req_http_rate": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "http 请求速率",
			},
			"cpu_cores": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "每个数据节点的 CPU 总核数",
			},

			"vnodes_num": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "每个数据节点的虚拟节点总数",
			},

			"cpu_engine": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "每个数据节点的 CPU 使用率",
			},

			"disk_used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeGB,
				Desc: "数据节点的磁盘使用量",
			},

			"disk_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeGB,
				Desc: "数据节点磁盘总量",
			},

			"disk_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "数据节点磁盘使用率",
			},

			"cpu_system": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "数据节点的 CPU 系统使用率",
			},

			"mem_engine": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeMB,
				Desc: "TDEngine 占用内存量",
			},

			"mem_system": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeMB,
				Desc: "数据节点系统占用总内存量",
			},

			"mem_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeGB,
				Desc: "数据节点总内存量",
			},

			"mem_engine_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "`taosd` 占用内存率",
			},

			"io_read_taosd": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeMB,
				Desc: "平均每秒 IO read 的数据大小",
			},

			"io_write_taosd": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeMB,
				Desc: "平均每秒 IO write 的数据大小",
			},

			"net_in": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeKB,
				Desc: "入口网络的 IO 速率",
			},

			"net_out": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeKB,
				Desc: "出口网络的 IO 速率",
			},

			"total_req_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "adapter 总请求量",
			},

			"status_code": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "请求返回的状态码",
			},

			"client_ip_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "客户端 IP 请求次数统计",
			},

			"request_in_flight": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "正在梳理的请求数量",
			},

			"cpu_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "adapter 占用 CPU 使用率",
			},

			"mem_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "adapter 占用 MEM 使用率",
			},
		},
		Tags: map[string]interface{}{
			"host":          &inputs.TagInfo{Desc: "主机名"},
			"cluster_name":  &inputs.TagInfo{Desc: "集群名称"},
			"end_point":     &inputs.TagInfo{Desc: "远端地址名称，一般命名规则是(host:port)"},
			"dnode_ep":      &inputs.TagInfo{Desc: "数据节点名称，一般情况下与 end_point 等价"},
			"database_name": &inputs.TagInfo{Desc: "数据库名称"},
			"vgroup_id":     &inputs.TagInfo{Desc: "虚拟组 ID"},
			"client_ip":     &inputs.TagInfo{Desc: "请求端 IP"},
			"version":       &inputs.TagInfo{Desc: "version"},
			"first_ep":      &inputs.TagInfo{Desc: "first endpoint"},
		},
	}
}

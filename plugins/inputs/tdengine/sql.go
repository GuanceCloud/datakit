// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package tdengine is input for tdengine SQL

package tdengine

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type selectSQL struct {
	desc   string // 说明
	title  string // measurement
	sql    string
	unit   string   // 单位
	fields []string // field 类型
	tags   []string // tag 类型

	plugInFun tdPlugIn
}

// metric 指标集结构.
type metric struct {
	metricName string        // 指标集名称
	doc        string        // 说明
	MetricList []selectSQL   // 指标结构（单指标中包含多个子指标数据）
	TimeSeries time.Duration // 查询间隔时间
}

//nolint:lll
var metrics = []metric{
	{
		metricName: "td_cluster",
		doc:        "集群状态",
		TimeSeries: time.Minute * 5,
		MetricList: []selectSQL{
			{
				desc:   "First EP Name",
				title:  "",
				sql:    "select last(first_ep),last(version),last(master_uptime) from log.cluster_info",
				unit:   "",
				fields: []string{"master_uptime"},
				tags:   []string{"first_ep", "version"},
			},
			{
				desc:   "企业版授权到期时间", // 非企业版没有该指标
				title:  "",
				sql:    "select last(expire_time) from log.grants_info",
				unit:   "s",
				fields: []string{"expire_time"},
				tags:   []string{},
			},
			{
				desc:   "企业版已使用的测点数", // 非企业版没有该指标
				title:  "",
				sql:    "select max(timeseries_used) as used ,max(timeseries_total) as total from log.grants_info where ts >= now-5m and ts <= now ",
				unit:   "s",
				fields: []string{"timeseries_used", "timeseries_total"},
				tags:   []string{},
			},
			{
				desc:      "数据库个数",
				title:     "",
				sql:       "show databases",
				unit:      "s",
				fields:    []string{},
				tags:      []string{},
				plugInFun: &databaseCount{},
			},
			{
				desc:      "所有数据库的表数量之和",
				title:     "",
				sql:       "show databases",
				unit:      "s",
				fields:    []string{},
				tags:      []string{},
				plugInFun: &tablesCount{},
			},
			{
				desc:   "dnode,mnode,vnode状态信息",
				title:  "",
				sql:    "select * from log.cluster_info where ts >= now-5m and ts <= now",
				unit:   "",
				fields: []string{"dnodes_total", "dnodes_alive", "mnodes_total", "mnodes_alive", "vgroups_total", "vgroups_alive", "vnodes_total", "vnodes_alive", "connections_total"},
				tags:   []string{},
			},
		},
	},
	{
		metricName: "td_node",
		doc:        "dnode状态",
		TimeSeries: time.Minute * 5,
		MetricList: []selectSQL{
			{
				desc:   "dnode状态",
				title:  "",
				sql:    "show dnodes",
				unit:   "",
				fields: []string{"id", "vnodes"},
				tags:   []string{"end_point", "status", "offline_reason"},
			},
		},
	},
	{
		metricName: "td_node",
		doc:        "master状态",
		TimeSeries: time.Minute * 5,
		MetricList: []selectSQL{
			{
				desc:   "通过选举成为master节点，保证高可用性",
				title:  "",
				sql:    "show mnodes",
				unit:   "",
				fields: []string{"id"},
				tags:   []string{"end_point", "role", "role_time"},
			},
		},
	},
	{
		metricName: "td_request",
		doc:        "请求",
		TimeSeries: time.Minute * 1,
		MetricList: []selectSQL{
			{
				desc:   "数据插入次数",
				title:  "",
				sql:    "select ts,req_insert_rate,req_insert_batch_rate,dnode_ep from log.dnodes_info where ts >= now-1m and ts <= now",
				unit:   "",
				fields: []string{"req_insert_rate", "req_insert_batch_rate"},
				tags:   []string{"dnode_ep"},
			},
			{
				desc:   "查询次数",
				title:  "",
				sql:    "select ts,req_select,req_select_rate,dnode_ep from log.dnodes_info where ts >= now-1m and ts <= now",
				unit:   "",
				fields: []string{"req_select", "req_select_rate"},
				tags:   []string{"dnode_ep"},
			},
		},
	},
	{
		metricName: "td_database",
		doc:        "数据库指标集",
		TimeSeries: time.Second * 30,
		MetricList: []selectSQL{
			{
				desc:   "VGroups 变化图",
				title:  "",
				sql:    "select last(ts),last(tables_num) as tables_count,database_name from log.vgroups_info where ts > now-30s and tables_num>1 and status='ready' group by database_name",
				unit:   "",
				fields: []string{"tables_count"},
				tags:   []string{"database_name"},
			},
		},
	},
	{
		metricName: "td_node_usage",
		doc:        "资源使用情况",
		TimeSeries: time.Minute * 1,
		MetricList: []selectSQL{
			{
				desc:   "dnode 启动时间",
				title:  "",
				sql:    "select last(ts),last(uptime)  from log.dnodes_info where errors=0 group by dnode_ep",
				unit:   "",
				fields: []string{"uptime"},
				tags:   []string{"dnode_ep"},
			},
			{
				desc:   "dnode 的 VNodes 数量",
				title:  "",
				sql:    "select last(ts),last(cpu_cores),last(vnodes_num) from log.dnodes_info where ts >= now-1m and ts <= now group by dnode_ep",
				unit:   "",
				fields: []string{"cpu_cores", "vnodes_num"},
				tags:   []string{"dnode_ep"},
			},
			{
				desc:   "磁盘使用率",
				title:  "",
				sql:    "select last(ts),last(disk_used),last(disk_total), last(disk_used)*100 / last(disk_total) as disk_percent from log.dnodes_info where ts >= now-1m and ts <= now group by dnode_ep",
				unit:   inputs.Percent,
				fields: []string{"disk_used", "disk_total", "disk_percent"},
				tags:   []string{"dnode_ep"},
			},
			{
				desc:   "CPU使用率",
				title:  "",
				sql:    "select last(ts),avg(cpu_engine) as cpu_engine, avg(cpu_system) as cpu_system from log.dnodes_info where ts >= now-1m and ts <= now group by dnode_ep",
				unit:   "",
				fields: []string{"cpu_engine", "cpu_system"},
				tags:   []string{"dnode_ep"},
			},
			{
				desc:   "RAM使用视图",
				title:  "",
				sql:    "select last(ts),last(mem_engine),last(mem_system),last(mem_total),last(mem_engine)*100/last(mem_total) as mem_engine_percent from log.dnodes_info where ts >= now-1m and ts <= now group by dnode_ep",
				unit:   "",
				fields: []string{"mem_engine", "mem_system", "mem_total", "mem_engine_percent"},
				tags:   []string{"dnode_ep"},
			},
			{
				desc:   "io使用情况-磁盘读写",
				title:  "",
				sql:    "select last(ts),avg(io_read_disk) as io_read_taosd, avg(io_write_disk) as io_write_taosd from log.dnodes_info where ts >= now-1m and ts <= now group by dnode_ep",
				unit:   "MBs",
				fields: []string{"io_read_taosd", "io_write_taosd"},
				tags:   []string{"dnode_ep"},
			},
			{
				desc:   "网络 IO，除本机网络之外的总合网络 IO 速率",
				title:  "",
				sql:    "select last(ts),avg(net_in) as net_in,avg(net_out) as net_out from log.dnodes_info where ts >= now-1m and ts <= now group by dnode_ep",
				unit:   "Mbits",
				fields: []string{"net_in", "net_out"},
				tags:   []string{"dnode_ep"},
			},
		},
	},
	{
		metricName: "td_adapter",
		doc:        "taosAdapter 监控",
		TimeSeries: time.Minute * 1,
		MetricList: []selectSQL{
			{
				desc:   "总请求数",
				title:  "",
				sql:    "select last(ts),sum(count) as total_req_count from log.taosadapter_restful_http_total where ts >= now-1m and ts <= now  group by endpoint,status_code",
				unit:   "",
				fields: []string{"total_req_count"},
				tags:   []string{"endpoint", "status_code"},
			},
			{
				desc:   "client ip 统计",
				title:  "",
				sql:    "select last(ts),count(ts) as client_ip_count from log.taosadapter_restful_http_total where ts >= now-1m and ts <= now  group by client_ip",
				unit:   "",
				fields: []string{"client_ip_count"},
				tags:   []string{"client_ip"},
			},
			{
				desc:   "CPU和内存使用情况",
				title:  "",
				sql:    "select ts,cpu_percent,mem_percent,endpoint from log.taosadapter_system where ts >= now-1m and ts <= now",
				unit:   "",
				fields: []string{"cpu_percent", "mem_percent"},
				tags:   []string{"endpoint"},
			},
		},
	},
}

var checkHealthSQL = selectSQL{
	desc: "检查数据库连接并使用用户名密码登陆",
	sql:  "show databases",
}

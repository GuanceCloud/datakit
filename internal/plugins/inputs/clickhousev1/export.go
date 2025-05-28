// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package clickhousev1

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			//nolint:lll
			"var_instance":              "实例",
			"group_overview":            "概览",
			"version":                   "版本",
			"version_desc":              "以 1000 为基数的单个整数表示。例如，版本 11.22.33 转换为 11022033",
			"http_connections":          "HTTP 连接数",
			"http_connections_desc":     "与 HTTP 服务器的连接数",
			"current_value_title":       "当前值",
			"max_value_title":           "最大值",
			"avg_value_title":           "平均值",
			"interserver_title":         "Interserver 连接数",
			"interserver_desc":          "表示其他副本用于获取部分数据的连接数量",
			"insertion_title":           "延迟 insert 数",
			"insert_bytes":              "Insert 字节",
			"insert_lines":              "Insert 行",
			"query_count":               "Query 数",
			"memory_tracking":           "内存使用量",
			"select_query_count":        "Select Query 数",
			"insert_query_count":        "Insert Query 数",
			"merge_count":               "Merge 数",
			"readonly_replica":          "只读状态的 Replicated 表的数量",
			"replicated_checks":         "数据块一致性检查的次数",
			"replicated_fetch":          "从副本中获取的数据块（part）数量",
			"replicated_sends":          "发送到副本的数据块（part）数量",
			"background_pool_schedule":  "后台线程池任务数",
			"dict_cache_keys_requested": "缓存类型字典的数据源中的请求数",
			"read_write":                "系统调用量",
			"ZooKeeperRequest":          "向 ZooKeeper 发送的请求数量",
			"ZooKeeperBytesReceived":    "从 ZooKeeper 接收到的字节数",
			"ZooKeeperBytesSent":        "向 ZooKeeper 发送的字节数",
			"tcp_connections":           "TCP 连接数",
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
			"var_instance":              "Instance",
			"group_overview":            "Overview",
			"version":                   "Version",
			"version_desc":              "A single integer representation with 1000 as the base. For example, version 11.22.33 is converted to 11022033",
			"http_connections":          "HTTP Connections",
			"http_connections_desc":     "Number of connections to the HTTP server",
			"current_value_title":       "Current Value",
			"max_value_title":           "Max Value",
			"avg_value_title":           "Average Value",
			"interserver_title":         "Interserver Connections",
			"interserver_desc":          "Number of connections used by other replicas to fetch data parts",
			"insertion_title":           "Delayed Inserts Count",
			"insert_bytes":              "Insert Bytes",
			"insert_lines":              "Insert Lines",
			"query_count":               "Query Count",
			"memory_tracking":           "Memory Usage",
			"select_query_count":        "Select Query Count",
			"insert_query_count":        "Insert Query Count",
			"merge_count":               "Merge Count",
			"readonly_replica":          "Number of Replicated Tables in Read-Only State",
			"replicated_checks":         "Data Part Consistency Check Count",
			"replicated_fetch":          "Number of Data Parts (Part) Fetched from Replicas",
			"replicated_sends":          "Number of Data Parts (Part) Sent to Replicas",
			"background_pool_schedule":  "Background Thread Pool Task Count",
			"dict_cache_keys_requested": "Number of Requests from Dictionary Cache Data Sources",
			"read_write":                "System Call Count",
			"ZooKeeperRequest":          "Number of Requests Sent to ZooKeeper",
			"ZooKeeperBytesReceived":    "Bytes Received from ZooKeeper",
			"ZooKeeperBytesSent":        "Bytes Sent to ZooKeeper",
			"tcp_connections":           "TCP Connections",
		}
	default:
		return nil
	}
}

func (ipt *Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"clickhouse_http_connections_title":   `ClickHouse实例{{instance}} HTTP 连接数告警`,
			"clickhouse_http_connections_message": `>主机：{{host}}\n>ClickHouse实例：{{instance}}\n>告警级别：{{df_status}}\n>HTTP连接数为：{{df_monitor_checker_value}}，超过设置的阈值，请关注！`,
			"clickhoust_tcp_title":                `ClickHouse实例{{instance}} TCP 连接数告警`,
			"clickhoust_tcp_message":              `>主机：{{host}}\n>ClickHouse实例：{{instance}}\n>告警级别：{{df_status}}\n>TCP连接数为：{{df_monitor_checker_value}}，超过设置的阈值，请关注！`,
			"replicated_checks_title":             `ClickHouse实例{{instance}} 副本一致性检查次数异常告警`,
			"replicated_checks_message":           `>主机：{{host}}\n>ClickHouse实例：{{instance}}\n>告警级别：{{df_status}}\n>副本之间数据块一致性检查的次数最近15分钟比最近30分钟差值百分比为：{{df_monitor_checker_value}} % ，请排查服务是否有异常！`,
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"clickhouse_http_connections_title":   `ClickHouse Instance {{instance}} HTTP Connections Alert`,
			"clickhouse_http_connections_message": `>Host: {{host}}\n>ClickHouse Instance: {{instance}}\n>Alert Level: {{df_status}}\n>HTTP connections count is: {{df_monitor_checker_value}}, which exceeds the threshold. Please check!`,
			"clickhoust_tcp_title":                `ClickHouse Instance {{instance}} TCP Connections Alert`,
			"clickhoust_tcp_message":              `>Host: {{host}}\n>ClickHouse Instance: {{instance}}\n>Alert Level: {{df_status}}\n>TCP connections count is: {{df_monitor_checker_value}}, which exceeds the threshold. Please check!`,
			"replicated_checks_title":             `ClickHouse Instance {{instance}} Replica Consistency Check Count Abnormal Alert`,
			"replicated_checks_message":           `>Host: {{host}}\n>ClickHouse Instance: {{instance}}\n>Alert Level: {{df_status}}\n>The percentage difference of data part consistency checks between the last 15 minutes and the last 30 minutes is: {{df_monitor_checker_value}}%. Please check if the service is abnormal!`,
		}
	default:
		return nil
	}
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"host_name":                        "主机名",
			"title":                            "MySQL 监控视图",
			"overview":                         "概览",
			"description":                      "说明",
			"description_content":              `使用这个仪表板，您可以获得连接数、QPS、TPS、吞吐量、异常连接数、每秒无索引join 查询次数、Schema 大小、慢查询、行锁等待时长、活动用户信息、缓冲情况、锁信息等。\n\n了解更多关于MySQL集成的信息：\n[我们的官方集成文档](https://docs.guance.com/integrations/datastorage/mysql/)\n[监控MySQL性能指标](https://docs.guance.com/datakit/mysql/#measurement)`,
			"connection_number":                "连接数",
			"abnormal_connection_number":       "异常连接数",
			"non_index_join_per_sec":           "每秒无索引 join 查询次数",
			"schema_size":                      "Schema 大小",
			"slow_query":                       "慢查询",
			"write_counts":                     "写入次数",
			"active_user_info":                 "活动用户信息",
			"active_user_list":                 "MySQL 活动用户信息列表",
			"user_connection_distribution":     "当前活动用户 Connections 分布",
			"user_slow_query_distribution":     "当前活动用户慢查询分布",
			"bytes_received":                   "接收字节数",
			"bytes_sent":                       "发送字节数",
			"slow_queries":                     "慢查询数量",
			"cache_hits":                       "缓存命中数",
			"disk_read_per_sec":                "每秒缓冲池磁盘读请求",
			"disk_read_alias":                  "读取请求",
			"buffer_pool_pages":                "缓冲池页数",
			"read_write_pages_per_sec":         "每秒读写页数",
			"row_lock_wait_time":               "行锁等待时长",
			"sql_exec_times_per_sec":           "每秒 SQL 执行次数",
			"read_write_requests_per_sec":      "每秒缓冲池内存读写请求",
			"net_traffic":                      "网络流量",
			"tables_opened_per_sec":            "每秒打开的表数",
			"temporary_tables_created_per_sec": "每秒自动创建的临时表数",
			"lock_information":                 "锁信息",
			"row_locks":                        "行锁数",
			"row_lock_time":                    "行锁耗时",
			"table_locks_per_sec":              "每秒表锁数",
			"aborted_clients_alias":            "由于客户端没有正确关闭连接而中止的连接数",
			"aborted_connections_alias":        "尝试连接到服务器失败的次数",
			"mysql_connect_error":              "尝试与服务器连接导致",
			"avg_lock_wait_time":               "等待锁平均耗时",
			"slow_sql":                         "慢 SQL",
			"transactions_per_sec_alias":       "每秒事务数",
			"locks_waited_alias":               "不能立即获得锁的数量",
			"locks_immediate_alias":            "立即获得锁的数量",
			"left_pages":                       "剩余页数",
			"total_pages_alias":                "总页数",
			"free_pages_alias":                 "可用页数",
			"avg_wait_time_alias":              "等待锁平均耗时",
			"avg_locks_time_alias":             "等待锁总时长",
			"avg_locks_max_time_alias":         "等待锁最大耗时",
			"read_counts_alias":                "读取次数",
			"time_unit":                        "次",
			"open_tables_alias":                "打开表数",
			"tempoary_files_alias":             "临时文件",
			"tempoary_tables_alias":            "临时表",
			"tempoary_disk_tables_alias":       "磁盘临时表",
			"write_pages_alias":                "写页面数",
			"read_pages_alias":                 "读页面数",
			"wait_locks_alias":                 "当前正在等待锁的数量",
			"wait_locks_total_alias":           "等待锁总数",
			"open_connections_alias":           "当前打开的连接数",
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"host_name":                        "host",
			"title":                            "MySQL-dashboard-template",
			"overview":                         "Overview",
			"description":                      "Description",
			"description_content":              `Using this dashboard, you can obtain connection counts, QPS, TPS, throughput, abnormal connection counts, number of non-index join queries per second, schema size, slow queries, row lock wait time, active user information, buffering status, lock information, and more.\n\nLearn more about MySQL integration:\n[Our official integration documentation](https://docs.guance.com/integrations/datastorage/mysql/)\n[Monitor MySQL performance metrics](https://docs.guance.com/datakit/mysql/#measurement)`,
			"connection_number":                "Connection number",
			"abnormal_connection_number":       "Abnormal connection number",
			"non_index_join_per_sec":           "Number of non-index join queries per second",
			"schema_size":                      "Schema size",
			"slow_query":                       "Slow query",
			"write_counts":                     "Write counts",
			"active_user_info":                 "Active user information",
			"active_user_list":                 "Active user list",
			"user_connection_distribution":     "Active user connections distribution",
			"user_slow_query_distribution":     "Active user slow query distribution",
			"bytes_received":                   "bytes received",
			"bytes_sent":                       "bytes sent",
			"slow_queries":                     "slow queries",
			"cache_hits":                       "cache hits",
			"disk_read_per_sec":                "Disk read requests per second",
			"disk_read_alias":                  "disk_read",
			"buffer_pool_pages":                "Buffer pool pages",
			"read_write_pages_per_sec":         "Page reads/writes per second",
			"row_lock_wait_time":               "Time spent in acquiring row locks",
			"sql_exec_times_per_sec":           "SQL execution times per second",
			"read_write_requests_per_sec":      "Buffer pool read and write requests per second",
			"net_traffic":                      "Net traffic",
			"tables_opened_per_sec":            "Number of tables opened per second",
			"temporary_tables_created_per_sec": "Number of temporary tables created per second",
			"lock_information":                 "Lock information",
			"row_locks":                        "Number of row locks",
			"row_lock_time":                    "Row lock time",
			"table_locks_per_sec":              "Table locks per second",
			"aborted_clients_alias":            "aborted_clients",
			"aborted_connections_alias":        "aborted_connections",
			"mysql_connect_error":              "Failed attempts to connect to the server",
			"avg_lock_wait_time":               "avg_lock_wait_time",
			"slow_sql":                         "slow SQL",
			"transactions_per_sec_alias":       "transactions_per_second",
			"locks_waited_alias":               "locks_waited",
			"locks_immediate_alias":            "locks_immediate",
			"left_pages":                       "left_pages",
			"total_pages_alias":                "total_pages",
			"free_pages_alias":                 "free_pages",
			"avg_wait_time_alias":              "avg_wait_time",
			"avg_locks_time_alias":             "avg_locks_time",
			"avg_locks_max_time_alias":         "avg_locks_max_wait_time",
			"read_counts_alias":                "read_counts",
			"time_unit":                        "time",
			"open_tables_alias":                "open_tables",
			"tempoary_files_alias":             "tempoary_files",
			"tempoary_tables_alias":            "tempoary_tables",
			"tempoary_disk_tables_alias":       "tempoary_disk_tables",
			"write_pages_alias":                "write_pages",
			"read_pages_alias":                 "read_pages",
			"wait_locks_alias":                 "wait_locks",
			"wait_locks_total_alias":           "wait_locks_total",
			"open_connections_alias":           "open_connections",
		}
	default:
		return nil
	}
}

func (*Input) DashboardList() []string {
	return nil
}

func (*Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"default_monitor_name": "默认",
			"message":              `- 所属空间：{{ df_workspace_name }}\n- 等级：{{ df_status | to_status_human }} \n- 检测对象：MySQL\n- 内容：server 为 {{server}} user 为 {{user}} 的慢查询数为 {{ Result }} ，已超出设置范围，请重点关注。\n- 建议：查看 MySQL 是否有异常。`,
			"title":                `MySQL server 为 {{server}} user 为 {{user}} 的慢查询数过多`,
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"default_monitor_name": "default",
			"message":              `- Workspace：{{ df_workspace_name }}\n- Status：{{ df_status | to_status_human }} \n- Detection object：MySQL\n- Content：The number of slow queries for server {{server}} and user {{user}} is {{Result}}, which has exceeded the set range. Please pay attention to it.\n- Suggestion：Check whether there is an exception in MySQL.`,
			"title":                `Too many slow queries for user {{user}} on MySQL server {{server}}`,
		}
	default:
		return nil
	}
}

func (*Input) MonitorList() []string {
	return nil
}

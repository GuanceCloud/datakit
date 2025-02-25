// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"title":                      "SQL Server 监控视图",
			"group_overview":             "基础监控",
			"group_base":                 "基础信息",
			"group_cache":                "缓存监控",
			"group_locks":                "锁监控",
			"group_request":              "吞吐量",
			"group_sql_quality":          "SQL 质量监控",
			"group_file":                 "文件系统",
			"group_db_file":              "数据库和文件系统",
			"host_name":                  "主机名",
			"request_branch_per_second":  "每秒批量请求数",
			"alias_db":                   "数据库",
			"alias_file_type":            "文件类型",
			"alias_logical_file_name":    "逻辑文件名",
			"alias_physical_file_name":   "物理文件名",
			"alias_server":               "服务器",
			"alias_log":                  "日志",
			"avg_write_data_bytes":       "每秒平均写入字节数",
			"avg_read_data_bytes":        "每秒平均读取字节数",
			"log_file_read_write":        "日志文件读写情况",
			"log_file_avg_write_bytes":   "平均写入数据大小",
			"log_file_avg_read_bytes":    "平均读取数据大小",
			"log_file_avg_write_count":   "平均写入数量",
			"log_file_avg_read_count":    "平均读取数量",
			"write_bytes":                "写入量",
			"read_bytes":                 "读取量",
			"data_file_read_write":       "数据文件读写大小",
			"data_file_read_write_bytes": "日志文件读写量",
			"data_avg_write_bytes":       "平均写入数据大小",
			"data_avg_read_bytes":        "平均读取数据大小",
			"data_avg_write_count":       "平均写入数量",
			"data_avg_read_count":        "平均读取数量",
			"unit_db":                    "数据库",
			"unit_log":                   "日志",
			"unit_physical_memory":       "物理内存",
			"unit_commit_memory":         "已提交的内存",
			"unit_manager_memory_free":   "内存管理器可使用的内存",
			"unit_online_time":           "在线时间",
			"database_file_size":         "数据库大小",
			"transaction_active":         "活动事务数",
			"sql_compiles_per_second":    "每秒 SQL 编译数",
			"sql_recompiles_per_second":  "每秒 SQL 重新编译数",
			"connection_user":            "用户连接数",
			"server":                     "服务器",
			"cpu_count":                  "CPU 逻辑核心数",
			"database_state_statics":     "数据库状态统计",
			"page_life_expectancy":       "页面寿命期望",
			"cache_hit_ratio":            "缓存命中率",
			"checked_point_page_count":   "检查点页面数量",
			"locks_wait_pre_second":      "每秒锁等待数",
			"blocked_processes":          "被阻塞的进程数",
		}
	case inputs.I18nEn:
		return map[string]string{
			"title":                      "SQL Server Monitor View",
			"group_overview":             "Overview",
			"group_base":                 "Base Info",
			"group_cache":                "Cache Monitor",
			"group_locks":                "Lock Monitor",
			"group_request":              "Throughput",
			"group_sql_quality":          "SQL Quality Monitor",
			"group_file":                 "File System",
			"group_db_file":              "Database and File System",
			"host_name":                  "Host Name",
			"request_branch_per_second":  "Batch requests per second",
			"alias_db":                   "Database",
			"alias_file_type":            "File Type",
			"alias_logical_file_name":    "Logical File Name",
			"alias_physical_file_name":   "Physical File Name",
			"alias_server":               "Server",
			"alias_log":                  "Log",
			"avg_write_data_bytes":       "Average write data bytes per second",
			"avg_read_data_bytes":        "Average read data bytes per second",
			"log_file_read_write":        "Log file read/write",
			"log_file_avg_write_bytes":   "Average write data size",
			"log_file_avg_read_bytes":    "Average read data size",
			"log_file_avg_write_count":   "Average write count",
			"log_file_avg_read_count":    "Average read count",
			"write_bytes":                "Write Bytes",
			"read_bytes":                 "Read Bytes",
			"data_file_read_write":       "Data file read/write",
			"data_file_read_write_bytes": "Data file read/write bytes",
			"data_avg_write_bytes":       "Average write data size",
			"data_avg_read_bytes":        "Average read data size",
			"data_avg_write_count":       "Average write count",
			"data_avg_read_count":        "Average read count",
			"unit_db":                    "Database",
			"unit_log":                   "Log",
			"unit_physical_memory":       "Physical Memory",
			"unit_commit_memory":         "Committed Memory",
			"unit_manager_memory_free":   "Memory Manager Free Memory",
			"unit_online_time":           "Online Time",
			"database_file_size":         "Database Size",
			"transaction_active":         "Active Transaction",
			"sql_compiles_per_second":    "SQL Compiles per second",
			"sql_recompiles_per_second":  "SQL Recompiles per second",
			"connection_user":            "User Connection",
			"server":                     "Server",
			"cpu_count":                  "CPU Logical Core Number",
			"database_state_statics":     "Database State Statics",
			"page_life_expectancy":       "Page Life Expectancy",
			"cache_hit_ratio":            "Cache Hit Ratio",
			"checked_point_page_count":   "Checked Point Page Count",
			"locks_wait_pre_second":      "Lock Waits per second",
			"blocked_processes":          "Blocked Processes",
		}
	default:
		return nil
	}
}

func (*Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"default_monitor_name":    "默认",
			"cache_hit_ratio_title":   "SQL Server - 缓存命中率较低",
			"cache_hit_ratio_message": `{% if df_status != 'ok' %}\n级别状态：{{df_status | to_status_human }}\n>主机：{{sqlserver_host}}\n>缓存命中率较低：{{Result}}\n>触发时间：{{date | to_datetime }}\n\n{% else %}\n级别状态：{{df_status | to_status_human }}\n>主机：{{sqlserver_host}}\n>内容：缓存命中率已经恢复\n>恢复时间：{{date | to_datetime }}\n\n{% endif %}`,
			"db_offline_title":        "SQL Server - 有数据库处于离线状态",
			"db_offline_message":      `{% if df_status != 'ok' %}\n级别状态：{{df_status | to_status_human }}\n>主机：{{sqlserver_host}}\n>处于离线的数据库数量：{{Result}}\n>触发时间：{{date | to_datetime }}\n\n{% else %}\n级别状态：{{df_status | to_status_human }}\n>主机：{{sqlserver_host}}\n>内容：已无离线数据库\n>恢复时间：{{date | to_datetime }}\n\n{% endif %}`,
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"default_monitor_name":    "Default",
			"cache_hit_ratio_title":   "SQL Server - Cache Hit Ratio Low",
			"cache_hit_ratio_message": `{% if df_status != 'ok' %}\nLevel Status: {{df_status | to_status_human }}\n>Host: {{sqlserver_host}}\n>Cache Hit Ratio Low: {{Result}}\n>Trigger Time: {{date | to_datetime }}\n\n{% else %}\nLevel Status: {{df_status | to_status_human }}\n>Host: {{sqlserver_host}}\n>Content: Cache Hit Ratio has recovered\n>Recovery Time: {{date | to_datetime }}\n\n{% endif %}`,
			"db_offline_title":        "SQL Server - Database Offline",
			"db_offline_message":      `{% if df_status != 'ok' %}\nLevel Status: {{df_status | to_status_human }}\n>Host: {{sqlserver_host}}\n>Database Offline: {{Result}}\n>Trigger Time: {{date | to_datetime }}\n\n{%else%}\nLevel Status: {{df_status | to_status_human }}\n>Host: {{sqlserver_host}}\n>Content: No offline database\n>Recovery Time: {{date | to_datetime }}\n\n{%endif%}`,
		}
	default:
		return nil
	}
}

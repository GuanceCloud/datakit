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
			"title":                          "SQL Server 监控视图",
			"overview":                       "基础监控",
			"cpu_count":                      "CPU 逻辑核心",
			"memory":                         "内存",
			"cpu_usage":                      "CPU 使用率",
			"mem_usage":                      "内存使用率",
			"active_txn":                     "活跃事务数",
			"active_mem":                     "活跃内存",
			"file_system":                    "文件系统",
			"logical_database":               "逻辑数据库",
			"data_files":                     "数据文件",
			"db_name":                        "数据库",
			"data_write":                     "数据写入",
			"file_type":                      "文件类型",
			"logical_filename":               "逻辑文件名",
			"physical_filename":              "物理文件名",
			"log_file":                       "日志文件",
			"data_file_read_write":           "数据文件读写量",
			"write_bytes":                    "写入量",
			"read_bytes":                     "读取量",
			"log_file_read_write":            "数据文件读写量",
			"database_size":                  "库存储量",
			"data_size":                      "数据量",
			"log_size":                       "日志量",
			"group_locks":                    "锁信息",
			"db_locks":                       "库锁",
			"table_locks":                    "表锁",
			"row_locks":                      "行锁",
			"dead_locks":                     "死锁",
			"dead_locks_db_name":             "死锁所在库",
			"dead_locks_session_id":          "阻塞源事物ID",
			"dead_locks_blocking_text":       "阻塞事物",
			"dead_locks_blocking_session_id": "阻塞事物ID",
			"dead_locks_request_text":        "被阻塞事物",
			"group_sql_quality":              "SQL 质量",
			"logical_io":                     "逻辑 IO",
			"sql_query":                      "SQL 详情",
			"total_logical_io":               "总逻辑 IO",
			"exec_count":                     "执行次数",
			"avg_logical_io":                 "平均逻辑 IO",
			"cpu_time":                       "CPU 时间",
			"total_cpu_time":                 "CPU 总执行时间",
			"avg_cpu_time":                   "CPU 平均执行时间",
		}
	case inputs.I18nEn:
		return map[string]string{
			"title":                          "SQL Server Monitor View",
			"overview":                       "overview",
			"cpu_count":                      "CPU logical core",
			"memory":                         "memory",
			"cpu_usage":                      "CPU usage",
			"mem_usage":                      "memory usage",
			"active_txn":                     "active transactions",
			"active_mem":                     "active memory",
			"file_system":                    "file system",
			"logical_database":               "logical database",
			"data_files":                     "data files",
			"db_name":                        "database",
			"data_write":                     "data write",
			"file_type":                      "file type",
			"logical_filename":               "logical filename",
			"physical_filename":              "physical filename",
			"log_file":                       "log file",
			"data_file_read_write":           "data file read write",
			"write_bytes":                    "write",
			"read_bytes":                     "read",
			"log_file_read_write":            "log file read write",
			"database_size":                  "database size",
			"data_size":                      "data size",
			"log_size":                       "log size",
			"group_locks":                    "locks",
			"db_locks":                       "database locks",
			"table_locks":                    "table locks",
			"row_locks":                      "row locks",
			"dead_locks":                     "dead locks",
			"dead_locks_db_name":             "database name",
			"dead_locks_session_id":          "blocking session id",
			"dead_locks_blocking_text":       "blocking text",
			"dead_locks_blocking_session_id": "blocking session id",
			"dead_locks_request_text":        "request text",
			"group_sql_quality":              "SQL quality",
			"logical_io":                     "logical IO",
			"sql_query":                      "SQL detail",
			"total_logical_io":               "total logical IO",
			"exec_count":                     "execution count",
			"avg_logical_io":                 "average logical IO",
			"cpu_time":                       "CPU time",
			"total_cpu_time":                 "CPU total execution time",
			"avg_cpu_time":                   "CPU average execution time",
		}
	default:
		return nil
	}
}

func (*Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			//nolint:lll
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
		}
	default:
		return nil
	}
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package db2

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

var (
	_ inputs.Dashboard = (*Input)(nil)
	_ inputs.Monitor   = (*Input)(nil)
)

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"alarm":                         "告警",
			"application_active":            "数据库目前连接着的程序数量",
			"application_executing":         "数据库目前在处理的程序数量",
			"backup_latest":                 "距离上次备份的时间间隔",
			"bufferpool_column_reads_total": "总的 Column 读取数",
			"bufferpool_data_reads_total":   "总的 Data 读取数",
			"bufferpool_hit_percent":        "Buffer pool 缓存命中率",
			"bufferpool_index_reads_total":  "总的 Index 读取数",
			"bufferpool_reads_total":        "总的 Bufferpool 读取数",
			"bufferpool_xda_reads_total":    "总的 XDA 读取数",
			"category_bufferpool":           "Bufferpool",
			"category_connections":          "连接情况",
			"category_database":             "数据库",
			"category_locks":                "锁统计",
			"category_log":                  "日志统计",
			"category_overview":             "告警",
			"category_readwrite":            "读/写",
			"category_tablespace":           "表空间",
			"connection_active":             "连接数",
			"connection_max":                "最大同时连接数",
			"lock_dead":                     "死锁发生的总数",
			"lock_timeouts":                 "请求锁定一个对象的超时而非允许的次数",
			"lock_wait":                     "平均锁的等待时间",
			"log_reads_writes":              "日志页的读写情况",
			"log_used":                      "目前使用的活动日志空间的磁盘块数量（一块 4 KB）",
			"log_utilized":                  "活动日志空间的利用率",
			"reads_returned_modified":       "数据行的读、写、更新情况",
			"status":                        "数据库状态 (0: 正常  1: 警告  2: 严重  3: 未知)",
			"tablespace_usable_used":        "平均表空间的使用情况",
			"tablespace_utilized":           "表空间的利用率",
		}
	case inputs.I18nEn:
		return map[string]string{
			"alarm":                         "Alarm",
			"application_active":            "Number of applications that are currently connected",
			"application_executing":         "Number of applications DB is currently processing",
			"backup_latest":                 "Time since last backup",
			"bufferpool_column_reads_total": "Total Column Reads",
			"bufferpool_data_reads_total":   "Total Data Reads",
			"bufferpool_hit_percent":        "Buffer pool cache hit ratio",
			"bufferpool_index_reads_total":  "Total Index Reads",
			"bufferpool_reads_total":        "Total Bufferpool Reads",
			"bufferpool_xda_reads_total":    "Total XDA Reads",
			"category_bufferpool":           "Bufferpool",
			"category_connections":          "Connections",
			"category_database":             "Database",
			"category_locks":                "Lock Statistics",
			"category_log":                  "Log Statistics",
			"category_overview":             "Overview",
			"category_readwrite":            "Reads/Writes",
			"category_tablespace":           "Tablespace",
			"connection_active":             "The current number of connections",
			"connection_max":                "Max simultaneous connections",
			"lock_dead":                     "Deadlocks",
			"lock_timeouts":                 "Lock timeouts",
			"lock_wait":                     "Average lock wait",
			"log_reads_writes":              "Log Reads and Writes",
			"log_used":                      "Disk blocks (4 KiB each) of active log space currently used",
			"log_utilized":                  "Log utilization",
			"reads_returned_modified":       "Total Row Reads and Updates",
			"status":                        "Database status (0: OK  1: WARNING  2: CRITICAL  3: UNKNOWN)",
			"tablespace_utilized":           "Table space utilization",
			"tablespace_usable_used":        "Average Tablespace Usage",
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
		return map[string]string{
			"message": "[db2_status_exception] IBM Db2 数据库状态异常，请检查",
			"title":   "db2_status_exception",
		}
	case inputs.I18nEn:
		return map[string]string{
			"message": "[db2_status_exception] Abnormal IBM Db2 database status, please check",
			"title":   "db2_status_exception",
		}
	default:
		return nil
	}
}

func (*Input) MonitorList() []string {
	return nil
}

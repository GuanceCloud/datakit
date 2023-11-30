// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"title":                        "PostgreSQL 监控视图",
			"dashboard_desc":               "此仪表板展示了 PostgreSQL 数据库的运行状况，可以跟踪每个服务器的吞吐量，主从复制和锁情况等其他指标。 ",
			"max_connections_percent":      "最大连接数百分比",
			"group_overview":               "概览",
			"group_resource":               "资源使用情况",
			"group_locks":                  "锁",
			"group_connection":             "连接监控",
			"group_throughput":             "吞吐量",
			"group_transaction":            "事务",
			"group_checkpoint":             "检查点",
			"resource_table":               "表资源",
			"temp_bytes":                   "临时文件字节数",
			"temp_files":                   "临时文件数",
			"most_scan_index":              "最常被扫描的索引",
			"least_scan_index":             "最少被扫描的索引",
			"table_disk_usage_top":         "表占用磁盘空间",
			"table_live_rows_top":          "表行数",
			"dead_rows":                    "待回收的行数",
			"lock_by_mode":                 "按模式统计锁的个数",
			"lock_top_10":                  "锁的模式",
			"dead_lock":                    "死锁",
			"title_not_contain_postgresql": "不包含 postgresql 库",
			"actions":                      "操作",
			"connections_per_db":           "数据库连接数",
			"most_connected_db":            "数据库连接数",
			"seq_idx_scan":                 "顺序扫描 vs 索引扫描",
			"seq_scan_7":                   "顺序扫描 7d",
			"table_scan":                   "表扫描",
			"this_week":                    "本周",
			"table":                        "表",
			"disk_usage":                   "磁盘占用",
			"index_scan":                   "索引扫描",
			"live_rows":                    "行数",
			"inserted":                     "插入",
			"updated":                      "更新",
			"deleted":                      "删除",
			"db":                           "数据库",
			"idx_scan":                     "索引扫描",
			"seq_scan":                     "顺序扫描",
			"commit":                       "提交",
			"rollback":                     "回滚",
			"checkpoints_timed":            "时间触发",
			"checkpoints_requested":        "请求触发",
			"sync_time":                    "同步时间",
			"write_time":                   "写入时间",
			"rows_fetch_return":            "获取/返回的行比例",
			"group_rows":                   "记录行",
			"heap_only_updates":            "仅堆更新",
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"title":                        "PostgreSQL Monitor View",
			"dashboard_desc":               "This dashboard provides a high-level overview of your PostgreSQL databases, so you can track throughput, replication, locks, and other metrics from all your servers and spot potential issues. ",
			"max_connections_percent":      "max connections in use",
			"group_overview":               "Current Activity",
			"group_resource":               "Resource Utilization",
			"group_locks":                  "Locks",
			"group_connection":             "Connection Monitor",
			"group_throughput":             "Throughput",
			"group_transaction":            "Transactions",
			"group_checkpoint":             "Checkpoint",
			"resource_table":               "resources by table",
			"temp_bytes":                   "temp bytes",
			"temp_files":                   "temp files",
			"most_scan_index":              "most frequently scanned indexes",
			"least_scan_index":             "least frequently scanned indexes",
			"table_disk_usage_top":         "tables with most disk usage",
			"table_live_rows_top":          "tables with most live rows",
			"dead_rows":                    "dead rows",
			"lock_by_mode":                 "locks by lock mode",
			"lock_top_10":                  "locks",
			"dead_lock":                    "deadlocks count",
			"title_not_contain_postgresql": "not contain db postgresql",
			"actions":                      "actions",
			"connections_per_db":           "connections per database",
			"most_connected_db":            "most connected databases",
			"seq_idx_scan":                 "sequential scans vs index scans",
			"seq_scan_7":                   "sequential scans 7d",
			"table_scan":                   "table scans",
			"this_week":                    "This week",
			"table":                        "table",
			"disk_usage":                   "disk usage",
			"index_scan":                   "index scans",
			"live_rows":                    "live rows",
			"inserted":                     "inserted",
			"updated":                      "updated",
			"deleted":                      "deleted",
			"db":                           "db",
			"idx_scan":                     "index scan",
			"seq_scan":                     "sequential scan",
			"commit":                       "commit",
			"rollback":                     "rollback",
			"checkpoints_timed":            "timed checkpoints",
			"checkpoints_requested":        "requested checkpoints",
			"sync_time":                    "sync time",
			"write_time":                   "write time",
			"rows_fetch_return":            "Rows fetched/returned",
			"group_rows":                   "Rows",
			"heap_only_updates":            "heap-only updates",
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
			"default_monitor_name": "默认",
			"message":              `>等级：{{df_status}}  \n>数据库：{{db}}\n>服务器：{{server}}\n>内容：PostgreSQL 连接数使用率为 {{ Result |  to_fixed(2) }}%  \n>建议：检查 PostgreSQL 是否有异常`,
			"title":                "PostgreSQL 连接数使用率过高",
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"default_monitor_name": "default",
			"message":              `>Status：{{df_status}}  \n>Database：{{db}}\n>Server：{{server}}\n>Content：PostgreSQL connection usage is {{ Result |  to_fixed(2) }}%  \n>Suggestion：Check if PostgreSQL is abnormal`,
			"title":                "High connection usage of PostgreSQL",
		}
	default:
		return nil
	}
}

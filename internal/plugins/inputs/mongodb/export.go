// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"title":                           "Mongodb 监控视图",
			"instance_name":                   "实例名称",
			"host_name":                       "主机名",
			"group_resource":                  "资源概览",
			"group_connection":                "连接",
			"group_agg":                       "聚合",
			"group_assertion":                 "断言",
			"group_document":                  "文档",
			"group_log":                       "日志",
			"document_removed_number":         "每秒使用ttl索引从集合中删除的文档总数",
			"process_document_removed_number": "每秒后台进程使用ttl索引从集合中删除文档的次数",
			"performance":                     "性能",
			"ttl_index":                       "TTL索引",
			"dirty_page_per":                  "内存中的脏页占比",
			"used_cache_per":                  "已使用缓存占比",
			"resource_overview":               "资源概览",
			"cache":                           "缓存",
			"cache_usage":                     "缓存使用率",
			"cache_dirty_usage":               "脏页使用率",
			"inflow":                          "入流量",
			"outflow":                         "出流量",
			"network_traffic":                 "网络流量",
			"delete_execution_per_sec":        "每秒执行删除次数",
			"refresh_time_per_sec":            "每秒刷新次数",
			"query_executed_per_sec":          "每秒执行查询次数",
			"command_executed_per_sec":        "每秒执行命令次数",
			"insertion_performed_per_sec":     "每秒执行插入次数",
			"operation_per_sec":               "每秒执行操作次数",
			"updated_performed_per_sec":       "每秒执行更新次数",
			"command":                         "命令",
			"command_fail_number":             "失败命令数",
			"connection":                      "连接数",
			"available_connection":            "可用连接数",
			"current_connection":              "当前连接数",
			"connection_created_per_sec":      "每秒创建连接次数",
			"pcs":                             "个",
			"qps":                             "次",
			"document_deleted_number":         "删除的文档总数",
			"document_inserted_number":        "插入的文档总数",
			"document_updated_number":         "更新的文档总数",
			"document_returned_number":        "返回的文档总数",
			"document_operation_per_sec":      "每秒文档操作",
			"key_scanned_number":              "扫描的key总数",
			"document_scan":                   "文档扫描",
			"document_scanned_number":         "扫描的文档总数",
			"cursor_pinned_number":            "固定游标数",
			"cursor_not_timed_out_number":     "游标未超时数",
			"cursor":                          "游标",
			"cursor_open":                     "打开的游标数",
			"cursor_number":                   "游标总数",
			"cursor_timeout_number":           "超时游标数",
			"occupied_memory":                 "已占用内存",
			"resident_memory":                 "常驻内存",
			"virtual_memory":                  "虚拟内存",
			"memory":                          "内存",
			"running_time":                    "运行时间",
			"lock_queue":                      "锁队列",
			"lock_request":                    "锁请求",
			"waiting_read_lock_queue_length":  "等待读锁队列长度",
			"waiting_write_lock_queue_length": "等待写锁队列长度",
			"client_requesting_read_lock":     "当前请求读锁的活动客户端",
			"client_requesting_write_lock":    "当前请求写锁的活动客户端",
			"cpu_usage":                       "CPU 使用率",
			"mem_usage":                       "内存使用率",
			"disk_usage":                      "磁盘使用率",
			"aggregation_fail_number":         "聚合失败次数",
			"assertion_message_number":        "断言消息次数",
			"assertion_user_number":           "断言用户次数",
			"log_time":                        "时间",
			"log_message":                     "内容",
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"title":                           "Mongodb Monitor View",
			"instance_name":                   "Instance Name",
			"host_name":                       "Host Name",
			"group_resource":                  "Resource Overview",
			"group_connection":                "Connection",
			"group_agg":                       "Aggregation",
			"group_assertion":                 "Assertion",
			"group_document":                  "Document",
			"group_log":                       "Log",
			"document_removed_number":         "Total number of documents removed from the collection per second using the ttl index",
			"process_document_removed_number": "Number of documents removed from the collection per second using the ttl index in background process",
			"performance":                     "Performance",
			"ttl_index":                       "TTL Index",
			"dirty_page_per":                  "Percentage of dirty pages in memory",
			"used_cache_per":                  "Percentage of used cache",
			"resource_overview":               "Resource Overview",
			"cache":                           "Cache",
			"cache_usage":                     "Cache Usage",
			"cache_dirty_usage":               "Dirty page usage",
			"inflow":                          "Inflow",
			"outflow":                         "Outflow",
			"network_traffic":                 "Network Traffic",
			"delete_execution_per_sec":        "Delete executions per second",
			"refresh_time_per_sec":            "Refresh times per second",
			"query_executed_per_sec":          "Queries executed per second",
			"command_executed_per_sec":        "Commands executed per second",
			"insertion_performed_per_sec":     "Insertions performed per second",
			"operation_per_sec":               "Operations performed per second",
			"updated_performed_per_sec":       "Updates performed per second",
			"command":                         "Command",
			"command_fail_number":             "Failed command number",
			"connection":                      "Connections",
			"available_connection":            "Available Connections",
			"current_connection":              "Current Connections",
			"connection_created_per_sec":      "Connections created per second",
			"pcs":                             "pcs",
			"qps":                             "qps",
			"document_deleted_number":         "Total number of documents deleted",
			"document_inserted_number":        "Total number of documents inserted",
			"document_updated_number":         "Total number of documents updated",
			"document_returned_number":        "Total number of documents returned",
			"document_operation_per_sec":      "Document operations per second",
			"key_scanned_number":              "Total number of keys scanned",
			"document_scan":                   "Document Scan",
			"document_scanned_number":         "Total number of documents scanned",
			"cursor_pinned_number":            "Number of pinned cursors",
			"cursor_timeout_number":           "Number of timed out cursors",
			"cursor_not_timed_out_number":     "Number of cursors not timed out",
			"cursor":                          "Cursor",
			"cursor_open":                     "Number of open cursors",
			"cursor_number":                   "Total number of cursors",
			"occupied_memory":                 "Occupied Memory",
			"resident_memory":                 "Resident Memory",
			"virtual_memory":                  "Virtual Memory",
			"memory":                          "Memory",
			"running_time":                    "Running Time",
			"lock_queue":                      "Lock Queue",
			"lock_request":                    "Lock Request",
			"waiting_read_lock_queue_length":  "Current length of waiting read lock queue",
			"waiting_write_lock_queue_length": "Current length of waiting write lock queue",
			"client_requesting_read_lock":     "Active clients currently requesting read locks",
			"client_requesting_write_lock":    "Active clients currently requesting write locks",
			"cpu_usage":                       "CPU Usage",
			"mem_usage":                       "Memory Usage",
			"disk_usage":                      "Disk Usage",
			"aggregation_fail_number":         "Aggregation Fail Number",
			"assertion_message_number":        "Assertion Message Number",
			"assertion_user_number":           "Assertion User Number",
			"log_time":                        "Time",
			"log_message":                     "Content",
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
			"mongodb_connection_title":   `MongoDB 的连接数过高`,
			"mongodb_connection_message": `数据库实例：{{instance}}\n当前连接数达到最大连接数的：{{ Result }}`,
			"mongodb_master_title":       `MongoDB 主从写操作超过{{Result}}`,
			"mongodb_master_message":     `当前{{rs_name}}集群写操作延迟过高，请关注！\n写操作延迟:{{ Result }}\n集群名称:{{rs_name}}`,
			"mongodb_agg_fail_title":     `MongoDB 聚合命令失败率过多`,
			"mongodb_agg_fail_message":   `当前MongoDB {{instance}}聚合命令失败过去，请关注！\n失败次数:{{ Result }}\n数据库实例:{{instance}}`,
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"mongodb_connection_title":   `MongoDB connection number is too high`,
			"mongodb_connection_message": `Database instance: {{instance}} \nThe current connection number reaches the maximum connection number: {{ Result }}`,
			"mongodb_master_title":       `MongoDB master-slave write operation is too high {{Result}}`,
			"mongodb_master_message":     `Current {{rs_name}} cluster write operation delay is too high, please pay attention to it!\nWrite operation delay: {{Result}}\nCluster name: {{rs_name}}`,
			"mongodb_agg_fail_title":     `MongoDB aggregation command failure rate is too high`,
			"mongodb_agg_fail_message":   `Current MongoDB {{instance}} aggregation command failure, please pay attention to it!\nFailure times: {{Result}}\nDatabase instance: {{instance}}`,
		}
	default:
		return nil
	}
}

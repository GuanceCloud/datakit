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
			"host_name":                       "主机名",
			"document_removed_number":         "每秒使用ttl索引从集合中删除的文档总数",
			"process_document_removed_number": "每秒后台进程使用ttl索引从集合中删除文档的次数",
			"performance":                     "性能",
			"ttl_index":                       "TTL索引",
			"dirty_page_per":                  "内存中的脏页占比",
			"used_cache_per":                  "已使用缓存占比",
			"resource_overview":               "资源概览",
			"cache":                           "缓存",
			"inflow":                          "入流量",
			"outflow":                         "出流量",
			"network_traffic":                 "网络流量",
			"delete_execution_per_sec":        "每秒执行删除次数",
			"refresh_time_per_sec":            "每秒刷新次数",
			"query_executed_per_sec":          "每秒执行查询次数",
			"command_executed_per_sec":        "每秒执行命令次数",
			"insertion_performed_per_sec":     "每秒执行插入次数",
			"operation_per_sec":               "每秒执行操作次数",
			"connection":                      "连接数",
			"available_connection":            "可用连接数",
			"current_connection":              "当前连接数",
			"connection_created_per_sec":      "每秒创建连接次数",
			"pcs":                             "个",
			"document_deleted_number":         "删除的文档总数",
			"document_inserted_number":        "插入的文档总数",
			"document_updated_number":         "更新的文档总数",
			"document_returned_number":        "返回的文档总数",
			"document_operation_per_sec":      "每秒文档操作",
			"key_scanned_number":              "扫描的key总数",
			"document_scan":                   "文档扫描",
			"document_scanned_number":         "扫描的文档总数",
			"cursor_pinned_number":            "固定游标数",
			"cursor_timeout_number":           "超时游标数",
			"cursor_not_timed_out_number":     "游标未超时数",
			"cursor":                          "游标",
			"cursor_number":                   "游标总数",
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
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"title":                           "Mongodb Monitor View",
			"host_name":                       "Host Name",
			"document_removed_number":         "Total number of documents removed from the collection per second using the ttl index",
			"process_document_removed_number": "Number of documents removed from the collection per second using the ttl index in background process",
			"performance":                     "Performance",
			"ttl_index":                       "TTL Index",
			"dirty_page_per":                  "Percentage of dirty pages in memory",
			"used_cache_per":                  "Percentage of used cache",
			"resource_overview":               "Resource Overview",
			"cache":                           "Cache",
			"inflow":                          "Inflow",
			"outflow":                         "Outflow",
			"network_traffic":                 "Network Traffic",
			"delete_execution_per_sec":        "Delete executions per second",
			"refresh_time_per_sec":            "Refresh times per second",
			"query_executed_per_sec":          "Queries executed per second",
			"command_executed_per_sec":        "Commands executed per second",
			"insertion_performed_per_sec":     "Insertions performed per second",
			"operation_per_sec":               "Operations performed per second",
			"connection":                      "Connections",
			"available_connection":            "Available Connections",
			"current_connection":              "Current Connections",
			"connection_created_per_sec":      "Connections created per second",
			"pcs":                             "pcs",
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

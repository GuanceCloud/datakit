// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package neo4j

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

//nolint:lll
func (i *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"dbms_page_cache_usage_ratio_note":                     "已用页面数与可用页面总数的比率",
			"dbms_page_cache_usage_ratio":                          "已用页面数比率",
			"database_cypher_replan_events_total_note":             "Cypher 决定重新计划查询的总次数",
			"database_cypher_replan_events_total":                  "Cypher 重新计划时间",
			"dbms_page_cache_page_faults_total_note":               "页面缓存中发生的页面错误总数",
			"dbms_page_cache_page_faults_total":                    "页缓存错误",
			"dbms_page_cache_flushes_total_note":                   "页缓存执行的页刷新总数",
			"dbms_page_cache_flushes_total":                        "页缓存刷新",
			"dbms_bolt_connections_note":                           "打开的 Bolt 连接总数<br>关闭的 Bolt 连接总数<br>正在执行`Cypher`并返回结果的 Bolt 连接总数<br>Bolt 连接空闲数",
			"dbms_bolt_connections":                                "Bolt 连接",
			"dbms_vm_file_descriptors_note":                        "打开的文件描述符数<br>打开的文件描述符的上限",
			"dbms_vm_file_descriptors":                             "虚拟机文件描述符",
			"database_ids_in_use_note":                             "不同关系类型的总数<br>不同属性名称的总数<br>ID 总数<br>节点总数",
			"database_ids_in_use":                                  "在用数据库 ID",
			"database_db_query_execution_latency_millis_note":      "成功执行查询的执行时间(毫秒)",
			"database_db_query_execution_latency_millis":           "查询执行延迟(毫秒)",
			"database_transaction_last_committed_tx_id_total_note": "",
			"database_transaction_last_committed_tx_id_total":      "上次提交的事务的ID",
			"dbms_vm_heap_committed_note":                          "保证可供 JVM 使用的内存量(字节)",
			"dbms_vm_heap_committed":                               "保证 JVM 内存量",
			"dbms_vm_heap_used_note":                               "当前使用的内存量(字节)",
			"dbms_vm_heap_used":                                    "使用的内存量",
		}
	case inputs.I18nEn:
		return map[string]string{
			"dbms_page_cache_usage_ratio_note":                     "The ratio of number of used pages to total number of available pages",
			"dbms_page_cache_usage_ratio":                          "Page Cache Usage Ratio",
			"database_cypher_replan_events_total_note":             "The total number of times Cypher has decided to re-plan a query",
			"database_cypher_replan_events_total":                  "Cypher Replan Events",
			"dbms_page_cache_page_faults_total_note":               "The total number of page faults happened in the page cache",
			"dbms_page_cache_page_faults_total":                    "Page Cache Faults",
			"dbms_page_cache_flushes_total_note":                   "The total number of page merges executed by the page cache",
			"dbms_page_cache_flushes_total":                        "Page Cache Flushes",
			"dbms_bolt_connections_note":                           "Number of Bolt connections opened<br>Number of Bolt connections closed<br>Number of Bolt connections running<br>Number of Bolt connections idle",
			"dbms_bolt_connections":                                "Bolt Connections",
			"dbms_vm_file_descriptors_note":                        "Number of open file descriptors<br>Maximum number of open file descriptors",
			"dbms_vm_file_descriptors":                             "Vm File Descriptors",
			"database_ids_in_use_note":                             "Number of different relationship types<br>Number of different property names<br>Number of relationships<br>Number of nodes",
			"database_ids_in_use":                                  "IDs In use",
			"database_db_query_execution_latency_millis_note":      "Execution time in milliseconds of queries executed successfully",
			"database_db_query_execution_latency_millis":           "Query Execution Latency(ms)",
			"database_transaction_last_committed_tx_id_total_note": "The ID of the last committed transaction",
			"database_transaction_last_committed_tx_id_total":      "Transaction  Last Committed ID",
			"dbms_vm_heap_committed_note":                          "Amount of memory (in bytes) guaranteed to be available for use by the JVM",
			"dbms_vm_heap_committed":                               "JVM Heap Committed",
			"dbms_vm_heap_used_note":                               "Amount of memory (in bytes) currently used",
			"dbms_vm_heap_used":                                    "JVM Heap Used",
		}
	default:
		return nil
	}
}

func (i *Input) DashboardList() []string {
	return nil
}

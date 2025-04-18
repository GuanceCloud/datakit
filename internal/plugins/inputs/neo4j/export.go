// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package neo4j

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

//nolint:lll
func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"var_instance_name":                                    "实例名称",
			"group_database_performance":                           "数据库性能指标",
			"group_storage_memory":                                 "存储和内存指标",
			"group_transaction_query":                              "事务和查询指标",
			"group_gc_system":                                      "垃圾回收和系统指标",
			"group_cache_io":                                       "页面缓存和I/O指标",
			"database_query_duration":                              "数据库查询执行时间",
			"active_read_transaction_count":                        "当前活跃的读事务数",
			"active_write_transaction_count":                       "当前活跃的写事务数",
			"transaction_peak_count":                               "最高并发事务数",
			"page_cache_hit_ratio":                                 "页面缓存命中率",
			"database_storage_size":                                "数据库存储大小",
			"database_pool_used_heap":                              "数据库池已使用的堆大小",
			"database_pool_used_native":                            "数据库池已使用本地大小",
			"dbms_vm_heap_used":                                    "已使用 JVM 堆内存",
			"transaction_committed_total":                          "事务已提交的总数",
			"transaction_rollbacks_total":                          "事务回滚总数",
			"db_query_execution_success_total":                     "数据库查询执行成功总数",
			"db_query_execution_success_total_desc":                "该图表展示了 Neo4j 数据库中查询执行失败总数的最新值随时间的变化趋势。通过 last 函数获取最新的 database_db_query_execution_failure_total 指标值，并以折线图的形式展示其变化情况。",
			"db_query_execution_failure_total":                     "数据库查询执行失败总数",
			"vm_gc_time_total":                                     "数据库VM垃圾回收总时间",
			"vm_file_descriptors_count":                            "当前打开文件描述符数量",
			"cypher_replan_events_total":                           "Cypher 重规划事件总数",
			"page_cache_hit_total":                                 "页面缓存命中总数",

		}
	case inputs.I18nEn:
		return map[string]string{
			"var_instance_name":                                    "Instance Name",
			"group_database_performance":                           "Database Performance",
			"group_storage_memory":                                 "Storage Memory",
			"group_transaction_query":                              "Transaction Query",
			"group_gc_system":                                      "GC System",
			"group_cache_io":                                       "Cache IO",
			"database_query_duration":                              "Database Query Duration",
			"active_read_transaction_count":                        "Active Read Transaction Count",
			"active_write_transaction_count":                       "Active Write Transaction Count",
			"transaction_peak_count":                               "Transaction Peak Count",
			"page_cache_hit_ratio":                                 "Page Cache Hit Ratio",
			"database_storage_size":                                "Database Storage Size",
			"database_pool_used_heap":                              "Database Pool Used Heap",
			"database_pool_used_native":                            "Database Pool Used Native",
			"dbms_vm_heap_used":                                    "DBMS VM Heap Used",
			"transaction_committed_total":                          "Transaction Committed Total",
			"transaction_rollbacks_total":                          "Transaction Rollbacks Total",
			"db_query_execution_success_total":                     "DB Query Execution Success Total",
			"db_query_execution_success_total_desc":                "This chart shows the latest value of database_db_query_execution_failure_total, which represents the total number of failed database queries, as a line graph over time. The last function is used to retrieve the latest value of this metric, and the data is presented as a line graph.",
			"db_query_execution_failure_total":                     "DB Query Execution Failure Total",
			"vm_gc_time_total":                                     "VM GC Time Total",
			"vm_file_descriptors_count":                            "VM File Descriptors Count",
			"cypher_replan_events_total":                           "Cypher Replan Events Total",
			"page_cache_hit_total":                                 "Page Cache Hit Total",
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
			"database_db_query_execution_title":                `查询的执行时间过长`,
			"database_db_query_execution_message":              `查询的执行时间过长，请关注：\n{{instance}}的{{db}}的查询时间超过{{Result}}请关注`,
			"page_cache_hit_ratio_title":                       `页面缓存的命中率过低`,
			"page_cache_hit_ratio_message":                     `{{instance}}的页面缓存的命中率低于{{df_status}},请关注`,
			"vm_gc_time_total_title":                           `垃圾回收的时间过长`,
			"vm_gc_time_total_message":                         `{{instance}}的垃圾回收时间超过{{df_status}}`,

		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
		    "database_db_query_execution_title":                `Query execution time is too long`,
		    "database_db_query_execution_message":              `Query execution time is too long, please pay attention to: \n{{instance}} of {{db}} query time exceeds {{Result}}, please pay attention to`,
		    "page_cache_hit_ratio_title":                       `Page cache hit ratio is too low`,
		    "page_cache_hit_ratio_message":                     `{{instance}} page cache hit ratio is below {{df_status}}, please pay attention to`,
		    "vm_gc_time_total_title":                           `GC time is too long`,
		    "vm_gc_time_total_message":                         `{{instance}} GC time exceeds {{df_status}}`,
		}
	default:
		return nil
	}
}


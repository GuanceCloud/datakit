// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oceanbase

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

var (
	_ inputs.Dashboard = (*Input)(nil)
	_ inputs.Monitor   = (*Input)(nil)
)

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"alarm":                              "告警面板",
			"category_alarm":                     "告警",
			"category_lock":                      "锁情况",
			"category_log":                       "日志区",
			"category_memory":                    "内存情况",
			"category_plancache":                 "计划缓存 (Plan Cache) 情况",
			"category_status":                    "状态情况",
			"category_workarea":                  "Work area 情况",
			"log_slow":                           "慢查询日志",
			"ob_concurrent_limit_sql_count":      "被限流的 SQL 个数",
			"ob_database_status":                 "数据库状态 (1: 正常)",
			"ob_lock_count":                      "行锁的个数",
			"ob_lock_max_ctime":                  "最大加锁耗时 (单位: 秒)",
			"ob_mem_sum_count":                   "所有租户使用中的内存单元个数",
			"ob_mem_sum_used":                    "所有租户当前使用的内存数值 (单位: Byte)",
			"ob_memstore_active_rate":            "所有服务器上所有租户的 Memtable 的内存活跃率",
			"ob_plancache_avg_hit_rate":          "所有 Server 上 plan_cache 的平均命中率",
			"ob_plancache_mem_used_rate":         "所有 Server 上 plan_cache 的总体内存使用率",
			"ob_plancache_sum_plan_num":          "所有 Server 上 plan 的总数",
			"ob_ps_hit_rate":                     "PS (Prepared Statement) Cache 的命中率",
			"ob_session_avg_wait_time":           "所有服务器上所有 Session 的当前或者上一次等待事件的平均等待耗时 (单位: 微秒)",
			"ob_workarea_global_mem_bound":       "auto 模式下，全局最大可用内存大小",
			"ob_workarea_max_auto_workarea_size": "预计最大可用内存大小，表示当前 workarea 情况下，auto 管理的最大内存大小",
			"ob_workarea_mem_target":             "当前 workarea 可用内存的目标大小",
		}
	case inputs.I18nEn:
		return map[string]string{
			"alarm":                              "Alarm panel",
			"category_alarm":                     "Alarm",
			"category_lock":                      "Lock",
			"category_log":                       "Log",
			"category_memory":                    "Memory",
			"category_plancache":                 "Plan Cache",
			"category_status":                    "Status",
			"category_workarea":                  "Work area",
			"log_slow":                           "Long Running Queries",
			"ob_concurrent_limit_sql_count":      "Concurrent limit sql count",
			"ob_database_status":                 "Database status (1: Active)",
			"ob_lock_count":                      "Lock count",
			"ob_lock_max_ctime":                  "Lock maximum time cost (Unit: second)",
			"ob_mem_sum_count":                   "All tenants used memory unit count",
			"ob_mem_sum_used":                    "All tenants used memory (Unit: Byte)",
			"ob_memstore_active_rate":            "All server all tenants Memtable memory active rate",
			"ob_plancache_avg_hit_rate":          "All server plan_cache average hit rate",
			"ob_plancache_mem_used_rate":         "All server plan_cache used rate",
			"ob_plancache_sum_plan_num":          "All server plan count",
			"ob_ps_hit_rate":                     "PS (Prepared Statement) cache hit rate",
			"ob_session_avg_wait_time":           "All server all session current or previous wait event average wait time (Unit: microsecond)",
			"ob_workarea_global_mem_bound":       "Global available memory size in auto mode",
			"ob_workarea_max_auto_workarea_size": "Estimated maximum available memory size, represents auto manage maximum memory size in current workarea",
			"ob_workarea_mem_target":             "Current workarea available memory target size",
		}
	default:
		return nil
	}
}

func (*Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"message": "[oceanbase_status_exception] OceanBase 数据库状态异常，请检查",
			"title":   "oceanbase_status_exception",
		}
	case inputs.I18nEn:
		return map[string]string{
			"message": "[oceanbase_status_exception] Abnormal OceanBase database status, please check",
			"title":   "oceanbase_status_exception",
		}
	default:
		return nil
	}
}

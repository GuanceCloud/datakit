// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"title":                         "Oracle 监控视图",
			"host_name":                     "主机名",
			"service_name":                  "服务名",
			"cache_hit_ratio":               "缓存命中率",
			"buffer_cache_hit_ratio":        "缓冲区缓存命中率",
			"library_cache_hit_ratio":       "库缓存命中率",
			"cursor_cache_hit_ratio":        "游标缓存命中率",
			"instance_information":          "实例基础信息",
			"instance_process_information":  "实例进程信息",
			"service_response_time":         "服务响应时间",
			"response_time":                 "响应时间",
			"temporary_space_size":          "临时空间大小",
			"activity_sorting_block_number": "活动排序的总块数",
			"temporary_space":               "TEMP空间",
			"used":                          "已使用",
			"total_size":                    "总大小",
			"percentage_occupied":           "百分比占用",
			"table_place":                   "表空间",
			"mount_state":                   "挂载状态",
			"wait_session":                  "等待会话",
			"session":                       "会话",
			"session_number":                "会话总数",
			"active_session_number":         "活跃会话数",
			"user_sort":                     "用户排序",
			"disk_sort":                     "磁盘排序",
			"data_sort":                     "数据排序",
			"time_per_second":               "次/秒",
			"physical_read":                 "物理读取",
			"physical_write":                "物理写入",
			"physical_read_write_number":    "物理读写次数",
			"pga_mem_alloc":                 "按进程分配的 PGA 内存",
			"pga_freeable":                  "按进程释放的 PGA 内存",
			"pga_mem_max":                   "进程分配的 PGA 最大内存",
			"pga_mem_used":                  "进程分配的 PGA 已使用内存",
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"title":                         "Oracle Monitoring View",
			"host_name":                     "Host Name",
			"service_name":                  "Service Name",
			"cache_hit_ratio":               "Cache hit ratio",
			"buffer_cache_hit_ratio":        "Buffer cache hit ratio",
			"library_cache_hit_ratio":       "Library cache hit ratio",
			"cursor_cache_hit_ratio":        "Cursor cache hit ratio",
			"instance_information":          "Basic information of instance",
			"instance_process_information":  "Basic information of instance process",
			"service_response_time":         "Service response time",
			"response_time":                 "Response time",
			"temporary_space_size":          "Temporary space size",
			"activity_sorting_block_number": "Total number of blocks for active sorting",
			"temporary_space":               "TEMP space",
			"used":                          "Used",
			"total_size":                    "Total size",
			"percentage_occupied":           "Percentage occupied",
			"table_place":                   "Tablespace",
			"mount_state":                   "Mount state",
			"wait_session":                  "Waiting for session",
			"session":                       "Sessions",
			"session_number":                "Total number of sessions",
			"active_session_number":         "Number of active sessions",
			"user_sort":                     "User sort",
			"disk_sort":                     "Disk sort",
			"data_sort":                     "Data sort",
			"time_per_second":               "times per second",
			"physical_read":                 "Physical read",
			"physical_write":                "Physical write",
			"physical_read_write_number":    "Number of physical reads and writes",
			"pga_mem_alloc":                 "PGA memory allocated by process",
			"pga_freeable":                  "PGA memory freeable by process",
			"pga_mem_max":                   "PGA memory maximum allocated by process",
			"pga_mem_used":                  "PGA memory used by process",
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
			"oracle_table_place_title":      `Oracle 表空间不足告警`,
			"oracle_table_place_message":    `当前主机{{host}}的{{tablespace_name}}表空间不足，请及时关注：\n>主机名：{{host}}\n>表空间名：{{tablespace_name}}\n>当前使用率：{{df_monitor_checker_value}} %\n`,
			"oracle_active_session_title":   `Oracle 活跃会话数突变告警`,
			"oracle_active_session_message": `当前主机{{host}}的Oracle服务{{oracle_service}} 活跃会话数突变告警，请及时关注：\n>主机名：{{host}}\n>Oracle服务名：{{oracle_service}}\n>最近15分钟到与最近30分钟两个时间段内活跃会话数差值为：{{df_monitor_checker_value}}，请及时关注是否有异常。`,
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"oracle_table_place_title":      `Oracle Table Place Warning`,
			"oracle_table_place_message":    `Current host {{host}}'s {{tablespace_name}} tablespace is insufficient, please pay attention to it: \n>Host name: {{host}}\n>Tablespace name: {{tablespace_name}}\n>Current usage rate: {{df_monitor_checker_value}} %\n`,
			"oracle_active_session_title":   `Oracle Active Session Number Mutation Warning`,
			"oracle_active_session_message": `Current host {{host}}'s Oracle service {{oracle_service}} active session number mutation warning, please pay attention to it: \n>Host name: {{host}}\n>Oracle service name: {{oracle_service}}\n>The difference between the active session number in the last 15 minutes and thelast 30 minutes is: {{df_monitor_checker_value`,
		}
	default:
		return nil
	}
}

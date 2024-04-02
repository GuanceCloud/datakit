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

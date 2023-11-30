// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"otel_version":            "otel 版本",
			"host":                    "主机",
			"service_name":            "服务名",
			"process":                 "进程",
			"server_duration":         "请求用时",
			"start_time":              "启动时间",
			"ready_time":              "初始化使用时间",
			"pool_size":               "当前池数量",
			"pool_max":                "允许最大线程",
			"pool_core":               "核心线程",
			"pool_task":               "任务 pool",
			"queue_remaining":         "可接受元素数量",
			"queued":                  "排队中",
			"completed":               "已完成",
			"active":                  "正在执行",
			"sequence":                "任务队列",
			"promoted":                "老年代",
			"allocated":               "年轻代",
			"uptime":                  "运行时间",
			"threads_states":          "线程状态",
			"cpu_utilization":         "使用率",
			"cpu_load_1m":             "一分钟使用率",
			"threads_count":           "线程数量",
			"threads":                 "线程",
			"http_request":            "http 请求",
			"memory":                  "内存",
			"memory_init":             "初始大小",
			"memory_committed":        "申请大小",
			"memory_usage":            "进程使用内存大小",
			"buffer_count":            "缓存区数量",
			"buffer_limit":            "总缓存区",
			"buffer_usage":            "已用缓冲区内存",
			"jvm_buffer":              "进程  JVM buffer",
			"disk_total":              "磁盘总大小",
			"disk_free":               "空闲大小",
			"sessions_rejected":       "拒绝个数",
			"sessions_expired":        "已过期数",
			"sessions_created":        "已过期数",
			"sessions_alive_max":      "最多活跃数",
			"sessions_active_max":     "最大 session 数",
			"sessions_active_current": "活跃 session 数",
			"files_open":              "当前句柄数量",
			"files_max":               "允许最大句柄",
			"open_file":               "句柄",
			"load_average_1m":         "负载",
			"cpu_usage":               "cpu 使用率",
			"cpu_core":                "cpu 核心数",
			"threads_peak":            "峰值线程数",
			"threads_live":            "当前活跃数",
			"threads_daemon":          "当前活跃数",
			"memory_used":             "已用内存",
			"memory_max":              "最大内存",
			"gc_committed":            "JVM GC 分配空间",
			"gc_overhead":             "GC 使用 cpu",
			"gc_max_data_size":        "老年代最大空间",
			"gc_live_data_size":       "老年代内存空间",
			"classes_unloaded":        "未加载数量",
			"classes_loaded":          "已加载数量",
			"buffer_memory_used":      "jvm 已用内存",
			"buffer_total_capacity":   "缓冲器容量",
		}
	case inputs.I18nEn:
		return map[string]string{
			"otel_version":            "otel version",
			"host":                    "host",
			"service_name":            "service name",
			"process":                 "process",
			"server_duration":         "request duration",
			"start_time":              "process start time",
			"ready_time":              "process ready time",
			"pool_size":               "pool size",
			"pool_max":                "pool max",
			"pool_core":               "pool core",
			"pool_task":               "task pool",
			"queue_remaining":         "queue remaining",
			"queued":                  "queued",
			"completed":               "completed",
			"active":                  "active",
			"sequence":                "sequence",
			"promoted":                "promoted",
			"allocated":               "allocated",
			"uptime":                  "process uptime",
			"threads_states":          "threads states",
			"cpu_utilization":         "cpu utilization",
			"cpu_load_1m":             "cpu load 1m",
			"threads_count":           "threads count",
			"threads":                 "threads",
			"http_request":            "http request",
			"memory":                  "memory",
			"memory_init":             "memory init",
			"memory_committed":        "memory committed",
			"memory_usage":            "memory usage",
			"buffer_count":            "buffer count",
			"buffer_limit":            "buffer limit",
			"buffer_usage":            "buffer usage",
			"jvm_buffer":              "JVM buffer",
			"disk_total":              "disk total",
			"disk_free":               "disk free",
			"sessions_rejected":       "sessions rejected",
			"sessions_expired":        "sessions expired",
			"sessions_created":        "sessions created",
			"sessions_alive_max":      "sessions alive_max",
			"sessions_active_max":     "sessions active_max",
			"sessions_active_current": "sessions active_current",
			"files_open":              "open files",
			"files_max":               "files max",
			"open_file":               "open file",
			"load_average_1m":         "load average_1m",
			"cpu_usage":               "CPU usage",
			"cpu_core":                "CPU core",
			"threads_peak":            "threads peak",
			"threads_live":            "threads live",
			"threads_daemon":          "threads daemon",
			"memory_used":             "memory used",
			"memory_max":              "memory max",
			"gc_committed":            "JVM GC committed",
			"gc_overhead":             "GC overhead",
			"gc_max_data_size":        "GC max_data_size",
			"gc_live_data_size":       "GC live_data_size",
			"classes_unloaded":        "classes unloaded",
			"classes_loaded":          "classes loaded",
			"buffer_memory_used":      "buffer memory used",
			"buffer_total_capacity":   "buffer total capacity",
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

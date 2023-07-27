// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dk

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

//nolint:lll
func (i *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"alive_collectors":            "活跃采集器",
			"alive_goroutines":            "活跃 Goroutine",
			"alive_goroutines_note":       "总 Goroutines >= 所有模块 Goroutines 之和",
			"cpu_ctx_switch":              "CPU Context 切换",
			"cpu_ctx_switch_note":         "仅支持 Linux",
			"cpu_usage":                   "Datakit CPU%",
			"cpu_usage_note":              "值同 top 命令中的进程 CPU 占用",
			"data_usage":                  "工作空间数据用量",
			"data_usage_note":             "如果用量异常则无法上报数据",
			"diskcache_read_bytes":        "读字节数",
			"diskcache_read_count":        "读次数",
			"diskcache_write_bytes":       "写字节数",
			"diskcache_write_count":       "写次数",
			"dw_api_request":              "API 请求耗时",
			"dw_api_retry":                "API 重试",
			"dw_upload_bytes":             "上传字节数（gzip）",
			"dw_upload_pts":               "上传点数",
			"election":                    "选举情况",
			"filter_errors":               "规则错误",
			"filter_last_update_time":     "黑名单更新日期",
			"filter_pts":                  "过滤点数",
			"filter_update_count":         "黑名单刷新次数",
			"filter_update_letency":       "黑名单更新耗时",
			"group_blacklist":             "黑名单",
			"group_collectors":            "采集器",
			"group_diskcache":             "磁盘缓存",
			"group_overview":              "概览",
			"http_api_latency":            "API 耗时",
			"http_bad_api":                "非正常 API 请求",
			"http_client_dns_latency":     "DNS 耗时",
			"http_client_tcp_connections": "TCP 连接",
			"http_req_size":               "API 请求体大小",
			"intro_content":               `Datakit 所有指标的展示（Datakit >= 1.11.0），需要开启更多的指标收集功能。详情参见[这里](https://docs.guance.com/datakit/datakit-metrics/)暴露的指标。`,
			"mem_usage":                   "Datakit 内存占用",
			"mem_usage_note":              "来自 Golang runtime 的内存占用",
			"open_files":                  "打开文件数",
			"open_files_note":             "仅支持 Linux/Windows",
			"pl_cost_per_pt":              "耗时/点",
			"pl_pts":                      "处理点数",
			"pts_per_collect":             "每次采集点数",
			"rw_bytes":                    "Datakit 进程读写字节数",
			"rw_bytes_note":               "仅支持 Linux/Windows",
			"rw_count":                    "Datakit 进程读写次数",
			"rw_count_note":               "仅支持 Linux/Windows",
			"stopped_goroutine_cost":      "已停止 Goroutines 耗时",
			"stopped_goroutines":          "已停止 Goroutines",
			"top_n_cpu_usage":             "Top(n) CPU%",
			"top_n_mem_usage":             "Top(n) 内存占用",
			"uptime":                      "启动时长",
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
			"alive_collectors":            "Alive Collectors",
			"alive_goroutines":            "Alive Goroutines",
			"alive_goroutines_note":       "Total goroutines >= Sum of all module goroutines",
			"cpu_ctx_switch":              "CPU Context Switch",
			"cpu_ctx_switch_note":         "Only for Linux",
			"cpu_usage":                   "Datakit CPU%",
			"cpu_usage_note":              "Same as process CPU usage in top command",
			"data_usage":                  "Workspace Data Usage",
			"data_usage_note":             "If not OK points will upload fail",
			"diskcache_read_bytes":        "Read Bytes",
			"diskcache_read_count":        "Read Count",
			"diskcache_write_bytes":       "Write Bytes",
			"diskcache_write_count":       "Write Count",
			"dw_api_request":              "API Latency",
			"dw_api_retry":                "API Retried",
			"dw_upload_bytes":             "Upload Bytes(gzip)",
			"dw_upload_pts":               "Uploaded Points",
			"election":                    "Election",
			"filter_errors":               "Rule Errors",
			"filter_last_update_time":     "Rules Update Date",
			"filter_pts":                  "Points",
			"filter_update_count":         "Rule Refresh Count",
			"filter_update_letency":       "Rule Pull Latency",
			"group_blacklist":             "Black List",
			"group_collectors":            "Collectors",
			"group_diskcache":             "Disk Cache",
			"group_overview":              "Overview",
			"http_api_latency":            "API Cost",
			"http_bad_api":                "Bad API Request",
			"http_client_dns_latency":     "DNS Cost",
			"http_client_tcp_connections": "TCP Connections",
			"http_req_size":               "API Request Size",
			"intro_content":               `This dashboard showing all metrics about Datakit(require Datakit >= 1.11.0). For more metrics, see [here](https://docs.guance.com/en/datakit/datakit-metrics/).`,
			"mem_usage":                   "Datakit Memory Usage",
			"mem_usage_note":              "From Golang runtime",
			"open_files":                  "Open Files",
			"open_files_note":             "Only for Linux/Windows",
			"pl_cost_per_pt":              "Cost/Point",
			"pl_pts":                      "Points",
			"pts_per_collect":             "Points/Time",
			"rw_bytes":                    "Datakit Read Bytes",
			"rw_bytes_note":               "Only for Linux/Windows",
			"rw_count":                    "Datakit Read/Write Count",
			"rw_count_note":               "Only for Linux/Windows",
			"stopped_goroutine_cost":      "Stopped Goroutines Cost",
			"stopped_goroutines":          "Stopped Goroutines",
			"top_n_cpu_usage":             "Top(n) CPU%",
			"top_n_mem_usage":             "Top(n) Memory Usage",
			"uptime":                      "Uptime",
		}
	default:
		return nil
	}
}

func (i *Input) DashboardList() []string {
	return nil
}

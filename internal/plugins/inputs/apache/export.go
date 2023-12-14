// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package apache

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"title":                  "Apache 监控视图",
			"host_name":              "主机名",
			"open_slot":              "空闲槽",
			"waiting_for_connection": "正在等待连接",
			"sending_reply":          "正在发送回复",
			"closing_connection":     "正在关闭连接",
			"reading_request":        "正在读取请求",
			"dns_lookup":             "查找 DNS",
			"gracefully_finishing":   "Gracefully 完成",
			"keepalive":              "Keepalive 机制",
			"logging":                "日志记录",
			"idle_cleanup":           "闲置清理",
			"starting_up":            "启动中",
			"cpu_load":               "CPU 负载",
			"workers":                "工作线程",
			"network_request_bytes":  "网络请求字节数/s",
			"network_request_count":  "网络请求数/s",
			"score_board":            "状态板",
			"general":                "通用指标",
		}
	case inputs.I18nEn:
		return map[string]string{
			"title":                  "Apache Monitor View",
			"host_name":              "Host Name",
			"open_slot":              "Open Slot",
			"waiting_for_connection": "Waiting for Connection",
			"sending_reply":          "Sending Reply",
			"closing_connection":     "Closing Connection",
			"reading_request":        "Reading Request",
			"dns_lookup":             "DNS Lookup",
			"gracefully_finishing":   "Gracefully Finishing",
			"keepalive":              "Keepalive",
			"logging":                "Logging",
			"idle_cleanup":           "Idle Cleanup",
			"starting_up":            "Starting Up",
			"cpu_load":               "CPU Load",
			"workers":                "Workers",
			"network_request_bytes":  "Network Request Bytes/s",
			"network_request_count":  "Network Request Count/s",
			"score_board":            "Score Board",
			"general":                "General",
		}
	default:
		return nil
	}
}

func (*Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"title":           "Apache 5XX 错误数过多",
			"monitorName":     "默认",
			"level":           "等级",
			"host":            "主机",
			"content":         "内容",
			"content_info":    "Apache 5XX 错误数为",
			"suggestion":      "建议",
			"suggestion_info": "检查 Apache 日志查看详细信息",
		}
	case inputs.I18nEn:
		return map[string]string{
			"title":           "Apache has too many 5XX Errors",
			"monitorName":     "Default",
			"level":           "Level",
			"host":            "Host",
			"content":         "Content",
			"content_info":    "Apache 5XX error count is",
			"suggestion":      "Suggestion",
			"suggestion_info": "Check Apache log for detail information",
		}
	default:
		return nil
	}
}

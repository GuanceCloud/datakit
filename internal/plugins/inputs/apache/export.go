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
			"title":                 "Apache 监控视图",
			"host_name":             "主机名",
			openSlot:                "空闲槽",
			waitingForConnection:    "正在等待连接",
			sendingReply:            "正在发送回复",
			closingConnection:       "正在关闭连接",
			readingRequest:          "正在读取请求",
			dnsLookup:               "查找 DNS",
			gracefullyFinishing:     "Gracefully 完成",
			keepAlive:               "Keepalive 机制",
			logging:                 "日志记录",
			idleCleanup:             "闲置清理",
			startingUp:              "启动中",
			cpuLoad:                 "CPU 负载",
			"workers":               "工作线程",
			"network_request_bytes": "网络请求字节数/s",
			"network_request_count": "网络请求数/s",
			"score_board":           "状态板",
			"general":               "通用指标",
		}
	case inputs.I18nEn:
		return map[string]string{
			"title":                 "Apache Monitor View",
			"host_name":             "Host Name",
			openSlot:                "Open Slot",
			waitingForConnection:    "Waiting for Connection",
			sendingReply:            "Sending Reply",
			closingConnection:       "Closing Connection",
			readingRequest:          "Reading Request",
			dnsLookup:               "DNS Lookup",
			gracefullyFinishing:     "Gracefully Finishing",
			keepAlive:               "Keepalive",
			logging:                 "Logging",
			idleCleanup:             "Idle Cleanup",
			startingUp:              "Starting Up",
			cpuLoad:                 "CPU Load",
			"workers":               "Workers",
			"network_request_bytes": "Network Request Bytes/s",
			"network_request_count": "Network Request Count/s",
			"score_board":           "Score Board",
			"general":               "General",
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

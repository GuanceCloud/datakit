// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nginx

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"active_per_sec":  "当前活跃连接数/s",
			"handled_per_sec": "处理连接总数/s",
			"reading_per_sec": "正在读取的连接数/s",
			"request_per_sec": "客户端请求总数/s",
			"waiting_per_sec": "正在等待的连接数/s",
			"writing_per_sec": "正在写入的连接数/s",
			"accepts_per_sec": "接受连接总数/s",
			"host_name":       "主机名",
			"title":           "Nginx 监控视图",
		}
	case inputs.I18nEn:
		return map[string]string{
			"active_per_sec":  "Active/s",
			"handled_per_sec": "Handled/s",
			"reading_per_sec": "Reading/s",
			"request_per_sec": "Requests/s",
			"waiting_per_sec": "Waiting/s",
			"writing_per_sec": "Writing/s",
			"accepts_per_sec": "Accepts/s",
			"host_name":       "Host name",
			"title":           "Nginx Dashboard",
		}
	default:
		return nil
	}
}

func (*Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"title":           "Nginx 5XX 错误数过多",
			"level":           "等级",
			"hostname":        "主机",
			"content":         "内容",
			"content_info":    "Nginx 5XX 错误数为",
			"suggestion":      "建议",
			"suggestion_info": "检查 Nginx 日志查看详细信息",
			"monitorName":     "默认",
		}
	case inputs.I18nEn:
		return map[string]string{
			"title":           "Too many Nginx 5XX errors",
			"level":           "Level",
			"hostname":        "Host name",
			"content":         "Content",
			"content_info":    "The number of Nginx 5XX Error is",
			"suggestion":      "Suggestion",
			"suggestion_info": "Please check Nginx log for detail information",
			"monitorName":     "Default",
		}
	default:
		return nil
	}
}

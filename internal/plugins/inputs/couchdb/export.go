// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package couchdb

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

//nolint:lll
func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"var_host":                      "主机",
			"var_instance":                  "实例",
			"group_overview":                "概览",
			"group_db_request":              "数据库请求",
			"group_replicator":              "复制器状态",
			"httpd_requests_total":          "请求数 HTTP requests/s",
			"request_time_seconds":          "请求耗时",
			"database_reads_total":          "数据库读(Reads/s)",
			"httpd_bulk_requests":           "批量请求数(requests/s)",
			"httpd_status_codes":            "HTTP 响应状态码趋势",
			"httpd_request_methods":         "HTTP 分类型请求数",
			"httpd_request_methods_pie":     "HTTP 分类型请求数分布",
			"database_writes_total":         "数据库写(Writes/s)",
			"httpd_status_codes_pie":        "HTTP 响应状态码分布",
			"couch_replicator_jobs_running": "复制器调度程序中运行的作业数量",
			"couch_replicator_jobs_pending": "复制器调度程序中待处理的作业数量",
			"couch_replicator_jobs_crashed": "复制器调度程序中崩溃的作业数量",
			"auth_cache_requests_total":     "身份验证缓存请求数",
			"auth_cache_misses_total":       "身份验证缓存未命中数",
			"alias_instance":                "实例",
			"alias_host":                    "所在主机",
			"alias_online":                  "在线时长",
			"instance_overview":             "实例概览",
		}
	case inputs.I18nEn:
		return map[string]string{
			"var_host":                      "Host",
			"var_instance":                  "Instance",
			"group_overview":                "Overview",
			"group_db_request":              "Database Request",
			"group_replicator":              "Replicator Status",
			"httpd_requests_total":          "HTTP requests/s",
			"request_time_seconds":          "Request Time",
			"database_reads_total":          "Database Read(Reads/s)",
			"httpd_bulk_requests":           "Bulk Requests/s",
			"httpd_status_codes":            "HTTP Status Codes",
			"httpd_request_methods":         "HTTP Request Methods",
			"httpd_request_methods_pie":     "HTTP Request Methods trend",
			"database_writes_total":         "Database Write(Writes/s)",
			"httpd_status_codes_pie":        "HTTP Status Codes trend",
			"couch_replicator_jobs_running": "Couch Replicator Jobs Running",
			"couch_replicator_jobs_pending": "Couch Replicator Jobs Pending",
			"couch_replicator_jobs_crashed": "Couch Replicator Jobs Crashed",
			"auth_cache_requests_total":     "Auth Cache Requests",
			"auth_cache_misses_total":       "Auth Cache Misses",
			"alias_instance":                "Instance",
			"alias_host":                    "Host",
			"alias_online":                  "Online Time",
			"instance_overview":             "Instance Overview",
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
			"httpd_status_codes_4xx_title":   `CouchDB 请求 4xx 错误率异常告警`,
			"httpd_status_codes_4xx_message": `> 告警等级：{{status}}\n> CouchDB 服务地址: 主机 {{host}}，实例 {{instance}}\n> 服务错误率：{{ Result | to_fixed(2) }}%\n> 建议：请求 4xx 错误率等于单位时间的 4xx 错误请求数(例如 '401 Unauthorized') 除以 总请求数。如果错误率过高，请关注。`,
			"httpd_status_codes_5xx_title":   `CouchDB 请求 5xx 错误率异常告警`,
			"httpd_status_codes_5xx_message": `> 告警等级：{{status}}\n> CouchDB 服务地址: 主机 {{host}}，实例 {{instance}}\n> 服务错误率：{{ Result | to_fixed(2) }}%\n> 建议：请求 5xx 错误率等于单位时间的 5xx 错误请求数(例如 '502 Bad Gateway') 除以 总请求数。如果错误率过高，请关注。`,
			"couchdb_p90_title":              `CouchDB P90 响应时长过高`,
			"couchdb_p90_message":            `> 告警等级：{{status}}\n> CouchDB 服务地址: 主机 {{host}}，实例 {{instance}}\n> P90 响应时长：{{Result}}ms\n> 建议：请求响应时长过高，请关注。`,
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"httpd_status_codes_4xx_title":   `CouchDB Request 4xx Error Rate Alert`,
			"httpd_status_codes_4xx_message": `> Alert Level: {{status}}\n> CouchDB Service Address: Host {{host}}, Instance {{instance}}\n> Service Error Rate: {{ Result | to_fixed(2) }}%\n> Suggestion: The request 4xx error rate equals the number of 4xx error requests (e.g., '401 Unauthorized') per unit time divided by the total number of requests. If the error rate is too high, please pay attention.`,
			"httpd_status_codes_5xx_title":   `CouchDB Request 5xx Error Rate Alert`,
			"httpd_status_codes_5xx_message": `> Alert Level: {{status}}\n> CouchDB Service Address: Host {{host}}, Instance {{instance}}\n> Service Error Rate: {{ Result | to_fixed(2) }}%\n> Suggestion: The request 5xx error rate equals the number of 5xx error requests (e.g., '502 Bad Gateway') per unit time divided by the total number of requests. If the error rate is too high, please pay attention.`,
			"couchdb_p90_title":              `CouchDB P90 Response Time Too High`,
			"couchdb_p90_message":            `> Alert Level: {{status}}\n> CouchDB Service Address: Host {{host}}, Instance {{instance}}\n> P90 Response Time: {{Result}}ms\n> Suggestion: The request response time is too high, please investigate.`,
		}
	default:
		return nil
	}
}

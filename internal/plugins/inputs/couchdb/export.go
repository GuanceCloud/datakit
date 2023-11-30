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
			"uptime_seconds_note":               "正常运行时间",
			"uptime_seconds":                    "正常运行时间",
			"erlang_memory_bytes_note":          "Erlang内存大小",
			"erlang_memory_bytes":               "Erlang内存大小",
			"httpd_request_methods_note":        "HTTP 分类型请求数量",
			"httpd_request_methods":             "HTTP 分类型请求数",
			"open_databases_note":               "打开的数据库数量",
			"open_databases":                    "数据库数量",
			"open_os_files_note":                "打开的文件数量",
			"open_os_files":                     "文件数量",
			"auth_cache_hits_note":              "身份验证缓存命中数",
			"auth_cache_hits":                   "缓存命中数",
			"auth_cache_misses_note":            "身份验证缓存未命中数",
			"auth_cache_misses":                 "缓存未命中数",
			"httpd_requests_note":               "HTTP 请求数",
			"httpd_requests":                    "HTTP 请求数",
			"httpd_bulk_requests_note":          "批量请求数",
			"httpd_bulk_requests":               "批量请求数",
			"collect_results_time_seconds_note": "呼叫延迟",
			"collect_results_time_seconds":      "呼叫延迟",
			"database_reads_writes_note":        "数据库读次数<br>数据库写次数<br>数据库被清除次数",
			"database_reads_writes":             "数据库读写",
			"httpd_status_codes_note":           "HTTP 分响应状态码统计次数",
			"httpd_status_codes":                "HTTP 响应状态码",
		}
	case inputs.I18nEn:
		return map[string]string{
			"uptime_seconds_note":               "CouchDB uptime",
			"uptime_seconds":                    "CouchDB uptime",
			"erlang_memory_bytes_note":          "Size of memory dynamically allocated by the Erlang emulator",
			"erlang_memory_bytes":               "Erlang memory",
			"httpd_request_methods_note":        "Number of HTTP option requests",
			"httpd_request_methods":             "HTTP option requests",
			"open_databases_note":               "Number of open databases",
			"open_databases":                    "Open databases",
			"open_os_files_note":                "Number of file descriptors CouchDB has open",
			"open_os_files":                     "Open files",
			"auth_cache_hits_note":              "Number of authentication cache hits",
			"auth_cache_hits":                   "Auth cache hits",
			"auth_cache_misses_note":            "Number of authentication cache misses",
			"auth_cache_misses":                 "Auth cache misses",
			"httpd_requests_note":               "HTTP requests",
			"httpd_requests":                    "HTTP requests",
			"httpd_bulk_requests_note":          "Number of bulk requests",
			"httpd_bulk_requests":               "Bulk requests",
			"collect_results_time_seconds_note": "Microsecond latency for calls to couch_db:collect_results",
			"collect_results_time_seconds":      "Latency for calls",
			"database_reads_writes_note":        "DB read times<br>DB write times<br>DB was purged times",
			"database_reads_writes":             "DB reads and writes",
			"httpd_status_codes_note":           "Number of HTTP status_codes responses",
			"httpd_status_codes":                "HTTP status_codes responses",
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

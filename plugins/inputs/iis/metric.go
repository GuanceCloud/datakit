package iis

var (
	PerfObjMetricMap = map[string]map[string]map[string]string{
		"iis_web_service": {
			"Web Service": {
				"Service Uptime":                            "service_uptime",
				"Bytes Sent/sec":                            "bytes_sent",
				"Bytes Received/sec":                        "bytes_received",
				"Bytes Total/sec":                           "bytes_total",
				"Current Connections":                       "current_connections",
				"Files Sent/sec":                            "files_sent",
				"Files Received/sec":                        "files_received",
				"Total Connection Attempts (all instances)": "total_connection_attempts",
				"Connection Attempts/sec":                   "connection_attempts",
				"Get Requests/sec":                          "http_requests_get",
				"Post Requests/sec":                         "http_requests_post",
				"Head Requests/sec":                         "http_requests_head",
				"Put Requests/sec":                          "http_requests_put",
				"Delete Requests/sec":                       "http_requests_delete",
				"Options Requests/sec":                      "http_requests_options",
				"Trace Requests/sec":                        "http_requests_trace",
				"Not Found Errors/sec":                      "error_not_found",
				"Locked Errors/sec":                         "error_locked",
				"Anonymous Users/sec":                       "anonymous_users",
				"NonAnonymous Users/sec":                    "non_anonymous_users",
				"CGI Requests/sec":                          "requests_cgi",
				"ISAPI Extension Requests/sec":              "requests_isapi_extension",
			},
		},
		"iis_app_pool_was": {
			"APP_POOL_WAS": {
				"Current Application Pool Uptime": "current_app_pool_uptime",
				"Current Application Pool State":  "current_app_pool_state",
				"Total Application Pool Recycles": "total_app_pool_recycles",
			},
		},
	}
)

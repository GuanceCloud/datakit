// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package coredns

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (i *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"var_cluster":                      "集群",
			"group_global_stats":               "全局统计",
			"group_local":                      "本地",
			"group_upstream":                   "上游",
			"request_by_instance":              "请求 (按实例)",
			"upstream_health_check_fails":      "上游健康检查失败",
			"panics":                           "Panics",
			"failed_reloads":                   "失败的重新加载",
			"cpu_time":                         "CPU 时间",
			"request_total":                    "请求 (总计)",
			"request_by_zone":                  "请求 (按区域)",
			"responses_zone":                   "响应 (延迟，互联网区域)",
			"request_by_type":                  "请求 (按类型)",
			"cache_hit_rate":                   "缓存 (命中率)",
			"request_dnssec_by_zone":           "请求 (DNSSEC 按区域)",
			"responses_by_code":                "响应 (按代码)",
			"requests_size_zone":               "请求 (大小，互联网区域)",
			"response_size_zone":               "响应 (大小，互联网区域)",
			"cache_size":                       "缓存 (大小)",
			"responses_latency":                "响应 (延迟)",
			"requests_by_upstream":             "请求 (按上游)",
			"responses_by_upstream":            "响应 (按上游)",
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"var_cluster":                      "Cluster",
			"group_global_stats":               "Global Stats",
			"group_local":                      "Local",
			"group_upstream":                   "Upstream",
			"request_by_instance":              "Requests (by instance)",
			"upstream_health_check_fails":      "Upstream health check fails",
			"panics":                           "Panics",
			"failed_reloads":                   "Failed reloads",
			"cpu_time":                         "CPU Time",
			"request_total":                    "Requests (total)",
			"request_by_zone":                  "Requests (by zone)",
			"responses_zone":                   "Responses (latency, internet zone)",
			"request_by_type":                  "Requests (by type)",
			"cache_hit_rate":                   "Cache (hitrate)",
			"request_dnssec_by_zone":           "Requests (DNSSEC by zone)",
			"responses_by_code":                "Responses (by code)",
			"requests_size_zone":               "Requests (size, internet zone)",
			"response_size_zone":               "Responses (size, internet zone)",
			"cache_size":                       "Cache (size)",
			"responses_latency":                "Responses (latency)",
			"requests_by_upstream":             "Requests (by upstream)",

		}
	default:
		return nil
	}
}

func (i *Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"default_monitor_name":                          "默认",
			"coredns_response_time_message":                 `coreDNS {{ instance }} 响应时间过高\n当前平均响应时间：{{Result}}\ncoreDNS负载过大，请调整优化相关资源限制或检查异常请求。`,
			"coredns_response_time_title":                   "coreDNS {{ instance }} 响应时间过高",
			"coredns_cache_hit_title":                       "coreDNS {{instance}} 缓存命中率异常",
			"coredns_cache_hit_message":                     `coreDNS {{instance}} 缓存命中率异常\n当前平均命中率：{{Result}}\n缓存命中率较低，可能导致上游DNS负载过高。请检查是否存在异常请求，并优化网络和缓存策略。`,
            "coredns_forward_request_duration_title":        "coreDNS {{instance}} 上游响应延迟异常",
            "coredns_forward_request_duration_message":      `coreDNS {{instance}} 上游响应延迟异常\n平均响应耗时：{{Result}}\n请检查与上游DNS网络情况或上游DNS服务健康状态`,
            "coredns_response_error_title":                  "coreDNS {{instance}} 大量错误响应",
            "coredns_response_error_message":                `coreDNS {{instance}} 大量错误响应\n错误码：{{rcode}}\n出现次数：{{Result}}\ncoreDNS错误可能导致pod访问受限，请检查相关错误日志。`,


		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"default_monitor_name":                          "Default",
			"coredns_response_time_message":                 `coreDNS {{ instance }} response time is too high\nCurrent average response time: {{Result}}\ncoreDNS load is too high, please adjust resource limits or check for abnormal requests.`,
			"coredns_response_time_title":                   "coreDNS {{ instance }} response time is too high",
			"coredns_cache_hit_title":                       "coreDNS {{instance}} cache hit rate is abnormal",
			"coredns_cache_hit_message":                     `coreDNS {{instance}} cache hit rate is abnormal\nCurrent average hit rate: {{Result}}\nCache hit rate is low, which may cause upstream DNS load to be too high. Please check for any abnormal requests and optimize network and cache strategies.`,
			"coredns_forward_request_duration_title":        "coreDNS {{instance}} upstream response delay is abnormal",
			"coredns_forward_request_duration_message":      `coreDNS {{instance}} upstream response delay is abnormal\nAverage response delay: {{Result}}\nPlease check the network status with the upstream DNS or the health status of the upstream DNS service.`,
			"coredns_response_error_title":                  "coreDNS {{instance}} has many error responses",
			"coredns_response_error_message":                `coreDNS {{instance}} has many error responses\nError code: {{rcode}}\nOccurrence: {{Result}}\ncoreDNS errors may cause pod access restrictions, please check the error logs.`,

		}
	default:
		return nil
	}
}

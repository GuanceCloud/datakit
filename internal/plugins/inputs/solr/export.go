// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package solr

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"title":                   "Solr 监控视图",
			"introduction":            "简介",
			"description":             "说明",
			"description_content":     `采集器可以从 Solr 实例中采取很多指标，比如cache、request数等多种指标，并将指标采集到观测云，帮助监控分析 Solr 各种异常情况。\n\nhttps://docs.guance.com/datakit/solr/`,
			"metric":                  "指标",
			"deleted_document_number": "删除文档数",
			"document_number":         "文档数",
			"maximum_document_number": "最大文档数",
			"request_overview":        "请求总览",
			"fifteen_min_request":     "15分钟请求总和",
			"five_min_request":        "5分钟请求总和",
			"one_min_request":         "1分钟请求总和",
			"index_cache_hit_number":  "索引缓存命中数",
			"insertion_cache_number":  "插入缓存数",
			"cache_lookup_number":     "缓存查找数",
			"request_count":           "请求总数",
			"request_rate":            "请求速率",
			"request_p99":             "P95 请求处理时间",
			"request_p95":             "P95 请求处理时间",
			"request_p75":             "P75 请求处理时间",
			"index_cache_hit_rate":    "索引缓存命中率",
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"title":                   "Solr Monitoring View",
			"introduction":            "Introduction",
			"description":             "Description",
			"description_content":     `The collector can take many metrics from the Solr instance, such as cache, request number and other metrics, and collect the metrics to the Guance Cloud to help monitor and analyze various abnormal situations of Solr.\n\nhttps://docs.guance.com/datakit/solr/`,
			"metric":                  "Metric",
			"deleted_document_number": "Number of deleted documents",
			"document_number":         "Number of documents",
			"maximum_document_number": "Maximum number of documents",
			"request_overview":        "Request Overview",
			"fifteen_min_request":     "15-minute requests sum",
			"five_min_request":        "5-minute requests sum",
			"one_min_request":         "1-minute requests sum",
			"index_cache_hit_number":  "Number of index cache hits",
			"insertion_cache_number":  "Number of insertion caches",
			"cache_lookup_number":     "Number of cache lookups",
			"request_count":           "Number of requests",
			"request_rate":            "Request rate",
			"request_p99":             "P95 request processing time",
			"request_p95":             "P95 request processing time",
			"request_p75":             "P75 request processing time",
			"index_cache_hit_rate":    "Index cache hit rate",
		}
	default:
		return nil
	}
}

func (ipt *Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"cache_hit_rate_title":   `Solr 实例{{core}} 缓存命中率过低告警`,
			"cache_hit_rate_message": `>实例名称：{{core}}\n>所在主机：{{host}}\n>缓存命中率为：{{df_monitor_checker_value * 100}} %\n>Solr缓存命中率过低，请尽快优化缓存配置。`,
			"request_p95_title":      `Solr 实例{{core}} P95请求响应时间告警`,
			"request_p95_message":    `>实例名称：{{core}}\n>所在主机：{{host}}\n>缓存命中率为：{{df_monitor_checker_value }} ms\n>SolrP95请求响应时间过高，系统性能可能存在问题，请尽快排查。`,
			"request_avg_title":      `Solr 实例{{core}} 所有请求处理平均时间告警`,
			"request_avg_message":    `>实例名称：{{core}}\n>所在主机：{{host}}\n>缓存命中率为：{{df_monitor_checker_value }} ms\n>Solr所有请求处理平均时间过高，系统性能可能存在问题，请尽快排查。`,
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"cache_hit_rate_title":   `Solr instance {{core}} cache hit rate is low alarm`,
			"cache_hit_rate_message": `>Instance name: {{core}}\n>Host: {{host}}\n>Cache hit rate: {{df_monitor_checker_value * 100}} %\n>Solr cache hit rate is low, please optimize the cache configuration as soon as possible.`,
			"request_p95_title":      `Solr instance {{core}} P95 request response time alarm`,
			"request_p95_message":    `>Instance name: {{core}}\n>Host: {{host}}\n>Cache hit rate: {{df_monitor_checker_value }} ms\n>Solr P95 request response time is high, which may affect the system performance. Please check as soon as possible.`,
			"request_avg_title":      `Solr instance {{core}} all request processing average time alarm`,
			"request_avg_message":    `>Instance name: {{core}}\n>Host: {{host}}\n>Cache hit rate: {{df_monitor_checker_value }} ms\n>Solr all request processing average time is high, which may affect the system performance. Please check as soon as possible.`,
		}
	default:
		return nil
	}
}

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
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"introduction":            "Introduction",
			"description":             "Description",
			"description_content":     `The collector can take many metrics from the Solr instance, such as cache, request number and other metrics, and collect the metrics to the observation cloud to help monitor and analyze various abnormal situations of Solr.\n\nhttps://docs.guance.com/datakit/solr/`,
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
			"message":              `>等级：{{df_status}}  \n>事件：{{ df_dimension_tags }}\n>监控器：{{ df_monitor_checker_name }}\n>告警策略：{{ df_monitor_name }}\n>事件状态： {{ df_status }}\n>内容：slor 15分钟请求消息数过高\n>建议：登录集群查看是否有异常`,
			"title":                "slor 15分钟请求消息数过高",
			"default_monitor_name": "默认",
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"message":              `>Level: {{df_status}}  \n>Event: {{ df_dimension_tags }}\n>Monitor: {{ df_monitor_checker_name }}\n>Alarm policy: {{ df_monitor_name }}\n>Event status: {{ df_status }}\n>Content: slor 15-minute request message count is too high\n>Suggestion: Log in to the cluster to see if there are any abnormalities`,
			"title":                "The number of slor 15-minute request messages is too high",
			"default_monitor_name": "Default",
		}
	default:
		return nil
	}
}

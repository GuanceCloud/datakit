// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ddtrace handle DDTrace APM traces.
package ddtrace

import (
	"encoding/json"
	"strconv"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
)

var _ inputs.ReadEnv = &Input{}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Endpoints", Type: doc.JSON, Example: "`[\"/v0.3/traces\", \"/v0.4/traces\", \"/v0.5/traces\"]`", Desc: "Agent endpoints", DescZh: "代理端点"},
		{FieldName: "CustomerTags", Type: doc.JSON, Example: "`[\"sink_project\", \"custom_dd_tag\"]`", Desc: "Whitelist to tags", DescZh: "标签白名单"},
		{FieldName: "KeepRareResource", Type: doc.Boolean, Default: `false`, Desc: "Keep rare tracing resources list switch", DescZh: "保持稀有跟踪资源列表"},
		{FieldName: "CompatibleOTEL", ENVName: "COMPATIBLE_OTEL", Type: doc.Boolean, Default: `false`, Desc: "Compatible `OTEL Trace` with `DDTrace trace`", DescZh: "将 `otel Trace` 与 `DDTrace Trace` 兼容"},
		{FieldName: "TraceID64BitHex", ENVName: "TRACE_ID_64_BIT_HEX", Type: doc.Boolean, Default: `false`, Desc: "Compatible `B3/B3Multi TraceID` with `DDTrace`", DescZh: "将 `B3/B3Multi-TraceID` 与 `DDTrace` 兼容"},
		{FieldName: "Trace128BitID", ENVName: "TRACE_128_BIT_ID", Type: doc.Boolean, Default: `true`, Desc: "Trace IDs as 32 lowercase hexadecimal", DescZh: "将链路 ID 转成长度为 32 的 16 进制编码的字符串"},
		{FieldName: "DelMessage", Type: doc.Boolean, Default: `false`, Desc: "Delete trace message", DescZh: "删除 trace 消息"},
		{FieldName: "TracingMetricEnable", Type: doc.Boolean, Default: `true`, Desc: "These metrics capture request counts, error counts, and latency measures.", DescZh: "开启请求计数，错误计数和延迟指标的采集"},
		{FieldName: "TracingMetricTagBlacklist", Type: doc.JSON, Example: "`[\"tag_a\", \"tag_b\"]`", Desc: "Blacklist of tags in the metric: \"tracing_metrics\"", DescZh: "指标集 tracing_metrics 中标签的黑名单"},
		{FieldName: "OmitErrStatus", Type: doc.JSON, Example: "`[\"404\", \"403\", \"400\"]`", Desc: "Whitelist to error status", DescZh: "错误状态白名单"},
		{FieldName: "CloseResource", Type: doc.JSON, Example: "`{\"service1\":[\"resource1\",\"other\"],\"service2\":[\"resource2\",\"other\"]}`", Desc: "Ignore tracing resources that service (regular)", DescZh: "忽略指定服务器的 tracing（正则匹配）"},
		{FieldName: "Sampler", Type: doc.Float, Example: `0.3`, Desc: "Global sampling rate", DescZh: "全局采样率"},
		{FieldName: "WPConfig", ENVName: "THREADS", Type: doc.JSON, Example: "`{\"buffer\":1000, \"threads\":100}`", Desc: "Total number of threads and buffer", DescZh: "线程和缓存的数量"},
		{FieldName: "LocalCacheConfig", ENVName: "STORAGE", Type: doc.JSON, Example: "`{\"storage\":\"./ddtrace_storage\", \"capacity\": 5120}`", Desc: "Local cache file path and size (MB) ", DescZh: "本地缓存路径和大小（MB）"},
		{FieldName: "Tags", Type: doc.JSON, Example: "`{\"k1\":\"v1\", \"k2\":\"v2\", \"k3\":\"v3\"}`"},

		{
			FieldName: "TraceMaxSpans",
			ENVName:   "ENV_INPUT_DDTRACE_MAX_SPANS",
			Type:      doc.Int,
			Example:   `1000`,
			Default:   "100000",
			Desc:      "Max spans of single trace. Set to -1 to remove this limit",
			DescZh:    "单个 trace 最大 span 个数，如果超过该限制，多余的 span 将截断，置为 -1 可关闭该限制",
		},

		{
			FieldName: "MaxTraceBodyMB",
			ENVName:   "ENV_INPUT_DDTRACE_MAX_BODY_MB",
			Type:      doc.JSON,
			Example:   `32`,
			Default:   "10",
			Desc:      "Max body(in MiB) of single trace POST. Set to -1 to remove this limit",
			DescZh:    "单个 trace API 请求最大 body 字节数（单位 MiB），置为 -1 可关闭该限制",
		},
	}

	return doc.SetENVDoc("ENV_INPUT_DDTRACE_", infos)
}

// ReadEnv load config from environment values
// ENV_INPUT_DDTRACE_ENDPOINTS : JSON string
// ENV_INPUT_DDTRACE_IGNORE_TAGS : JSON string
// ENV_INPUT_DDTRACE_COMPATIBLE_OTEL : bool
// ENV_INPUT_DDTRACE_TRACE_ID_64_BIT_HEX : bool
// ENV_INPUT_DDTRACE_KEEP_RARE_RESOURCE : bool
// ENV_INPUT_DDTRACE_OMIT_ERR_STATUS : JSON string
// ENV_INPUT_DDTRACE_CLOSE_RESOURCE : JSON string
// ENV_INPUT_DDTRACE_SAMPLER : float
// ENV_INPUT_DDTRACE_TAGS : JSON string
// ENV_INPUT_DDTRACE_THREADS : JSON string
// ENV_INPUT_DDTRACE_STORAGE : JSON string
// ENV_INPUT_DDTRACE_MAX_SPANS: int
// below is a complete example for env in shell
// export ENV_INPUT_DDTRACE_ENDPOINTS=`["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]`
// export ENV_INPUT_DDTRACE_IGNORE_TAGS=`["block1", "block2"]`
// export ENV_INPUT_DDTRACE_KEEP_RARE_RESOURCE=true
// export ENV_INPUT_DDTRACE_OMIT_ERR_STATUS=`["404", "403", "400"]`
// export ENV_INPUT_DDTRACE_CLOSE_RESOURCE=`{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}`
// export ENV_INPUT_DDTRACE_SAMPLER=0.3
// export ENV_INPUT_DDTRACE_TAGS=`{"k1":"v1", "k2":"v2", "k3":"v3"}`
// export ENV_INPUT_DDTRACE_THREADS=`{"buffer":1000, "threads":100}`
// export ENV_INPUT_DDTRACE_STORAGE=`{"storage":"./ddtrace_storage", "capacity": 5120}`.
func (ipt *Input) ReadEnv(envs map[string]string) {
	log = logger.SLogger(inputName)

	for _, key := range []string{
		"ENV_INPUT_DDTRACE_ENDPOINTS",
		"ENV_INPUT_DDTRACE_COMPATIBLE_OTEL",
		"ENV_INPUT_DDTRACE_CUSTOMER_TAGS",
		"ENV_INPUT_DDTRACE_KEEP_RARE_RESOURCE",
		"ENV_INPUT_DDTRACE_OMIT_ERR_STATUS",
		"ENV_INPUT_DDTRACE_CLOSE_RESOURCE",
		"ENV_INPUT_DDTRACE_SAMPLER",
		"ENV_INPUT_DDTRACE_TAGS",
		"ENV_INPUT_DDTRACE_THREADS",
		"ENV_INPUT_DDTRACE_STORAGE",
		"ENV_INPUT_DDTRACE_DEL_MESSAGE",
		"ENV_INPUT_DDTRACE_TRACE_ID_64_BIT_HEX",
		"ENV_INPUT_DDTRACE_MAX_SPANS",
		"ENV_INPUT_DDTRACE_MAX_BODY_MB",
		"ENV_INPUT_DDTRACE_TRACE_128_BIT_ID",
		"ENV_INPUT_DDTRACE_TRACING_METRIC_ENABLE",
		"ENV_INPUT_DDTRACE_TRACING_METRIC_TAG_BLACKLIST",
	} {
		value, ok := envs[key]
		if !ok {
			continue
		}
		switch key {
		case "ENV_INPUT_DDTRACE_MAX_SPANS":
			if n, err := strconv.ParseInt(value, 10, 64); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.TraceMaxSpans = int(n)
			}

		case "ENV_INPUT_DDTRACE_MAX_BODY_MB":
			if n, err := strconv.ParseInt(value, 10, 64); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.MaxTraceBodyMB = n
			}

		case "ENV_INPUT_DDTRACE_ENDPOINTS":
			var list []string
			if err := json.Unmarshal([]byte(value), &list); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.Endpoints = list
			}
		case "ENV_INPUT_DDTRACE_COMPATIBLE_OTEL":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.CompatibleOTEL = ok
			}
		case "ENV_INPUT_DDTRACE_TRACE_ID_64_BIT_HEX":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.TraceID64BitHex = ok
			}
		case "ENV_INPUT_DDTRACE_TRACE_128_BIT_ID":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.Trace128BitID = ok
			}

		case "ENV_INPUT_DDTRACE_CUSTOMER_TAGS":
			var list []string
			if err := json.Unmarshal([]byte(value), &list); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				log.Infof("customerTags from ENV key=%s value=%s", key, value)
				ipt.CustomerTags = list
			}
		case "ENV_INPUT_DDTRACE_KEEP_RARE_RESOURCE":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.KeepRareResource = ok
			}
		case "ENV_INPUT_DDTRACE_OMIT_ERR_STATUS":
			var list []string
			if err := json.Unmarshal([]byte(value), &list); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.OmitErrStatus = list
			}
		case "ENV_INPUT_DDTRACE_CLOSE_RESOURCE":
			var closeRes map[string][]string
			if err := json.Unmarshal([]byte(value), &closeRes); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.CloseResource = closeRes
			}
		case "ENV_INPUT_DDTRACE_SAMPLER":
			if ratio, err := strconv.ParseFloat(value, 64); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				if ipt.Sampler == nil {
					ipt.Sampler = &itrace.Sampler{}
				}
				ipt.Sampler.SamplingRateGlobal = ratio
			}
		case "ENV_INPUT_DDTRACE_TAGS":
			var tags map[string]string
			if err := json.Unmarshal([]byte(value), &tags); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.Tags = tags
			}
		case "ENV_INPUT_DDTRACE_THREADS":
			var threads workerpool.WorkerPoolConfig
			if err := json.Unmarshal([]byte(value), &threads); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.WPConfig = &threads
			}
		case "ENV_INPUT_DDTRACE_STORAGE":
			var storage storage.StorageConfig
			if err := json.Unmarshal([]byte(value), &storage); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.LocalCacheConfig = &storage
			}
		case "ENV_INPUT_DDTRACE_DEL_MESSAGE":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.DelMessage = ok
			}
		case "ENV_INPUT_DDTRACE_TRACING_METRIC_ENABLE":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.TracingMetricEnable = ok
			}
		case "ENV_INPUT_DDTRACE_TRACING_METRIC_TAG_BLACKLIST":
			var list []string
			if err := json.Unmarshal([]byte(value), &list); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.TracingMetricTagBlacklist = list
			}
		}
	}
}

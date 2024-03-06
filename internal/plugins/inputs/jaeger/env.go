// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package jaeger handle Jaeger tracing metrics.
package jaeger

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
		{FieldName: "Endpoint", ENVName: "HTTP_ENDPOINT", ConfField: "endpoint", Type: doc.String, Example: `/apis/traces`, Desc: "Endpoint for receiving tracing span over HTTP", DescZh: "通过 HTTP 接收 tracing span 的端点"},
		{FieldName: "Address", ENVName: "UDP_ENDPOINT", ConfField: "address", Type: doc.String, Example: `127.0.0.1:6831`, Desc: "Agent URL for UDP transport", DescZh: "UDP 代理 URL"},
		{FieldName: "IgnoreTags", Type: doc.JSON, Example: `["block1","block2"]`, Desc: "Ignore tags", DescZh: "忽略的标签"},
		{FieldName: "KeepRareResource", Type: doc.Boolean, Default: `false`, Desc: "Keep rare tracing resources list switch", DescZh: "保持稀有跟踪资源列表"},
		{FieldName: "DelMessage", Type: doc.Boolean, Default: `false`, Desc: "Delete trace message", DescZh: "删除 trace 消息"},
		{FieldName: "CloseResource", Type: doc.JSON, Example: `{"service1":["resource1","other"],"service2":["resource2","other"]}`, Desc: "Ignore tracing resources that service (regular)", DescZh: "忽略指定服务器的 tracing（正则匹配）"},
		{FieldName: "Sampler", Type: doc.Float, Example: `0.3`, Desc: "Global sampling rate", DescZh: "全局采样率"},
		{FieldName: "WPConfig", ENVName: "THREADS", Type: doc.JSON, Example: `{"buffer":1000, "threads":100}`, Desc: "Total number of threads and buffer", DescZh: "线程和缓存的数量"},
		{FieldName: "LocalCacheConfig", ENVName: "STORAGE", Type: doc.JSON, Example: `{"storage":"./jaeger_storage", "capacity": 5120}`, Desc: "Local cache file path and size (MB) ", DescZh: "本地缓存路径和大小（MB）"},
		{FieldName: "Tags", Type: doc.JSON, Example: `{"k1":"v1", "k2":"v2", "k3":"v3"}`},
	}

	return doc.SetENVDoc("ENV_INPUT_JAEGER_", infos)
}

// ReadEnv load config from environment values
// ENV_INPUT_JAEGER_HTTP_ENDPOINT : string
// ENV_INPUT_JAEGER_UDP_ENDPOINT : string
// ENV_INPUT_JAEGER_IGNORE_TAGS : JSON string
// ENV_INPUT_JAEGER_KEEP_RARE_RESOURCE : bool
// ENV_INPUT_JAEGER_CLOSE_RESOURCE : JSON string
// ENV_INPUT_JAEGER_SAMPLER : float
// ENV_INPUT_JAEGER_TAGS : JSON string
// ENV_INPUT_JAEGER_THREADS : JSON string
// ENV_INPUT_JAEGER_STORAGE : JSON string
// below is a complete example for env in shell
// export ENV_INPUT_JAEGER_HTTP_ENDPOINT="/apis/traces"
// export ENV_INPUT_JAEGER_UDP_ENDPOINT="127.0.0.1:6831"
// export ENV_INPUT_JAEGER_IGNORE_TAGS=`["block1", "block2"]`
// export ENV_INPUT_JAEGER_KEEP_RARE_RESOURCE=true
// export ENV_INPUT_JAEGER_CLOSE_RESOURCE=`{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}`
// export ENV_INPUT_JAEGER_SAMPLER=0.3
// export ENV_INPUT_JAEGER_TAGS=`{"k1":"v1", "k2":"v2", "k3":"v3"}`
// export ENV_INPUT_JAEGER_THREADS=`{"buffer":1000, "threads":100}`
// export ENV_INPUT_JAEGER_STORAGE=`{"storage":"./jaeger_storage", "capacity": 5120}`.
func (ipt *Input) ReadEnv(envs map[string]string) {
	log = logger.SLogger(inputName)

	for _, key := range []string{
		"ENV_INPUT_JAEGER_HTTP_ENDPOINT", "ENV_INPUT_JAEGER_UDP_ENDPOINT", "ENV_INPUT_JAEGER_IGNORE_TAGS",
		"ENV_INPUT_JAEGER_KEEP_RARE_RESOURCE", "ENV_INPUT_JAEGER_CLOSE_RESOURCE", "ENV_INPUT_JAEGER_SAMPLER",
		"ENV_INPUT_JAEGER_TAGS", "ENV_INPUT_JAEGER_THREADS", "ENV_INPUT_JAEGER_STORAGE", "ENV_INPUT_JAEGER_DEL_MESSAGE",
	} {
		value, ok := envs[key]
		if !ok {
			continue
		}
		switch key {
		case "ENV_INPUT_JAEGER_HTTP_ENDPOINT":
			ipt.Endpoint = value
		case "ENV_INPUT_JAEGER_UDP_ENDPOINT":
			ipt.Address = value
		case "ENV_INPUT_JAEGER_IGNORE_TAGS":
			var list []string
			if err := json.Unmarshal([]byte(value), &list); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.IgnoreTags = list
			}
		case "ENV_INPUT_JAEGER_KEEP_RARE_RESOURCE":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.KeepRareResource = ok
			}
		case "ENV_INPUT_JAEGER_CLOSE_RESOURCE":
			var closeRes map[string][]string
			if err := json.Unmarshal([]byte(value), &closeRes); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.CloseResource = closeRes
			}
		case "ENV_INPUT_JAEGER_SAMPLER":
			if ratio, err := strconv.ParseFloat(value, 64); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				if ipt.Sampler == nil {
					ipt.Sampler = &itrace.Sampler{}
				}
				ipt.Sampler.SamplingRateGlobal = ratio
			}
		case "ENV_INPUT_JAEGER_TAGS":
			var tags map[string]string
			if err := json.Unmarshal([]byte(value), &tags); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.Tags = tags
			}
		case "ENV_INPUT_JAEGER_THREADS":
			var threads workerpool.WorkerPoolConfig
			if err := json.Unmarshal([]byte(value), &threads); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.WPConfig = &threads
			}
		case "ENV_INPUT_JAEGER_STORAGE":
			var storage storage.StorageConfig
			if err := json.Unmarshal([]byte(value), &storage); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.LocalCacheConfig = &storage
			}
		case "ENV_INPUT_JAEGER_DEL_MESSAGE":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.DelMessage = ok
			}
		}
	}
}

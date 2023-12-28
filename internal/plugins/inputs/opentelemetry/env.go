// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package opentelemetry handle OTEL APM trace
package opentelemetry

import (
	"encoding/json"
	"strconv"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
)

var _ inputs.ReadEnv = &Input{}

// ReadEnv load config from environment values
// ENV_INPUT_OTEL_CUSTOMER_TAGS : JSON string
// ENV_INPUT_OTEL_KEEP_RARE_RESOURCE : bool
// ENV_INPUT_OTEL_OMIT_ERR_STATUS : JSON string
// ENV_INPUT_OTEL_CLOSE_RESOURCE : JSON string
// ENV_INPUT_OTEL_SAMPLER : float
// ENV_INPUT_OTEL_TAGS : JSON string
// ENV_INPUT_OTEL_THREADS : JSON string
// ENV_INPUT_OTEL_STORAGE : JSON string
// ENV_INPUT_OTEL_HTTP : JSON string
// ENV_INPUT_OTEL_GRPC : JSON string
// ENV_INPUT_OTEL_EXPECTED_HEADERS : JSON string
// below is a complete example for env in shell
// export ENV_INPUT_OTEL_IGNORE_TAGS=`["block1", "block2"]`
// export ENV_INPUT_OTEL_KEEP_RARE_RESOURCE=true
// export ENV_INPUT_OTEL_OMIT_ERR_STATUS=`["404", "403", "400"]`
// export ENV_INPUT_OTEL_CLOSE_RESOURCE=`{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}`
// export ENV_INPUT_OTEL_SAMPLER=0.3
// export ENV_INPUT_OTEL_TAGS=`{"k1":"v1", "k2":"v2", "k3":"v3"}`
// export ENV_INPUT_OTEL_THREADS=`{"buffer":1000, "threads":100}`
// export ENV_INPUT_OTEL_STORAGE=`{"storage":"./otel_storage", "capacity": 5120}`
// export ENV_INPUT_OTEL_HTTP=`{"enable":true, "http_status_ok": 200, "trace_api": "/otel/v1/trace", "metric_api": "/otel/v1/metric"}`
// export ENV_INPUT_OTEL_GRPC=`{"trace_enable": true, "metric_enable": true, "addr": "127.0.0.1:4317"}`
// export ENV_INPUT_OTEL_EXPECTED_HEADERS=`{"ex_version": "1.2.3", "ex_name": "env_resource_name"}`.
func (ipt *Input) ReadEnv(envs map[string]string) {
	log = logger.SLogger(inputName)

	for _, key := range []string{
		"ENV_INPUT_OTEL_CUSTOMER_TAGS", "ENV_INPUT_OTEL_KEEP_RARE_RESOURCE", "ENV_INPUT_OTEL_OMIT_ERR_STATUS",
		"ENV_INPUT_OTEL_CLOSE_RESOURCE", "ENV_INPUT_OTEL_SAMPLER", "ENV_INPUT_OTEL_TAGS",
		"ENV_INPUT_OTEL_THREADS", "ENV_INPUT_OTEL_STORAGE", "ENV_INPUT_OTEL_HTTP",
		"ENV_INPUT_OTEL_GRPC", "ENV_INPUT_OTEL_EXPECTED_HEADERS", "ENV_INPUT_OTEL_DEL_MESSAGE",
	} {
		value, ok := envs[key]
		if !ok {
			continue
		}
		switch key {
		case "ENV_INPUT_OTEL_CUSTOMER_TAGS":
			var list []string
			if err := json.Unmarshal([]byte(value), &list); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.CustomerTags = list
			}
		case "ENV_INPUT_OTEL_KEEP_RARE_RESOURCE":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.KeepRareResource = ok
			}
		case "ENV_INPUT_OTEL_OMIT_ERR_STATUS":
			var list []string
			if err := json.Unmarshal([]byte(value), &list); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.OmitErrStatus = list
			}
		case "ENV_INPUT_OTEL_CLOSE_RESOURCE":
			var closeRes map[string][]string
			if err := json.Unmarshal([]byte(value), &closeRes); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.CloseResource = closeRes
			}
		case "ENV_INPUT_OTEL_SAMPLER":
			if ratio, err := strconv.ParseFloat(value, 64); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				if ipt.Sampler == nil {
					ipt.Sampler = &itrace.Sampler{}
				}
				ipt.Sampler.SamplingRateGlobal = ratio
			}
		case "ENV_INPUT_OTEL_TAGS":
			var tags map[string]string
			if err := json.Unmarshal([]byte(value), &tags); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.Tags = tags
			}
		case "ENV_INPUT_OTEL_THREADS":
			var threads workerpool.WorkerPoolConfig
			if err := json.Unmarshal([]byte(value), &threads); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.WPConfig = &threads
			}
		case "ENV_INPUT_OTEL_STORAGE":
			var storage storage.StorageConfig
			if err := json.Unmarshal([]byte(value), &storage); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.LocalCacheConfig = &storage
			}
		case "ENV_INPUT_OTEL_HTTP":
			var httpconf httpConfig
			if err := json.Unmarshal([]byte(value), &httpconf); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.HTTPConfig = &httpconf
			}
		case "ENV_INPUT_OTEL_GRPC":
			var grpcconf grpcConfig
			if err := json.Unmarshal([]byte(value), &grpcconf); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.GRPCConfig = &grpcconf
			}
		case "ENV_INPUT_OTEL_EXPECTED_HEADERS":
			var headers map[string]string
			if err := json.Unmarshal([]byte(value), &headers); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.ExpectedHeaders = headers
			}
		case "ENV_INPUT_OTEL_DEL_MESSAGE":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.DelMessage = ok
			}
		}
	}
}

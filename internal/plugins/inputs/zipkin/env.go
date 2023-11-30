// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package zipkin handle Zipkin APM traces.
package zipkin

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
// ENV_INPUT_ZIPKIN_PATH_V1 : string
// ENV_INPUT_ZIPKIN_PATH_V2 : string
// ENV_INPUT_ZIPKIN_IGNORE_TAGS : JSON string
// ENV_INPUT_ZIPKIN_KEEP_RARE_RESOURCE : bool
// ENV_INPUT_ZIPKIN_CLOSE_RESOURCE : JSON string
// ENV_INPUT_ZIPKIN_SAMPLER : float
// ENV_INPUT_ZIPKIN_TAGS : JSON string
// ENV_INPUT_ZIPKIN_THREADS : JSON string
// ENV_INPUT_ZIPKIN_STORAGE : JSON string
// below is a complete example for env in shell
// export ENV_INPUT_ZIPKIN_PATH_V1="/api/v1/spans"
// export ENV_INPUT_ZIPKIN_PATH_V2="/api/v2/spans"
// export ENV_INPUT_ZIPKIN_IGNORE_TAGS=`["block1", "block2"]`
// export ENV_INPUT_ZIPKIN_KEEP_RARE_RESOURCE=true
// export ENV_INPUT_ZIPKIN_CLOSE_RESOURCE=`{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}`
// export ENV_INPUT_ZIPKIN_SAMPLER=0.3
// export ENV_INPUT_ZIPKIN_TAGS=`{"k1":"v1", "k2":"v2", "k3":"v3"}`
// export ENV_INPUT_ZIPKIN_THREADS=`{"buffer":1000, "threads":100}`
// export ENV_INPUT_ZIPKIN_STORAGE=`{"storage":"./zipkin_storage", "capacity": 5120}`.
func (ipt *Input) ReadEnv(envs map[string]string) {
	log = logger.SLogger(inputName)

	for _, key := range []string{
		"ENV_INPUT_ZIPKIN_PATH_V1", "ENV_INPUT_ZIPKIN_PATH_V2", "ENV_INPUT_ZIPKIN_IGNORE_TAGS",
		"ENV_INPUT_ZIPKIN_KEEP_RARE_RESOURCE", "ENV_INPUT_ZIPKIN_CLOSE_RESOURCE", "ENV_INPUT_ZIPKIN_SAMPLER",
		"ENV_INPUT_ZIPKIN_TAGS", "ENV_INPUT_ZIPKIN_THREADS", "ENV_INPUT_ZIPKIN_STORAGE", "ENV_INPUT_ZIPKIN_DEL_MESSAGE",
	} {
		value, ok := envs[key]
		if !ok {
			continue
		}
		switch key {
		case "ENV_INPUT_ZIPKIN_PATH_V1":
			ipt.PathV1 = value
		case "ENV_INPUT_ZIPKIN_PATH_V2":
			ipt.PathV2 = value
		case "ENV_INPUT_ZIPKIN_IGNORE_TAGS":
			var list []string
			if err := json.Unmarshal([]byte(value), &list); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.IgnoreTags = list
			}
		case "ENV_INPUT_ZIPKIN_KEEP_RARE_RESOURCE":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.KeepRareResource = ok
			}
		case "ENV_INPUT_ZIPKIN_CLOSE_RESOURCE":
			var closeRes map[string][]string
			if err := json.Unmarshal([]byte(value), &closeRes); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.CloseResource = closeRes
			}
		case "ENV_INPUT_ZIPKIN_SAMPLER":
			if ratio, err := strconv.ParseFloat(value, 64); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				if ipt.Sampler == nil {
					ipt.Sampler = &itrace.Sampler{}
				}
				ipt.Sampler.SamplingRateGlobal = ratio
			}
		case "ENV_INPUT_ZIPKIN_TAGS":
			var tags map[string]string
			if err := json.Unmarshal([]byte(value), &tags); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.Tags = tags
			}
		case "ENV_INPUT_ZIPKIN_THREADS":
			var threads workerpool.WorkerPoolConfig
			if err := json.Unmarshal([]byte(value), &threads); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.WPConfig = &threads
			}
		case "ENV_INPUT_ZIPKIN_STORAGE":
			var storage storage.StorageConfig
			if err := json.Unmarshal([]byte(value), &storage); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.LocalCacheConfig = &storage
			}
		case "ENV_INPUT_ZIPKIN_DEL_MESSAGE":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.DelMessage = ok
			}
		}
	}
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalking handle SkyWalking tracing metrics.
package skywalking

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
// ENV_INPUT_SKYWALKING_HTTP_ENDPOINTS : JSON string
// ENV_INPUT_SKYWALKING_GRPC_ENDPOINT : string
// ENV_INPUT_SKYWALKING_PLUGINS : JSON string
// ENV_INPUT_SKYWALKING_IGNORE_TAGS : JSON string
// ENV_INPUT_SKYWALKING_KEEP_RARE_RESOURCE : bool
// ENV_INPUT_SKYWALKING_CLOSE_RESOURCE : JSON string
// ENV_INPUT_SKYWALKING_SAMPLER : float
// ENV_INPUT_SKYWALKING_TAGS : JSON string
// ENV_INPUT_SKYWALKING_THREADS : JSON string
// ENV_INPUT_SKYWALKING_STORAGE : JSON string
// below is a complete example for env in shell
// export ENV_INPUT_SKYWALKING_HTTP_ENDPOINTS=`["/v3/trace", "/v3/metric", "/v3/logging", "/v3/profiling"]`
// export ENV_INPUT_SKYWALKING_GRPC_ENDPOINT="127.0.0.1:11800"
// export ENV_INPUT_SKYWALKING_PLUGINS=`["db.type", "os.call"]`
// export ENV_INPUT_SKYWALKING_IGNORE_TAGS=`["block1", "block2"]`
// export ENV_INPUT_SKYWALKING_KEEP_RARE_RESOURCE=true
// export ENV_INPUT_SKYWALKING_CLOSE_RESOURCE=`{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}`
// export ENV_INPUT_SKYWALKING_SAMPLER=0.3
// export ENV_INPUT_SKYWALKING_TAGS=`{"k1":"v1", "k2":"v2", "k3":"v3"}`
// export ENV_INPUT_SKYWALKING_THREADS=`{"buffer":1000, "threads":100}`
// export ENV_INPUT_SKYWALKING_STORAGE=`{"storage":"./skywalking_storage", "capacity": 5120}`.
func (ipt *Input) ReadEnv(envs map[string]string) {
	log = logger.SLogger(inputName)

	for _, key := range []string{
		"ENV_INPUT_SKYWALKING_HTTP_ENDPOINTS", "ENV_INPUT_SKYWALKING_GRPC_ENDPOINT", "ENV_INPUT_SKYWALKING_PLUGINS",
		"ENV_INPUT_SKYWALKING_IGNORE_TAGS", "ENV_INPUT_SKYWALKING_KEEP_RARE_RESOURCE", "ENV_INPUT_SKYWALKING_CLOSE_RESOURCE",
		"ENV_INPUT_SKYWALKING_SAMPLER", "ENV_INPUT_SKYWALKING_TAGS", "ENV_INPUT_SKYWALKING_THREADS", "ENV_INPUT_SKYWALKING_STORAGE",
		"ENV_INPUT_SKYWALKING_DEL_MESSAGE",
	} {
		value, ok := envs[key]
		if !ok {
			continue
		}
		switch key {
		case "ENV_INPUT_SKYWALKING_HTTP_ENDPOINTS":
			var list []string
			if err := json.Unmarshal([]byte(value), &list); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.Endpoints = list
			}
		case "ENV_INPUT_SKYWALKING_GRPC_ENDPOINT":
			ipt.Address = value
		case "ENV_INPUT_SKYWALKING_PLUGINS":
			var list []string
			if err := json.Unmarshal([]byte(value), &list); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.Plugins = list
			}
		case "ENV_INPUT_SKYWALKING_IGNORE_TAGS":
			var list []string
			if err := json.Unmarshal([]byte(value), &list); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.IgnoreTags = list
			}
		case "ENV_INPUT_SKYWALKING_KEEP_RARE_RESOURCE":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.KeepRareResource = ok
			}
		case "ENV_INPUT_SKYWALKING_CLOSE_RESOURCE":
			var closeRes map[string][]string
			if err := json.Unmarshal([]byte(value), &closeRes); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.CloseResource = closeRes
			}
		case "ENV_INPUT_SKYWALKING_SAMPLER":
			if ratio, err := strconv.ParseFloat(value, 64); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				if ipt.Sampler == nil {
					ipt.Sampler = &itrace.Sampler{}
				}
				ipt.Sampler.SamplingRateGlobal = ratio
			}
		case "ENV_INPUT_SKYWALKING_TAGS":
			var tags map[string]string
			if err := json.Unmarshal([]byte(value), &tags); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.Tags = tags
			}
		case "ENV_INPUT_SKYWALKING_THREADS":
			var threads workerpool.WorkerPoolConfig
			if err := json.Unmarshal([]byte(value), &threads); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.WPConfig = &threads
			}
		case "ENV_INPUT_SKYWALKING_STORAGE":
			var storage storage.StorageConfig
			if err := json.Unmarshal([]byte(value), &storage); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.LocalCacheConfig = &storage
			}
		case "ENV_INPUT_SKYWALKING_DEL_MESSAGE":
			if ok, err := strconv.ParseBool(value); err != nil {
				log.Warnf("parse %s=%s failed: %s", key, value, err.Error())
			} else {
				ipt.DelMessage = ok
			}
		}
	}
}

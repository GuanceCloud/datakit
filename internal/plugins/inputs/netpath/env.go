// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (*Input) GetENVDoc() []*inputs.ENVInfo {
	infos := []*inputs.ENVInfo{
		{FieldName: "Protocol"},
		{FieldName: "Interval"},
		{FieldName: "Timeout"},
		{FieldName: "MaxTTL", ENVName: "MAX_TTL"},
		{FieldName: "TracerouteQueries", ENVName: "TRACEROUTE_QUERIES"},
		{FieldName: "E2EQueries", ENVName: "E2E_QUERIES"},
		{FieldName: "Tags"},
		{FieldName: "DynamicEnabled", ENVName: "DYNAMIC_ENABLED", ConfField: "dynamic.enabled", Type: doc.Boolean},
		{FieldName: "DynamicProtocol", ENVName: "DYNAMIC_PROTOCOL", ConfField: "dynamic.protocol"},
		{FieldName: "DynamicToken", ENVName: "DYNAMIC_TOKEN", ConfField: "dynamic.token", Type: doc.String},
		{FieldName: "DynamicTTL", ENVName: "DYNAMIC_TTL", ConfField: "dynamic.ttl", Type: doc.TimeDuration},
		{FieldName: "DynamicInterval", ENVName: "DYNAMIC_INTERVAL", ConfField: "dynamic.interval", Type: doc.TimeDuration},
		{FieldName: "DynamicFlushInterval", ENVName: "DYNAMIC_FLUSH_INTERVAL", ConfField: "dynamic.flush_interval", Type: doc.TimeDuration},
		{FieldName: "DynamicMaxPerMinute", ENVName: "DYNAMIC_MAX_PER_MINUTE", ConfField: "dynamic.max_per_minute", Type: doc.Int},
		{FieldName: "DynamicWorkers", ENVName: "DYNAMIC_WORKERS", ConfField: "dynamic.workers", Type: doc.Int},
		{FieldName: "DynamicContextsLimit", ENVName: "DYNAMIC_CONTEXTS_LIMIT", ConfField: "dynamic.contexts_limit", Type: doc.Int},
		{FieldName: "DynamicContextsBytesLimit", ENVName: "DYNAMIC_CONTEXTS_BYTES_LIMIT", ConfField: "dynamic.contexts_bytes_limit", Type: doc.Int},
		{FieldName: "DynamicE2EQueries", ENVName: "DYNAMIC_E2E_QUERIES", ConfField: "dynamic.e2e_queries", Type: doc.Int},
		{FieldName: "MonitorIPWithoutDomain", ENVName: "DYNAMIC_MONITOR_IP_WITHOUT_DOMAIN", ConfField: "dynamic.monitor_ip_without_domain", Type: doc.Boolean},
	}
	return doc.SetENVDoc("ENV_INPUT_NETPATH_", infos)
}

// ReadEnv applies Kubernetes-oriented environment overrides. A netpath input
// instance must still be declared in conf.d so DataKit can load the input.
func (ipt *Input) ReadEnv(envs map[string]string) {
	if ipt.Tags == nil {
		ipt.Tags = map[string]string{}
	}
	if tags, ok := envs["ENV_INPUT_NETPATH_TAGS"]; ok {
		for k, v := range config.ParseGlobalTags(tags) {
			ipt.Tags[k] = v
		}
	}
	if value, ok := envs["ENV_INPUT_NETPATH_PROTOCOL"]; ok {
		ipt.Protocol = value
	}
	setInputDurationFromEnv(envs, "ENV_INPUT_NETPATH_INTERVAL", &ipt.Interval)
	setInputDurationFromEnv(envs, "ENV_INPUT_NETPATH_TIMEOUT", &ipt.Timeout)
	setInputIntFromEnv(envs, "ENV_INPUT_NETPATH_MAX_TTL", &ipt.MaxTTL)
	setInputIntFromEnv(envs, "ENV_INPUT_NETPATH_TRACEROUTE_QUERIES", &ipt.TracerouteQueries)
	setInputIntFromEnv(envs, "ENV_INPUT_NETPATH_E2E_QUERIES", &ipt.E2EQueries)

	if ipt.Dynamic == nil {
		ipt.Dynamic = &DynamicConfig{}
	}
	dynamic := ipt.Dynamic
	if value, ok := envs["ENV_INPUT_NETPATH_DYNAMIC_ENABLED"]; ok {
		setInputBoolFromEnv(value, "ENV_INPUT_NETPATH_DYNAMIC_ENABLED", &dynamic.Enabled)
	}
	if value, ok := envs["ENV_INPUT_NETPATH_DYNAMIC_PROTOCOL"]; ok {
		dynamic.Protocol = value
	}
	if value, ok := envs["ENV_INPUT_NETPATH_DYNAMIC_TOKEN"]; ok {
		dynamic.Token = value
	}
	setInputDurationFromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_TTL", &dynamic.TTL)
	setInputDurationFromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_INTERVAL", &dynamic.Interval)
	setInputDurationFromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_FLUSH_INTERVAL", &dynamic.FlushInterval)
	setInputDurationFromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_TIMEOUT", &dynamic.Timeout)
	setInputIntFromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_CONTEXTS_LIMIT", &dynamic.ContextsLimit)
	setInputInt64FromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_CONTEXTS_BYTES_LIMIT", &dynamic.ContextsBytesLimit)
	setInputIntFromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_MAX_PER_MINUTE", &dynamic.MaxPerMinute)
	setInputIntFromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_WORKERS", &dynamic.Workers)
	setInputIntFromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_MAX_TTL", &dynamic.MaxTTL)
	setInputIntFromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_TRACEROUTE_QUERIES", &dynamic.Queries)
	setInputIntFromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_E2E_QUERIES", &dynamic.E2EQueries)
	setInputIntFromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_INPUT_QUEUE", &dynamic.InputQueue)
	setInputIntFromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_PROCESS_QUEUE", &dynamic.ProcessQueue)
	setInputIntFromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_MAX_TESTS_PER_REQUEST", &dynamic.MaxTestsPerRequest)
	setInputInt64FromEnv(envs, "ENV_INPUT_NETPATH_DYNAMIC_MAX_BODY_BYTES", &dynamic.MaxBodyBytes)
	if value, ok := envs["ENV_INPUT_NETPATH_DYNAMIC_MONITOR_IP_WITHOUT_DOMAIN"]; ok {
		setInputBoolFromEnv(value, "ENV_INPUT_NETPATH_DYNAMIC_MONITOR_IP_WITHOUT_DOMAIN", &dynamic.MonitorIPWithoutDomain)
	}
}

func setInputDurationFromEnv(envs map[string]string, key string, destination **datakit.Duration) {
	value, ok := envs[key]
	if !ok {
		return
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		l.Warnf("parse %s: %s, ignored", key, err)
		return
	}
	*destination = &datakit.Duration{Duration: duration}
}

func setInputIntFromEnv(envs map[string]string, key string, destination *int) {
	value, ok := envs[key]
	if !ok {
		return
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		l.Warnf("parse %s: %s, ignored", key, err)
		return
	}
	*destination = parsed
}

func setInputInt64FromEnv(envs map[string]string, key string, destination *int64) {
	value, ok := envs[key]
	if !ok {
		return
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		l.Warnf("parse %s: %s, ignored", key, err)
		return
	}
	*destination = parsed
}

func setInputBoolFromEnv(value, key string, destination *bool) {
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		l.Warnf("parse %s: %s, ignored", key, err)
		return
	}
	*destination = parsed
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
)

// ReadEnv , support envs：
//
// ENV_INPUT_CONTAINER_ENDPOINTS : []string
// ENV_INPUT_CONTAINER_DOCKER_ENDPOINT : string
// ENV_INPUT_CONTAINER_CONTAINERD_ADDRESS : string
// ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS : booler
// ENV_INPUT_CONTAINER_LOGGING_SEARCH_INTERVAL : string ("10s")
// ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC : booler
// ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC : booler
// ENV_INPUT_CONTAINER_ENABLE_POD_METRIC : booler
// ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL : booler
// ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS     booler
// ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS booler
// ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_MONITORS        booler
// ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_MONITORS    booler
// ENV_INPUT_CONTAINER_AUTO_DISCOVERY_OF_PROM_STREAM_SIZE : int e.g. "10"
// ENV_INPUT_CONTAINER_TAGS : "a=b,c=d"
// ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG : []string
// ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG : []string
// ENV_INPUT_CONTAINER_KUBERNETES_URL : string
// ENV_INPUT_CONTAINER_BEARER_TOKEN : string
// ENV_INPUT_CONTAINER_BEARER_TOKEN_STRING : string
// ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP : string
// ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON : string (JSON map)
// ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_DETECTION: booler
// ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON : string (JSON string array)
// ENV_INPUT_CONTAINER_LOGGING_MIN_FLUSH_INTERVAL: string ("10s")
// ENV_INPUT_CONTAINER_LOGGING_MAX_MULTILINE_LIFE_DURATION : string ("5s")
// ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES : booler.
func (ipt *Input) ReadEnv(envs map[string]string) {
	if endpointStr, ok := envs["ENV_INPUT_CONTAINER_ENDPOINTS"]; ok {
		arrays := strings.Split(endpointStr, ",")
		ipt.Endpoints = append(ipt.Endpoints, arrays...)
	}

	if endpoint, ok := envs["ENV_INPUT_CONTAINER_DOCKER_ENDPOINT"]; ok {
		ipt.DeprecatedDockerEndpoint = endpoint
	}

	if address, ok := envs["ENV_INPUT_CONTAINER_CONTAINERD_ADDRESS"]; ok {
		ipt.DeprecatedContainerdAddress = address
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP"]; ok {
		ipt.LoggingExtraSourceMap = config.ParseGlobalTags(str)
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON"]; ok {
		if err := json.Unmarshal([]byte(str), &ipt.LoggingSourceMultilineMap); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON to map: %s, ignore", err)
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS to bool: %s, ignore", err)
		} else {
			ipt.EnableExtractK8sLabelAsTags = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC to bool: %s, ignore", err)
		} else {
			ipt.EnableContainerMetric = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC to bool: %s, ignore", err)
		} else {
			ipt.EnableK8sMetric = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_POD_METRIC"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_POD_METRIC to bool: %s, ignore", err)
		} else {
			ipt.EnablePodMetric = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_K8S_EVENT"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_K8S_EVENT to bool: %s, ignore", err)
		} else {
			ipt.EnableK8sEvent = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL to bool: %s, ignore", err)
		} else {
			ipt.EnableK8sNodeLocal = b
		}
	}

	if sizeStr, ok := envs["ENV_INPUT_CONTAINER_AUTO_DISCOVERY_OF_PROM_STREAM_SIZE"]; ok {
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_AUTO_DISCOVERY_OF_PROM_STREAM_SIZE to int64: %s, ignore", err)
		} else {
			ipt.autoDiscoveryOfPromStreamSize = int(size)
		}
	}
	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS to bool: %s, ignore", err)
		} else {
			ipt.EnableAutoDiscoveryOfPrometheusServiceAnnotations = b
		}
	}
	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS to bool: %s, ignore", err)
		} else {
			ipt.EnableAutoDiscoveryOfPrometheusPodAnnotations = b
		}
	}
	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_MONITORS"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_MONITORS to bool: %s, ignore", err)
		} else {
			ipt.EnableAutoDiscoveryOfPrometheusPodMonitors = b
		}
	}
	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_MONITORS"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_MONITORS to bool: %s, ignore", err)
		} else {
			ipt.EnableAutoDiscoveryOfPrometheusServiceMonitors = b
		}
	}

	if open, ok := envs["ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_DETECTION"]; ok {
		b, err := strconv.ParseBool(open)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_DETECTION to bool: %s, ignore", err)
		} else {
			ipt.LoggingAutoMultilineDetection = b
		}
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON"]; ok {
		if err := json.Unmarshal([]byte(str), &ipt.LoggingAutoMultilineExtraPatterns); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON to map: %s, ignore", err)
		}
	}

	if tagsStr, ok := envs["ENV_INPUT_CONTAINER_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG"]; ok {
		arrays := strings.Split(str, ",")
		ipt.ContainerIncludeLog = append(ipt.ContainerIncludeLog, arrays...)
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG"]; ok {
		arrays := strings.Split(str, ",")
		ipt.ContainerExcludeLog = append(ipt.ContainerExcludeLog, arrays...)
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_KUBERNETES_URL"]; ok {
		ipt.K8sURL = str
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_BEARER_TOKEN"]; ok {
		ipt.K8sBearerToken = str
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_BEARER_TOKEN_STRING"]; ok {
		ipt.K8sBearerTokenString = str
	}

	if durStr, ok := envs["ENV_INPUT_CONTAINER_LOGGING_SEARCH_INTERVAL"]; ok {
		if dur, err := time.ParseDuration(durStr); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_SEARCH_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.LoggingSearchInterval = dur
		}
	}

	if durStr, ok := envs["ENV_INPUT_CONTAINER_LOGGING_MIN_FLUSH_INTERVAL"]; ok {
		if dur, err := time.ParseDuration(durStr); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_MIN_FLUSH_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.LoggingMinFlushInterval = dur
		}
	}

	if durStr, ok := envs["ENV_INPUT_CONTAINER_LOGGING_MAX_MULTILINE_LIFE_DURATION"]; ok {
		if dur, err := timex.ParseDuration(durStr); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_MAX_MULTILINE_LIFE_DURATION to time.Duration: %s, ignore", err)
		} else {
			ipt.LoggingMaxMultilineLifeDuration = dur
		}
	}

	if remove, ok := envs["ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES"]; ok {
		b, err := strconv.ParseBool(remove)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES to bool: %s, ignore", err)
		} else {
			ipt.LoggingRemoveAnsiEscapeCodes = b
		}
	}
}

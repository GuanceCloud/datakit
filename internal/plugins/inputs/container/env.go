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
// ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC : booler
// ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC : booler
// ENV_INPUT_CONTAINER_ENABLE_POD_METRIC : booler
// ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL : booler
// ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS : booler
// ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS     booler
// ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS booler
// ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_MONITORS        booler
// ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_MONITORS    booler
// ENV_INPUT_CONTAINER_AUTO_DISCOVERY_OF_PROM_STREAM_SIZE : int e.g. "10"
// ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG : []string
// ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG : []string
// ENV_INPUT_CONTAINER_KUBERNETES_URL : string
// ENV_INPUT_CONTAINER_BEARER_TOKEN : string
// ENV_INPUT_CONTAINER_BEARER_TOKEN_STRING : string
// ENV_INPUT_CONTAINER_LOGGING_FORCE_FLUSH_LIMIT : int
// ENV_INPUT_CONTAINER_LOGGING_SEARCH_INTERVAL : string ("10s")
// ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP : string
// ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON : string (JSON map)
// ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_DETECTION: booler
// ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON : string (JSON string array)
// ENV_INPUT_CONTAINER_LOGGING_MAX_MULTILINE_LIFE_DURATION : string ("5s")
// ENV_INPUT_CONTAINER_LOGGING_FILE_FROM_BEGINNING_THRESHOLD_SIZE : int.
// ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES : booler.
// ENV_INPUT_CONTAINER_TAGS : "a=b,c=d".
func (ipt *Input) ReadEnv(envs map[string]string) {
	///
	/// base configs
	///
	if str, ok := envs["ENV_INPUT_CONTAINER_ENDPOINTS"]; ok {
		ipt.Endpoints = append(ipt.Endpoints, strings.Split(str, ",")...)
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_DOCKER_ENDPOINT"]; ok {
		ipt.DeprecatedDockerEndpoint = str
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_CONTAINERD_ADDRESS"]; ok {
		ipt.DeprecatedContainerdAddress = str
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_TAGS"]; ok {
		tags := config.ParseGlobalTags(str)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	///
	/// container and k8s
	///
	if str, ok := envs["ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC to bool: %s, ignore", err)
		} else {
			ipt.EnableContainerMetric = b
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_ENABLE_POD_METRIC"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_POD_METRIC to bool: %s, ignore", err)
		} else {
			ipt.EnablePodMetric = b
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC to bool: %s, ignore", err)
		} else {
			ipt.EnableK8sMetric = b
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_ENABLE_K8S_EVENT"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_K8S_EVENT to bool: %s, ignore", err)
		} else {
			ipt.EnableK8sEvent = b
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL to bool: %s, ignore", err)
		} else {
			ipt.EnableK8sNodeLocal = b
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS to bool: %s, ignore", err)
		} else {
			ipt.EnableExtractK8sLabelAsTags = b
		}
	}

	///
	/// k8s connect
	///
	if str, ok := envs["ENV_INPUT_CONTAINER_KUBERNETES_URL"]; ok {
		ipt.K8sURL = str
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_BEARER_TOKEN"]; ok {
		ipt.K8sBearerToken = str
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_BEARER_TOKEN_STRING"]; ok {
		ipt.K8sBearerTokenString = str
	}

	///
	/// k8s autodiscoery
	///
	if str, ok := envs["ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS to bool: %s, ignore", err)
		} else {
			ipt.EnableAutoDiscoveryOfPrometheusServiceAnnotations = b
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS to bool: %s, ignore", err)
		} else {
			ipt.EnableAutoDiscoveryOfPrometheusPodAnnotations = b
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_MONITORS"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_MONITORS to bool: %s, ignore", err)
		} else {
			ipt.EnableAutoDiscoveryOfPrometheusPodMonitors = b
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_MONITORS"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_MONITORS to bool: %s, ignore", err)
		} else {
			ipt.EnableAutoDiscoveryOfPrometheusServiceMonitors = b
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_AUTO_DISCOVERY_OF_PROM_STREAM_SIZE"]; ok {
		if size, err := strconv.ParseInt(str, 10, 64); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_AUTO_DISCOVERY_OF_PROM_STREAM_SIZE to int64: %s, ignore", err)
		} else {
			ipt.autoDiscoveryOfPromStreamSize = int(size)
		}
	}

	///
	/// logging sample configs
	///
	if str, ok := envs["ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG"]; ok {
		arrays := strings.Split(str, ",")
		ipt.ContainerIncludeLog = append(ipt.ContainerIncludeLog, arrays...)
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG"]; ok {
		arrays := strings.Split(str, ",")
		ipt.ContainerExcludeLog = append(ipt.ContainerExcludeLog, arrays...)
	}

	///
	/// logging configs
	///
	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_FORCE_FLUSH_LIMIT"]; ok {
		if limit, err := strconv.ParseInt(str, 10, 64); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_FORCE_FLUSH_LIMIT to int64: %s, ignore", err)
		} else {
			ipt.LoggingForceFlushLimit = int(limit)
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON"]; ok {
		if err := json.Unmarshal([]byte(str), &ipt.LoggingSourceMultilineMap); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON to map: %s, ignore", err)
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_DETECTION"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
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
	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_SEARCH_INTERVAL"]; ok {
		if dur, err := time.ParseDuration(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_SEARCH_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.LoggingSearchInterval = dur
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_MAX_MULTILINE_LIFE_DURATION"]; ok {
		if dur, err := timex.ParseDuration(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_MAX_MULTILINE_LIFE_DURATION to time.Duration: %s, ignore", err)
		} else {
			ipt.LoggingMaxMultilineLifeDuration = dur
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_FILE_FROM_BEGINNING_THRESHOLD_SIZE"]; ok {
		if size, err := strconv.ParseInt(str, 10, 64); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_FILE_FROM_BEGINNING_THRESHOLD_SIZE to int64: %s, ignore", err)
		} else {
			ipt.LoggingFileFromBeginningThresholdSize = int(size)
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES to bool: %s, ignore", err)
		} else {
			ipt.LoggingRemoveAnsiEscapeCodes = b
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP"]; ok {
		ipt.LoggingExtraSourceMap = config.ParseGlobalTags(str)
	}
}

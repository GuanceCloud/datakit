// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"encoding/json"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
)

// ReadEnv , support envs：
//   ENV_INPUT_CONTAINER_DOCKER_ENDPOINT : string
//   ENV_INPUT_CONTAINER_CONTAINERD_ADDRESS : string
//   ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES : booler
//   ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC : booler
//   ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC : booler
//   ENV_INPUT_CONTAINER_ENABLE_POD_METRIC : booler
//   ENV_INPUT_CONTAINER_AUTO_DISCOVERY_OF_K8S_SERVICE_PROMETHEUS: booler
//   ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS: booler
//   ENV_INPUT_CONTAINER_TAGS : "a=b,c=d"
//   ENV_INPUT_CONTAINER_EXCLUDE_PAUSE_CONTAINER : booler
//   ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG : []string
//   ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG : []string
//   ENV_INPUT_CONTAINER_KUBERNETES_URL : string
//   ENV_INPUT_CONTAINER_BEARER_TOKEN : string
//   ENV_INPUT_CONTAINER_BEARER_TOKEN_STRING : string
//   ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP : string
//   ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON : string (JSON map)
//   ENV_INPUT_CONTAINER_LOGGING_BLOCKING_MODE : booler
//   ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_DETECTION: booler
//   ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON : string (JSON string array)
//   ENV_INPUT_CONTAINER_LOGGING_MIN_FLUSH_INTERVAL: string ("10s")
//   ENV_INPUT_CONTAINER_LOGGING_MAX_MULTILINE_LIFE_DURATION : string ("5s")
func (i *Input) ReadEnv(envs map[string]string) {
	if endpoint, ok := envs["ENV_INPUT_CONTAINER_DOCKER_ENDPOINT"]; ok {
		i.DockerEndpoint = endpoint
	}

	if address, ok := envs["ENV_INPUT_CONTAINER_CONTAINERD_ADDRESS"]; ok {
		i.ContainerdAddress = address
	}

	if v, ok := envs["ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP"]; ok {
		i.LoggingExtraSourceMap = config.ParseGlobalTags(v)
	}

	if v, ok := envs["ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON"]; ok {
		if err := json.Unmarshal([]byte(v), &i.LoggingSourceMultilineMap); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON to map: %s, ignore", err)
		}
	}

	if remove, ok := envs["ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES"]; ok {
		b, err := strconv.ParseBool(remove)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES to bool: %s, ignore", err)
		} else {
			i.LoggingRemoveAnsiEscapeCodes = b
		}
	}

	if disable, ok := envs["ENV_INPUT_CONTAINER_LOGGING_BLOCKING_MODE"]; ok {
		b, err := strconv.ParseBool(disable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_BLOCKING_MODE to bool: %s, ignore", err)
		} else {
			i.LoggingBlockingMode = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC to bool: %s, ignore", err)
		} else {
			i.EnableContainerMetric = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS to bool: %s, ignore", err)
		} else {
			i.ExtractK8sLabelAsTags = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC to bool: %s, ignore", err)
		} else {
			i.EnableK8sMetric = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_POD_METRIC"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_POD_METRIC to bool: %s, ignore", err)
		} else {
			i.EnablePodMetric = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_AUTO_DISCOVERY_OF_K8S_SERVICE_PROMETHEUS"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_AUTO_DISCOVERY_OF_K8S_SERVICE_PROMETHEUS to bool: %s, ignore", err)
		} else {
			i.AutoDiscoveryOfK8sServicePrometheus = b
		}
	}

	if exclude, ok := envs["ENV_INPUT_CONTAINER_EXCLUDE_PAUSE_CONTAINER"]; ok {
		b, err := strconv.ParseBool(exclude)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_EXCLUDE_PAUSE_CONTAINER to bool: %s, ignore", err)
		} else {
			i.ExcludePauseContainer = b
		}
	}

	if open, ok := envs["ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_DETECTION"]; ok {
		b, err := strconv.ParseBool(open)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_DETECTION to bool: %s, ignore", err)
		} else {
			i.LoggingAutoMultilineDetection = b
		}
	}

	if v, ok := envs["ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON"]; ok {
		if err := json.Unmarshal([]byte(v), &i.LoggingAutoMultilineExtraPatterns); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON to map: %s, ignore", err)
		}
	}

	if tagsStr, ok := envs["ENV_INPUT_CONTAINER_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			i.Tags[k] = v
		}
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add CONTAINER_INCLUDE_LOG from ENV: %v", arrays)
		i.ContainerIncludeLog = append(i.ContainerIncludeLog, arrays...)
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add CONTAINER_EXCLUDE_LOG from ENV: %v", arrays)
		i.ContainerExcludeLog = append(i.ContainerExcludeLog, arrays...)
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_KUBERNETES_URL"]; ok {
		i.K8sURL = str
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_BEARER_TOKEN"]; ok {
		i.K8sBearerToken = str
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_BEARER_TOKEN_STRING"]; ok {
		i.K8sBearerTokenString = str
	}

	if durStr, ok := envs["ENV_INPUT_CONTAINER_LOGGING_MIN_FLUSH_INTERVAL"]; ok {
		if dur, err := timex.ParseDuration(durStr); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_MIN_FLUSH_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			i.LoggingMinFlushInterval = dur
		}
	}

	if durStr, ok := envs["ENV_INPUT_CONTAINER_LOGGING_MAX_MULTILINE_LIFE_DURATION"]; ok {
		if dur, err := timex.ParseDuration(durStr); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_MAX_MULTILINE_LIFE_DURATION to time.Duration: %s, ignore", err)
		} else {
			i.LoggingMaxMultilineLifeDuration = dur
		}
	}
}

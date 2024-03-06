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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
)

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Endpoints", Type: doc.List, Example: "\"`unix:///var/run/docker.sock,unix:///var/run/containerd/containerd.sock,unix:///var/run/crio/crio.sock`\"", Desc: "Append to container endpoints", DescZh: "追加多个容器运行时的 endpoint"},
		{FieldName: "DeprecatedDockerEndpoint", ENVName: "DOCKER_ENDPOINT", Type: doc.String, Example: "`unix:///var/run/docker.sock`", Desc: "Deprecated. Specify the endpoint of Docker Engine", DescZh: "已废弃，指定 Docker Engine 的 endpoint"},
		{FieldName: "DeprecatedContainerdAddress", ENVName: "CONTAINERD_ADDRESS", Type: doc.String, Example: "`/var/run/containerd/containerd.sock`", Desc: "Deprecated. Specify the endpoint of `Containerd`", DescZh: "已废弃，指定 `Containerd` 的 endpoint"},
		{FieldName: "EnableContainerMetric", Type: doc.Boolean, Default: "true", Desc: "Start container index collection", DescZh: "开启容器指标采集"},
		{FieldName: "EnableK8sMetric", ENVName: "ENABLE_K8S_METRIC", Type: doc.Boolean, Default: "true", Desc: "Start k8s index collection", DescZh: "开启 k8s 指标采集"},
		{FieldName: "EnablePodMetric", Type: doc.Boolean, Default: "false", Desc: `Turn on Pod index collection`, DescZh: `是否开启 Pod 指标采集（CPU 和内存使用情况），需要安装[Kubernetes-metrics-server](https://github.com/kubernetes-sigs/metrics-server){:target="_blank"}`},
		{FieldName: "EnableK8sEvent", ENVName: "ENABLE_K8S_EVENT", Type: doc.Boolean, Default: "true", Desc: "Enable event collection mode", DescZh: "是否开启分时间采集模式"},
		{FieldName: "EnableK8sNodeLocal", ENVName: "ENABLE_K8S_NODE_LOCAL", Type: doc.Boolean, Default: "true", Desc: "Enable sub-Node collection mode, where the Datakit deployed on each Node independently collects the resources of the current Node.[:octicons-tag-24: Version-1.5.7](../datakit/changelog.md#cl-1.5.7) Need new `RABC` [link](#rbac-nodes-stats)", DescZh: "是否开启分 Node 采集模式，由部署在各个 Node 的 Datakit 独立采集当前 Node 的资源。[:octicons-tag-24: Version-1.19.0](../datakit/changelog.md#cl-1.19.0) 需要额外的 `RABC` 权限，见[此处](#rbac-nodes-stats)"},
		{FieldName: "EnableExtractK8sLabelAsTags", ENVName: "EXTRACT_K8S_LABEL_AS_TAGS", Type: doc.Boolean, Default: "false", Desc: `Should the labels of the resources be appended to the tags collected? Only Pod metrics, objects, and Node objects will be added, and the labels of container logs belonging to the Pod will also be added. If the key of a label contains a dot character, it will be replaced with a hyphen`, DescZh: `是否追加资源的 labels 到采集的 tag 中。只有 Pod 指标、对象和 Node 对象会添加，另外容器日志也会添加其所属 Pod 的 labels。如果 label 的 key 有 dot 字符，会将其变为横线`},
		{FieldName: "EnableAutoDiscoveryOfPrometheusPodAnnotations", Type: doc.Boolean, Default: "false", Desc: `Whether to turn on Prometheus Pod Annotations and collect metrics automatically`, DescZh: `是否开启自动发现 Prometheus Pod Annotations 并采集指标`},
		{FieldName: "EnableAutoDiscoveryOfPrometheusServiceAnnotations", Type: doc.Boolean, Default: "false", Desc: `Whether to turn on Prometheus Service Annotations and collect metrics automatically`, DescZh: `是否开启自动发现 Prometheus 服务 Annotations 并采集指标`},
		{FieldName: "EnableAutoDiscoveryOfPrometheusPodMonitors", Type: doc.Boolean, Default: "false", Desc: `Whether to turn on automatic discovery of Prometheus PodMonitor CRD and collection of metrics, see [Prometheus-Operator CRD doc](kubernetes-prometheus-operator-crd`, DescZh: `是否开启自动发现 Prometheus Pod Monitor CRD 并采集指标，详见[Prometheus-Operator CRD 文档](kubernetes-prometheus-operator-crd.md#config)`},
		{FieldName: "EnableAutoDiscoveryOfPrometheusServiceMonitors", Type: doc.Boolean, Default: "false", Desc: `Whether to turn on automatic discovery of Prometheus ServiceMonitor CRD and collection of metrics, see [Prometheus-Operator CRD doc](kubernetes-prometheus-operator-crd`, DescZh: `是否开启自动发现 Prometheus ServiceMonitor CRD 并采集指标，详见[Prometheus-Operator CRD 文档](kubernetes-prometheus-operator-crd.md#config)`},
		{FieldName: "ContainerIncludeLog", Type: doc.List, Example: `"image:pubrepo.jiagouyun.com/datakit/logfwd*"`, Desc: `Include condition of container log, filtering with image`, DescZh: `容器日志白名单，使用 image 过滤`},
		{FieldName: "ContainerExcludeLog", Type: doc.List, Example: `"image:pubrepo.jiagouyun.com/datakit/logfwd*"`, Desc: `Exclude condition of container log, filtering with image`, DescZh: `容器日志黑名单，使用 image 过滤`},
		{FieldName: "K8sURL", ENVName: "KUBERNETES_URL", Type: doc.String, Example: `https://kubernetes.default:443`, Desc: `k8s api-server access address`, DescZh: `k8s API 服务访问地址`},
		{FieldName: "K8sBearerToken", ENVName: "BEARER_TOKEN", Type: doc.String, Example: "`/run/secrets/kubernetes.io/serviceaccount/token`", Desc: `The path to the token file required to access k8s api-server`, DescZh: `访问 k8s 服务所需的 token 文件路径`},
		{FieldName: "K8sBearerTokenString", ENVName: "BEARER_TOKEN_STRING", Type: doc.String, Example: "your-token-string", Desc: `Token string required to access k8s api-server`, DescZh: `访问 k8s 服务所需的 token 字符串`},
		{FieldName: "LoggingSearchInterval", Type: doc.TimeDuration, Default: `60s`, Desc: `The time interval of log discovery, that is, how often logs are retrieved. If the interval is too long, some logs with short survival will be ignored`, DescZh: `日志发现的时间间隔，即每隔多久检索一次日志，如果间隔太长，会导致忽略了一些存活较短的日志`},
		{FieldName: "LoggingExtraSourceMap", Type: doc.Map, Example: `source_regex*=new_source,regex*=new_source2`, Desc: `Log collection configures additional source matching, and the regular source will be renamed`, DescZh: `日志采集配置额外的 source 匹配，符合正则的 source 会被改名`},
		{FieldName: "LoggingSourceMultilineMap", ENVName: "LOGGING_SOURCE_MULTILINE_MAP_JSON", ConfField: "logging_source_multiline_map", Type: doc.JSON, Example: `{"source_nginx":"^\\d{4}", "source_redis":"^[A-Za-z_]"}`, Desc: `Log collection configures additional source matching, and the regular source will be renamed`, DescZh: `日志采集配置额外的 source 匹配，符合正则的 source 会被改名`},
		{FieldName: "LoggingAutoMultilineDetection", Type: doc.Boolean, Default: `false`, Desc: `Whether the automatic multi-line mode is turned on for log collection; the applicable multi-line rules will be matched in the patterns list after it is turned on`, DescZh: `日志采集是否开启自动多行模式，开启后会在 patterns 列表中匹配适用的多行规则`},
		{FieldName: "LoggingAutoMultilineExtraPatterns", ENVName: "LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON", ConfField: "logging_auto_multiline_extra_patterns", Type: doc.JSON, Default: `For more default rules, see [doc](logging.md#auto-multiline)`, Example: `["^\\d{4}-\\d{2}", "^[A-Za-z_]"]`, Desc: `Automatic multi-line pattern pattens list for log collection, supporting manual configuration of multiple multi-line rules`, DescZh: `日志采集的自动多行模式 pattens 列表，支持手动配置多个多行规则`},
		{FieldName: "LoggingMaxMultilineLifeDuration", Type: doc.TimeDuration, Default: `3s`, Desc: `Maximum single multi-row life cycle of log collection. At the end of this cycle, existing multi-row data will be emptied and uploaded to avoid accumulation`, DescZh: `日志采集的单次多行最大生命周期，此周期结束将清空和上传现存的多行数据，避免堆积`},
		{FieldName: "LoggingRemoveAnsiEscapeCodes", Type: doc.Boolean, Default: `false`, Desc: "Remove `ansi` escape codes and color characters, referred to [`ansi-decode` doc](logging.md#ansi-decode)", DescZh: `日志采集删除包含的颜色字符，详见[日志特殊字符处理说明](logging.md#ansi-decode)`},
		{FieldName: "LoggingForceFlushLimit", Type: doc.Int, Default: `5`, Desc: `If there are consecutive N empty collections, the existing data will be uploaded to prevent memory occupation caused by accumulated`, DescZh: `日志采集上传限制，如果连续 N 次都采集为空，会将现有的数据上传，避免数据积攒占用内存`},
		{FieldName: "Tags"},
	}

	return doc.SetENVDoc("ENV_INPUT_CONTAINER_", infos)
}

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
// ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS_V2_FOR_METRIC : json arrry
// ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS_V2 : json arrry
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
			ipt.DeprecatedEnableExtractK8sLabelAsTags = b
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS_V2_FOR_METRIC"]; ok {
		var keys []string
		if err := json.Unmarshal([]byte(str), &keys); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS_V2_FOR_METRIC to jsonarray: %s, ignore", err)
		} else {
			ipt.ExtractK8sLabelAsTagsV2ForMetric = keys
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS_V2"]; ok {
		var keys []string
		if err := json.Unmarshal([]byte(str), &keys); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS_V2 to jsonarray: %s, ignore", err)
		} else {
			ipt.ExtractK8sLabelAsTagsV2 = keys
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
		ipt.ContainerIncludeLog = arrays
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG"]; ok {
		arrays := strings.Split(str, ",")
		ipt.ContainerExcludeLog = arrays
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

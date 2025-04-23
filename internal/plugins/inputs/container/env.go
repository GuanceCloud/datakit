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
)

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Endpoints", Type: doc.List, Example: "\"`unix:///var/run/docker.sock,unix:///var/run/containerd/containerd.sock,unix:///var/run/crio/crio.sock`\"", Desc: "Append to container endpoints", DescZh: "追加多个容器运行时的 endpoint"},
		{FieldName: "EnableContainerMetric", Type: doc.Boolean, Default: "true", Desc: "Start container index collection", DescZh: "开启容器指标采集"},
		{FieldName: "EnableK8sMetric", ENVName: "ENABLE_K8S_METRIC", Type: doc.Boolean, Default: "true", Desc: "Start k8s index collection", DescZh: "开启 k8s 指标采集"},
		{FieldName: "EnablePodMetric", Type: doc.Boolean, Default: "false", Desc: `Turn on Pod index collection`, DescZh: `是否开启 Pod 指标采集（CPU 和内存使用情况）`},
		{FieldName: "EnableK8sEvent", ENVName: "ENABLE_K8S_EVENT", Type: doc.Boolean, Default: "true", Desc: "Enable event collection mode", DescZh: "是否开启分时间采集模式"},
		{FieldName: "EnableK8sNodeLocal", ENVName: "ENABLE_K8S_NODE_LOCAL", Type: doc.Boolean, Default: "true", Desc: "Enable sub-Node collection mode, where the Datakit deployed on each Node independently collects the resources of the current Node.[:octicons-tag-24: Version-1.5.7](../datakit/changelog.md#cl-1.5.7) Need new `RABC` [link](#rbac-nodes-stats)", DescZh: "是否开启分 Node 采集模式，由部署在各个 Node 的 Datakit 独立采集当前 Node 的资源。[:octicons-tag-24: Version-1.19.0](../datakit/changelog.md#cl-1.19.0) 需要额外的 `RABC` 权限，见[此处](#rbac-nodes-stats)"},
		{FieldName: "EnableCollectKubeJob", Type: doc.Boolean, Default: `true`, Desc: `Turn off collection of Kubernetes Job resources (including metrics data and object data)`, DescZh: `开启对 Kubernetes Job 资源的采集（包括指标数据和对象数据）`},

		{FieldName: "EnableExtractK8sLabelAsTagsV2", ENVName: "EXTRACT_K8S_LABEL_AS_TAGS_V2", Type: doc.JSON, Example: "`[\"app\",\"name\"]`", Desc: `Append the labels of the resource to the tag of the non-metric (like object and logging) data. Label keys should be specified, if there is only one key and it is an empty string (e.g. [""]), all labels will be added to the tag. The container will inherit the Pod labels. If the key of the label has the dot character, it will be changed to a horizontal line`, DescZh: `追加资源的 labels 到数据（不包括指标数据）的 tag 中。需指定 label keys，如果只有一个 key 且为空字符串（例如 [""]），会添加所有 labels 到 tag。容器会继承 Pod labels。如果 label 的 key 有 dot 字符，会将其变为横线`},
		{FieldName: "EnableExtractK8sLabelAsTagsV2ForMetric", ENVName: "EXTRACT_K8S_LABEL_AS_TAGS_V2_FOR_METRIC", Type: doc.JSON, Example: "`[\"app\",\"name\"]`", Desc: `Append the labels of the resource to the tag of the metric data. Label keys should be specified, if there is only one key and it is an empty string (e.g. [""]), all labels will be added to the tag. The container will inherit the Pod labels. If the key of the label has the dot character, it will be changed to a horizontal line`, DescZh: `追加资源的 labels 到指标数据的 tag 中。需指定 label keys，如果只有一个 key 且为空字符串（例如 [""]），会添加所有 labels 到 tag。容器会继承 Pod labels。如果 label 的 key 有 dot 字符，会将其变为横线`},

		{FieldName: "EnableAutoDiscoveryOfPrometheusPodAnnotations", Type: doc.Boolean, Default: "false", Desc: `Deprecated. Whether to turn on Prometheus Pod Annotations and collect metrics automatically`, DescZh: `是否开启自动发现 Prometheus Pod Annotations 并采集指标`},
		{FieldName: "EnableAutoDiscoveryOfPrometheusServiceAnnotations", Type: doc.Boolean, Default: "false", Desc: `Deprecated. Whether to turn on Prometheus Service Annotations and collect metrics automatically`, DescZh: `是否开启自动发现 Prometheus 服务 Annotations 并采集指标`},
		{FieldName: "EnableAutoDiscoveryOfPrometheusPodMonitors", Type: doc.Boolean, Default: "false", Desc: `Deprecated. Whether to turn on automatic discovery of Prometheus PodMonitor CRD and collection of metrics, see [Prometheus-Operator CRD doc](kubernetes-prometheus-operator-crd`, DescZh: `是否开启自动发现 Prometheus Pod Monitor CRD 并采集指标，详见 [Prometheus-Operator CRD 文档](kubernetes-prometheus-operator-crd.md#config)`},
		{FieldName: "EnableAutoDiscoveryOfPrometheusServiceMonitors", Type: doc.Boolean, Default: "false", Desc: `Deprecated. Whether to turn on automatic discovery of Prometheus ServiceMonitor CRD and collection of metrics, see [Prometheus-Operator CRD doc](kubernetes-prometheus-operator-crd`, DescZh: `是否开启自动发现 Prometheus ServiceMonitor CRD 并采集指标，详见 [Prometheus-Operator CRD 文档](kubernetes-prometheus-operator-crd.md#config)`},

		{FieldName: "ContainerMaxConcurrent", Type: doc.Int, Default: `cpu cores + 1`, Desc: `Maximum number of concurrency when collecting container data, recommended to be turned on only when the collection delay is large`, DescZh: `采集容器数据时的最大并发数，推荐只在采集延迟较大时开启`},
		{FieldName: "ContainerIncludeLog", Type: doc.List, Example: "`\"image:pubrepo.jiagouyun.com/datakit/logfwd*\"`", Desc: `Include condition of container log, filtering with image`, DescZh: `容器日志白名单，使用 image/namespace 过滤`},
		{FieldName: "ContainerExcludeLog", Type: doc.List, Example: "`\"image:pubrepo.jiagouyun.com/datakit/logfwd*\"`", Desc: `Exclude condition of container log, filtering with image`, DescZh: `容器日志黑名单，使用 image/namespace 过滤`},
		{FieldName: "PodIncludeMetric", Type: doc.List, Example: "`\"namespace:datakit*\"`", Desc: `Include condition of pod metrics, filtering with namespace`, DescZh: `Pod 指标白名单，使用 namespace 过滤`},
		{FieldName: "PodExcludeMetric", Type: doc.List, Example: "`\"namespace:kube-system\"`", Desc: `Exclude condition of pod metrics, filtering with namespace`, DescZh: `Pod 指标黑名单，使用 namespace 过滤`},

		{FieldName: "LoggingSearchInterval", Type: doc.TimeDuration, Default: `60s`, Desc: `The time interval of log discovery, that is, how often logs are retrieved. If the interval is too long, some logs with short survival will be ignored`, DescZh: `日志发现的时间间隔，即每隔多久检索一次日志，如果间隔太长，会导致忽略了一些存活较短的日志`},
		{FieldName: "LoggingExtraSourceMap", Type: doc.Map, Example: "`source_regex*=new_source,regex*=new_source2`", Desc: `Log collection configures additional source matching, and the regular source will be renamed`, DescZh: `日志采集配置额外的 source 匹配，符合正则的 source 会被改名`},
		{FieldName: "LoggingSourceMultilineMap", ENVName: "LOGGING_SOURCE_MULTILINE_MAP_JSON", ConfField: "logging_source_multiline_map", Type: doc.JSON, Example: "`{\"source_nginx\":\"^\\d{4}\", \"source_redis\":\"^[A-Za-z_]\"}`", Desc: `Log collection with multiline configuration as specified by the source`, DescZh: `日志采集根据 source 指定多行配置`},
		{FieldName: "LoggingAutoMultilineDetection", Type: doc.Boolean, Default: `false`, Desc: `Whether the automatic multi-line mode is turned on for log collection; the applicable multi-line rules will be matched in the patterns list after it is turned on`, DescZh: `日志采集是否开启自动多行模式，开启后会在 patterns 列表中匹配适用的多行规则`},
		{FieldName: "LoggingAutoMultilineExtraPatterns", ENVName: "LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON", ConfField: "logging_auto_multiline_extra_patterns", Type: doc.JSON, Default: `For more default rules, see [doc](logging.md#auto-multiline)`, Example: "`[\"^\\d{4}-\\d{2}\", \"^[A-Za-z_]\"]`", Desc: `Automatic multi-line pattern pattens list for log collection, supporting manual configuration of multiple multi-line rules`, DescZh: `日志采集的自动多行模式 pattens 列表，支持手动配置多个多行规则`},
		{FieldName: "LoggingRemoveAnsiEscapeCodes", Type: doc.Boolean, Default: `false`, Desc: "Remove `ansi` escape codes and color characters, referred to [`ansi-decode` doc](logging.md#ansi-decode)", DescZh: `日志采集删除包含的颜色字符，详见[日志特殊字符处理说明](logging.md#ansi-decode)`},
		{FieldName: "LoggingFileFromBeginningThresholdSize", Type: doc.Int, Default: `20,000,000`, Desc: "Decide whether or not to from_beginning based on the file size, if the file size is smaller than this value when the file is found, start the collection from the begin", DescZh: `根据文件 size 决定是否 from_beginning，如果发现该文件时，文件 size 小于这个值，就使用 from_beginning 从头部开始采集`},
		{FieldName: "LoggingFileFromBeginning", Type: doc.Boolean, Default: `false`, Desc: "Whether to collect logs from the begin of the file", DescZh: `是否从文件首部采集日志`},
		{FieldName: "LoggingMaxOpenFiles", Type: doc.Int, Default: `500`, Desc: `The maximum allowed number of open files. If it is set to -1, it means there is no limit.`, DescZh: `日志采集最大打开文件个数，如果是 -1 则没有限制`},
		{FieldName: "LoggingFieldWhiteList", Type: doc.List, Example: "`'[\"service\",\"container_id\"]'`", Desc: `"Only retain the fields specified in the whitelist."`, DescZh: `指定保留白名单中的字段`},
		{FieldName: "Tags"},
	}

	return doc.SetENVDoc("ENV_INPUT_CONTAINER_", infos)
}

// ReadEnv , support envs：
//
// ENV_INPUT_CONTAINER_ENDPOINTS          : []string
// ENV_INPUT_CONTAINER_DOCKER_ENDPOINT    : string
// ENV_INPUT_CONTAINER_CONTAINERD_ADDRESS : string
// ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC : booler
// ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC       : booler
// ENV_INPUT_CONTAINER_ENABLE_POD_METRIC       : booler
// ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL   : booler
// ENV_INPUT_CONTAINER_ENABLE_K8S_EVENT        : booler
// ENV_INPUT_CONTAINER_ENABLE_COLLECT_KUBE_JOB : booler

// ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS_V2_FOR_METRIC : json arrry
// ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS_V2            : json arrry

// ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG : []string
// ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG : []string
// ENV_INPUT_CONTAINER_POD_INCLUDE_METRIC    : []string
// ENV_INPUT_CONTAINER_POD_EXCLUDE_METRIC    : []string

// ENV_INPUT_CONTAINER_LOGGING_SEARCH_INTERVAL  : string ("10s")
// ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP : string
// ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON : string (JSON map)
// ENV_INPUT_CONTAINER_LOGGING_ENABLE_MULTILINE        : booler
// ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_DETECTION: booler
// ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON : string (JSON string array)
// ENV_INPUT_CONTAINER_LOGGING_FILE_FROM_BEGINNING : booler
// ENV_INPUT_CONTAINER_LOGGING_FILE_FROM_BEGINNING_THRESHOLD_SIZE : int
// ENV_INPUT_CONTAINER_LOGGING_FIELD_WHITE_LIST : JSON string array
// ENV_INPUT_CONTAINER_LOGGING_MAX_OPEN_FILES: int
// ENV_INPUT_CONTAINER_TAGS : "a=b,c=d"

//nolint:funlen
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

	// Deprecated
	if str, ok := envs["ENV_INPUT_CONTAINER_DISABLE_COLLECT_KUBE_JOB"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_DISABLE_COLLECT_KUBE_JOB to bool: %s, ignore", err)
		} else {
			ipt.EnableCollectK8sJob = !b
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_ENABLE_COLLECT_KUBE_JOB"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_COLLECT_KUBE_JOB to bool: %s, ignore", err)
		} else {
			ipt.EnableCollectK8sJob = b
		}
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_CONTAINER_MAX_CONCURRENT"]; ok {
		if size, err := strconv.ParseInt(str, 10, 64); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_CONTAINER_MAX_CONCURRENT to int64: %s, ignore", err)
		} else {
			ipt.ContainerMaxConcurrent = int(size)
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
	/// pod metric configs
	///
	if str, ok := envs["ENV_INPUT_CONTAINER_POD_INCLUDE_METRIC"]; ok {
		arrays := strings.Split(str, ",")
		ipt.PodIncludeMetric = arrays
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_POD_EXCLUDE_METRIC"]; ok {
		arrays := strings.Split(str, ",")
		ipt.PodExcludeMetric = arrays
	}

	whitelistStr := ""
	// This code is to maintain compatibility.
	if str, ok := envs["ENV_LOGGING_FIELD_WHITE_LIST"]; ok && str != "" {
		whitelistStr = str
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_FIELD_WHITE_LIST"]; ok && str != "" {
		whitelistStr = str
	}
	if whitelistStr != "" {
		if err := json.Unmarshal([]byte(whitelistStr), &ipt.LoggingFieldWhiteList); err != nil {
			l.Warnf("parse ENV_LOGGING_FIELD_WHITE_LIST/ENV_INPUT_CONTAINER_LOGGING_FIELD_WHITE_LIST to slice: %s, ignore", err)
		}
	}

	openfilesStr := ""
	// This code is to maintain compatibility.
	if str, ok := envs["ENV_LOGGING_MAX_OPEN_FILES"]; ok && str != "" {
		openfilesStr = str
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_MAX_OPEN_FILES"]; ok && str != "" {
		openfilesStr = str
	}
	if openfilesStr != "" {
		if limit, err := strconv.ParseInt(openfilesStr, 10, 64); err != nil {
			l.Warnf("parse ENV_LOGGING_MAX_OPEN_FILES/ENV_INPUT_CONTAINER_LOGGING_MAX_OPEN_FILES to int64: %s, ignore", err)
		} else {
			ipt.LoggingMaxOpenFiles = int(limit)
		}
	}

	///
	/// logging configs
	///
	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON"]; ok {
		if err := json.Unmarshal([]byte(str), &ipt.LoggingSourceMultilineMap); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON to map: %s, ignore", err)
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_ENABLE_MULTILINE"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_ENABLE_MULTILINE to bool: %s, ignore", err)
		} else {
			ipt.LoggingEnableMultline = b
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
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON to slice: %s, ignore", err)
		}
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_SEARCH_INTERVAL"]; ok {
		if dur, err := time.ParseDuration(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_SEARCH_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.LoggingSearchInterval = dur
		}
	}
	if str, ok := envs["ENV_INPUT_CONTAINER_LOGGING_FILE_FROM_BEGINNING"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_FILE_FROM_BEGINNING to bool: %s, ignore", err)
		} else {
			ipt.LoggingFileFromBeginning = b
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

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pushgateway

import (
	"strconv"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "RoutePrefix", Type: doc.String, Example: "`/v1/pushgateway`", Desc: "Prefix for the internal routes of web endpoints.", DescZh: "配置 endpoints 路由前缀"},
		{FieldName: "MeasurementName", Type: doc.String, Desc: `Set measurement name.`, DescZh: `配置指标集名称`},
		{FieldName: "JobAsMeasurement", Type: doc.Boolean, Default: "false", Desc: `Whether to use the job field for the measurement name.`, DescZh: `是否使用 job 标签值作为指标集名称`},
		{FieldName: "KeepExistMetricName", Type: doc.Boolean, Default: "true", Desc: `Whether to keep the raw field names for Prometheus, see [Kubernetes Prometheus doc](kubernetes-prom.md#measurement-and-tags)`, DescZh: `是否保留原始的 Prometheus 字段名，详见 [Kubernetes Prometheus doc](kubernetes-prom.md#measurement-and-tags)`},
	}
	return doc.SetENVDoc("ENV_INPUT_PUSHGATEWAY_", infos)
}

func (ipt *Input) ReadEnv(envs map[string]string) {
	if str, ok := envs["ENV_INPUT_PUSHGATEWAY_ROUTE_PREFIX"]; ok {
		ipt.RoutePrefix = str
	}

	if str, ok := envs["ENV_INPUT_PUSHGATEWAY_MEASUREMENT_NAME"]; ok {
		ipt.MeasurementName = str
	}

	if str, ok := envs["ENV_INPUT_PUSHGATEWAY_JOB_AS_MEASUREMENT"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			log.Warnf("parse ENV_INPUT_PUSHGATEWAY_JOB_AS_MEASUREMENT to bool: %s, ignore", err)
		} else {
			ipt.JobAsMeasurement = b
		}
	}

	if str, ok := envs["ENV_INPUT_PUSHGATEWAY_KEEP_EXIST_METRIC_NAME"]; ok {
		if b, err := strconv.ParseBool(str); err != nil {
			log.Warnf("parse ENV_INPUT_PUSHGATEWAY_KEEP_EXIST_METRIC_NAME to bool: %s, ignore", err)
		} else {
			ipt.KeepExistMetricName = b
		}
	}
}

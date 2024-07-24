// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var _ inputs.ReadEnv = &Input{}

func (ipt *Input) ReadEnv(m map[string]string) {
	ipt.initConf(m)
}

const (
	UseNowTimeInstead      = "ENV_USE_NOW_TIME_INSTEAD"
	EnableLogCollection    = "ENV_ENABLE_LOG_COLLECTION"
	EnableMetricCollection = "ENV_ENABLE_METRIC_COLLECTION"
)

func (ipt *Input) initConf(m map[string]string) {
	if v, ok := m[UseNowTimeInstead]; ok {
		ipt.UseNowTimeInstead = v == "true"
	}
	if v, ok := m[EnableLogCollection]; ok {
		ipt.EnableLogCollection = v != "false"
	}
	if v, ok := m[EnableMetricCollection]; ok {
		ipt.EnableMetricCollection = v != "false"
	}
}

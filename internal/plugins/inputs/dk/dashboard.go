// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dk

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (i *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			//nolint:lll
			"intro_content": `Datakit 所有指标的展示，需要开启更多的指标收集功能。详情参见[这里](https://docs.guance.com/datakit/datakit-metrics/)暴露的指标。`,
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
			"intro_content": `This dashboard showing all metrics about Datakit. For more metrics, see [here](https://docs.guance.com/en/datakit/datakit-metrics/).`,
		}
	default:
		return nil
	}
}

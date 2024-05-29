// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package swap

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"fifteen_min_load": "15分钟负载",
			"host_name":        "主机名",
			"title":            "Swap 监控视图",
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"fifteen_min_load": "15-minutes Load",
			"host_name":        "Host Name",
			"title":            "Swap Monitor View",
		}
	default:
		return nil
	}
}

func (*Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			//nolint:lll
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
		}
	default:
		return nil
	}
}

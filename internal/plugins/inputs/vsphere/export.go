// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package vsphere

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"title": "vSphere 监控视图",
		}

	case inputs.I18nEn:
		return map[string]string{
			"title": "vSphere Monitor View",
		}
	default:
		return nil
	}
}

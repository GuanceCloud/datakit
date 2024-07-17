// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oceanbase

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

var (
	_ inputs.Dashboard = (*Input)(nil)
	_ inputs.Monitor   = (*Input)(nil)
)

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"left_curly":   "{{",
			"right_curly":  "}}",
			"tenant_name":  "租户名称",
			"cluster_name": "集群名称",
			"title":        "Oceanbase 监控视图",
		}
	case inputs.I18nEn:
		return map[string]string{
			"left_curly":   "{{",
			"right_curly":  "}}",
			"tenant_name":  "tenant",
			"cluster_name": "cluster",
			"title":        "Oceanbase Monitor View",
		}
	default:
		return nil
	}
}

func (*Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"message": "[oceanbase_status_exception] OceanBase 数据库状态异常，请检查",
			"title":   "oceanbase_status_exception",
		}
	case inputs.I18nEn:
		return map[string]string{
			"message": "[oceanbase_status_exception] Abnormal OceanBase database status, please check",
			"title":   "oceanbase_status_exception",
		}
	default:
		return nil
	}
}

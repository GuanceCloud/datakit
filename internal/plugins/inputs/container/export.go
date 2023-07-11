// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (i *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			//nolint:lll
			// TODO
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
			// TODO
		}
	default:
		return nil
	}
}

func (i *Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			//nolint:lll
			// TODO
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
			// TODO
		}
	default:
		return nil
	}
}

func (i *Input) MonitorList() []string {
	return nil
}

func (i *Input) DashboardList() []string {
	return []string{
		"kubernetes",
		"kubernetes_events",
		"kubernetes_nodes_overview",
		"kubernetes_pods_overview",
		"kubernetes_service",
	}
}

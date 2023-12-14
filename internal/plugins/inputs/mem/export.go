// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mem

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"memory_usage":         "Memory usage",
			"memory_buffered":      "Memory buffered",
			"memory_buffered_note": "仅支持 Linux",
			"memory_total":         "Memory total",
			"memory_cached":        "Memory cached",
			"memory_cached_note":   "仅支持 Linux",
			"memory_free":          "Memory free",
			"memory_free_note":     "仅支持 Darwin/Linux",
			"memory_used":          "Memory used",
			"memory_available":     "Memory available",
			"memory_shared":        "Memory shared",
			"memory_shared_note":   "仅支持 Linux",
			"memory_active":        "Memory active",
			"memory_active_note":   "仅支持 Darwin/Linux",
			"load_15_minute":       "15分钟负载",
			"host_name":            "主机名",
		}
	case inputs.I18nEn:
		return map[string]string{
			"memory_usage":         "Memory usage",
			"memory_buffered":      "Memory buffered",
			"memory_buffered_note": "Only for Linux",
			"memory_total":         "Memory total",
			"memory_cached":        "Memory cached",
			"memory_cached_note":   "Only for Linux",
			"memory_free":          "Memory free",
			"memory_free_note":     "Only for Darwin/Linux",
			"memory_used":          "Memory used",
			"memory_available":     "Memory available",
			"memory_shared":        "Memory shared",
			"memory_shared_note":   "Only for Linux",
			"memory_active":        "Memory active",
			"memory_active_note":   "Only for Darwin/Linux",
			"load_15_minute":       "15 minute load",
			"host_name":            "Host name",
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

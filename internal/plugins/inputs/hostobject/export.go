// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"x_title_cpu":     "CPU",
			"title_cpu_usage": "CPU 使用率",
			"title_cpu_load":  "CPU 负载",

			"desc_win_some_cpu_metric_zero": "部分 CPU 指标在 Windows 中不存在",

			"x_title_mem":                   "内存",
			"title_mem_usage_pct":           "内存使用率",
			"title_mem_used_bytes":          "内存用量",
			"desc_win_some_mem_metric_miss": "部分内存指标在 Windows 中不存在",

			"x_title_disk":                "磁盘",
			"title_disk_rw_duration":      "磁盘读写耗时",
			"title_disk_rw_bytes_per_sec": "每秒读写字节数",
			"title_disk_used":             "磁盘用量",
			"desc_linux_only":             "仅 Linux 支持",

			"x_title_net":           "网络",
			"title_network_traffic": "网络流量/sec",
			"title_network_packets": "网络数据包/sec",
		}
	case inputs.I18nEn:
		return map[string]string{
			"x_title_cpu":     "CPU",
			"title_cpu_usage": "CPU usage",
			"title_cpu_load":  "CPU load",

			"desc_win_some_cpu_metric_zero": "Some CPU metrics not available under Windows",

			"x_title_mem":                   "Memory",
			"title_mem_usage_pct":           "Memory usage",
			"title_mem_used_bytes":          "Memory used bytes",
			"desc_win_some_mem_metric_miss": "Some memory metrics not available under Windows",

			"x_title_disk":                "Disk",
			"title_disk_rw_duration":      "R/W cost",
			"title_disk_rw_bytes_per_sec": "R/W bytes/sec",
			"title_disk_used":             "Usage",
			"desc_linux_only":             "Only for Linux",

			"x_title_net":           "Network",
			"title_network_traffic": "Traffic/sec",
			"title_network_packets": "Packetes/sec",
		}
	default:
		return nil
	}
}

func (ipt *Input) Monitor(lang inputs.I18n) map[string]string {
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

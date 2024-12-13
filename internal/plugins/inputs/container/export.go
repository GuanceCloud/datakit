// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			// for single pod
			"spod_x_title_cpu":                  "CPU",
			"spod_title_node_cpu_usage_base100": "CPU 使用率(node)",

			"spod_title_base_100_cpu_usage": "CPU 使用率",
			"spod_title_per_cpu_usage":      "CPU 总使用率",
			"spod_title_cpu_limit":          "CPU 限额（核心数）",

			"spod_desc_node_cpu_usage_base100": "基于 node 的 CPU 使用率（0~100）",
			"spod_desc_base_100_cpu_usage":     "基于 limit 的 CPU 使用率（0~100）",
			"spod_desc_per_cpu_usage":          "基于 limit 的 CPU 使用率 (0~100*cores)",

			"spod_x_title_net":    "网络",
			"spod_title_net_recv": "接收字节数",
			"spod_title_net_sent": "发送字节数",

			"spod_x_title_mem":     "内存",
			"spod_title_mem_usage": "内存用量",
			"spod_title_mem_limit": "内存限量",

			"spod_x_title_storage":             "Pod 磁盘",
			"spod_title_pod_used_storage":      "磁盘用量",
			"spod_title_pod_available_storage": "磁盘可用量",

			// for single container
			"scont_x_title_cpu":      "CPU",
			"scont_title_cpu_usage":  "CPU 使用率",
			"scont_title_cpu_limit":  "CPU 限额",
			"scont_desc_cpu_base100": "0~100%",

			"scont_x_title_mem":     "内存",
			"scont_title_mem_usage": "内存使用率",
			"scont_title_mem_used":  "内存用量",
			"scont_title_mem_limit": "内存限额",
			"scont_desc_mem_usage":  "0~100%",

			"scont_x_title_disk":     "磁盘",
			"scont_title_disk_read":  "读取",
			"scont_title_disk_write": "写入",

			"scont_x_title_net":    "网络",
			"scont_title_net_recv": "接收",
			"scont_title_net_sent": "发送",

			// for single service
			"ss_title_service_port": "端口数",

			// others...
		}
	case inputs.I18nEn:
		return map[string]string{
			// for single pod
			"spod_x_title_cpu":                  "CPU",
			"spod_title_node_cpu_usage_base100": "CPU usage(node)",
			"spod_title_base_100_cpu_usage":     "CPU usage",
			"spod_title_per_cpu_usage":          "CPU total usage",
			"spod_title_cpu_limit":              "CPU limited(cores)",

			"spod_desc_node_cpu_usage_base100": "CPU usage base on all node CPUs (0~100)",
			"spod_desc_base_100_cpu_usage":     "CPU usage based on limit (0~100)",
			"spod_desc_per_cpu_usage":          "CPU usage based on limit (0~100*cores)",

			"spod_x_title_net":    "Network",
			"spod_title_net_recv": "Recv",
			"spod_title_net_sent": "Send",

			"spod_x_title_mem":     "Memory",
			"spod_title_mem_usage": "Used",
			"spod_title_mem_limit": "Limit",

			"spod_x_title_storage":             "Disk",
			"spod_title_pod_used_storage":      "Used",
			"spod_title_pod_available_storage": "Available",

			// for single container
			"scont_x_title_cpu":      "CPU",
			"scont_title_cpu_usage":  "Usage",
			"scont_title_cpu_limit":  "Limit",
			"scont_desc_cpu_base100": "0~100%",

			"scont_x_title_mem":     "Memory",
			"scont_title_mem_usage": "Usage",
			"scont_title_mem_used":  "Used",
			"scont_title_mem_limit": "Limit",
			"scont_desc_mem_usage":  "0~100%",

			"scont_x_title_disk":     "Disk",
			"scont_title_disk_read":  "Read",
			"scont_title_disk_write": "Write",

			"scont_x_title_net":    "Network",
			"scont_title_net_recv": "Receive",
			"scont_title_net_sent": "Send",

			// for single service
			"ss_title_service_port": "Ports",

			// others...
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

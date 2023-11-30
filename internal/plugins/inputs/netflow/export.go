// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netflow

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	_ inputs.Dashboard = (*Input)(nil)
	_ inputs.Monitor   = (*Input)(nil)
)

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"alarm":                  "NetFlow 告警面板",
			"category_conversations": "网络会话",
			"category_detail":        "NetFlow 详情",
			"category_netflow":       "NetFlow 自身监测",
			"category_overview":      "Netflow 告警",
			"category_ports":         "源 IP/目标 IP 端口",
			"dest_device":            "以目标 IP 和设备 IP 为组的流量排行",
			"dest_port_ip_device":    "以目标端口、目标 IP 和设备 IP 为组的流量排行",
			"interval_increment":     "在指定间隔内 NetFlow 捕捉到的流量增长量",
			"message":                "NetFlow 流详情",
			"netflow_type":           "NetFlow 协议类型",
			"network_protocol":       "网络协议",
			"src_device_dest":        "以源 IP、设备 IP 和目标 IP 为组的流量排行",
			"src_device":             "以源 IP 和设备 IP 为组的流量排行",
			"src_port_ip_device":     "以源端口、源 IP 和设备 IP 为组的流量排行",
			"total":                  "NetFlow 捕捉到的总流量",
		}
	case inputs.I18nEn:
		return map[string]string{
			"alarm":                  "NetFlow Alarm Panel",
			"category_conversations": "Network Conversations",
			"category_detail":        "NetFlow Detail",
			"category_netflow":       "NetFlow Self",
			"category_overview":      "Netflow Alarm",
			"category_ports":         "Source/Destination Ports",
			"dest_device":            "Top bytes ranking by destination IP and device IP",
			"dest_port_ip_device":    "Top bytes ranking by destination port, destination IP and device IP",
			"interval_increment":     "Increment bytes captured by NetFlow in specified time interval",
			"message":                "NetFlow flow detail",
			"netflow_type":           "NetFlow type",
			"network_protocol":       "Network protocol",
			"src_device_dest":        "Top bytes ranking by source IP, device IP and destination IP",
			"src_device":             "Top bytes ranking by source IP and device IP",
			"src_port_ip_device":     "Top bytes ranking by source port, source IP and device IP",
			"total":                  "Total bytes captured by NetFlow",
		}
	default:
		return nil
	}
}

func (*Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			"message": "未检测到网络信号, 请检查:\\n1. 网络连接状态是否正常;\\n2. NetFlow 是否开启;",
			"title":   "no_network",
		}
	case inputs.I18nEn:
		return map[string]string{
			"message": "No network signal is detected, please check: \\n1. If the network connection status is normal; \\n2. If NetFlow is enabled;",
			"title":   "no_network",
		}
	default:
		return nil
	}
}

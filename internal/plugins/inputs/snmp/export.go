// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package snmp

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

//nolint:lll
func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return getMapping(true)
	case inputs.I18nEn:
		return getMapping(false)
	default:
		return nil
	}
}

func (ipt *Input) DashboardList() []string {
	return nil
}

func getMapping(zh bool) map[string]string {
	out := make(map[string]string)
	for k, v := range templateNamingMapping {
		if zh {
			out[k] = v[1]
		} else {
			out[k] = v[0]
		}
	}
	return out
}

//nolint:lll
var templateNamingMapping = map[string]([2]string){
	"Device_Alive": {"Device Alive", "活跃设备数"},
	"Uptime":       {"Uptime", "启动时长"},
	"Power":        {"Power", "电源"},
	"Fan":          {"Fan", "风扇"},
	"Temperature":  {"Temperature", "温度"},
	"CPU":          {"CPU", "CPU"},
	"Memory":       {"Memory", "内存"},
	"Disk":         {"Disk", "磁盘"},
	"Net":          {"Net", "网络"},
	"Interface":    {"Interface", "接口"},
	"Discards":     {"Discards", "丢包数"},
	"Errors":       {"Errors", "错误包数"},
	"Packets":      {"Packets", "数据包数"},
	"Bytes":        {"Bytes", "数据字节数"},
}

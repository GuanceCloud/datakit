// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mem

import (
	"github.com/shirou/gopsutil/mem"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type VMStat func() (*mem.VirtualMemoryStat, error)

func VirtualMemoryStat() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

// NewFieldInfoB new ibyte data type.
func NewFieldInfoB(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

// NewFieldInfoP new precent filed.
func NewFieldInfoP(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Float,
		Unit:     inputs.Percent,
		Desc:     desc,
	}
}

// NewFieldInfoC new count field.
func NewFieldInfoC(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

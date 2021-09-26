package mem

import (
	"github.com/shirou/gopsutil/mem"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type VMStat func() (*mem.VirtualMemoryStat, error)

func VirtualMemoryStat() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

// ibyte data type.
func NewFieldInfoB(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.SizeIByte,
		Desc:     desc,
	}
}

// precent.
func NewFieldInfoP(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Float,
		Unit:     inputs.Percent,
		Desc:     desc,
	}
}

// count.
func NewFieldInfoC(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.Count,
		Desc:     desc,
	}
}

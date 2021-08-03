package system

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}

func (m *measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

type conntrackMeasurement measurement

func (m *conntrackMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *conntrackMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameConntrack,
		Desc: "系统网络连接指标（仅 Linux 支持）",
		Fields: map[string]interface{}{
			"entries":             newFieldInfoCount("当前连接数量"),
			"entries_limit":       newFieldInfoCount("连接跟踪表的大小"),
			"stat_found":          newFieldInfoCount("成功的搜索条目数目"),
			"stat_invalid":        newFieldInfoCount("不能被跟踪的包数目"),
			"stat_ignore":         newFieldInfoCount("已经被跟踪的报数目"),
			"stat_insert":         newFieldInfoCount("插入的包数目"),
			"stat_insert_failed":  newFieldInfoCount("插入失败的包数目"),
			"stat_drop":           newFieldInfoCount("跟踪失败被丢弃的包数目"),
			"stat_early_drop":     newFieldInfoCount("由于跟踪表满而导致部分已跟踪包条目被丢弃的数目"),
			"stat_search_restart": newFieldInfoCount("由于hash表大小修改而导致跟踪表查询重启的数目"),
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "主机名"},
		},
	}
}

type filefdMeasurement measurement

func (m *filefdMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *filefdMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameFilefd,
		Desc: "系统文件句柄指标（仅 Linux 支持）",
		Fields: map[string]interface{}{
			"allocated": newFieldInfoCount("已分配文件句柄的数目"),
			// "maximum":      newFieldInfoCount("文件句柄的最大数目； 当值为 2^63-1 则页面显示 9223372036854776000(若大于此值，均会造成精度损失)"),
			"maximum_mega": newFieldInfoMega("文件句柄的最大数目, 单位 M(10^6)"),
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "主机名"},
		},
	}
}

type systemMeasurement measurement

func (m *systemMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameSystem,
		Desc: "系统运行基础信息",
		Fields: map[string]interface{}{
			"load1":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "过去 1 分钟的 CPU 平均负载"},
			"load5":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "过去 5 分钟的 CPU 平均负载"},
			"load15":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "过去 15 分钟的 CPU 平均负载"},
			"load1_per_core":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "每个核心过去 1 分钟的 CPU 平均负载"},
			"load5_per_core":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "每个核心过去 5 分钟的 CPU 平均负载"},
			"load15_per_core": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "每个核心过去 15 分钟的 CPU 平均负载"},
			"n_cpus":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "CPU 逻辑核心数"},
			"n_users":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "用户数"},
			"uptime":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "系统运行时间"},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "主机名"},
		},
	}
}

func (m *systemMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func newFieldInfoCount(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func newFieldInfoMega(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Float,
		Unit:     inputs.Mega,
		Desc:     desc,
	}
}

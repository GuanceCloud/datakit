package process

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "host_processes"
	category  = "host"

	sampleConfig = `
[[inputs.host_processes]]
  ## process name support regexp
  # process_name = [".*datakit.*"]
  ## process min run time default 10m,Collection  the process of running more than ten minutes
  min_run_time     = "10m"
  ## open collection metric
  open_metric = false
  ## pipeline path
  # pipeline = ""
  [inputs.host_processes.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
`

	pipelineSample = ``
)

type ProcessMetric struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *ProcessMetric) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *ProcessMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Desc: "采集进程指标数据,包括cpu内存使用率等",
		Type: "metric",
		Fields: map[string]interface{}{
			"threads":          newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "线程数"),
			"rss":              newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.SizeIByte, "Resident Set Size （常驻内存大小）"),
			"cpu_usage":        newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "cpu使用占比（%*100），进程自启动以来所占 CPU 百分比，该值相对会比较稳定（跟 top 的瞬时百分比不同）"),
			"mem_used_percent": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "mem使用占比（%*100）"),
			"open_files":       newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "open_files 个数(仅支持linux)"),
		},
		Tags: map[string]interface{}{
			"username":     inputs.NewTagInfo("用户名"),
			"host":         inputs.NewTagInfo("主机名"),
			"process_name": inputs.NewTagInfo("进程名"),
			"pid":          inputs.NewTagInfo("进程id"),
		},
	}
}

type ProcessObject struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *ProcessObject) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *ProcessObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Desc: "采集进程对象的数据，包括进程名，cmd等",
		Type: "object",
		Fields: map[string]interface{}{
			"message":          newOtherFieldInfo(inputs.String, inputs.Gauge, inputs.UnknownUnit, "进程详细信息"),
			"start_time":       newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.DurationMS, "进程启动时间"),
			"threads":          newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "线程数"),
			"rss":              newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.SizeIByte, "Resident Set Size （常驻内存大小）"),
			"pid":              newOtherFieldInfo(inputs.Int, inputs.UnknownType, inputs.UnknownUnit, "进程id"),
			"cpu_usage":        newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "cpu使用占比（%*100），进程自启动以来所占 CPU 百分比，该值相对会比较稳定（跟 top 的瞬时百分比不同）"),
			"mem_used_percent": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "mem使用占比（%*100）"),
			"open_files":       newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "open_files 个数(仅支持linux)"),
			"work_directory":   newOtherFieldInfo(inputs.String, inputs.Gauge, inputs.UnknownUnit, "工作目录(仅支持linux)"),
			"cmdline":          newOtherFieldInfo(inputs.String, inputs.Gauge, inputs.UnknownUnit, "进程的命令行参数"),
			"state_zombie":     newOtherFieldInfo(inputs.Bool, inputs.Gauge, inputs.UnknownUnit, "是否是僵尸进程"),
		},
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("name 字段，由 host_pid 组成"),
			"class":        inputs.NewTagInfo("分类: host_processes"),
			"username":     inputs.NewTagInfo("用户名"),
			"host":         inputs.NewTagInfo("主机名"),
			"state":        inputs.NewTagInfo("进程状态，暂不支持 windows"),
			"process_name": inputs.NewTagInfo("进程名"),
		},
	}
}

func newOtherFieldInfo(datatype, ftype, unit, desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: datatype,
		Type:     ftype,
		Unit:     unit,
		Desc:     desc,
	}
}

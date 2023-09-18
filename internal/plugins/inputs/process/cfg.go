// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package process

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName = "host_processes"
	category  = "host"

	sampleConfig = `
[[inputs.host_processes]]
  # Only collect these matched process' metrics. For process objects
  # these white list not applied. Process name support regexp.
  # process_name = [".*nginx.*", ".*mysql.*"]

  # Process minimal run time(default 10m)
  # If process running time less than the setting, we ignore it(both for metric and object)
  min_run_time = "10m"

  ## Enable process metric collecting
  open_metric = false

  ## Enable listen ports tag, default is false
  enable_listen_ports = false

  ## Enable open files field, default is false
  enable_open_files = false

  # Extra tags
  [inputs.host_processes.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
`
)

type ProcessMetric struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *ProcessMetric) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *ProcessMetric) LineProto() (*dkpt.Point, error) {
	// return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
	return nil, fmt.Errorf("not implement")
}

//nolint:lll
func (m *ProcessMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Desc: "采集进程指标数据，包括 CPU/内存使用率等",
		Type: "metric",
		Fields: map[string]interface{}{
			"threads":          newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "线程数"),
			"rss":              newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.SizeByte, "Resident Set Size （常驻内存大小）"),
			"cpu_usage":        newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "CPU 使用占比，进程自启动以来所占 CPU 百分比，该值相对会比较稳定（跟 `top` 的瞬时百分比不同）"),
			"cpu_usage_top":    newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "CPU 使用占比，一个采集周期内的进程的 CPU 使用率均值"),
			"mem_used_percent": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "内存使用占比"),
			"open_files":       newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "打开文件个数(仅支持 Linux, 且需开启 `enable_open_files` 选项)"),
		},
		Tags: map[string]interface{}{
			"username":     inputs.NewTagInfo("用户名"),
			"host":         inputs.NewTagInfo("主机名"),
			"process_name": inputs.NewTagInfo("进程名"),
			"pid":          inputs.NewTagInfo("进程 ID"),
		},
	}
}

type ProcessObject struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *ProcessObject) Point() *point.Point {
	opts := point.DefaultObjectOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *ProcessObject) LineProto() (*dkpt.Point, error) {
	// return point.NewPoint(m.name, m.tags, m.fields, point.OOpt())
	return nil, fmt.Errorf("not implement")
}

//nolint:lll
func (m *ProcessObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Desc: "采集进程对象的数据，包括进程名，进程命令等",
		Type: "object",
		Fields: map[string]interface{}{
			"message":          newOtherFieldInfo(inputs.String, inputs.Gauge, inputs.UnknownUnit, "进程详细信息"),
			"start_time":       newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.TimestampMS, "进程启动时间"),
			"started_duration": newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.TimestampSec, "进程启动时长"),
			"threads":          newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "线程数"),
			"rss":              newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.SizeByte, "Resident Set Size （常驻内存大小）"),
			"pid":              newOtherFieldInfo(inputs.Int, inputs.UnknownType, inputs.UnknownUnit, "进程 ID"),
			"cpu_usage":        newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "CPU 使用占比（%*100），进程自启动以来所占 CPU 百分比，该值相对会比较稳定（跟 `top` 的瞬时百分比不同）"),
			"cpu_usage_top":    newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "CPU 使用占比（%*100）, 一个采集周期内的进程的 CPU 使用率均值"),
			"mem_used_percent": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "内存使用占比（%*100）"),
			"open_files":       newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "打开的文件个数(仅支持 Linux, 且需开启 `enable_open_files` 选项)"),
			"work_directory":   newOtherFieldInfo(inputs.String, inputs.Gauge, inputs.UnknownUnit, "工作目录(仅支持 Linux)"),
			"cmdline":          newOtherFieldInfo(inputs.String, inputs.Gauge, inputs.UnknownUnit, "进程的命令行参数"),
			"state_zombie":     newOtherFieldInfo(inputs.Bool, inputs.Gauge, inputs.UnknownUnit, "是否是僵尸进程"),
		},
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("name 字段，由 `[host-name]_[pid]` 组成"),
			"class":        inputs.NewTagInfo("固定为 `host_processes`"),
			"username":     inputs.NewTagInfo("用户名"),
			"host":         inputs.NewTagInfo("主机名"),
			"state":        inputs.NewTagInfo("进程状态，暂不支持 Windows"),
			"process_name": inputs.NewTagInfo("进程名"),
			"listen_ports": inputs.NewTagInfo("进程正在监听的端口。对应配置文件的 `enable_listen_ports`，默认为 false，不携带此字段"),
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

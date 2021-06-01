package container

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	containerName = "container"
)

type containersMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (c *containersMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(c.name, c.tags, c.fields, c.ts)
}

func (c *containersMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: containerName,
		Desc: "容器指标或对象数据",
		Tags: map[string]interface{}{
			"container_id":   inputs.NewTagInfo(`容器 ID`),
			"container_name": inputs.NewTagInfo(`容器名称`),
			"images_name":    inputs.NewTagInfo(`容器镜像名称`),
			"docker_image":   inputs.NewTagInfo(`镜像名称+版本号`),
			"name":           inputs.NewTagInfo(`对象数据的指定 ID，（仅在对象数据中存在）`),
			"container_host": inputs.NewTagInfo(`容器内部的主机名（仅在对象数据中存在）`),
			"host":           inputs.NewTagInfo(`容器宿主机的主机名`),
			"state":          inputs.NewTagInfo(`运行状态，running/exited/removed`),
			"pod_name":       inputs.NewTagInfo(`pod名称`),
			"pod_namesapce":  inputs.NewTagInfo(`pod命名空间`),
			// "kube_container_name": inputs.NewTagInfo(`TODO`),
			// "kube_daemon_set":     inputs.NewTagInfo(`TODO`),
			// "kube_deployment":     inputs.NewTagInfo(`TODO`),
			// "kube_namespace":      inputs.NewTagInfo(`TODO`),
			// "kube_ownerref_kind":  inputs.NewTagInfo(`TODO`),
		},
		Fields: map[string]interface{}{
			"from_kubernetes":    &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "该容器是否由 Kubernetes 创建"},
			"cpu_usage":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "CPU 占主机总量的使用率"},
			"cpu_delta":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "容器 CPU 增量"},
			"cpu_system_delta":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "系统 CPU 增量"},
			"cpu_numbers":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "CPU 核心数"},
			"message":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "容器对象详情，（仅在对象数据中存在）"},
			"mem_limit":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "内存可用总量，如果未对容器做内存限制，则为主机内存容量"},
			"mem_usage":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "内存使用量"},
			"mem_used_percent":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "内存使用率，使用量除以可用总量"},
			"mem_failed_count":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "内存分配失败的次数"},
			"network_bytes_rcvd": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "从网络接收到的总字节数"},
			"network_bytes_sent": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "向网络发送出的总字节数"},
			"block_read_byte":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "从容器文件系统读取的总字节数"},
			"block_write_byte":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "向容器文件系统写入的总字节数"},
		},
	}
}

type containersLogMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (c *containersLogMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(c.name, c.tags, c.fields, c.ts)
}

func (c *containersLogMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "日志数据",
		Desc: "默认使用容器名，如果容器名能匹配 `log_option.container_name_match` 正则，则使用对应的 `source` 字段值",
		Tags: map[string]interface{}{
			"container_name": inputs.NewTagInfo(`容器名称`),
			"container_id":   inputs.NewTagInfo(`容器ID`),
			"image_name":     inputs.NewTagInfo(`容器镜像名称`),
			"stream":         inputs.NewTagInfo(`数据流方式，stdout/stderr/tty`),
		},
		Fields: map[string]interface{}{
			"from_kubernetes": &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "该容器是否由 Kubernetes 创建"},
			"service":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "服务名称"},
			"status":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志状态，info/emerg/alert/critical/error/warning/debug/OK"},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志源数据"},
		},
	}
}

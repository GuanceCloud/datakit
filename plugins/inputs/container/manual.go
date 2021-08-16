package container

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	containerName  = "docker_containers"
	kubeletPodName = "kubelet_pod"
)

type containerMetricMeasurement struct{}

func (c *containerMetricMeasurement) LineProto() (*io.Point, error) { return nil, nil }

func (c *containerMetricMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: containerName,
		Desc: "容器指标数据（忽略 k8s pause 容器），只采集正在运行的容器",
		Tags: map[string]interface{}{
			"container_id":      inputs.NewTagInfo(`容器 ID（该字段默认被删除）`),
			"container_name":    inputs.NewTagInfo(`容器名称`),
			"docker_image":      inputs.NewTagInfo("镜像全称，例如 `nginx.org/nginx:1.21.0`"),
			"images_name":       inputs.NewTagInfo("镜像名称，例如 `nginx.org/nginx`"),
			"images_short_name": inputs.NewTagInfo("镜像名称精简版，例如 `nginx`"),
			"images_tag":        inputs.NewTagInfo("镜像 tag，例如 `1.21.0`"),
			"container_type":    inputs.NewTagInfo(`容器类型，表明该容器由谁创建，kubernetes/docker`),
			"state":             inputs.NewTagInfo(`运行状态，running`),
			"pod_name":          inputs.NewTagInfo(`pod 名称（容器由 k8s 创建时存在）`),
			"pod_namesapce":     inputs.NewTagInfo(`pod 命名空间（容器由 k8s 创建时存在）`),
			"deployment":        inputs.NewTagInfo(`deployment 名称（容器由 k8s 创建时存在）`),
		},
		Fields: map[string]interface{}{
			"cpu_usage":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "CPU 占主机总量的使用率"},
			"cpu_delta":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "容器 CPU 增量"},
			"cpu_system_delta":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "系统 CPU 增量"},
			"cpu_numbers":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "CPU 核心数"},
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

type containerObjectMeasurement struct{}

func (c *containerObjectMeasurement) LineProto() (*io.Point, error) { return nil, nil }

func (c *containerObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: containerName,
		Desc: "容器对象数据（忽略 k8s pause 容器），如果容器处于非 running 状态，则`cpu_usage`等指标将不存在",
		Tags: map[string]interface{}{
			"container_id":      inputs.NewTagInfo(`容器 ID（该字段默认被删除）`),
			"name":              inputs.NewTagInfo(`对象数据的指定 ID`),
			"status":            inputs.NewTagInfo("容器状态，例如 `Up 5 hours`"),
			"container_name":    inputs.NewTagInfo(`容器名称`),
			"docker_image":      inputs.NewTagInfo("镜像全称，例如 `nginx.org/nginx:1.21.0`"),
			"images_name":       inputs.NewTagInfo("镜像名称，例如 `nginx.org/nginx`"),
			"images_short_name": inputs.NewTagInfo("镜像名称精简版，例如 `nginx`"),
			"images_tag":        inputs.NewTagInfo("镜像tag，例如 `1.21.0`"),
			"container_host":    inputs.NewTagInfo(`容器内部的主机名`),
			"container_type":    inputs.NewTagInfo(`容器类型，表明该容器由谁创建，kubernetes/docker`),
			"state":             inputs.NewTagInfo(`运行状态，running/exited/removed`),
			"pod_name":          inputs.NewTagInfo(`pod 名称（容器由 k8s 创建时存在）`),
			"pod_namesapce":     inputs.NewTagInfo(`pod 命名空间（容器由 k8s 创建时存在）`),
			"deployment":        inputs.NewTagInfo(`deployment 名称（容器由 k8s 创建时存在）`),
		},
		Fields: map[string]interface{}{
			"process":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "容器进程列表，即运行命令`ps -ef`所得，内容为 JSON 字符串，格式是 map 数组"},
			"age":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: `该容器创建时长，单位秒`},
			"from_kubernetes":    &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "该容器是否由 Kubernetes 创建（deprecated）"},
			"cpu_usage":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "CPU 占主机总量的使用率"},
			"cpu_delta":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "容器 CPU 增量"},
			"cpu_system_delta":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "系统 CPU 增量"},
			"cpu_numbers":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "CPU 核心数"},
			"message":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "容器对象详情"},
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

type containerLogMeasurement struct{}

func (c *containerLogMeasurement) LineProto() (*io.Point, error) { return nil, nil }

func (c *containerLogMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "日志数据",
		Desc: "默认使用容器名，如果容器名能匹配 `log_option.container_name_match` 正则，则使用对应的 `source` 字段值（只采集正在运行的容器日志且忽略 k8s pause 容器）",
		Tags: map[string]interface{}{
			"container_name": inputs.NewTagInfo(`容器名称`),
			"container_id":   inputs.NewTagInfo(`容器ID`),
			"container_type": inputs.NewTagInfo(`容器类型，表明该容器由谁创建，kubernetes/docker`),
			"stream":         inputs.NewTagInfo(`数据流方式，stdout/stderr/tty`),
			"pod_name":       inputs.NewTagInfo(`pod 名称（容器由 k8s 创建时存在）`),
			"pod_namesapce":  inputs.NewTagInfo(`pod 命名空间（容器由 k8s 创建时存在）`),
			"deployment":     inputs.NewTagInfo(`deployment 名称（容器由 k8s 创建时存在）`),
			"service":        inputs.NewTagInfo(`服务名称`),
		},
		Fields: map[string]interface{}{
			"status":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志状态，info/emerg/alert/critical/error/warning/debug/OK"},
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志源数据"},
		},
	}
}

type kubeletPodMetricMeasurement struct{}

func (k *kubeletPodMetricMeasurement) LineProto() (*io.Point, error) { return nil, nil }

func (k *kubeletPodMetricMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubeletPodName,
		Desc: "kubelet pod 指标数据（只采集正在运行的 pod）",
		Tags: map[string]interface{}{
			"node_name": inputs.NewTagInfo(`所在 kubelet node 名字`),
			"pod_name":  inputs.NewTagInfo(`pod 名称`),
			"namespace": inputs.NewTagInfo(`pod 所属命名空间`),
		},
		Fields: map[string]interface{}{
			"cpu_usage":                  &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage of cpu used"},
			"mem_usage_percent":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage of memory used"},
			"cpu_usage_nanocores":        &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The number of cpu usage nanocores"},
			"cpu_usage_core_nanoseconds": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The number of cpu usage core nanoseconds"},
			"memory_available_bytes":     &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "The number of memory available in bytes"},
			"memory_usage_bytes":         &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "The number of memory used in bytes"},
			"memory_working_set_bytes":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "Current working set in bytes "},
			"memory_rss_bytes":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "Size of RSS in bytes"},
			"memory_page_faults":         &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The count of memory page faults"},
			"memory_major_page_faults":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The count of memory major page faults"},
			"network_rx_bytes":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "The number of bytes per second received"},
			"network_rx_errors":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The number of rx errors per second"},
			"network_tx_bytes":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "The number of bytes per second transmitted"},
			"network_tx_errors":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The number of tx errors per second"},
		},
	}
}

type kubeletPodObjectMeasurement struct{}

func (k *kubeletPodObjectMeasurement) LineProto() (*io.Point, error) { return nil, nil }

func (k *kubeletPodObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubeletPodName,
		Desc: "kubelet pod 对象数据，如果 pod 处于非 Running 状态，则`cpu_usage`等指标将不存在",
		Tags: map[string]interface{}{
			"node_name": inputs.NewTagInfo(`所在 kubelet node 名字`),
			"name":      inputs.NewTagInfo(`pod UID`),
			"pod_name":  inputs.NewTagInfo(`pod 名称`),
			"namespace": inputs.NewTagInfo(`pod 所属命名空间`),
			"ready":     inputs.NewTagInfo(`可用副本数，就绪个数/期望个数，例如 "1/1"`),
			"state":     inputs.NewTagInfo(`当前阶段的状态，Running/Failed/Pending`),
			"labels":    inputs.NewTagInfo(`pod labels，格式为 JSON 字符串`),
		},
		Fields: map[string]interface{}{
			"age":                        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: `该 pod start 至今的时长，单位秒`},
			"restart":                    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "pod 重启次数"},
			"message":                    &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "pod 对象详情"},
			"cpu_usage":                  &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage of cpu used"},
			"mem_usage_percent":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage of memory used"},
			"cpu_usage_nanocores":        &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The number of cpu usage nanocores"},
			"cpu_usage_core_nanoseconds": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The number of cpu usage core nanoseconds"},
			"memory_available_bytes":     &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "The number of memory available in bytes"},
			"memory_usage_bytes":         &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "The number of memory used in bytes"},
			"memory_working_set_bytes":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "Current working set in bytes "},
			"memory_rss_bytes":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "Size of RSS in bytes"},
			"memory_page_faults":         &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The count of memory page faults"},
			"memory_major_page_faults":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The count of memory major page faults"},
			"network_rx_bytes":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "The number of bytes per second received"},
			"network_rx_errors":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The number of rx errors per second"},
			"network_tx_bytes":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "The number of bytes per second transmitted"},
			"network_tx_errors":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The number of tx errors per second"},
		},
	}
}

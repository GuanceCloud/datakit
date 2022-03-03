package container

import (
	"context"
	"encoding/json"
	"time"

	"github.com/docker/docker/api/types"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func gatherDockerContainerObject(client dockerClientX, k8sClient k8sClientX, container *types.Container) (*containerObject, error) {
	m := &containerObject{}
	m.tags = getContainerInfo(container, k8sClient)
	m.tags["name"] = container.ID
	m.tags["status"] = container.Status

	if hostname, err := getContainerHostname(client, container.ID); err == nil {
		m.tags["container_host"] = hostname
	}

	f, err := getContainerStats(client, container.ID)
	if err != nil {
		return nil, err
	}
	m.fields = f

	// 毫秒除以1000得秒数，不使用Second()因为它返回浮点
	m.fields["age"] = time.Since(time.Unix(container.Created, 0)).Milliseconds() / 1e3
	m.fields["from_kubernetes"] = containerIsFromKubernetes(getContainerName(container.Names))
	m.fields.mergeToMessage(m.tags)

	if process, err := getContainerProcessToJSON(client, container.ID); err == nil {
		m.fields["process"] = process
	}
	return m, nil
}

func getContainerHostname(client dockerClientX, containerID string) (string, error) {
	containerJSON, err := client.ContainerInspect(context.TODO(), containerID)
	if err != nil {
		return "", err
	}
	return containerJSON.Config.Hostname, nil
}

func getContainerProcessToJSON(client dockerClientX, containerID string) (string, error) {
	process, err := getContainerProcess(client, containerID)
	if err != nil {
		return "", err
	}

	j, err := json.Marshal(process)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func getContainerProcess(client dockerClientX, containerID string) ([]map[string]string, error) {
	// query parameters: top
	// default "-ef"
	// The arguments to pass to ps. For example, aux
	top, err := client.ContainerTop(context.TODO(), containerID, nil)
	if err != nil {
		return nil, err
	}

	var res []map[string]string
	for _, proc := range top.Processes {
		if len(proc) != len(top.Titles) {
			continue
		}

		p := make(map[string]string)

		for idx, title := range top.Titles {
			p[title] = proc[idx]
		}

		res = append(res, p)
	}
	return res, nil
}

type containerObject struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func (c *containerObject) LineProto() (*io.Point, error) {
	return io.NewPoint(dockerContainerName, c.tags, c.fields, &io.PointOption{Time: c.time, Category: datakit.Object})
}

//nolint:lll
func (c *containerObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dockerContainerName,
		Desc: "容器对象数据，如果容器处于非 running 状态，则`cpu_usage`等指标将不存在",
		Type: "object",
		Tags: map[string]interface{}{
			"container_id":     inputs.NewTagInfo(`容器 ID（该字段默认被删除）`),
			"name":             inputs.NewTagInfo(`对象数据的指定 ID`),
			"status":           inputs.NewTagInfo("容器状态，例如 `Up 5 hours`"),
			"container_name":   inputs.NewTagInfo(`容器名称`),
			"docker_image":     inputs.NewTagInfo("镜像全称，例如 `nginx.org/nginx:1.21.0` （Depercated, use image）"),
			"image":            inputs.NewTagInfo("镜像全称，例如 `nginx.org/nginx:1.21.0`"),
			"image_name":       inputs.NewTagInfo("镜像名称，例如 `nginx.org/nginx`"),
			"image_short_name": inputs.NewTagInfo("镜像名称精简版，例如 `nginx`"),
			"image_tag":        inputs.NewTagInfo("镜像tag，例如 `1.21.0`"),
			"container_host":   inputs.NewTagInfo(`容器内部的主机名`),
			"container_type":   inputs.NewTagInfo(`容器类型，表明该容器由谁创建，kubernetes/docker`),
			"state":            inputs.NewTagInfo(`运行状态，running/exited/removed`),
			"pod_name":         inputs.NewTagInfo(`pod 名称（容器由 k8s 创建时存在）`),
			"pod_namesapce":    inputs.NewTagInfo(`pod 命名空间（容器由 k8s 创建时存在）`),
			"deployment":       inputs.NewTagInfo(`deployment 名称（容器由 k8s 创建时存在）`),
		},
		Fields: map[string]interface{}{
			"process":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "容器进程列表，即运行命令`ps -ef`所得，内容为 JSON 字符串，格式是 map 数组"},
			"age":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: `该容器创建时长，单位秒`},
			"from_kubernetes":    &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "该容器是否由 Kubernetes 创建（deprecated）"},
			"cpu_usage":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "CPU 占主机总量的使用率"},
			"cpu_delta":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "容器 CPU 增量"},
			"cpu_system_delta":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "系统 CPU 增量"},
			"cpu_numbers":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "CPU 核心数"},
			"message":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "容器对象详情"},
			"mem_limit":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "内存可用总量，如果未对容器做内存限制，则为主机内存容量"},
			"mem_usage":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "内存使用量"},
			"mem_used_percent":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "内存使用率，使用量除以可用总量"},
			"mem_failed_count":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "内存分配失败的次数"},
			"network_bytes_rcvd": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "从网络接收到的总字节数"},
			"network_bytes_sent": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "向网络发送出的总字节数"},
			"block_read_byte":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "从容器文件系统读取的总字节数"},
			"block_write_byte":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "向容器文件系统写入的总字节数"},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&containerObject{})
}

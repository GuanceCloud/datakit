// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"encoding/json"
	"time"

	"github.com/docker/docker/api/types"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"sigs.k8s.io/yaml"
)

func gatherDockerContainerObject(client dockerClientX, k8sClient k8sClientX, container *types.Container) (*containerObject, error) {
	m := &containerObject{}
	m.tags = getContainerInfo(container, k8sClient)
	m.tags["name"] = container.ID
	m.tags["linux_namespace"] = "moby"
	m.tags["status"] = container.Status

	if hostname, err := getContainerHostname(client, container.ID); err == nil {
		m.tags["container_host"] = hostname
	}

	f, err := getContainerStats(client, container.ID)
	if err != nil {
		return nil, err
	}
	m.fields = f

	if y, err := yaml.Marshal(container); err != nil {
		l.Debugf("failed to get container yaml %s, ID: %s, ignored", err.Error(), container.ID)
	} else {
		m.fields["yaml"] = string(y)
	}

	// 毫秒除以1000得秒数，不使用Second()因为它返回浮点
	m.fields["age"] = time.Since(time.Unix(container.Created, 0)).Milliseconds() / 1e3
	m.fields["from_kubernetes"] = containerIsFromKubernetes(getContainerName(container.Names))
	m.fields.mergeToMessage(m.tags)

	if process, err := getContainerProcessToJSON(client, container.ID); err == nil {
		m.fields["process"] = process
	}
	m.fields.delete("yaml")

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
}

func (c *containerObject) LineProto() (*point.Point, error) {
	return point.NewPoint(dockerContainerName, c.tags, c.fields, point.OOpt())
}

//nolint:lll
func (c *containerObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dockerContainerName,
		Desc: "The object of containers, only supported Running status.",
		Type: "object",
		Tags: map[string]interface{}{
			"container_id":           inputs.NewTagInfo(`Container ID`),
			"container_name":         inputs.NewTagInfo(`Container name from k8s (label 'io.kubernetes.container.name'). If empty then use $container_runtime_name.`),
			"container_runtime_name": inputs.NewTagInfo(`Container name from runtime (like 'docker ps'). If empty then use 'unknown' ([:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)).`),
			"name":                   inputs.NewTagInfo(`The ID of the contaienr.`),
			"docker_image":           inputs.NewTagInfo("The full name of the container image, example `nginx.org/nginx:1.21.0` (Depercated, use image)."),
			"linux_namespace":        inputs.NewTagInfo(`The [linux namespace](https://man7.org/linux/man-pages/man7/namespaces.7.html) where this container is located.`),
			"image":                  inputs.NewTagInfo("The full name of the container image, example `nginx.org/nginx:1.21.0`."),
			"image_name":             inputs.NewTagInfo("The name of the container image, example `nginx.org/nginx`."),
			"image_short_name":       inputs.NewTagInfo("The short name of the container image, example `nginx`."),
			"image_tag":              inputs.NewTagInfo("The tag of the container image, example `1.21.0`."),
			"container_host":         inputs.NewTagInfo(`The name of the container host (unsupported containerd).`),
			"container_type":         inputs.NewTagInfo(`The type of the container (this container is created by kubernetes/docker/containerd).`),
			"status":                 inputs.NewTagInfo("The status of the container，example `Up 5 hours` (unsupported containerd)."),
			"state":                  inputs.NewTagInfo(`The state of the Container (only Running, unsupported containerd).`),
			"pod_name":               inputs.NewTagInfo(`The pod name of the container (label 'io.kubernetes.pod.name').`),
			"namespace":              inputs.NewTagInfo(`The pod namespace of the container (label 'io.kubernetes.pod.namespace').`),
			"deployment":             inputs.NewTagInfo(`The deployment name of the container's pod (unsupported containerd).`),
		},
		Fields: map[string]interface{}{
			"process":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "List processes running inside a container (like `ps -ef`, unsupported containerd)."},
			"age":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: `The age of the container.`},
			"from_kubernetes":    &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "Is the container created by k8s (deprecated)."},
			"cpu_usage":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of CPU on system host."},
			"cpu_delta":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationNS, Desc: "The delta of the CPU (unsupported containerd)."},
			"cpu_system_delta":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationNS, Desc: "The delta of the system CPU, only supported Linux (unsupported containerd)."},
			"cpu_numbers":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of the CPU core (unsupported containerd)."},
			"mem_limit":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The available usage of the memory, if there is container limit, use host memory."},
			"mem_usage":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The usage of the memory."},
			"mem_used_percent":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory."},
			"mem_failed_count":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The count of memory allocation failures (unsupported containerd)."},
			"network_bytes_rcvd": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes received from the network (unsupported containerd)."},
			"network_bytes_sent": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes send to the network (unsupported containerd)."},
			"block_read_byte":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes read from the container file system (unsupported containerd)."},
			"block_write_byte":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes wrote to the container file system (unsupported containerd)."},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&containerObject{})
}

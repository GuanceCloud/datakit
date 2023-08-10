// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const containerMeasurement = "docker_containers"

type containerMetric struct{ typed.PointKV }

func (c *containerMetric) LineProto() (*point.Point, error) {
	return point.NewPoint(containerMeasurement, c.Tags(), c.Fields(), point.MOpt())
}

//nolint:lll
func (c *containerMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: containerMeasurement,
		Type: "metric",
		Desc: "The metric of containers, only supported Running status.",
		Tags: map[string]interface{}{
			"container_id":           inputs.NewTagInfo(`Container ID`),
			"container_name":         inputs.NewTagInfo("Container name from k8s (label `io.kubernetes.container.name`). If empty then use $container_runtime_name."),
			"container_runtime_name": inputs.NewTagInfo(`Container name from runtime (like 'docker ps'). If empty then use 'unknown' ([:octicons-tag-24: Version-1.4.6](../datakit/changelog.md#cl-1.4.6)).`),
			"container_runtime":      inputs.NewTagInfo(`Container runtime (this container from Docker/Containerd/cri-o).`),
			"container_type":         inputs.NewTagInfo(`The type of the container (this container is created by Kubernetes/Docker/Containerd/cri-o).`),
			"image":                  inputs.NewTagInfo("The full name of the container image, example `nginx.org/nginx:1.21.0`."),
			"image_name":             inputs.NewTagInfo("The name of the container image, example `nginx.org/nginx`."),
			"image_short_name":       inputs.NewTagInfo("The short name of the container image, example `nginx`."),
			"image_tag":              inputs.NewTagInfo("The tag of the container image, example `1.21.0`."),
			"state":                  inputs.NewTagInfo(`Container status (only Running).`),
			"pod_name":               inputs.NewTagInfo("The pod name of the container (label `io.kubernetes.pod.name`)."),
			"pod_uid":                inputs.NewTagInfo("The pod uid of the container (label `io.kubernetes.pod.uid`)."),
			"namespace":              inputs.NewTagInfo("The namespace of the container (label `io.kubernetes.pod.namespace`)."),
			"deployment":             inputs.NewTagInfo("The name of the Deployment which the object belongs to."),
			"daemonset":              inputs.NewTagInfo("The name of the DaemonSet which the object belongs to."),
			"statefulset":            inputs.NewTagInfo("The name of the StatefulSet which the object belongs to."),
		},
		Fields: map[string]interface{}{
			"cpu_usage":                   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of CPU on system host."},
			"cpu_numbers":                 &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of the CPU core."},
			"cpu_usage_base100":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized cpu usage, with a maximum of 100%."},
			"mem_usage":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The usage of the memory."},
			"mem_limit":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The limit memory in the container."},
			"mem_capacity":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory in the host machine."},
			"mem_used_percent":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory is calculated based on the capacity of host machine."},
			"mem_used_percent_base_limit": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory is calculated based on the limit."},
			"network_bytes_rcvd":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes received from the network (only count the usage of the main process in the container, excluding loopback)."},
			"network_bytes_sent":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes send to the network (only count the usage of the main process in the container, excluding loopback)."},
			"block_read_byte":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes read from the container file system (only supported docker)."},
			"block_write_byte":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes wrote to the container file system (only supported docker)."},
		},
	}
}

type containerObject struct{ typed.PointKV }

func (c *containerObject) LineProto() (*point.Point, error) {
	return point.NewPoint(containerMeasurement, c.Tags(), c.Fields(), point.OOpt())
}

//nolint:lll
func (c *containerObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: containerMeasurement,
		Desc: "The object of containers, only supported Running status.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":                   inputs.NewTagInfo(`The ID of the container.`),
			"container_id":           inputs.NewTagInfo(`Container ID`),
			"container_name":         inputs.NewTagInfo("Container name from k8s (label `io.kubernetes.container.name`). If empty then use $container_runtime_name."),
			"container_runtime_name": inputs.NewTagInfo(`Container name from runtime (like 'docker ps'). If empty then use 'unknown' ([:octicons-tag-24: Version-1.4.6](../datakit/changelog.md#cl-1.4.6)).`),
			"container_runtime":      inputs.NewTagInfo(`Container runtime (this container from Docker/Containerd/cri-o).`),
			"container_type":         inputs.NewTagInfo(`The type of the container (this container is created by Kubernetes/Docker/Containerd/cri-o).`),
			"image":                  inputs.NewTagInfo("The full name of the container image, example `nginx.org/nginx:1.21.0`."),
			"image_name":             inputs.NewTagInfo("The name of the container image, example `nginx.org/nginx`."),
			"image_short_name":       inputs.NewTagInfo("The short name of the container image, example `nginx`."),
			"image_tag":              inputs.NewTagInfo("The tag of the container image, example `1.21.0`."),
			"status":                 inputs.NewTagInfo("The status of the container，example `Up 5 hours`."),
			"state":                  inputs.NewTagInfo(`The state of the Container (only Running).`),
			"pod_name":               inputs.NewTagInfo("The pod name of the container (label `io.kubernetes.pod.name`)."),
			"pod_uid":                inputs.NewTagInfo("The pod uid of the container (label `io.kubernetes.pod.uid`)."),
			"namespace":              inputs.NewTagInfo("The namespace of the container (label `io.kubernetes.pod.namespace`)."),
			"deployment":             inputs.NewTagInfo("The name of the Deployment which the object belongs to."),
			"daemonset":              inputs.NewTagInfo("The name of the DaemonSet which the object belongs to."),
			"statefulset":            inputs.NewTagInfo("The name of the StatefulSet which the object belongs to."),
		},
		Fields: map[string]interface{}{
			"age":                         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"cpu_usage":                   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of CPU on system host."},
			"cpu_numbers":                 &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of the CPU core."},
			"cpu_usage_base100":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized cpu usage, with a maximum of 100%."},
			"mem_usage":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The usage of the memory."},
			"mem_limit":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The limit memory in the container."},
			"mem_capacity":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory in the host machine."},
			"mem_used_percent":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory is calculated based on the capacity of host machine."},
			"mem_used_percent_base_limit": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory is calculated based on the limit."},
			"network_bytes_rcvd":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes received from the network (only count the usage of the main process in the container, excluding loopback)."},
			"network_bytes_sent":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes send to the network (only count the usage of the main process in the container, excluding loopback)."},
			"block_read_byte":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes read from the container file system (only supported docker)."},
			"block_write_byte":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes wrote to the container file system (only supported docker)."},
		},
	}
}

type containerLog struct{}

func (c *containerLog) LineProto() (*point.Point, error) { return nil, nil }

//nolint:lll
func (c *containerLog) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "Use Logging Source",
		Desc: "The logging of the container.",
		Type: "logging",
		Tags: map[string]interface{}{
			"container_id":   inputs.NewTagInfo(`Container ID`),
			"container_name": inputs.NewTagInfo("Container name from k8s (label `io.kubernetes.container.name`). If empty then use $container_runtime_name."),
			"service":        inputs.NewTagInfo("The name of the service, if `service` is empty then use `source`."),
			"pod_name":       inputs.NewTagInfo("The pod name of the container (label `io.kubernetes.pod.name`)."),
			"namespace":      inputs.NewTagInfo("The namespace of the container (label `io.kubernetes.pod.namespace`)."),
			"deployment":     inputs.NewTagInfo("The name of the Deployment which the object belongs to."),
			"daemonset":      inputs.NewTagInfo("The name of the DaemonSet which the object belongs to."),
		},
		Fields: map[string]interface{}{
			"status":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The status of the logging, only supported `info/emerg/alert/critical/error/warning/debug/OK/unknown`."},
			"log_read_lines":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The lines of the read file ([:octicons-tag-24: Version-1.4.6](../datakit/changelog.md#cl-1.4.6))."},
			"log_read_offset": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The offset of the read file ([:octicons-tag-24: Version-1.4.8](../datakit/changelog.md#cl-1.4.8) · [:octicons-beaker-24: Experimental](../datakit/index.md#experimental))."},
			"log_read_time":   &inputs.FieldInfo{DataType: inputs.DurationSecond, Unit: inputs.UnknownUnit, Desc: "The timestamp of the read file."},
			"message_length":  &inputs.FieldInfo{DataType: inputs.SizeByte, Unit: inputs.NCount, Desc: "The length of the message content."},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The text of the logging."},
		},
	}
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const containerMeasurement = "docker_containers"

type containerMetric struct{}

//nolint:lll
func (*containerMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: containerMeasurement,
		Type: "metric",
		Desc: "The metric of containers, only supported Running status.",
		Tags: map[string]interface{}{
			"container_id":              inputs.NewTagInfo(`Container ID`),
			"container_name":            inputs.NewTagInfo("Container name from k8s (label `io.kubernetes.container.name`). If empty then use $container_runtime_name."),
			"container_runtime_name":    inputs.NewTagInfo(`Container name from runtime (like 'docker ps'). If empty then use 'unknown'.`),
			"container_runtime":         inputs.NewTagInfo(`Container runtime (this container from Docker/Containerd/cri-o).`),
			"container_runtime_version": inputs.NewTagInfo(`Container runtime version.`),
			"container_type":            inputs.NewTagInfo(`The type of the container (this container is created by Kubernetes/Docker/Containerd/cri-o).`),
			"image":                     inputs.NewTagInfo("The full name of the container image, example `nginx.org/nginx:1.21.0`."),
			"image_name":                inputs.NewTagInfo("The name of the container image, example `nginx.org/nginx`."),
			"image_short_name":          inputs.NewTagInfo("The short name of the container image, example `nginx`."),
			"image_tag":                 inputs.NewTagInfo("The tag of the container image, example `1.21.0`."),
			"state":                     inputs.NewTagInfo(`Container status (only Running).`),
			"pod_name":                  inputs.NewTagInfo("The pod name of the container (label `io.kubernetes.pod.name`)."),
			"pod_uid":                   inputs.NewTagInfo("The pod uid of the container (label `io.kubernetes.pod.uid`)."),
			"namespace":                 inputs.NewTagInfo("The namespace of the container (label `io.kubernetes.pod.namespace`)."),
			"deployment":                inputs.NewTagInfo("The name of the Deployment which the object belongs to."),
			"daemonset":                 inputs.NewTagInfo("The name of the DaemonSet which the object belongs to."),
			"statefulset":               inputs.NewTagInfo("The name of the StatefulSet which the object belongs to."),
			"cluster_name_k8s":          inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
			"aws_ecs_cluster_name":      inputs.NewTagInfo("Cluster name of the AWS ECS."),
			"task_family":               inputs.NewTagInfo("The task family of the AWS fargate."),
			"task_version":              inputs.NewTagInfo("The task version of the AWS fargate."),
			"task_arn":                  inputs.NewTagInfo("The task arn of the AWS Fargate."),
		},
		Fields: map[string]interface{}{
			"cpu_usage":                     &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The actual CPU usage on the system host (percentage)."},
			"cpu_usage_millicores":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Millicores, Desc: "The CPU usage of the container, measured in milli-cores."},
			"cpu_usage_base100":             &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized CPU usage, with a maximum value of 100%. It is calculated as the number of CPU cores multiplied by 100."},
			"cpu_usage_base_limit":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The CPU usage based on the CPU limit (percentage)."},
			"cpu_usage_base_request":        &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The CPU usage based on the CPU request (percentage)."},
			"cpu_numbers":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of CPU cores on the system host."},
			"cpu_limit_millicores":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Millicores, Desc: "The CPU limit of the container, measured in milli-cores."},
			"cpu_request_millicores":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Millicores, Desc: "The CPU request of the container, measured in milli-cores."},
			"mem_usage":                     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The actual memory usage of the container."},
			"mem_capacity":                  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory on the system host."},
			"mem_used_percent":              &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The memory usage percentage based on the total memory of the system host."},
			"mem_used_percent_base_limit":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The memory usage percentage based on the memory limit."},
			"mem_used_percent_base_request": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The memory usage percentage based on the memory request."},
			"mem_limit":                     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The memory limit of the container."},
			"mem_request":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The memory request of the container."},
			"network_bytes_rcvd":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes received from the network (only count the usage of the main process in the container, excluding loopback)."},
			"network_bytes_sent":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes send to the network (only count the usage of the main process in the container, excluding loopback)."},
			"block_read_byte":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes read from the container file system (only supported docker)."},
			"block_write_byte":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes wrote to the container file system (only supported docker)."},
		},
	}
}

type containerObject struct{}

//nolint:lll
func (*containerObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: containerMeasurement,
		Desc: "The object of containers, only supported Running status.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":                      inputs.NewTagInfo(`The ID of the container.`),
			"container_id":              inputs.NewTagInfo(`Container ID`),
			"container_name":            inputs.NewTagInfo("Container name from k8s (label `io.kubernetes.container.name`). If empty then use $container_runtime_name."),
			"container_runtime_name":    inputs.NewTagInfo(`Container name from runtime (like 'docker ps'). If empty then use 'unknown'.`),
			"container_runtime_version": inputs.NewTagInfo(`Container runtime version.`),
			"container_runtime":         inputs.NewTagInfo(`Container runtime (this container from Docker/Containerd/cri-o).`),
			"container_type":            inputs.NewTagInfo(`The type of the container (this container is created by Kubernetes/Docker/Containerd/cri-o).`),
			"image":                     inputs.NewTagInfo("The full name of the container image, example `nginx.org/nginx:1.21.0`."),
			"image_name":                inputs.NewTagInfo("The name of the container image, example `nginx.org/nginx`."),
			"image_short_name":          inputs.NewTagInfo("The short name of the container image, example `nginx`."),
			"image_tag":                 inputs.NewTagInfo("The tag of the container image, example `1.21.0`."),
			"status":                    inputs.NewTagInfo("The status of the container，example `Up 5 hours`."),
			"state":                     inputs.NewTagInfo(`The state of the Container (only Running).`),
			"pod_name":                  inputs.NewTagInfo("The pod name of the container (label `io.kubernetes.pod.name`)."),
			"pod_uid":                   inputs.NewTagInfo("The pod uid of the container (label `io.kubernetes.pod.uid`)."),
			"namespace":                 inputs.NewTagInfo("The namespace of the container (label `io.kubernetes.pod.namespace`)."),
			"deployment":                inputs.NewTagInfo("The name of the Deployment which the object belongs to."),
			"daemonset":                 inputs.NewTagInfo("The name of the DaemonSet which the object belongs to."),
			"statefulset":               inputs.NewTagInfo("The name of the StatefulSet which the object belongs to."),
			"cluster_name_k8s":          inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
			"aws_ecs_cluster_name":      inputs.NewTagInfo("Cluster name of the AWS ECS."),
			"task_family":               inputs.NewTagInfo("The task family of the AWS fargate."),
			"task_version":              inputs.NewTagInfo("The task version of the AWS fargate."),
			"task_arn":                  inputs.NewTagInfo("The task arn of the AWS Fargate."),
		},
		Fields: map[string]interface{}{
			"age":                           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"cpu_usage":                     &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The actual CPU usage on the system host (percentage)."},
			"cpu_usage_millicores":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Millicores, Desc: "The CPU usage of the container, measured in milli-cores."},
			"cpu_usage_base100":             &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized CPU usage, with a maximum value of 100%. It is calculated as the number of CPU cores multiplied by 100."},
			"cpu_usage_base_limit":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The CPU usage based on the CPU limit (percentage)."},
			"cpu_usage_base_request":        &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The CPU usage based on the CPU request (percentage)."},
			"cpu_numbers":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of CPU cores on the system host."},
			"cpu_limit_millicores":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Millicores, Desc: "The CPU limit of the container, measured in milli-cores."},
			"cpu_request_millicores":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Millicores, Desc: "The CPU request of the container, measured in milli-cores."},
			"mem_usage":                     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The actual memory usage of the container."},
			"mem_capacity":                  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory on the system host."},
			"mem_used_percent":              &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The memory usage percentage based on the total memory of the system host."},
			"mem_used_percent_base_limit":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The memory usage percentage based on the memory limit."},
			"mem_used_percent_base_request": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The memory usage percentage based on the memory request."},
			"mem_limit":                     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The memory limit of the container."},
			"mem_request":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The memory request of the container."},
			"network_bytes_rcvd":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes received from the network (only count the usage of the main process in the container, excluding loopback)."},
			"network_bytes_sent":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes send to the network (only count the usage of the main process in the container, excluding loopback)."},
			"block_read_byte":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes read from the container file system (only supported docker)."},
			"block_write_byte":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total number of bytes wrote to the container file system (only supported docker)."},
			"message":                       &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}

type containerLog struct{}

//nolint:lll
func (*containerLog) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "Use Logging Source",
		Desc: "The logging of the container.",
		Type: "logging",
		Tags: map[string]interface{}{
			"container_id":   inputs.NewTagInfo(`Container ID`),
			"container_name": inputs.NewTagInfo("Container name from k8s (label `io.kubernetes.container.name`). If empty then use $container_runtime_name."),
			"service":        inputs.NewTagInfo("The name of the service, if `service` is empty then use `source`."),
			"pod_name":       inputs.NewTagInfo("The pod name of the container (label `io.kubernetes.pod.name`)."),
			"pod_ip":         inputs.NewTagInfo("The pod ip of the container."),
			"namespace":      inputs.NewTagInfo("The namespace of the container (label `io.kubernetes.pod.namespace`)."),
			"deployment":     inputs.NewTagInfo("The name of the Deployment which the object belongs to."),
			"daemonset":      inputs.NewTagInfo("The name of the DaemonSet which the object belongs to."),
		},
		Fields: map[string]interface{}{
			"status":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The status of the logging, only supported `info/emerg/alert/critical/error/warning/debug/OK/unknown`."},
			"log_read_lines": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The lines of the read file ([:octicons-tag-24: Version-1.4.6](../datakit/changelog.md#cl-1.4.6))."},
			"message":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The text of the logging."},
		},
	}
}

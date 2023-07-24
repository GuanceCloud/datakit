// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const dockerContainerName = "docker_containers"

func gatherDockerContainerMetric(client dockerClientX, k8sClient k8sClientX, container *types.Container) (*containerMetric, error) {
	m := &containerMetric{}
	m.tags = getContainerInfo(container, k8sClient)

	f, err := getContainerStats(client, container.ID)
	if err != nil {
		return nil, err
	}

	m.fields = f
	return m, nil
}

func getContainerTags(container *types.Container) tagsType {
	imageName, imageShortName, imageTag := ParseImage(container.Image)

	tags := map[string]string{
		"state":                  container.State,
		"docker_image":           container.Image,
		"image":                  container.Image,
		"image_name":             imageName,
		"image_short_name":       imageShortName,
		"image_tag":              imageTag,
		"container_runtime_name": getContainerName(container.Names),
		"container_id":           container.ID,
		"linux_namespace":        "moby",
	}

	if n := getContainerNameForLabels(container.Labels); n != "" {
		tags["container_name"] = n
	} else {
		tags["container_name"] = tags["container_runtime_name"]
	}

	if !containerIsFromKubernetes(getContainerName(container.Names)) {
		tags["container_type"] = "docker"
	} else {
		tags["container_type"] = "kubernetes"
	}

	if n := getPodNameForLabels(container.Labels); n != "" {
		tags["pod_name"] = n
	}
	if n := getPodNamespaceForLabels(container.Labels); n != "" {
		tags["namespace"] = n
	}

	return tags
}

func getContainerInfo(container *types.Container, k8sClient k8sClientX) tagsType {
	tags := getContainerTags(container)

	podname := getPodNameForLabels(container.Labels)
	podnamespace := getPodNamespaceForLabels(container.Labels)
	podContainerName := getContainerNameForLabels(container.Labels)

	if k8sClient == nil || podname == "" {
		return tags
	}

	meta, err := queryPodMetaData(k8sClient, podname, podnamespace)
	if err != nil {
		// ignore err
		return tags
	}

	if image := meta.containerImage(podContainerName); image != "" {
		// 如果能找到 pod image，则使用它
		imageName, imageShortName, imageTag := ParseImage(image)

		tags["docker_image"] = image
		tags["image"] = image
		tags["image_name"] = imageName
		tags["image_short_name"] = imageShortName
		tags["image_tag"] = imageTag
	}

	if replicaSet := meta.replicaSet(); replicaSet != "" {
		tags["replicaSet"] = replicaSet
	}

	labels := meta.labels()
	if deployment := getDeployment(labels["app"], podnamespace); deployment != "" {
		tags["deployment"] = deployment
	}

	return tags
}

func getContainerStats(client dockerClientX, containerID string) (fieldsType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := client.ContainerStatsOneShot(ctx, containerID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.OSType == datakit.OSWindows {
		return nil, fmt.Errorf("not support windows docker/container")
	}

	var v *types.StatsJSON
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	metrics := calculateContainerStats(v)

	if metrics["network_bytes_rcvd"] == int64(0) && metrics["network_bytes_sent"] == int64(0) {
		pid, err := getContainerPID(client, containerID)
		if err != nil {
			l.Warnf("unable to get container pid from ID %s, err: %s, ignored", containerID, err)
		} else {
			netRx, netTx, err := getNetworkMetricsWithProc(pid)
			if err != nil {
				l.Warnf("unable to get net/dev info from ID %s, err: %s, ignored", containerID, err)
			} else {
				l.Debugf("use net/dev info from ID %s, rx: %d, tx: %d", containerID, netRx, netTx)
				metrics["network_bytes_rcvd"] = netRx
				metrics["network_bytes_sent"] = netTx
			}
		}
	}

	return metrics, nil
}

func calculateContainerStats(v *types.StatsJSON) map[string]interface{} {
	mem := calculateMemUsageUnixNoCache(v.MemoryStats)
	memPercent := calculateMemPercentUnixNoCache(float64(v.MemoryStats.Limit), float64(mem))
	blkRead, blkWrite := calculateBlockIO(v.BlkioStats)
	netRx, netTx := calculateNetwork(v.Networks)

	return map[string]interface{}{
		"cpu_usage": calculateCPUPercentUnix(v.PreCPUStats.CPUUsage.TotalUsage,
			v.PreCPUStats.SystemUsage, v), /*float64*/
		"cpu_delta":          calculateCPUDelta(v),
		"cpu_system_delta":   calculateCPUSystemDelta(v),
		"cpu_numbers":        calculateCPUNumbers(v),
		"mem_limit":          int64(v.MemoryStats.Limit),
		"mem_usage":          mem,
		"mem_used_percent":   memPercent, /*float64*/
		"mem_failed_count":   int64(v.MemoryStats.Failcnt),
		"network_bytes_rcvd": netRx,
		"network_bytes_sent": netTx,
		"block_read_byte":    blkRead,
		"block_write_byte":   blkWrite,
	}
}

// getDeployment
// 	ex: deployment-func-work-0, namespace is ’func', deployment is 'work-0'
func getDeployment(appStr, namespace string) string {
	if !strings.HasPrefix(appStr, "deployment") {
		return ""
	}

	if !strings.HasPrefix(appStr, "deployment-"+namespace) {
		return ""
	}

	return strings.TrimPrefix(appStr, "deployment-"+namespace+"-")
}

func getContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return "unknown"
}

func isRunningContainer(state string) bool {
	return state == "running"
}

func isPauseContainer(command string) bool {
	return command == "/pause"
}

func getNetworkMetricsWithProc(pid int) (int64, int64, error) {
	file := fmt.Sprintf("/proc/%d/net/dev", pid)
	if datakit.Docker {
		file = "/rootfs" + file
	}

	netdev, err := NewNetDev(file)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to read net/dev file, err: %w", err)
	}

	total := netdev.Total()
	return int64(total.RxBytes), int64(total.TxBytes), nil
}

// nolint:lll
// containerIsFromKubernetes 判断该容器是否由kubernetes创建
// 所有kubernetes启动的容器的containerNamePrefix都是k8s，依据链接如下
// https://github.com/rootsongjc/kubernetes-handbook/blob/master/practice/monitor.md#%E5%AE%B9%E5%99%A8%E7%9A%84%E5%91%BD%E5%90%8D%E8%A7%84%E5%88%99
func containerIsFromKubernetes(containerName string) bool {
	const kubernetesContainerNamePrefix = "k8s"
	return strings.HasPrefix(containerName, kubernetesContainerNamePrefix)
}

type containerMetric struct {
	tags   tagsType
	fields fieldsType
}

func (c *containerMetric) LineProto() (*point.Point, error) {
	return point.NewPoint(dockerContainerName, c.tags, c.fields, point.MOpt())
}

//nolint:lll
func (c *containerMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dockerContainerName,
		Type: "metric",
		Desc: "The metric of containers, only supported Running status.",
		Tags: map[string]interface{}{
			"container_id":           inputs.NewTagInfo(`Container ID`),
			"container_name":         inputs.NewTagInfo("Container name from k8s (label `io.kubernetes.container.name`). If empty then use $container_runtime_name."),
			"container_runtime_name": inputs.NewTagInfo(`Container name from runtime (like 'docker ps'). If empty then use 'unknown'.`),
			"docker_image":           inputs.NewTagInfo("The full name of the container image, example `nginx.org/nginx:1.21.0` (Deprecated: use image)."),
			"linux_namespace":        inputs.NewTagInfo(`The [Linux namespace](https://man7.org/linux/man-pages/man7/namespaces.7.html){:target="_blank"} where this container is located.`),
			"image":                  inputs.NewTagInfo("The full name of the container image, example `nginx.org/nginx:1.21.0`."),
			"image_name":             inputs.NewTagInfo("The name of the container image, example `nginx.org/nginx`."),
			"image_short_name":       inputs.NewTagInfo("The short name of the container image, example `nginx`."),
			"image_tag":              inputs.NewTagInfo("The tag of the container image, example `1.21.0`."),
			"container_type":         inputs.NewTagInfo(`The type of the container (this container is created by Kubernetes/Docker/containerd).`),
			"state":                  inputs.NewTagInfo(`Container status (only Running, unsupported containerd).`),
			"pod_name":               inputs.NewTagInfo("The pod name of the container (label `io.kubernetes.pod.name`)."),
			"namespace":              inputs.NewTagInfo("The pod namespace of the container (label `io.kubernetes.pod.namespace`)."),
			"deployment":             inputs.NewTagInfo(`The deployment name of the container's pod (unsupported containerd).`),
		},
		Fields: map[string]interface{}{
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
	registerMeasurement(&containerMetric{})
}

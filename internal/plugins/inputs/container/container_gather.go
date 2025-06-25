// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/podutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *containerCollector) gatherMetric() {
	start := time.Now()
	list, err := c.runtime.ListContainers()
	if err != nil {
		l.Warn(err)
		return
	}

	pts := c.runGather(list, c.buildMetricPoints)

	nsCount := make(map[string]int)
	for _, pt := range pts {
		kvs := pt.KVs()
		if ns := kvs.GetTag("namespace"); ns != "" {
			nsCount[ns]++
		}
	}

	opts := point.DefaultMetricOptions()
	for ns, count := range nsCount {
		var kvs point.KVs
		kvs = kvs.AddTag("namespace", ns)
		kvs = kvs.AddV2("container", count, false)
		kvs = kvs.AddTag("node_name", c.localNodeName)

		pt := point.NewPointV2("kubernetes", kvs, append(opts, point.WithTime(c.ptsTime))...)
		pts = append(pts, pt)
	}

	collectPtsVec.WithLabelValues("metric").Add(float64(len(pts)))

	if err := c.feeder.Feed(
		point.Metric,
		pts,
		dkio.WithElection(false),
		dkio.WithSource("container-metric"),
	); err != nil {
		l.Warnf("container-metric feed failed, err: %s", err)
	}

	collectCostVec.WithLabelValues("metric").Observe(time.Since(start).Seconds())
}

func (c *containerCollector) gatherObject() {
	start := time.Now()
	list, err := c.runtime.ListContainers()
	if err != nil {
		l.Warn(err)
		return
	}

	pts := c.runGather(list, c.buildObjectPoint)
	collectPtsVec.WithLabelValues("object").Add(float64(len(pts)))

	if err := c.feeder.Feed(
		point.Object,
		pts,
		dkio.WithElection(false),
		dkio.WithSource("container-object"),
	); err != nil {
		l.Warnf("container-object feed failed, err: %s", err)
	}

	collectCostVec.WithLabelValues("object").Observe(time.Since(start).Seconds())
}

func (c *containerCollector) runGather(
	list []*runtime.Container,
	buildPointsFunc func(*runtime.Container) *point.Point,
) []*point.Point {
	var pts []*point.Point
	var mu sync.Mutex

	g := goroutine.NewGroup(goroutine.Option{Name: "container-gather"})

	for idx := range list {
		if isPauseContainer(list[idx]) {
			continue
		}

		func(item *runtime.Container) {
			g.Go(func(ctx context.Context) error {
				pt := buildPointsFunc(item)
				mu.Lock()
				pts = append(pts, pt)
				mu.Unlock()
				return nil
			})
		}(list[idx])

		if (idx+1)%c.maxConcurrent == 0 {
			if err := g.Wait(); err != nil {
				l.Warn("waiting err: %s", err)
			}
		}
	}

	if err := g.Wait(); err != nil {
		l.Warn("waiting err: %s", err)
	}

	return pts
}

func (c *containerCollector) buildMetricPoints(item *runtime.Container) *point.Point {
	var kvs point.KVs
	kvs = append(kvs, buildInfoKVs(item)...)
	kvs = append(kvs, buildECSFargateKVs(item)...)

	top, err := c.runtime.ContainerTop(item.ID)
	if err != nil {
		l.Warnf("query stats failed, err: %s", err)
	} else {
		kvs = append(kvs, buildTopKVs(top)...)
	}

	var (
		containerName = getContainerNameForLabels(item.Labels)
		podname       = getPodNameForLabels(item.Labels)
		namespace     = getPodNamespaceForLabels(item.Labels)
		image         = item.Image
	)

	if c.k8sClient != nil && podname != "" {
		pod, err := c.k8sClient.GetPods(namespace).Get(
			context.Background(),
			podname,
			metav1.GetOptions{ResourceVersion: "0"})
		if err != nil {
			l.Warnf("query pod failed, err: %s", err)
		} else {
			if img := podutil.ContainerImageFromPod(containerName, pod); img != "" {
				image = img
			}
			kvs = append(kvs, buildPodKVs(containerName, pod, top)...)
			kvs = append(kvs, pointutil.LabelsToPointKVs(pod.Labels,
				c.podLabelAsTagsForMetric.all,
				c.podLabelAsTagsForMetric.keys)...)
		}
	}

	kvs = kvs.AddTag("image", image)
	kvs = append(kvs, point.NewTags(c.extraTags)...)

	return point.NewPointV2(containerMeasurement, kvs,
		append(point.DefaultMetricOptions(), point.WithTime(c.ptsTime))...)
}

func (c *containerCollector) buildObjectPoint(item *runtime.Container) *point.Point {
	var kvs point.KVs
	kvs = append(kvs, buildInfoKVs(item)...)
	kvs = append(kvs, buildECSFargateKVs(item)...)
	kvs = append(kvs, pointutil.ConvertDFLabels(item.Labels)...)

	top, err := c.runtime.ContainerTop(item.ID)
	if err != nil {
		l.Warnf("query stats failed, err: %s", err)
	} else {
		kvs = append(kvs, buildTopKVs(top)...)
	}

	var (
		containerName = getContainerNameForLabels(item.Labels)
		podname       = getPodNameForLabels(item.Labels)
		namespace     = getPodNamespaceForLabels(item.Labels)
		image         = item.Image
	)

	if c.k8sClient != nil && podname != "" {
		pod, err := c.k8sClient.GetPods(namespace).Get(context.Background(), podname, metav1.GetOptions{ResourceVersion: "0"})
		if err != nil {
			l.Warnf("query pod failed, err: %s", err)
		} else {
			if img := podutil.ContainerImageFromPod(containerName, pod); img != "" {
				// 优先使用 Pod 存在的 image
				image = img
			}
			kvs = append(kvs, buildPodKVs(containerName, pod, top)...)

			// 容器的 message 包含 Labels，k8s message 不包含 Labels
			kvs = append(kvs, pointutil.LabelsToPointKVs(pod.Labels, c.podLabelAsTagsForNonMetric.all, c.podLabelAsTagsForNonMetric.keys)...)
		}
	}

	kvs = kvs.AddTag("name", item.ID)
	kvs = kvs.AddV2("age", time.Since(time.Unix(0, item.CreatedAt)).Milliseconds()/1e3, false)
	kvs = kvs.AddTag("image", image)

	msg := pointutil.PointKVsToJSON(kvs)
	kvs = kvs.AddV2("message", pointutil.TrimString(msg, maxMessageLength), false)
	kvs = append(kvs, point.NewTags(c.extraTags)...)
	return point.NewPointV2(containerMeasurement, kvs,
		append(point.DefaultObjectOptions(), point.WithTime(ntp.Now()))...)
}

func buildInfoKVs(item *runtime.Container) point.KVs {
	var kvs point.KVs

	kvs = kvs.AddTag("container_id", item.ID)
	kvs = kvs.AddTag("container_runtime", item.RuntimeName)
	kvs = kvs.AddTag("container_runtime_version", item.RuntimeVersion)
	kvs = kvs.AddTag("state", item.State)

	kvs = kvs.AddV2("image", item.Image, true /*force*/, point.WithKVTagSet(true))

	if item.Name != "" {
		kvs = kvs.AddTag("container_runtime_name", item.Name)
	} else {
		kvs = kvs.AddTag("container_runtime_name", "unknown")
	}

	if containerIsFromKubernetes(item.Labels) {
		kvs = kvs.AddTag("container_type", "kubernetes")
	} else {
		kvs = kvs.AddTag("container_type", item.RuntimeName)
	}

	if name := getContainerNameForLabels(item.Labels); name != "" {
		kvs = kvs.AddTag("container_name", name)
	} else {
		kvs = kvs.AddTag("container_name", item.Name)
	}

	if uid := getPodUIDForLabels(item.Labels); uid != "" {
		kvs = kvs.AddTag("pod_uid", uid)
	}
	if podname := getPodNameForLabels(item.Labels); podname != "" {
		kvs = kvs.AddTag("pod_name", podname)
	}
	if namespace := getPodNamespaceForLabels(item.Labels); namespace != "" {
		kvs = kvs.AddTag("namespace", namespace)
	}

	return kvs
}

func buildTopKVs(top *runtime.ContainerTop) point.KVs {
	var kvs point.KVs

	kvs = kvs.AddV2("cpu_usage", top.CPUPercent, false)
	kvs = kvs.AddV2("cpu_usage_millicores", top.CPUUsageMillicores, false)
	kvs = kvs.AddV2("cpu_numbers", top.CPUCores, false)

	if top.CPUCores != 0 {
		kvs = kvs.AddV2("cpu_usage_base100", top.CPUPercent/float64(top.CPUCores), false)
	}
	if top.CPULimitMillicores != 0 {
		kvs = kvs.AddV2("cpu_limit_millicores", top.CPULimitMillicores, false)
		kvs = kvs.AddV2("cpu_usage_base_limit", float64(top.CPUUsageMillicores)/float64(top.CPULimitMillicores)*100, false)
	}

	kvs = kvs.AddV2("mem_usage", top.MemoryWorkingSet, false)
	if top.MemoryCapacity != 0 && top.MemoryCapacity != math.MaxInt64 {
		kvs = kvs.AddV2("mem_capacity", top.MemoryCapacity, false)
		kvs = kvs.AddV2("mem_used_percent", float64(top.MemoryWorkingSet)/float64(top.MemoryCapacity)*100, false)
	}
	if top.MemoryLimitInBytes != 0 {
		kvs = kvs.AddV2("mem_limit", top.MemoryLimitInBytes, false)
		kvs = kvs.AddV2("mem_used_percent_base_limit", float64(top.MemoryWorkingSet)/float64(top.MemoryLimitInBytes)*100, false)
	}

	kvs = kvs.AddV2("network_bytes_rcvd", top.NetworkRcvd, false)
	kvs = kvs.AddV2("network_bytes_sent", top.NetworkSent, false)

	// only supported docker
	if top.BlockRead != 0 {
		kvs = kvs.AddV2("block_read_byte", top.BlockRead, false)
	}
	if top.BlockWrite != 0 {
		kvs = kvs.AddV2("block_write_byte", top.BlockWrite, false)
	}

	return kvs
}

func buildPodKVs(containerName string, pod *apicorev1.Pod, top *runtime.ContainerTop) point.KVs {
	var kvs point.KVs

	if ownerKind, ownerName := podutil.PodOwner(pod); ownerKind != "" && ownerName != "" {
		kvs = kvs.AddTag(ownerKind, ownerName)
	}

	if image := podutil.ContainerImageFromPod(containerName, pod); image != "" {
		kvs = kvs.AddV2("image", image, true /*force*/, point.WithKVTagSet(true))
	}

	cpuLimit, memLimit := podutil.ContainerLimitInPod(containerName, pod)
	if cpuLimit != 0 {
		kvs = kvs.AddV2("cpu_limit_millicores", cpuLimit, true) // use force
		if top != nil {
			kvs = kvs.AddV2("cpu_usage_base_limit", float64(top.CPUUsageMillicores)/float64(cpuLimit)*100, true)
		}
	}
	if memLimit != 0 {
		kvs = kvs.AddV2("mem_limit", memLimit, true)
		if top != nil {
			kvs = kvs.AddV2("mem_used_percent_base_limit", float64(top.MemoryWorkingSet)/float64(memLimit)*100, true)
		}
	}

	cpuRequest, memRequest := podutil.ContainerLimitInPod(containerName, pod)
	if cpuRequest != 0 {
		kvs = kvs.AddV2("cpu_request_millicores", cpuRequest, false)
		if top != nil {
			kvs = kvs.AddV2("cpu_usage_base_request", float64(top.CPUUsageMillicores)/float64(cpuRequest)*100, false)
		}
	}
	if memRequest != 0 {
		kvs = kvs.AddV2("mem_request", memRequest, false)
		if top != nil {
			kvs = kvs.AddV2("mem_used_percent_base_request", float64(top.MemoryWorkingSet)/float64(memRequest)*100, false)
		}
	}

	return kvs
}

func buildECSFargateKVs(item *runtime.Container) point.KVs {
	var kvs point.KVs

	if name := getAWSClusterNameForLabels(item.Labels); name != "" {
		kvs = kvs.AddTag("aws_ecs_cluster_name", name)
	}
	if family := getTaskFamilyForLabels(item.Labels); family != "" {
		kvs = kvs.AddTag("task_family", family)
	}
	if version := getTaskVersionForLabels(item.Labels); version != "" {
		kvs = kvs.AddTag("task_version", version)
	}
	if arn := getTaskARNForLabels(item.Labels); arn != "" {
		kvs = kvs.AddTag("task_arn", arn)
	}

	return kvs
}

// get container info for k8s

func getPodNameForLabels(labels map[string]string) string {
	return labels["io.kubernetes.pod.name"]
}

func getPodUIDForLabels(labels map[string]string) string {
	return labels["io.kubernetes.pod.uid"]
}

func containerIsFromKubernetes(labels map[string]string) bool {
	uid, ok := labels["io.kubernetes.pod.uid"]
	return ok && uid != ""
}

func getPodNamespaceForLabels(labels map[string]string) string {
	return labels["io.kubernetes.pod.namespace"]
}

func getContainerNameForLabels(labels map[string]string) string {
	return labels["io.kubernetes.container.name"]
}

func getDockerTypeForLabels(labels map[string]string) string {
	return labels["io.kubernetes.docker.type"]
}

func isPauseContainer(info *runtime.Container) bool {
	typ := getDockerTypeForLabels(info.Labels)
	return typ == "podsandbox"
}

/// get task info for ecs fargate

func getAWSClusterNameForLabels(labels map[string]string) string {
	return trimClusterName(labels["com.amazonaws.ecs.cluster"])
}

func getTaskFamilyForLabels(labels map[string]string) string {
	return labels["com.amazonaws.ecs.task-definition-family"]
}

func getTaskVersionForLabels(labels map[string]string) string {
	return labels["com.amazonaws.ecs.task-definition-version"]
}

func getTaskARNForLabels(labels map[string]string) string {
	return labels["com.amazonaws.ecs.task-arn"]
}

func trimClusterName(s string) string {
	// e.g. arn:aws-cn:ecs:cn-north-1:123123123:cluster/datakit-dev-cluster
	flag := "cluster/"
	index := strings.LastIndex(s, flag)
	if index == -1 {
		return ""
	}
	return s[index+len(flag):]
}

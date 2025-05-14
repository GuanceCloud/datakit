// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"math"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/podutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
)

const (
	podMetricMeasurement = "kube_pod"
	podObjectMeasurement = "kubelet_pod"
)

//nolint:gochecknoinits
func init() {
	registerResource("pod-local", true, newPodLocal)
	registerResource("pod-remote", false, newPodRemote)
	registerMeasurements(&podMetric{}, &podObject{})
}

type (
	podLocal  struct{ *pod }
	podRemote struct{ *pod }
)

func newPodLocal(client k8sClient, cfg *Config) resource {
	return &podLocal{pod: newPod(client, cfg)}
}

func (local *podLocal) gatherMetric(ctx context.Context, timestamp int64) {
	fieldSelector := "spec.nodeName=" + local.cfg.NodeName
	local.pod.gatherMetric(ctx, fieldSelector, timestamp, false)
}

func (local *podLocal) gatherObject(ctx context.Context) {
	fieldSelector := "spec.nodeName=" + local.cfg.NodeName
	local.pod.gatherObject(ctx, fieldSelector, false)
}

func (local *podLocal) addChangeInformer(_ informers.SharedInformerFactory) { /* nil */ }

func newPodRemote(client k8sClient, cfg *Config) resource {
	return &podRemote{pod: newPod(client, cfg)}
}

func (remote *podRemote) gatherMetric(ctx context.Context, timestamp int64) {
	fieldSelector := ""
	pending := false

	if remote.cfg.NodeLocal {
		fieldSelector = "status.phase==Pending"
		pending = true
	}
	remote.pod.gatherMetric(ctx, fieldSelector, timestamp, pending)
}

func (remote *podRemote) gatherObject(ctx context.Context) {
	fieldSelector := ""
	pending := false

	if remote.cfg.NodeLocal {
		fieldSelector = "status.phase==Pending"
		pending = true
	}
	remote.pod.gatherObject(ctx, fieldSelector, pending)
}

func (remote *podRemote) addChangeInformer(_ informers.SharedInformerFactory) { /* nil */ }

type pod struct {
	client  k8sClient
	cfg     *Config
	counter map[string]map[string]int // e.g. map["namespace"]["node_name"] = N
}

func newPod(client k8sClient, cfg *Config) *pod {
	return &pod{client: client, cfg: cfg, counter: make(map[string]map[string]int)}
}

func (p *pod) gatherMetric(ctx context.Context, fieldSelector string, timestamp int64, pending bool) {
	var continued string
	for {
		list, err := p.client.GetPods(allNamespaces).List(ctx, newListOptions(fieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		var metricsClient PodMetricsClient
		var nodeInfo nodeCapacity

		if p.cfg.NodeLocal {
			metricsClient = newPodMetricsFromKubelet(p.client)
			nodeInfo = getCapacityFromNode(context.Background(), p.client, p.cfg.NodeName)
		} else {
			metricsClient = newPodMetricsFromAPIServer(p.client)
		}

		if pending {
			metricsClient = nil
		}

		pts := p.buildMetricPoints(list, metricsClient, nodeInfo, timestamp)
		feedMetric("k8s-pod-metric", p.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}

	var counterPts []*point.Point
	opts := point.DefaultMetricOptions()

	for ns, count := range p.counter {
		for nodeName, n := range count {
			var kvs point.KVs
			kvs = kvs.AddTag("namespace", ns)
			kvs = kvs.AddTag("node_name", nodeName)
			kvs = kvs.AddV2("pod", n, false)

			pt := point.NewPointV2("kubernetes", kvs, append(opts, point.WithTimestamp(timestamp))...)
			counterPts = append(counterPts, pt)
		}
	}
	feedMetric("k8s-counter", p.cfg.Feeder, counterPts, true)
}

func (p *pod) gatherObject(ctx context.Context, fieldSelector string, pending bool) {
	var runners []*promRunner
	var continued string

	for {
		list, err := p.client.GetPods(allNamespaces).List(ctx, newListOptions(fieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		var metricsClient PodMetricsClient
		var nodeInfo nodeCapacity

		if p.cfg.NodeLocal {
			metricsClient = newPodMetricsFromKubelet(p.client)
			nodeInfo = getCapacityFromNode(context.Background(), p.client, p.cfg.NodeName)
		} else {
			metricsClient = newPodMetricsFromAPIServer(p.client)
		}

		if pending {
			metricsClient = nil
		}

		pts := p.buildObjectPoints(list, metricsClient, nodeInfo)
		feedObject("k8s-pod-object", p.cfg.Feeder, pts, true)

		runners = append(runners, p.newPromRunners(list)...)

		if continued == "" {
			break
		}
	}

	if pending {
		return
	}

	select {
	case promRunnersChan <- runners:
		// nil
	default:
		// nil
	}
}

func (p *pod) buildMetricPoints(list *apicorev1.PodList, metricsClient PodMetricsClient, nodeInfo nodeCapacity, timestamp int64) []*point.Point {
	var pts []*point.Point
	opts := point.DefaultMetricOptions()

	for idx, item := range list.Items {
		if p.counter[item.Namespace] == nil {
			p.counter[item.Namespace] = make(map[string]int)
		}
		p.counter[item.Namespace][item.Spec.NodeName]++

		if p.cfg.PodFilterForMetric != nil {
			if !p.cfg.PodFilterForMetric.Match(filter.FilterNamespace, item.Namespace) {
				continue
			}
		}

		var kvs point.KVs
		kvs = append(kvs, buildPodKVs(&list.Items[idx])...)
		kvs = kvs.AddTag("pod", item.Name)

		if p.cfg.EnablePodMetric && shouldCollectPodMetrics(&list.Items[idx], metricsClient) {
			if nodeInfo.nodeName != item.Spec.NodeName {
				nodeInfo = getCapacityFromNode(context.Background(), p.client, item.Spec.NodeName)
			}
			metKVs := queryPodMetrics(context.Background(), metricsClient, &list.Items[idx], nodeInfo)
			kvs = append(kvs, metKVs...)
		}

		if p.cfg.EnableExtractK8sLabelAsTagsV1 {
			kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, true /*all labels*/, nil)...)
		} else {
			kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, p.cfg.LabelAsTagsForMetric.All, p.cfg.LabelAsTagsForMetric.Keys)...)
		}

		kvs = append(kvs, point.NewTags(p.cfg.ExtraTags)...)
		pt := point.NewPointV2(podMetricMeasurement, kvs, append(opts, point.WithTimestamp(timestamp))...)
		pts = append(pts, pt)
	}

	return pts
}

func (p *pod) buildObjectPoints(list *apicorev1.PodList, metricsClient PodMetricsClient, nodeInfo nodeCapacity) []*point.Point {
	var pts []*point.Point
	opts := point.DefaultObjectOptions()

	for idx, item := range list.Items {
		var kvs point.KVs
		kvs = append(kvs, buildPodKVs(&list.Items[idx])...)

		kvs = kvs.AddTag("name", string(item.UID))
		kvs = kvs.AddTag("pod_ip", item.Status.PodIP)
		kvs = kvs.AddTag("host", item.Spec.NodeName) // Pointing to the node where the pod is located.
		kvs = kvs.AddTag("phase", string(item.Status.Phase))
		kvs = kvs.AddTag("qos_class", string(item.Status.QOSClass))
		kvs = kvs.AddTag("status", string(item.Status.Phase))

		kvs = kvs.AddV2("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3, false)
		kvs = kvs.AddV2("available", len(item.Status.ContainerStatuses), false)

		for _, containerStatus := range item.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				kvs = kvs.AddTag("status", containerStatus.State.Waiting.Reason)
				break
			}
		}

		if y, err := yaml.Marshal(item); err == nil {
			kvs = kvs.AddV2("yaml", string(y), false)
		}
		kvs = kvs.AddV2("annotations", pointutil.MapToJSON(item.Annotations), false)
		kvs = append(kvs, pointutil.ConvertDFLabels(item.Labels)...)

		msg := pointutil.PointKVsToJSON(kvs)
		kvs = kvs.AddV2("message", pointutil.TrimString(msg, maxMessageLength), false)

		kvs = kvs.Del("annotations")
		kvs = kvs.Del("yaml")

		// The Object does not require checking if EnablePodMetric is enabled.
		if shouldCollectPodMetrics(&list.Items[idx], metricsClient) {
			if nodeInfo.nodeName != item.Spec.NodeName {
				nodeInfo = getCapacityFromNode(context.Background(), p.client, item.Spec.NodeName)
			}
			metKVs := queryPodMetrics(context.Background(), metricsClient, &list.Items[idx], nodeInfo)
			kvs = append(kvs, metKVs...)
		}

		if p.cfg.EnableExtractK8sLabelAsTagsV1 {
			kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, true /*all labels*/, nil)...)
		} else {
			kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, p.cfg.LabelAsTagsForNonMetric.All, p.cfg.LabelAsTagsForNonMetric.Keys)...)
		}
		kvs = append(kvs, point.NewTags(p.cfg.ExtraTags)...)
		pt := point.NewPointV2(podObjectMeasurement, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

func (p *pod) newPromRunners(list *apicorev1.PodList) []*promRunner {
	var runners []*promRunner
	for idx, item := range list.Items {
		if item.Status.Phase != apicorev1.PodRunning {
			continue
		}

		inputConfig, exist := item.Annotations[annotationPromExport]
		if !exist {
			continue
		}

		runners = append(runners, newPromRunnersForPod(&list.Items[idx], inputConfig, p.cfg)...)
	}
	return runners
}

func buildPodKVs(item *apicorev1.Pod) point.KVs {
	var kvs point.KVs

	kvs = kvs.AddTag("uid", string(item.UID))
	kvs = kvs.AddTag("pod_name", item.Name)
	kvs = kvs.AddTag("namespace", item.Namespace)
	kvs = kvs.AddTag("node_name", item.Spec.NodeName)

	// "scheduled","unschedulable","volumes_persistentvolumeclaims_readonly"

	containerReadyCount := 0
	for _, cs := range item.Status.ContainerStatuses {
		if cs.State.Running != nil {
			containerReadyCount++
		}
	}
	kvs = kvs.AddV2("ready", containerReadyCount, false)

	maxRestarts := 0
	for _, containerStatus := range item.Status.ContainerStatuses {
		if int(containerStatus.RestartCount) > maxRestarts {
			maxRestarts = int(containerStatus.RestartCount)
		}
	}
	kvs = kvs.AddV2("restarts", maxRestarts, false)

	ownerKind, ownerName := podutil.PodOwner(item)
	if ownerKind != "" && ownerName != "" {
		kvs = kvs.AddTag(ownerKind, ownerName)
	}

	return kvs
}

func shouldCollectPodMetrics(item *apicorev1.Pod, metricsClient PodMetricsClient) bool {
	if ownerKind, _ := podutil.PodOwner(item); ownerKind == "job" || ownerKind == "cronjob" {
		return false
	}
	if item.Status.Phase != apicorev1.PodRunning {
		return false
	}
	return metricsClient != nil
}

func queryPodMetrics(ctx context.Context, client PodMetricsClient, item *apicorev1.Pod, node nodeCapacity) point.KVs {
	podMet, err := client.GetPodMetrics(ctx, item.Namespace, item.Name)
	if err != nil {
		klog.Warnf("query for pod-metrics failed, err: %s", err)
		return nil
	}

	var kvs point.KVs

	cpuUsage := float64(podMet.cpuUsageMilliCores) / 1e3 * 100.0
	kvs = kvs.AddV2("cpu_usage_millicores", podMet.cpuUsageMilliCores, false)
	kvs = kvs.AddV2("cpu_usage", cpuUsage, false)

	kvs = kvs.AddV2("mem_usage", podMet.memoryUsageBytes, false)
	kvs = kvs.AddV2("mem_rss", podMet.memoryRSSBytes, false)
	kvs = kvs.AddV2("mem_capacity", node.memCapacity, false)
	kvs = kvs.AddV2("memory_usage_bytes", podMet.memoryUsageBytes, false) // maintain compatibility
	kvs = kvs.AddV2("memory_capacity", node.memCapacity, false)           // maintain compatibility

	if node.cpuCapacityMillicores != 0 {
		cores := node.cpuCapacityMillicores / 1e3
		x := cpuUsage / float64(cores)
		if math.IsNaN(x) {
			x = 0.0
		}
		kvs = kvs.AddV2("cpu_number", cores, false)
		kvs = kvs.AddV2("cpu_usage_base100", x, false)
	}

	if cpuLimit := podutil.SumCPULimits(item); cpuLimit != 0 {
		kvs = kvs.AddV2("cpu_limit_millicores", cpuLimit, false)
		kvs = kvs.AddV2("cpu_usage_base_limit", float64(podMet.cpuUsageMilliCores)/float64(cpuLimit)*100, false)
	}
	if cpuRequest := podutil.SumCPURequests(item); cpuRequest != 0 {
		kvs = kvs.AddV2("cpu_request_millicores", cpuRequest, false)
		kvs = kvs.AddV2("cpu_usage_base_request", float64(podMet.cpuUsageMilliCores)/float64(cpuRequest)*100, false)
	}

	if node.memCapacity != 0 {
		x := float64(podMet.memoryUsageBytes) / float64(node.memCapacity) * 100
		kvs = kvs.AddV2("mem_used_percent_base_100", x, false)
		kvs = kvs.AddV2("memory_used_percent", x, false) // maintain compatibility
	}
	if memLimit := podutil.SumMemoryLimits(item); memLimit != 0 {
		kvs = kvs.AddV2("mem_limit", memLimit, false)
		kvs = kvs.AddV2("mem_used_percent_base_limit", float64(podMet.memoryUsageBytes)/float64(memLimit)*100, false)
	}
	if memRequest := podutil.SumMemoryRequests(item); memRequest != 0 {
		kvs = kvs.AddV2("mem_request", memRequest, false)
		kvs = kvs.AddV2("mem_used_percent_base_request", float64(podMet.memoryUsageBytes)/float64(memRequest)*100, false)
	}

	kvs = kvs.AddV2("network_bytes_rcvd", podMet.networkRcvdBytes, false)
	kvs = kvs.AddV2("network_bytes_sent", podMet.networkSentBytes, false)
	kvs = kvs.AddV2("ephemeral_storage_used_bytes", podMet.ephemeralStorageUsedBytes, false)
	kvs = kvs.AddV2("ephemeral_storage_available_bytes", podMet.ephemeralStorageAvailableBytes, false)
	kvs = kvs.AddV2("ephemeral_storage_capacity_bytes", podMet.ephemeralStorageCapacityBytes, false)

	return kvs
}

type podMetric struct{}

//nolint:lll
func (*podMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: podMetricMeasurement,
		Desc: "The metric of the Kubernetes Pod.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":              inputs.NewTagInfo("The UID of pod."),
			"pod":              inputs.NewTagInfo("Name must be unique within a namespace."),
			"pod_name":         inputs.NewTagInfo("Renamed from 'pod'."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"node_name":        inputs.NewTagInfo("NodeName is a request to schedule this pod onto a specific node."),
			"deployment":       inputs.NewTagInfo("The name of the Deployment which the object belongs to."),
			"daemonset":        inputs.NewTagInfo("The name of the DaemonSet which the object belongs to."),
			"statefulset":      inputs.NewTagInfo("The name of the StatefulSet which the object belongs to."),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"ready":                             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Describes whether the pod is ready to serve requests."},
			"restarts":                          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of times the container has been restarted."},
			"cpu_usage":                         &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The total CPU usage across all containers in this Pod."},
			"cpu_usage_millicores":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Millicores, Desc: "The total CPU usage (in millicores) averaged over the sample window for all containers."},
			"cpu_usage_base100":                 &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized CPU usage, with a maximum of 100%."},
			"cpu_usage_base_limit":              &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized CPU usage, with a maximum of 100%, based on the CPU limit."},
			"cpu_usage_base_request":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized CPU usage, with a maximum of 100%, based on the CPU request."},
			"cpu_number":                        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of CPUs on the node where the Pod is running."},
			"cpu_limit_millicores":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Millicores, Desc: "The total CPU limit (in millicores) across all containers in this Pod. Note: This value is the sum of all container limit values, as Pods do not have a direct limit value."},
			"cpu_request_millicores":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Millicores, Desc: "The total CPU request (in millicores) across all containers in this Pod.  Note: This value is the sum of all container request values, as Pods do not have a direct request value."},
			"mem_usage":                         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory usage of all containers in this Pod."},
			"mem_rss":                           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total RSS memory usage of all containers in this Pod, which is not supported by metrics-server."},
			"mem_used_percent":                  &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage of memory usage based on the host machine’s total memory capacity."},
			"mem_used_percent_base_limit":       &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage of memory usage based on the memory limit."},
			"mem_used_percent_base_request":     &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage of memory usage based on the memory request."},
			"mem_capacity":                      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory capacity of the host machine."},
			"mem_limit":                         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory limit across all containers in this Pod.  Note: This value is the sum of all container limit values, as Pods do not have a direct limit value."},
			"mem_request":                       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory request across all containers in this Pod.  Note: This value is the sum of all container request values, as Pods do not have a direct request value."},
			"network_bytes_rcvd":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Cumulative count of bytes received."},
			"network_bytes_sent":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Cumulative count of bytes transmitted."},
			"ephemeral_storage_used_bytes":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The bytes used for a specific task on the filesystem."},
			"ephemeral_storage_available_bytes": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The storage space available (bytes) for the filesystem."},
			"ephemeral_storage_capacity_bytes":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total capacity (bytes) of the filesystems underlying storage."},
			"memory_usage_bytes":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The sum of the memory usage of all containers in this Pod (Deprecated use `mem_usage`)."},
			"memory_capacity":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory in the host machine (Deprecated use `mem_capacity`)."},
			"memory_used_percent":               &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory (refer from `mem_used_percent`"},
		},
	}
}

type podObject struct{}

//nolint:lll
func (*podObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: podObjectMeasurement,
		Desc: "The object of the Kubernetes Pod.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":             inputs.NewTagInfo("The UID of Pod."),
			"uid":              inputs.NewTagInfo("The UID of Pod."),
			"pod_name":         inputs.NewTagInfo("Name must be unique within a namespace."),
			"node_name":        inputs.NewTagInfo("NodeName is a request to schedule this pod onto a specific node."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"phase":            inputs.NewTagInfo("The phase of a Pod is a simple, high-level summary of where the Pod is in its lifecycle.(Pending/Running/Succeeded/Failed/Unknown)"),
			"status":           inputs.NewTagInfo("Reason the container is not yet running."),
			"qos_class":        inputs.NewTagInfo("The Quality of Service (QOS) classification assigned to the pod based on resource requirements"),
			"deployment":       inputs.NewTagInfo("The name of the Deployment which the object belongs to."),
			"daemonset":        inputs.NewTagInfo("The name of the DaemonSet which the object belongs to."),
			"statefulset":      inputs.NewTagInfo("The name of the StatefulSet which the object belongs to."),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"age":                               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"restarts":                          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of times the container has been restarted."},
			"ready":                             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Describes whether the pod is ready to serve requests."},
			"available":                         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of containers"},
			"cpu_usage":                         &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The total CPU usage across all containers in this Pod."},
			"cpu_usage_millicores":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Millicores, Desc: "The total CPU usage (in millicores) averaged over the sample window for all containers."},
			"cpu_usage_base100":                 &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized CPU usage, with a maximum of 100%."},
			"cpu_usage_base_limit":              &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized CPU usage, with a maximum of 100%, based on the CPU limit."},
			"cpu_usage_base_request":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized CPU usage, with a maximum of 100%, based on the CPU request."},
			"cpu_number":                        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of CPUs on the node where the Pod is running."},
			"cpu_limit_millicores":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Millicores, Desc: "The total CPU limit (in millicores) across all containers in this Pod. Note: This value is the sum of all container limit values, as Pods do not have a direct limit value."},
			"cpu_request_millicores":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Millicores, Desc: "The total CPU request (in millicores) across all containers in this Pod.  Note: This value is the sum of all container request values, as Pods do not have a direct request value."},
			"mem_usage":                         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory usage of all containers in this Pod."},
			"mem_rss":                           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total RSS memory usage of all containers in this Pod, which is not supported by metrics-server."},
			"mem_used_percent":                  &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage of memory usage based on the host machine’s total memory capacity."},
			"mem_used_percent_base_limit":       &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage of memory usage based on the memory limit."},
			"mem_used_percent_base_request":     &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage of memory usage based on the memory request."},
			"mem_capacity":                      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory capacity of the host machine."},
			"mem_limit":                         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory limit across all containers in this Pod.  Note: This value is the sum of all container limit values, as Pods do not have a direct limit value."},
			"mem_request":                       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory request across all containers in this Pod.  Note: This value is the sum of all container request values, as Pods do not have a direct request value."},
			"memory_usage_bytes":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The sum of the memory usage of all containers in this Pod (Deprecated use `mem_usage`)."},
			"memory_capacity":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory in the host machine (Deprecated use `mem_capacity`)."},
			"memory_used_percent":               &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory (refer from `mem_used_percent`"},
			"network_bytes_rcvd":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Cumulative count of bytes received."},
			"network_bytes_sent":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Cumulative count of bytes transmitted."},
			"ephemeral_storage_used_bytes":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The bytes used for a specific task on the filesystem."},
			"ephemeral_storage_available_bytes": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The storage space available (bytes) for the filesystem."},
			"ephemeral_storage_capacity_bytes":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total capacity (bytes) of the filesystems underlying storage."},
			"message":                           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}

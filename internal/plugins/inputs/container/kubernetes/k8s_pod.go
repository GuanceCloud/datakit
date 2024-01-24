// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	podMetricMeasurement = "kube_pod"
	podObjectMeasurement = "kubelet_pod"
)

//nolint:gochecknoinits
func init() {
	registerResource("pod", true, true, newPod)
	registerMeasurements(&podMetric{}, &podObject{})
}

type pod struct {
	client    k8sClient
	continued string
	//    e.g. map["kube-system"]["node_name"] = 10
	counter map[string]map[string]int
}

func newPod(client k8sClient) resource {
	return &pod{client: client, counter: make(map[string]map[string]int)}
}

func (p *pod) count() []pointV2 {
	var pts []pointV2
	t := time.Now()

	for ns, nodes := range p.counter {
		for node, cnt := range nodes {
			pt := point.NewPointV2(
				"kubernetes",
				point.KVs{
					point.NewKV("namespace", ns, point.WithKVTagSet(true)),
					point.NewKV("node_name", node, point.WithKVTagSet(true)),
					point.NewKV("pod", cnt),
				},
				append(point.DefaultMetricOptions(), point.WithTime(t))...,
			)
			pts = append(pts, pt)
		}
	}
	return pts
}

func (p *pod) hasNext() bool { return p.continued != "" }

func (p *pod) getMetadata(ctx context.Context, ns, fieldSelector string) (metadata, error) {
	opt := metav1.ListOptions{
		Limit:         queryLimit,
		Continue:      p.continued,
		FieldSelector: fieldSelector,
	}

	list, err := p.client.GetPods(ns).List(ctx, opt)
	if err != nil {
		return nil, err
	}

	p.continued = list.Continue

	var metricsCollect PodMetricsCollect = newPodMetricsFromAPIServer(p.client)
	// use NodeLocal mode
	if strings.HasPrefix(fieldSelector, "spec.nodeName") {
		metricsCollect = newPodMetricsFromKubelet(p.client)
	}

	return &podMetadata{p, list, metricsCollect}, nil
}

type podMetadata struct {
	parent         *pod
	list           *apicorev1.PodList
	metricsCollect PodMetricsCollect
}

func (m *podMetadata) newMetric(conf *Config) pointKVs {
	var res pointKVs
	nodeInfo := nodeCapacity{}

	for idx, item := range m.list.Items {
		met := typed.NewPointKV(podMetricMeasurement)

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("pod", item.Name)
		met.SetTag("pod_name", item.Name)
		met.SetTag("namespace", item.Namespace)
		met.SetTag("node_name", item.Spec.NodeName)

		met.SetField("ready", 0)
		// "scheduled", "volumes_persistentvolumeclaims_readonly","unschedulable"

		containerReadyCount := 0
		for _, cs := range item.Status.ContainerStatuses {
			if cs.State.Running != nil {
				containerReadyCount++
			}
		}
		met.SetField("ready", containerReadyCount)

		maxRestarts := 0
		for _, containerStatus := range item.Status.ContainerStatuses {
			if int(containerStatus.RestartCount) > maxRestarts {
				maxRestarts = int(containerStatus.RestartCount)
			}
		}
		met.SetField("restarts", maxRestarts)

		cpuLimit := getMaxCPULimitFromResource(item.Spec.Containers)
		if cpuLimit != 0 {
			met.SetField("cpu_limit_millicores", cpuLimit)
		}
		memLimit := getMemoryLimitFromResource(item.Spec.Containers)
		if memLimit != 0 {
			met.SetField("mem_limit", memLimit)
		}

		ownerKind, ownerName := getOwner(item.OwnerReferences, item.Labels["pod-template-hash"])
		if ownerKind != "" && ownerName != "" {
			met.SetTag(ownerKind, ownerName)
		}

		if conf.EnablePodMetric &&
			shouldCollectPodMetric(&m.list.Items[idx]) &&
			m.metricsCollect != nil &&
			m.parent != nil {
			if nodeInfo.nodeName != item.Spec.NodeName {
				nodeInfo = getCapacityFromNode(context.Background(), m.parent.client, item.Spec.NodeName)
			}

			p, err := queryPodMetrics(context.Background(), m.metricsCollect, &m.list.Items[idx], nodeInfo, "metric")
			if err != nil {
				klog.Warnf("pod %s metric-pt fail, err: %s, skip", item.Name, err)
			} else {
				met.SetTags(p.Tags())
				met.SetFields(p.Fields())
			}
		}

		if conf.EnableExtractK8sLabelAsTagsV1 {
			met.SetLabelAsTags(item.Labels, true /*all labels*/, nil)
		} else {
			met.SetLabelAsTags(item.Labels, conf.LabelAsTagsForMetric.All, conf.LabelAsTagsForMetric.Keys)
		}
		res = append(res, met)

		if m.parent.counter[item.Namespace] == nil {
			m.parent.counter[item.Namespace] = make(map[string]int)
		}
		m.parent.counter[item.Namespace][item.Spec.NodeName]++
	}

	return res
}

func (m *podMetadata) newObject(conf *Config) pointKVs {
	var res pointKVs
	nodeInfo := nodeCapacity{}

	for idx, item := range m.list.Items {
		obj := typed.NewPointKV(podObjectMeasurement)

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("uid", fmt.Sprintf("%v", item.UID))
		obj.SetTag("pod_name", item.Name)
		obj.SetTag("pod_ip", item.Status.PodIP)
		obj.SetTag("node_name", item.Spec.NodeName)
		obj.SetTag("host", item.Spec.NodeName) // 指向 pod 所在的 node，便于关联
		obj.SetTag("phase", fmt.Sprintf("%v", item.Status.Phase))
		obj.SetTag("qos_class", fmt.Sprintf("%v", item.Status.QOSClass))
		obj.SetTag("status", fmt.Sprintf("%v", item.Status.Phase))
		obj.SetTag("namespace", item.Namespace)

		obj.SetField("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		obj.SetField("available", len(item.Status.ContainerStatuses))

		if y, err := yaml.Marshal(item); err == nil {
			obj.SetField("yaml", string(y))
		}

		for _, containerStatus := range item.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				obj.SetTag("status", containerStatus.State.Waiting.Reason)
				break
			}
		}

		cpuLimit := getMaxCPULimitFromResource(item.Spec.Containers)
		obj.SetField("cpu_limit_millicores", cpuLimit)

		memLimit := getMemoryLimitFromResource(item.Spec.Containers)
		obj.SetField("mem_limit", memLimit)

		containerReadyCount := 0
		for _, cs := range item.Status.ContainerStatuses {
			if cs.State.Running != nil {
				containerReadyCount++
			}
		}
		obj.SetField("ready", containerReadyCount)

		maxRestarts := 0
		for _, containerStatus := range item.Status.ContainerStatuses {
			if int(containerStatus.RestartCount) > maxRestarts {
				maxRestarts = int(containerStatus.RestartCount)
			}
		}
		obj.SetField("restarts", maxRestarts)

		obj.SetFields(transLabels(item.Labels))
		obj.SetField("annotations", typed.MapToJSON(item.Annotations))
		obj.SetField("message", typed.TrimString(obj.String(), maxMessageLength))
		obj.DeleteField("annotations")
		obj.DeleteField("yaml")

		ownerKind, ownerName := getOwner(item.OwnerReferences, item.Labels["pod-template-hash"])
		if ownerKind != "" && ownerName != "" {
			obj.SetTag(ownerKind, ownerName)
		}

		if conf.EnablePodMetric && shouldCollectPodMetric(&m.list.Items[idx]) && m.metricsCollect != nil {
			if nodeInfo.nodeName != item.Spec.NodeName {
				nodeInfo = getCapacityFromNode(context.Background(), m.parent.client, item.Spec.NodeName)
			}

			p, err := queryPodMetrics(context.Background(), m.metricsCollect, &m.list.Items[idx], nodeInfo, "object")
			if err != nil {
				klog.Warnf("pod %s object-pt fail, err: %s, skip", item.Name, err)
			} else {
				obj.SetTags(p.Tags())
				obj.SetFields(p.Fields())
			}
		}

		if conf.EnableExtractK8sLabelAsTagsV1 {
			obj.SetLabelAsTags(item.Labels, true /*all labels*/, nil)
		} else {
			obj.SetLabelAsTags(item.Labels, conf.LabelAsTagsForNonMetric.All, conf.LabelAsTagsForNonMetric.Keys)
		}
		res = append(res, obj)
	}

	return res
}

func getOwner(owners []metav1.OwnerReference, podTemplateHash string) (kind string, name string) {
	if len(owners) != 0 {
		switch owners[0].Kind {
		case "ReplicaSet":
			kind = "deployment"
			name = strings.TrimSuffix(owners[0].Name, "-"+podTemplateHash)
		case "DaemonSet":
			kind = "daemonset"
			name = owners[0].Name
		case "StatefulSet":
			kind = "statefulset"
			name = owners[0].Name
		case "Job":
			kind = "job"
			name = owners[0].Name
		case "CronJob":
			kind = "cronjob"
			name = owners[0].Name
		default:
			// nil
		}
	}
	return
}

func shouldCollectPodMetric(item *apicorev1.Pod) bool {
	for _, owner := range item.OwnerReferences {
		switch owner.Kind {
		case "Job", "CronJob":
			return false
		default:
		}
	}
	return item.Status.Phase == apicorev1.PodRunning
}

func queryPodMetrics(
	ctx context.Context,
	collect PodMetricsCollect,
	item *apicorev1.Pod,
	node nodeCapacity,
	category string,
) (*typed.PointKV, error) {
	p := typed.NewPointKV("NULL")

	podMet, err := collect.GetPodMetrics(ctx, item.Namespace, item.Name)
	if err != nil {
		return p, err
	}

	p.SetField("cpu_usage", podMet.cpuUsage)
	p.SetField("cpu_usage_millicores", podMet.cpuUsageMilliCores)
	p.SetField("mem_usage", podMet.memoryUsageBytes)

	p.SetField("cpu_capacity_millicores", node.cpuCapacityMillicores)
	p.SetField("mem_capacity", node.memCapacity)

	if node.cpuCapacityMillicores != 0 {
		cores := node.cpuCapacityMillicores / 1e3
		x := podMet.cpuUsage / float64(cores)
		if math.IsNaN(x) {
			x = 0.0
		}
		p.SetField("cpu_usage_base100", x)
	}

	if memLimit := getMemoryLimitFromResource(item.Spec.Containers); memLimit != 0 {
		p.SetField("mem_used_percent_base_limit", float64(podMet.memoryUsageBytes)/float64(memLimit)*100)
	}

	if node.memCapacity != 0 {
		p.SetField("mem_used_percent", float64(podMet.memoryUsageBytes)/float64(node.memCapacity)*100)
	}

	// maintain compatibility
	p.SetField("memory_usage_bytes", p.GetField("mem_usage"))
	p.SetField("memory_capacity", p.GetField("mem_capacity"))
	p.SetField("memory_used_percent", p.GetField("mem_used_percent"))

	switch category {
	case "metric":
		p.SetField("network_bytes_rcvd", podMet.networkRcvdBytes)
		p.SetField("network_bytes_sent", podMet.networkSentBytes)
		p.SetField("ephemeral_storage_used_bytes", podMet.ephemeralStorageUsedBytes)
		p.SetField("ephemeral_storage_available_bytes", podMet.ephemeralStorageAvailableBytes)
		p.SetField("ephemeral_storage_capacity_bytes", podMet.ephemeralStorageCapacityBytes)
	default:
		// nil
	}

	return p, nil
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
			"cpu_usage":                         &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The sum of the cpu usage of all containers in this Pod."},
			"cpu_usage_base100":                 &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized cpu usage, with a maximum of 100%. (Experimental)"},
			"cpu_usage_millicores":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationMS, Desc: "Total CPU usage (sum of all cores) averaged over the sample window."},
			"memory_usage_bytes":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The sum of the memory usage of all containers in this Pod (Deprecated use `mem_usage`)."},
			"memory_capacity":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory in the host machine (Deprecated use `mem_capacity`)."},
			"memory_used_percent":               &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory (refer from `mem_used_percent`"},
			"mem_usage":                         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The sum of the memory usage of all containers in this Pod."},
			"mem_limit":                         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The sum of the memory limit of all containers in this Pod."},
			"mem_capacity":                      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory in the host machine."},
			"mem_used_percent":                  &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory is calculated based on the capacity of host machine."},
			"mem_used_percent_base_limit":       &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory is calculated based on the limit."},
			"network_bytes_rcvd":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Cumulative count of bytes received."},
			"network_bytes_sent":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Cumulative count of bytes transmitted."},
			"ephemeral_storage_used_bytes":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The bytes used for a specific task on the filesystem."},
			"ephemeral_storage_available_bytes": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The storage space available (bytes) for the filesystem."},
			"ephemeral_storage_capacity_bytes":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total capacity (bytes) of the filesystems underlying storage."},
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
			"age":                         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"restarts":                    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of times the container has been restarted."},
			"ready":                       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Describes whether the pod is ready to serve requests."},
			"available":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of containers"},
			"cpu_usage":                   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The sum of the cpu usage of all containers in this Pod."},
			"cpu_usage_base100":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized cpu usage, with a maximum of 100%. (Experimental)"},
			"cpu_usage_millicores":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationMS, Desc: "Total CPU usage (sum of all cores) averaged over the sample window."},
			"cpu_limit_millicores":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationMS, Desc: "Max limits for CPU resources."},
			"memory_usage_bytes":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The sum of the memory usage of all containers in this Pod (Deprecated use `mem_usage`)."},
			"memory_capacity":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory in the host machine (Deprecated use `mem_capacity`)."},
			"memory_used_percent":         &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory (refer from `mem_used_percent`"},
			"mem_usage":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The sum of the memory usage of all containers in this Pod."},
			"mem_limit":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The sum of the memory limit of all containers in this Pod."},
			"mem_capacity":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory in the host machine."},
			"mem_used_percent":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory is calculated based on the capacity of host machine."},
			"mem_used_percent_base_limit": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory is calculated based on the limit."},
			"message":                     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}

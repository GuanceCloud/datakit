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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
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
	registerResource("pod", true, newPod)
	registerMeasurements(&podMetric{}, &podObject{})
}

type pod struct {
	client    k8sClient
	continued string
}

func newPod(client k8sClient) resource {
	return &pod{client: client}
}

func (p *pod) hasNext() bool { return p.continued != "" }

func (p *pod) getMetadata(ctx context.Context, ns string) (metadata, error) {
	opt := metav1.ListOptions{
		Limit:    queryLimit,
		Continue: p.continued,
	}

	list, err := p.client.GetPods(ns).List(ctx, opt)
	if err != nil {
		return nil, err
	}

	p.continued = list.Continue
	return &podMetadata{p, list}, nil
}

type podMetadata struct {
	opt  *pod
	list *apicorev1.PodList
}

func (m *podMetadata) transformMetric() pointKVs {
	var res pointKVs

	for idx, item := range m.list.Items {
		met := typed.NewPointKV(podMetricMeasurement)

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("pod", item.Name)
		met.SetTag("namespace", item.Namespace)

		met.SetField("ready", 0)
		// "scheduled", "volumes_persistentvolumeclaims_readonly","unschedulable"

		containerReadyCount := 0
		for _, cs := range item.Status.ContainerStatuses {
			if cs.State.Running != nil {
				containerReadyCount++
			}
		}
		met.SetField("ready", containerReadyCount)

		if len(item.OwnerReferences) != 0 {
			switch item.OwnerReferences[0].Kind {
			case "ReplicaSet":
				if hash, ok := item.Labels["pod-template-hash"]; ok {
					met.SetTag("deployment", strings.TrimRight(item.OwnerReferences[0].Name, "-"+hash))
				}
			case "DaemonSet":
				met.SetTag("daemonset", item.OwnerReferences[0].Name)
			case "StatefulSet":
				met.SetTag("statefulset", item.OwnerReferences[0].Name)
			default:
				// skip
			}
		}

		if setExtraK8sLabelAsTags() {
			for k, v := range item.Labels {
				met.SetTag(transLabelKey(k), v)
			}
		}

		if canCollectPodMetrics() {
			p, err := queryPodFromMetricsServer(context.Background(), m.opt.client, &m.list.Items[idx])
			if err != nil {
				klog.Warnf("pod %s from metrics-server fail, err: %s, skip", item.Name, err)
			} else {
				met.SetTags(p.Tags())
				met.SetFields(p.Fields())
			}
		}

		met.SetCustomerTags(item.Labels, getGlobalCustomerKeys())
		res = append(res, met)
	}

	return res
}

func (m *podMetadata) transformObject() pointKVs {
	var res pointKVs

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
		obj.SetCustomerTags(item.Labels, getGlobalCustomerKeys())

		if setExtraK8sLabelAsTags() {
			for k, v := range item.Labels {
				obj.SetTag(transLabelKey(k), v)
			}
		}

		if canCollectPodMetrics() {
			p, err := queryPodFromMetricsServer(context.Background(), m.opt.client, &m.list.Items[idx])
			if err != nil {
				klog.Warnf("pod %s from metrics-server fail, err: %s, skip", item.Name, err)
			} else {
				obj.SetTags(p.Tags())
				obj.SetFields(p.Fields())
			}
		}

		obj.SetCustomerTags(item.Labels, getGlobalCustomerKeys())
		res = append(res, obj)
	}

	return res
}

func queryPodFromMetricsServer(ctx context.Context, client k8sClient, item *apicorev1.Pod) (*typed.PointKV, error) {
	p := typed.NewPointKV("NULL")

	podMet, err := queryPodMetrics(ctx, client, item.Name, item.Namespace)
	if err != nil {
		return p, err
	}

	p.SetField("cpu_usage", podMet.cpuUsage)
	p.SetField("cpu_usage_millicores", podMet.cpuUsageMilliCores)
	p.SetField("mem_usage", podMet.memoryUsageBytes)

	cpuCapacityMillicores, memCapacity := getCapacityFromNode(ctx, client, item.Spec.NodeName)
	p.SetField("cpu_capacity_millicores", cpuCapacityMillicores)
	p.SetField("mem_capacity", memCapacity)

	if cpuCapacityMillicores != 0 {
		cores := cpuCapacityMillicores / 1e3
		x := podMet.cpuUsage / float64(cores)
		if math.IsNaN(x) {
			x = 0.0
		}
		p.SetField("cpu_usage_base100", x)
	}

	if memCapacity != 0 {
		p.SetField("mem_used_percent", float64(podMet.memoryUsageBytes)/float64(memCapacity)*100)
	}

	memLimit := getMemoryLimitFromResource(item.Spec.Containers)
	if memLimit != 0 {
		p.SetField("mem_used_percent_base_limit", float64(podMet.memoryUsageBytes)/float64(memLimit)*100)
	}

	// maintain compatibility
	p.SetField("memory_usage_bytes", p.GetField("mem_usage"))
	p.SetField("memory_capacity", p.GetField("mem_capacity"))
	p.SetField("memory_used_percent", p.GetField("mem_used_percent"))

	return p, nil
}

type podMetric struct{}

func (*podMetric) LineProto() (*dkpt.Point, error) { return nil, nil }

//nolint:lll
func (*podMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: podMetricMeasurement,
		Desc: "The metric of the Kubernetes Pod.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":         inputs.NewTagInfo("The UID of pod."),
			"pod":         inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":   inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"deployment":  inputs.NewTagInfo("The name of the Deployment which the object belongs to."),
			"daemonset":   inputs.NewTagInfo("The name of the DaemonSet which the object belongs to."),
			"statefulset": inputs.NewTagInfo("The name of the StatefulSet which the object belongs to."),
		},
		Fields: map[string]interface{}{
			"ready":                       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Describes whether the pod is ready to serve requests."},
			"cpu_usage":                   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The sum of the cpu usage of all containers in this Pod."},
			"cpu_usage_base100":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized cpu usage, with a maximum of 100%. (Experimental)"},
			"memory_usage_bytes":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The sum of the memory usage of all containers in this Pod (Deprecated use `mem_usage`)."},
			"memory_capacity":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory in the host machine (Deprecated use `mem_capacity`)."},
			"memory_used_percent":         &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory (refer from `mem_used_percent`"},
			"mem_usage":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The sum of the memory usage of all containers in this Pod."},
			"mem_limit":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The sum of the memory limit of all containers in this Pod."},
			"mem_capacity":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The total memory in the host machine."},
			"mem_used_percent":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory is calculated based on the capacity of host machine."},
			"mem_used_percent_base_limit": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage usage of the memory is calculated based on the limit."},
		},
	}
}

type podObject struct{}

func (*podObject) LineProto() (*dkpt.Point, error) { return nil, nil }

//nolint:lll
func (*podObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: podObjectMeasurement,
		Desc: "The object of the Kubernetes Pod.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":        inputs.NewTagInfo("The UID of Pod."),
			"uid":         inputs.NewTagInfo("The UID of Pod."),
			"pod_name":    inputs.NewTagInfo("Name must be unique within a namespace."),
			"node_name":   inputs.NewTagInfo("NodeName is a request to schedule this pod onto a specific node."),
			"namespace":   inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"phase":       inputs.NewTagInfo("The phase of a Pod is a simple, high-level summary of where the Pod is in its lifecycle.(Pending/Running/Succeeded/Failed/Unknown)"),
			"status":      inputs.NewTagInfo("Reason the container is not yet running."),
			"qos_class":   inputs.NewTagInfo("The Quality of Service (QOS) classification assigned to the pod based on resource requirements"),
			"deployment":  inputs.NewTagInfo("The name of the Deployment which the object belongs to."),
			"daemonset":   inputs.NewTagInfo("The name of the DaemonSet which the object belongs to."),
			"statefulset": inputs.NewTagInfo("The name of the StatefulSet which the object belongs to."),
		},
		Fields: map[string]interface{}{
			"age":                         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"restart":                     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of times the container has been restarted. (Deprecated, use restarts)"},
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

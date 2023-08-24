// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apicorev1 "k8s.io/api/core/v1"
)

//nolint:gochecknoinits
func init() {
	registerMetricResource("pod", gatherPodMetric)
	registerObjectResource("pod", gatherPodObject)
	registerMeasurement(&podMetric{})
	registerMeasurement(&podObject{})
}

func gatherPodMetric(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetPods().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}

	var res []measurement

	for idx, pod := range list.Items {
		met := composePodMetric(&list.Items[idx])

		if b, ok := ctx.Value(canCollectPodMetricsKey).(bool); ok && b {
			p, err := queryPodFromMetricsServer(ctx, client, &list.Items[idx])
			if err != nil {
				klog.Warnf("pod %s from metrics-server fail, err: %s, skip", pod.Name, err)
			} else {
				met.SetTags(p.Tags())
				met.SetFields(p.Fields())
			}
		}

		if set, ok := ctx.Value(setExtraK8sLabelAsTagsKey).(bool); ok && set {
			for k, v := range list.Items[idx].Labels {
				met.SetTag(transLabelKey(k), v)
			}
		}

		if len(pod.OwnerReferences) != 0 {
			switch pod.OwnerReferences[0].Kind {
			case "ReplicaSet":
				if hash, ok := pod.Labels["pod-template-hash"]; ok {
					met.SetTag("deployment", strings.TrimRight(pod.OwnerReferences[0].Name, "-"+hash))
				}
			case "DaemonSet":
				met.SetTag("daemonset", pod.OwnerReferences[0].Name)
			case "StatefulSet":
				met.SetTag("statefulset", pod.OwnerReferences[0].Name)
			default:
				// skip
			}
		}

		met.SetCustomerTags(pod.Labels, getGlobalCustomerKeys())
		res = append(res, &podMetric{*met})
	}

	return res, nil
}

func composePodMetric(item *apicorev1.Pod) *typed.PointKV {
	met := typed.NewPointKV()

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

	return &met
}

func gatherPodObject(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetPods().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}

	var res []measurement

	for idx, pod := range list.Items {
		obj := composePodObject(&list.Items[idx])

		if b, ok := ctx.Value(canCollectPodMetricsKey).(bool); ok && b {
			p, err := queryPodFromMetricsServer(ctx, client, &list.Items[idx])
			if err != nil {
				klog.Warnf("pod %s from metrics-server fail, err: %s, skip", pod.Name, err)
			} else {
				obj.SetTags(p.Tags())
				obj.SetFields(p.Fields())
			}
		}

		if set, ok := ctx.Value(setExtraK8sLabelAsTagsKey).(bool); ok && set {
			for k, v := range pod.Labels {
				obj.SetTag(transLabelKey(k), v)
			}
		}

		if len(pod.OwnerReferences) != 0 {
			switch pod.OwnerReferences[0].Kind {
			case "ReplicaSet":
				if hash, ok := pod.Labels["pod-template-hash"]; ok {
					obj.SetTag("deployment", strings.TrimRight(pod.OwnerReferences[0].Name, "-"+hash))
				}
			case "DaemonSet":
				obj.SetTag("daemonset", pod.OwnerReferences[0].Name)
			case "StatefulSet":
				obj.SetTag("statefulset", pod.OwnerReferences[0].Name)
			default:
				// skip
			}
		}

		obj.SetCustomerTags(pod.Labels, getGlobalCustomerKeys())
		res = append(res, &podObject{*obj})
	}

	return res, nil
}

func composePodObject(item *apicorev1.Pod) *typed.PointKV {
	obj := typed.NewPointKV()

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

	return &obj
}

func queryPodFromMetricsServer(ctx context.Context, client k8sClient, item *apicorev1.Pod) (typed.PointKV, error) {
	p := typed.NewPointKV()

	podMet, err := queryPodMetrics(ctx, client, item.Name, item.Namespace)
	if err != nil {
		return p, err
	}

	p.SetField("cpu_usage", float64(podMet.cpuUsage)*100) // percentage
	p.SetField("mem_usage", podMet.memoryUsageBytes)

	memLimit := getMemoryCapacityFromResourceLimit(item.Spec.Containers)
	p.SetField("mem_limit", memLimit)
	if memLimit != 0 {
		p.SetField("mem_used_percent_base_limit", float64(podMet.memoryUsageBytes)/float64(memLimit))
	}

	cpuCapacity, memCapacity := getCapacityFromNode(ctx, client, item.Spec.NodeName)
	p.SetField("cpu_capacity", cpuCapacity)
	p.SetField("mem_capacity", memCapacity)

	if cpuCapacity != 0 {
		cores := cpuCapacity / 1000
		p.SetField("cpu_usage_base100", float64(podMet.cpuUsage)*100/float64(cores))
	}

	if memCapacity != 0 {
		p.SetField("mem_used_percent", float64(podMet.memoryUsageBytes)/float64(memCapacity))
	}

	// maintain compatibility
	p.SetField("memory_usage_bytes", p.GetField("mem_usage"))
	p.SetField("memory_capacity", p.GetField("mem_capacity"))
	p.SetField("memory_used_percent", p.GetField("mem_used_percent"))

	return p, nil
}

type podMetric struct{ typed.PointKV }

func (p *podMetric) namespace() string { return p.GetTag("namespace") }

func (p *podMetric) addExtraTags(m map[string]string) { p.SetTags(m) }

func (p *podMetric) LineProto() (*point.Point, error) {
	return point.NewPoint("kube_pod", p.Tags(), p.Fields(), metricOpt)
}

//nolint:lll
func (*podMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_pod",
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
			"cpu_usage_base100":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized cpu usage, with a maximum of 100%."},
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

type podObject struct{ typed.PointKV }

func (p *podObject) namespace() string { return p.GetTag("namespace") }

func (p *podObject) addExtraTags(m map[string]string) { p.SetTags(m) }

func (p *podObject) LineProto() (*point.Point, error) {
	return point.NewPoint("kubelet_pod", p.Tags(), p.Fields(), objectOpt)
}

//nolint:lll
func (*podObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubelet_pod",
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
			"cpu_usage_base100":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The normalized cpu usage, with a maximum of 100%."},
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

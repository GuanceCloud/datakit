package container

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/core/v1"
)

const k8sPodObjectName = "kubelet_pod"

func gatherPod(client k8sClientX, extraTags map[string]string) (k8sResourceStats, error) {
	list, err := client.getPods().List(context.Background(), metaV1ListOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods resource: %w", err)
	}
	if len(list.Items) == 0 {
		return nil, nil
	}
	return exportPod(list.Items, extraTags), nil
}

func exportPod(items []v1.Pod, extraTags tagsType) k8sResourceStats {
	res := newK8sResourceStats()

	for idx, item := range items {
		obj := newPod()
		obj.tags = map[string]string{
			"name":      fmt.Sprintf("%v", item.UID),
			"pod_name":  item.Name,
			"node_name": item.Spec.NodeName,
			"host":      item.Spec.NodeName, // 指定 pod 所在的 host
			"phase":     fmt.Sprintf("%v", item.Status.Phase),
			"qos_class": fmt.Sprintf("%v", item.Status.QOSClass),
			"state":     fmt.Sprintf("%v", item.Status.Phase), // Depercated
			"status":    fmt.Sprintf("%v", item.Status.Phase),
		}

		for _, ref := range item.OwnerReferences {
			if ref.Kind == "ReplicaSet" {
				obj.tags["replica_set"] = ref.Name
				break
			}
		}
		if deployment := getDeployment(item.Labels["app"], item.Namespace); deployment != "" {
			obj.tags["deployment"] = deployment
		}

		obj.tags.addValueIfNotEmpty("cluster_name", item.ClusterName)
		obj.tags.addValueIfNotEmpty("namespace", defaultNamespace(item.Namespace))
		obj.tags.append(extraTags)

		for _, containerStatus := range item.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				obj.tags["state"] = containerStatus.State.Waiting.Reason // Depercated
				obj.tags["status"] = containerStatus.State.Waiting.Reason
				break
			}
		}

		containerAllCount := len(item.Status.ContainerStatuses)
		containerReadyCount := 0
		for _, cs := range item.Status.ContainerStatuses {
			if cs.State.Running != nil {
				containerReadyCount++
			}
		}

		obj.fields = map[string]interface{}{
			"age":         int64(time.Since(item.CreationTimestamp.Time).Seconds()),
			"ready":       containerReadyCount,
			"availale":    containerAllCount,
			"create_time": item.CreationTimestamp.Time.Unix(),
		}

		restartCount := 0
		for _, containerStatus := range item.Status.InitContainerStatuses {
			restartCount += int(containerStatus.RestartCount)
		}
		for _, containerStatus := range item.Status.ContainerStatuses {
			restartCount += int(containerStatus.RestartCount)
		}
		for _, containerStatus := range item.Status.EphemeralContainerStatuses {
			restartCount += int(containerStatus.RestartCount)
		}
		obj.fields["restart"] = restartCount // Depercated
		obj.fields["restarts"] = restartCount

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)

		obj.time = time.Now()
		res.set(defaultNamespace(item.Namespace), obj)

		if err := tryRunInput(&items[idx]); err != nil {
			l.Warnf("failed to run input(discovery), %w", err)
		}
	}

	return res
}

//nolint:deadcode,unused
func getPodLables(k8sClient k8sClientX, podname, podnamespace string) (map[string]string, error) {
	pod, err := queryPodMetaData(k8sClient, podname, podnamespace)
	if err != nil {
		return nil, err
	}
	return pod.labels(), nil
}

func getPodAnnotations(k8sClient k8sClientX, podname, podnamespace string) (map[string]string, error) {
	pod, err := queryPodMetaData(k8sClient, podname, podnamespace)
	if err != nil {
		return nil, err
	}
	return pod.annotations(), nil
}

type podMeta struct{ *v1.Pod }

func queryPodMetaData(k8sClient k8sClientX, podname, podnamespace string) (*podMeta, error) {
	pod, err := k8sClient.getPodsForNamespace(podnamespace).Get(context.Background(), podname, metaV1GetOption)
	if err != nil {
		return nil, err
	}
	return &podMeta{pod}, nil
}

func (item *podMeta) labels() map[string]string { return item.Labels }

func (item *podMeta) annotations() map[string]string { return item.Annotations }

func (item *podMeta) containerImage(name string) string {
	for _, container := range item.Spec.Containers {
		if container.Name == name {
			return container.Image
		}
	}
	return ""
}

func (item *podMeta) replicaSet() string {
	for _, ref := range item.OwnerReferences {
		if ref.Kind == "ReplicaSet" {
			return ref.Name
		}
	}
	return ""
}

type pod struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func newPod() *pod {
	return &pod{
		tags:   make(tagsType),
		fields: make(fieldsType),
	}
}

func (p *pod) LineProto() (*io.Point, error) {
	return io.NewPoint(k8sPodObjectName, p.tags, p.fields, &io.PointOption{Time: p.time, Category: datakit.Object})
}

//nolint:lll
func (*pod) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: k8sPodObjectName,
		Desc: "Kubernetes pod 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("UID"),
			"pod_name":     inputs.NewTagInfo("Name must be unique within a namespace."),
			"node_name":    inputs.NewTagInfo("NodeName is a request to schedule this pod onto a specific node."),
			"cluster_name": inputs.NewTagInfo("The name of the cluster which the object belongs to."),
			"namespace":    inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"phase":        inputs.NewTagInfo("The phase of a Pod is a simple, high-level summary of where the Pod is in its lifecycle.(Pending/Running/Succeeded/Failed/Unknown)"),
			"state":        inputs.NewTagInfo("Reason the container is not yet running. (Depercated, use status)"),
			"status":       inputs.NewTagInfo("Reason the container is not yet running."),
			"qos_class":    inputs.NewTagInfo("The Quality of Service (QOS) classification assigned to the pod based on resource requirements"),
			"deployment":   inputs.NewTagInfo("The name of the deployment which the object belongs to. (Probably empty)"),
			"replica_set":  inputs.NewTagInfo("The name of the replicaSet which the object belongs to. (Probably empty)"),
		},
		Fields: map[string]interface{}{
			"age":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"create_time": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampSec, Desc: "CreationTimestamp is a timestamp representing the server time when this object was created.(second)"},
			"restart":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of times the container has been restarted. (Depercated, use restarts)"},
			"restarts":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of times the container has been restarted."},
			"ready":       &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "container ready"},
			"available":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "container count"},
			"annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "kubernetes annotations"},
			"message":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&pod{})
}

package kubernetes

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const kubernetesPodName = "kubernetes_pods"

type pod struct {
	client interface {
		getPods() (*corev1.PodList, error)
	}
}

func (p pod) Gather() {
	list, err := p.client.getPods()
	if err != nil {
		l.Errorf("failed of get pods resource: %s", err)
		return
	}

	for _, obj := range list.Items {
		containerAllCount := len(obj.Status.ContainerStatuses)
		containerReadyCount := 0
		for _, cs := range obj.Status.ContainerStatuses {
			if cs.State.Running != nil {
				containerReadyCount++
			}
		}

		tags := map[string]string{
			"name":         fmt.Sprintf("%v", obj.UID),
			"pod_name":     obj.Name,
			"cluster_name": obj.ClusterName,
			"node_name":    obj.Spec.NodeName,
			"status":       fmt.Sprintf("%v", obj.Status.Phase),
			"phase":        fmt.Sprintf("%v", obj.Status.Phase),
			"qos_class":    fmt.Sprintf("%v", obj.Status.QOSClass),
			"namespace":    obj.Namespace,
		}

		fields := map[string]interface{}{
			"ready":       fmt.Sprintf("%d/%d", containerReadyCount, containerAllCount),
			"age":         int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"create_time": obj.CreationTimestamp.Time,
		}

		restartCount := 0
		for _, containerStatus := range obj.Status.InitContainerStatuses {
			restartCount += int(containerStatus.RestartCount)
		}
		for _, containerStatus := range obj.Status.ContainerStatuses {
			restartCount += int(containerStatus.RestartCount)
		}
		for _, containerStatus := range obj.Status.EphemeralContainerStatuses {
			restartCount += int(containerStatus.RestartCount)
		}
		fields["restarts"] = restartCount

		addJSONStringToMap("kubernetes_labels", obj.Labels, fields)
		addJSONStringToMap("kubernetes_annotations", obj.Annotations, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesPodName, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
		} else {
			if err := io.Feed(inputName, datakit.Object, []*io.Point{pt}, nil); err != nil {
				l.Error(err)
			}
		}
	}
}

func (*pod) LineProto() (*io.Point, error) {
	return nil, nil
}

func (*pod) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesPodName,
		Desc: kubernetesPodName,
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo(""),
			"pod_name":     inputs.NewTagInfo(""),
			"cluster_name": inputs.NewTagInfo(""),
			"node_name":    inputs.NewTagInfo(""),
			"status":       inputs.NewTagInfo(""),
			"phase":        inputs.NewTagInfo(""),
			"namespace":    inputs.NewTagInfo(""),
			"qos_class":    inputs.NewTagInfo(""),
		},
		Fields: map[string]interface{}{
			"age":                    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: ""},
			"restarts":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: ""},
			"ready":                  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubernetes_labels":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubernetes_annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"message":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}

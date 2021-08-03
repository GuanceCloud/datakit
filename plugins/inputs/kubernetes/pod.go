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
	tags map[string]string
}

func (p *pod) Gather() {
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
			"phase":        fmt.Sprintf("%v", obj.Status.Phase),
			"qos_class":    fmt.Sprintf("%v", obj.Status.QOSClass),
			"namespace":    obj.Namespace,
			"status":       fmt.Sprintf("%v", obj.Status.Phase),
		}
		for k, v := range p.tags {
			tags[k] = v
		}

		for _, containerStatus := range obj.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				tags["status"] = containerStatus.State.Waiting.Reason
				break
			}
		}

		fields := map[string]interface{}{
			"ready":       fmt.Sprintf("%d/%d", containerReadyCount, containerAllCount),
			"age":         int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"create_time": obj.CreationTimestamp.Time.Unix(),
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

		addMapToFields("annotations", obj.Annotations, fields)
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

func (*pod) Resource() { /*empty interface*/ }

func (*pod) LineProto() (*io.Point, error) { return nil, nil }

func (*pod) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesPodName,
		Desc: fmt.Sprintf("%s 对象数据", kubernetesPodName),
		Type: "object",
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("pod UID"),
			"pod_name":     inputs.NewTagInfo("pod 名称"),
			"node_name":    inputs.NewTagInfo("所在 node"),
			"cluster_name": inputs.NewTagInfo("所在 cluster"),
			"namespace":    inputs.NewTagInfo("所在命名空间"),
			"phase":        inputs.NewTagInfo("所处阶段，Pending/Running/Succeeded/Failed/Unknown"),
			"status":       inputs.NewTagInfo("当前状态"),
			"qos_class":    inputs.NewTagInfo("QOS Class"),
		},
		Fields: map[string]interface{}{
			"age":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "存活时长，单位为秒"},
			"create_time": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "创建时间戳，精度为秒"},
			"restarts":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "所有容器的重启次数"},
			"ready":       &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "就绪"},
			"annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "kubernetes annotations"},
			"message":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "详情数据"},
		},
	}
}

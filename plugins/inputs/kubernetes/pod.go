package kubernetes

import (
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	corev1 "k8s.io/api/core/v1"
)

const kubernetesPodName = "kubernetes_pods"

type pod struct {
	client interface {
		getPods() (*corev1.PodList, error)
	}
	tags         map[string]string
	promExporter *PromExporter
}

func (p *pod) Gather() {
	start := time.Now()
	var pts []*io.Point

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
			"age":         int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"ready":       containerReadyCount,
			"availale":    containerAllCount,
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
		addLabelToFields(obj.Labels, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesPodName, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
		} else {
			pts = append(pts, pt)
		}
	}

	if err := io.Feed(inputName, datakit.Object, pts, &io.Option{CollectCost: time.Since(start)}); err != nil {
		l.Error(err)
	}
}

const annotationExportKey = "datakit/prom.exporter"

func (p *pod) Export() {
	if p.promExporter == nil {
		p.promExporter = NewPromExporter()
	}

	list, err := p.client.getPods()
	if err != nil {
		l.Errorf("failed of get pods resource: %s", err)
		return
	}

	for _, obj := range list.Items {
		config := obj.Annotations[annotationExportKey]
		strings.ReplaceAll(config, "$IP", obj.Status.PodIP)
		strings.ReplaceAll(config, "$NAMESPACE", obj.Namespace)
		strings.ReplaceAll(config, "$PODNAME", obj.Name)

		if err := p.promExporter.TryRun(config); err != nil {
			l.Warn(err)
		}
	}
}

func (p *pod) Stop() {
	if p.promExporter != nil {
		p.promExporter.Stop()
	}
}

func (*pod) Resource() { /*empty interface*/ }

func (*pod) LineProto() (*io.Point, error) { return nil, nil }

func (*pod) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesPodName,
		Desc: "Kubernetes pod 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("UID"),
			"pod_name":     inputs.NewTagInfo("Name must be unique within a namespace."),
			"node_name":    inputs.NewTagInfo("NodeName is a request to schedule this pod onto a specific node."),
			"cluster_name": inputs.NewTagInfo("The name of the cluster which the object belongs to."),
			"namespace":    inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"phase":        inputs.NewTagInfo("The phase of a Pod is a simple, high-level summary of where the Pod is in its lifecycle.(Pending/Running/Succeeded/Failed/Unknown)"),
			"status":       inputs.NewTagInfo("Reason the container is not yet running."),
			"qos_class":    inputs.NewTagInfo("The Quality of Service (QOS) classification assigned to the pod based on resource requirements"),
		},
		Fields: map[string]interface{}{
			"age":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"create_time": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "CreationTimestamp is a timestamp representing the server time when this object was created.(second)"},
			"restarts":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of times the container has been restarted"},
			"ready":       &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "TODO"},
			"available":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "TODO"},
			"annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "kubernetes annotations"},
			"message":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
		},
	}
}

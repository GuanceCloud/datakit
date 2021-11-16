package kubernetes

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	corev1 "k8s.io/api/core/v1"
	kubev1core "k8s.io/client-go/kubernetes/typed/core/v1"
)

const kubernetesPodName = "kubernetes_pods"

type podclient interface {
	getPods(namespace string) kubev1core.PodInterface
}

type pod struct {
	client    podclient
	tags      map[string]string
	discovery *Discovery
}

func (p *pod) Gather() {
	start := time.Now()
	var pts []*io.Point

	list, err := p.client.getPods("").List(context.Background(), metav1ListOption)
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
			"name":      fmt.Sprintf("%v", obj.UID),
			"pod_name":  obj.Name,
			"node_name": obj.Spec.NodeName,
			"phase":     fmt.Sprintf("%v", obj.Status.Phase),
			"qos_class": fmt.Sprintf("%v", obj.Status.QOSClass),
			"status":    fmt.Sprintf("%v", obj.Status.Phase),
		}
		if obj.ClusterName != "" {
			tags["cluster_name"] = obj.ClusterName
		}
		if obj.Namespace != "" {
			tags["namespace"] = obj.Namespace
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
			"age":         int64(time.Since(obj.CreationTimestamp.Time).Seconds()),
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

	if len(pts) == 0 {
		l.Debug("no points")
		return
	}

	if err := io.Feed(inputName, datakit.Object, pts, &io.Option{CollectCost: time.Since(start)}); err != nil {
		l.Error(err)
	}
}

func (p *pod) Export() {
	if p.discovery == nil {
		p.discovery = NewDiscovery()
	}

	l.Debug("k8s export")

	list, err := p.client.getPods("").List(context.Background(), metav1ListOption)
	if err != nil {
		l.Errorf("failed of get pods resource: %s", err)
		return
	}

	p.run(list)
}

const (
	annotationPromExport  = "datakit/prom.instances"
	annotationPromIPIndex = "datakit/prom.instances.ip_index"

	annotationPodLogging = "datakit/pod.logging"
)

func (p *pod) run(list *corev1.PodList) {
	for idx, obj := range list.Items {
		func() {
			config, ok := obj.Annotations[annotationPromExport]
			if !ok {
				return
			}
			l.Info("k8s export, find prom export")
			if !shouldForkInput(obj.Spec.NodeName) {
				l.Debugf("should not fork input, pod-nodeName:%s", obj.Spec.NodeName)
				return
			}

			config = complatePromConfig(config, &list.Items[idx])
			if err := p.discovery.TryRun("prom", config); err != nil {
				l.Warn(err)
			}
		}()

		func() {
			config, ok := obj.Annotations[annotationPodLogging]
			if !ok {
				return
			}
			l.Infof("k8s export, find podlogging, namespace:%s, UID:%s", obj.Namespace, string(obj.UID))
			exist, md5Str := p.discovery.IsExist(string(obj.UID))
			if exist {
				return
			}

			podlog := podlogging{}
			if err := toml.Unmarshal([]byte(config), &podlog); err != nil {
				l.Errorf("podlogging config unmarshal err: %s", err)
				return
			}

			p.discovery.addList(md5Str)

			go func(namespace, name string) {
				l.Infof("discovery: add pod-logging, namespace:%s, podName:%s", namespace, name)
				podlog.run(p.client, namespace, name)
			}(obj.Namespace, obj.Name)
		}()
	}
}

func (*pod) Stop() {}

func (*pod) Resource() { /*empty interface*/ }

func (*pod) LineProto() (*io.Point, error) { return nil, nil }

//nolint:lll
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

func complatePromConfig(config string, podObj *corev1.Pod) string {
	podIP := podObj.Status.PodIP

	func() {
		indexStr, ok := podObj.Annotations[annotationPromIPIndex]
		if !ok {
			return
		}
		idx, err := strconv.Atoi(indexStr)
		if err != nil {
			l.Warnf("annotation prom.ip_index parse err: %s", err)
			return
		}
		if !(0 <= idx && idx < len(podObj.Status.PodIPs)) {
			l.Warnf("annotation prom.ip_index %d outrange, len(PodIPs) %d", idx, len(podObj.Status.PodIPs))
			return
		}
		podIP = podObj.Status.PodIPs[idx].IP
	}()

	config = strings.ReplaceAll(config, "$IP", podIP)
	config = strings.ReplaceAll(config, "$NAMESPACE", podObj.Namespace)
	config = strings.ReplaceAll(config, "$PODNAME", podObj.Name)

	return config
}

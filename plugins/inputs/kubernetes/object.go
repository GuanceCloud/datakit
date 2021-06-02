package kubernetes

import (
	"encoding/json"
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	corev1 "k8s.io/api/core/v1"
	"time"
)

var objectName = "kube_pod"

type Pod struct {
	Name       string            `json:"name,omitempty"`
	Namespace  string            `json:"namespace,omitempty"`
	Ready      string            `json:"ready,omitempty"`
	Status     string            `json:"Status,omitempty"`
	Restarts   int32             `json:"restarts,omitempty"`
	Age        string            `json:"age,omitempty"`
	PodIp      string            `json:"podIp,omitempty"`
	NodeName   string            `json:"nodeName,omitempty"`
	CreateTime time.Time         `json:"startTime,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

func (i *Input) collectPodsObject() error {
	list, err := i.client.getPods()
	if err != nil {
		i.lastErr = err
		return err
	}

	i.gatherPod(list)
	return nil
}

type podObject struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (o *podObject) LineProto() (*io.Point, error) {
	return io.MakePoint(o.name, o.tags, o.fields, o.ts)
}

func (o *podObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: objectName,
		Desc: "kubernet pod 对象",
		Tags: map[string]interface{}{
			"name":      &inputs.TagInfo{Desc: "pod name"},
			"namespace": &inputs.TagInfo{Desc: "namespace"},
			"nodeName":  &inputs.TagInfo{Desc: "node name"},
		},
		Fields: map[string]interface{}{
			"ready": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "容器ready数/总数",
			},
			"status": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod 状态",
			},
			"restarts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "重启次数",
			},
			"age": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod存活时长",
			},
			"podIp": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod ip",
			},
			"createTime": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod 创建时间",
			},
			"label_xxx": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod lable",
			},
		},
	}
}

func (i *Input) gatherPod(p *corev1.PodList) {
	for _, pod := range p.Items {
		containerAllCount := len(pod.Status.ContainerStatuses)
		containerReadyCount := 0
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Running != nil {
				containerReadyCount++
			}
		}

		pd := &Pod{
			Name:       pod.Name,
			Ready:      fmt.Sprintf("%d/%d", containerReadyCount, containerAllCount),
			Status:     fmt.Sprintf("%v", pod.Status.Phase),
			Restarts:   pod.Status.ContainerStatuses[0].RestartCount,
			Age:        time.Now().Sub(pod.Status.StartTime.Time).String(),
			CreateTime: pod.Status.StartTime.Time,
			PodIp:      pod.Status.PodIP,
			Labels:     pod.Labels,
			NodeName:   pod.Spec.NodeName,
			Namespace:  pod.Namespace,
		}

		tags := map[string]string{
			"name":      pd.Name,
			"status":    pd.Status,
			"nodeName":  pd.NodeName,
			"namespace": pd.Namespace,
		}

		for key, val := range i.Tags {
			tags[key] = val
		}

		fields := map[string]interface{}{
			"ready":      pd.Ready,
			"restarts":   pd.Restarts,
			"age":        pd.Age,
			"createTime": pd.CreateTime.Unix(),
			"podIp":      pd.PodIp,
		}

		for key, item := range pod.Labels {
			key = "label_" + key
			fields[key] = item
		}

		m, err := json.Marshal(pd)
		if err == nil {
			fields["message"] = string(m)
		} else {
			l.Errorf("marshal message err:%s", err.Error())
		}

		obj := &podObject{
			name:   objectName,
			tags:   tags,
			fields: fields,
			ts:     time.Now(),
		}

		i.collectObjectCache = append(i.collectObjectCache, obj)
	}
}

func (i *Input) CollectObject() error {
	if err := i.collectPodsObject(); err != nil {
		return err
	}
	return nil
}

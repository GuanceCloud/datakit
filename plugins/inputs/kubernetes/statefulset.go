package kubernetes

import (
	"context"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/apps/v1"
	"time"
)

var statefulSetMeasurement = "kube_daemonset"

type statefulSet struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *statefulSet) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *statefulSet) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: daemonSetMeasurement,
		Desc: "kubernet daemonSet 对象",
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

func (i *Input) collectStatefulSets(ctx context.Context) error {
	list, err := i.client.getStatefulSets(ctx)
	if err != nil {
		return err
	}
	for _, s := range list.Items {
		i.gatherStatefulSet(s)
	}

	return nil
}

func (i *Input) gatherStatefulSet(s v1.StatefulSet) {
	status := s.Status
	fields := map[string]interface{}{
		"created":             s.GetCreationTimestamp().UnixNano(),
		"generation":          s.Generation,
		"replicas":            status.Replicas,
		"replicas_current":    status.CurrentReplicas,
		"replicas_ready":      status.ReadyReplicas,
		"replicas_updated":    status.UpdatedReplicas,
		"spec_replicas":       *s.Spec.Replicas,
		"observed_generation": s.Status.ObservedGeneration,
	}
	tags := map[string]string{
		"statefulset_name": s.Name,
		"namespace":        s.Namespace,
	}
	// for key, val := range s.Spec.Selector.MatchLabels {
	// 	if ki.selectorFilter.Match(key) {
	// 		tags["selector_"+key] = val
	// 	}
	// }

	m := &statefulSet{
		name:   deploymentMeasurement,
		tags:   tags,
		fields: fields,
		ts:     time.Now(),
	}

	i.collectCache = append(i.collectCache, m)
}

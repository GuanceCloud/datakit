package kubernetes

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/apps/v1"
)

var statefulSetMeasurement = "kube_statefulSet"

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
		Name: statefulSetMeasurement,
		Desc: "kubernetes statefulSet 对象",
		Tags: map[string]interface{}{
			"statefulset_name": &inputs.TagInfo{Desc: "statefulset name"},
			"namespace":        &inputs.TagInfo{Desc: "namespace"},
		},
		Fields: map[string]interface{}{
			"created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "created time",
			},
			"generation": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "A sequence number representing a specific generation of the desired state",
			},
			"replicas": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "replicas is the number of Pods created by the StatefulSet controller",
			},
			"replicas_current": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "currentReplicas is the number of Pods created by the StatefulSet controller from the StatefulSet version indicated by currentRevision",
			},
			"replicas_ready": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "readyReplicas is the number of Pods created by the StatefulSet controller that have a Ready Condition",
			},
			"replicas_updated": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "updatedReplicas is the number of Pods created by the StatefulSet controller from the StatefulSet version",
			},
			"spec_replicas": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "replicas is the desired number of replicas of the given Template",
			},
			"observed_generation": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "observedGeneration is the most recent generation observed for this StatefulSet. It corresponds to the StatefulSet's generation, which is updated on mutation by the API Server",
			},
		},
	}
}

func (i *Input) collectStatefulSets(collector string) error {
	list, err := i.client.getStatefulSets()
	if err != nil {
		return err
	}
	for _, s := range list.Items {
		i.gatherStatefulSet(collector, s)
	}

	return nil
}

func (i *Input) gatherStatefulSet(collector string, s v1.StatefulSet) {
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

	for key, val := range s.Spec.Selector.MatchLabels {
		tags["selector_"+key] = val
	}

	m := &statefulSet{
		name:   statefulSetMeasurement,
		tags:   tags,
		fields: fields,
		ts:     time.Now(),
	}

	i.collectCache[collector] = append(i.collectCache[collector], m)
}

package kubernetes

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const kubernetesReplicaSetName = "kubernetes_replica_sets"

type replicaSet struct {
	client interface {
		getReplicaSets() (*appsv1.ReplicaSetList, error)
	}
	tags map[string]string
}

func (r *replicaSet) Gather() {
	var start = time.Now()
	var pts []*io.Point

	list, err := r.client.getReplicaSets()
	if err != nil {
		l.Errorf("failed of get replicaSet resource: %s", err)
		return
	}

	for _, obj := range list.Items {
		tags := map[string]string{
			"name":             fmt.Sprintf("%v", obj.UID),
			"replica_set_name": obj.Name,
			"cluster_name":     obj.ClusterName,
			"namespace":        obj.Namespace,
		}
		for k, v := range r.tags {
			tags[k] = v
		}

		fields := map[string]interface{}{
			"age":       int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"ready":     obj.Status.ReadyReplicas,
			"available": obj.Status.AvailableReplicas,
		}

		// addMapToFields("selectors", obj.Spec.Selector, fields)
		addMapToFields("annotations", obj.Annotations, fields)
		addLabelToFields(obj.Labels, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesReplicaSetName, tags, fields, time.Now())
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

func (*replicaSet) resource() { /*empty interface*/ }

func (*replicaSet) LineProto() (*io.Point, error) { return nil, nil }

func (*replicaSet) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesReplicaSetName,
		Desc: "Kubernetes replicaSet 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":             inputs.NewTagInfo("UID"),
			"replica_set_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"cluster_name":     inputs.NewTagInfo("The name of the cluster which the object belongs to."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"age":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"ready":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of ready replicas for this replica set."},
			"available":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of available replicas (ready for at least minReadySeconds) for this replica set."},
			"annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "kubernetes annotations"},
			"message":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
			//TODO:
			// "selectors": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "current/desired":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}

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
}

func (r replicaSet) Gather() {
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
		fields := map[string]interface{}{
			"age":   int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"ready": obj.Status.ReadyReplicas,
		}

		addJSONStringToMap("selectors", obj.Spec.Selector, fields)
		addJSONStringToMap("kubernetes_labels", obj.Labels, fields)
		addJSONStringToMap("kubernetes_annotations", obj.Annotations, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesReplicaSetName, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
		} else {
			if err := io.Feed(inputName, datakit.Object, []*io.Point{pt}, nil); err != nil {
				l.Error(err)
			}
		}
	}
}

func (*replicaSet) LineProto() (*io.Point, error) {
	return nil, nil
}

func (*replicaSet) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesReplicaSetName,
		Desc: kubernetesReplicaSetName,
		Tags: map[string]interface{}{
			"name":             inputs.NewTagInfo(""),
			"replica_set_name": inputs.NewTagInfo(""),
			"cluster_name":     inputs.NewTagInfo(""),
			"namespace":        inputs.NewTagInfo(""),
		},
		Fields: map[string]interface{}{
			"age":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: ""},
			"ready":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"selectors": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "current/desired":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubernetes_labels":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubernetes_annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"message":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}

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
			"age":   int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"ready": obj.Status.ReadyReplicas,
		}

		// addJSONStringToMap("selectors", obj.Spec.Selector, fields)
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

func (*replicaSet) resource() { /*empty interface*/ }

func (*replicaSet) LineProto() (*io.Point, error) { return nil, nil }

func (*replicaSet) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesReplicaSetName,
		Desc: fmt.Sprintf("%s 对象数据", kubernetesReplicaSetName),
		Type: datakit.Object,
		Tags: map[string]interface{}{
			"name":             inputs.NewTagInfo("replicaSet UID"),
			"replica_set_name": inputs.NewTagInfo("replicaSet 名称"),
			"cluster_name":     inputs.NewTagInfo("所在 cluster"),
			"namespace":        inputs.NewTagInfo("所在命名空间"),
		},
		Fields: map[string]interface{}{
			"age":                    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "存活时长，单位为秒"},
			"ready":                  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "ready replicas"},
			"kubernetes_annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "k8s annotations"},
			"message":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "详情数据"},
			//TODO:
			// "selectors": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "current/desired":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}

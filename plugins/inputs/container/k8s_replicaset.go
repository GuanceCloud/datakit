package container

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/apps/v1"
)

const k8sReplicaSetName = "kubernetes_replica_sets"

func gatherReplicaSet(client k8sClientX, extraTags map[string]string) (k8sResourceStats, error) {
	list, err := client.getReplicaSets().List(context.Background(), metaV1ListOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get replicaSet resource: %w", err)
	}

	if len(list.Items) == 0 {
		return nil, nil
	}
	return exportReplicaSet(list.Items, extraTags), nil
}

func exportReplicaSet(items []v1.ReplicaSet, extraTags tagsType) k8sResourceStats {
	res := newK8sResourceStats()

	for _, item := range items {
		obj := newReplicaSet()
		obj.tags["name"] = fmt.Sprintf("%v", item.UID)
		obj.tags["replica_set_name"] = item.Name

		obj.tags.addValueIfNotEmpty("cluster_name", item.ClusterName)
		obj.tags.addValueIfNotEmpty("namespace", defaultNamespace(item.Namespace))
		obj.tags.append(extraTags)

		obj.fields["age"] = int64(time.Since(item.CreationTimestamp.Time).Seconds())
		obj.fields["ready"] = item.Status.ReadyReplicas
		obj.fields["available"] = item.Status.AvailableReplicas

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)

		obj.time = time.Now()
		res.set(defaultNamespace(item.Namespace), obj)
	}
	return res
}

type replicaSet struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func newReplicaSet() *replicaSet {
	return &replicaSet{
		tags:   make(tagsType),
		fields: make(fieldsType),
	}
}

func (r *replicaSet) LineProto() (*io.Point, error) {
	return io.NewPoint(k8sReplicaSetName, r.tags, r.fields, &io.PointOption{Time: r.time, Category: datakit.Object})
}

//nolint:lll
func (*replicaSet) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: k8sReplicaSetName,
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
			// TODO:
			// "selectors": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "current/desired":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&replicaSet{})
}

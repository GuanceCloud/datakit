// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apiappsv1 "k8s.io/api/apps/v1"
)

//nolint:gochecknoinits
func init() {
	registerMetricResource("replicaset", gatherReplicasetMetric)
	registerObjectResource("replicaset", gatherReplicasetObject)
	registerMeasurement(&replicasetMetric{})
	registerMeasurement(&replicasetObject{})
}

func gatherReplicasetMetric(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetReplicaSets().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeReplicasetMetric(list), nil
}

func composeReplicasetMetric(list *apiappsv1.ReplicaSetList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		met := typed.NewPointKV()

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("replica_set", item.Name)
		met.SetTag("namespace", item.Namespace)

		met.SetField("fully_labeled_replicas", item.Status.FullyLabeledReplicas)
		met.SetField("replicas", item.Status.Replicas)
		met.SetField("replicas_ready", item.Status.ReadyReplicas)

		if item.Spec.Replicas != nil {
			met.SetField("replicas_desired", *item.Spec.Replicas)
		}

		res = append(res, &replicasetMetric{met})
	}

	return res
}

func gatherReplicasetObject(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetReplicaSets().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeReplicasetObject(list), nil
}

func composeReplicasetObject(list *apiappsv1.ReplicaSetList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		obj := typed.NewPointKV()

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("replica_set_name", item.Name)
		obj.SetTag("namespace", item.Namespace)

		for _, ref := range item.OwnerReferences {
			if ref.Kind == "Deployment" {
				obj.SetTag("deployment", ref.Name)
				break
			}
		}

		obj.SetField("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		obj.SetField("ready", item.Status.ReadyReplicas)
		obj.SetField("available", item.Status.AvailableReplicas)

		if y, err := yaml.Marshal(item); err == nil {
			obj.SetField("yaml", string(y))
		}

		obj.SetFields(transLabels(item.Labels))
		obj.SetField("annotations", typed.MapToJSON(item.Annotations))
		obj.SetField("message", typed.TrimString(obj.String(), maxMessageLength))
		obj.DeleteField("annotations")
		obj.DeleteField("yaml")

		res = append(res, &replicasetObject{obj})
	}

	return res
}

type replicasetMetric struct{ typed.PointKV }

func (r *replicasetMetric) namespace() string { return r.GetTag("namespace") }

func (r *replicasetMetric) addExtraTags(m map[string]string) { r.SetTags(m) }

func (r *replicasetMetric) LineProto() (*point.Point, error) {
	return point.NewPoint("kube_replicaset", r.Tags(), r.Fields(), metricOpt)
}

//nolint:lll
func (*replicasetMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_replicaset",
		Desc: "The metric of the Kubernetes ReplicaSet.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":         inputs.NewTagInfo("The UID of ReplicaSet."),
			"replica_set": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":   inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"replicas_desired":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Replicas is the number of desired replicas."},
			"fully_labeled_replicas": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of fully labeled replicas per ReplicaSet."},
			"replicas_ready":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of ready replicas for this replica set."},
			"replicas":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Replicas is the most recently observed number of replicas."},
		},
	}
}

type replicasetObject struct{ typed.PointKV }

func (r *replicasetObject) namespace() string { return r.GetTag("namespace") }

func (r *replicasetObject) addExtraTags(m map[string]string) { r.SetTags(m) }

func (r *replicasetObject) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes_replica_sets", r.Tags(), r.Fields(), objectOpt)
}

//nolint:lll
func (*replicasetObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_replica_sets",
		Desc: "The object of the Kubernetes ReplicaSet.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":             inputs.NewTagInfo("The UID of ReplicaSet."),
			"replica_set_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"deployment":       inputs.NewTagInfo("The name of the deployment which the object belongs to."),
		},
		Fields: map[string]interface{}{
			"age":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"ready":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of ready replicas for this replica set."},
			"available": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of available replicas (ready for at least minReadySeconds) for this replica set."},
			"message":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}

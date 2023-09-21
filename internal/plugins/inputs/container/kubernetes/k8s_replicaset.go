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
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apiappsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	replicasetMetricMeasurement = "kube_replicaset"
	replicasetObjectMeasurement = "kubernetes_replica_sets"
)

//nolint:gochecknoinits
func init() {
	registerResource("replicaset", true, newReplicaset)
	registerMeasurements(&replicasetMetric{}, &replicasetObject{})
}

type replicaset struct {
	client    k8sClient
	continued string
}

func newReplicaset(client k8sClient) resource {
	return &replicaset{client: client}
}

func (r *replicaset) hasNext() bool { return r.continued != "" }

func (r *replicaset) getMetadata(ctx context.Context, ns string) (metadata, error) {
	opt := metav1.ListOptions{
		Limit:    queryLimit,
		Continue: r.continued,
	}

	list, err := r.client.GetReplicaSets(ns).List(ctx, opt)
	if err != nil {
		return nil, err
	}

	r.continued = list.Continue
	return &replicasetMetadata{list}, nil
}

type replicasetMetadata struct {
	list *apiappsv1.ReplicaSetList
}

func (m *replicasetMetadata) transformMetric() pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		met := typed.NewPointKV(replicasetMetricMeasurement)

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("replicaset", item.Name)
		met.SetTag("replica_set", item.Name) // Deprecated
		met.SetTag("namespace", item.Namespace)

		met.SetField("replicas", item.Status.Replicas)
		met.SetField("replicas_ready", item.Status.ReadyReplicas)
		met.SetField("replicas_available", item.Status.AvailableReplicas)
		met.SetField("fully_labeled_replicas", item.Status.FullyLabeledReplicas)

		if item.Spec.Replicas != nil {
			met.SetField("replicas_desired", *item.Spec.Replicas)
		}

		met.SetCustomerTags(item.Labels, getGlobalCustomerKeys())
		res = append(res, met)
	}

	return res
}

func (m *replicasetMetadata) transformObject() pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		obj := typed.NewPointKV(replicasetObjectMeasurement)

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("uid", fmt.Sprintf("%v", item.UID))
		obj.SetTag("replicaset_name", item.Name)
		obj.SetTag("replica_set_name", item.Name) // Deprecated
		obj.SetTag("namespace", item.Namespace)

		if len(item.OwnerReferences) != 0 {
			switch item.OwnerReferences[0].Kind {
			case "Deployment":
				obj.SetTag("deployment", item.OwnerReferences[0].Name)
			case "StatefulSet":
				obj.SetTag("statefulset", item.OwnerReferences[0].Name)
			default:
				// nil
			}
		}

		obj.SetField("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		obj.SetField("replicas", item.Status.Replicas)
		obj.SetField("replicas_ready", item.Status.ReadyReplicas)
		obj.SetField("replicas_available", item.Status.AvailableReplicas)
		obj.SetField("ready", item.Status.ReadyReplicas)         // Deprecated
		obj.SetField("available", item.Status.AvailableReplicas) // Deprecated

		if item.Spec.Replicas != nil {
			obj.SetField("replicas_desired", *item.Spec.Replicas)
		}

		if y, err := yaml.Marshal(item); err == nil {
			obj.SetField("yaml", string(y))
		}

		obj.SetFields(transLabels(item.Labels))
		obj.SetField("annotations", typed.MapToJSON(item.Annotations))
		obj.SetField("message", typed.TrimString(obj.String(), maxMessageLength))
		obj.DeleteField("annotations")
		obj.DeleteField("yaml")

		obj.SetCustomerTags(item.Labels, getGlobalCustomerKeys())
		res = append(res, obj)
	}

	return res
}

type replicasetMetric struct{}

func (*replicasetMetric) LineProto() (*dkpt.Point, error) { return nil, nil }

//nolint:lll
func (*replicasetMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: replicasetMetricMeasurement,
		Desc: "The metric of the Kubernetes ReplicaSet.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":              inputs.NewTagInfo("The UID of ReplicaSet."),
			"replicaset_name":  inputs.NewTagInfo("Name must be unique within a namespace."),
			"replica_set_name": inputs.NewTagInfo("Name must be unique within a namespace. (Deprecated)"),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"replicas":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The most recently observed number of replicas."},
			"replicas_desired":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of desired replicas."},
			"replicas_ready":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of ready replicas for this replica set."},
			"replicas_available":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of available replicas (ready for at least minReadySeconds) for this replica set."},
			"fully_labeled_replicas": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of fully labeled replicas per ReplicaSet."},
		},
	}
}

type replicasetObject struct{}

func (*replicasetObject) LineProto() (*dkpt.Point, error) { return nil, nil }

//nolint:lll
func (*replicasetObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: replicasetObjectMeasurement,
		Desc: "The object of the Kubernetes ReplicaSet.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":             inputs.NewTagInfo("The UID of ReplicaSet."),
			"uid":              inputs.NewTagInfo("The UID of ReplicaSet."),
			"replicaset_name":  inputs.NewTagInfo("Name must be unique within a namespace."),
			"replica_set_name": inputs.NewTagInfo("Name must be unique within a namespace. (Deprecated)"),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"deployment":       inputs.NewTagInfo("The name of the Deployment which the object belongs to."),
			"statefulset":      inputs.NewTagInfo("The name of the StatefulSet which the object belongs to."),
		},
		Fields: map[string]interface{}{
			"age":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"replicas":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The most recently observed number of replicas."},
			"replicas_desired":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of desired replicas."},
			"replicas_ready":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of ready replicas for this replica set."},
			"replicas_available": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of available replicas (ready for at least minReadySeconds) for this replica set."},
			"message":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
			"ready":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of ready replicas for this replica set. (Deprecated)"},
			"available":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of available replicas (ready for at least minReadySeconds) for this replica set. (Deprecated)"},
		},
	}
}

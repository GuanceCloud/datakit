// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

var (
	_ k8sResourceMetricInterface = (*replicaset)(nil)
	_ k8sResourceObjectInterface = (*replicaset)(nil)
)

type replicaset struct {
	client    k8sClientX
	extraTags map[string]string
	items     []v1.ReplicaSet
}

func newReplicaset(client k8sClientX, extraTags map[string]string) *replicaset {
	return &replicaset{
		client:    client,
		extraTags: extraTags,
	}
}

func (r *replicaset) name() string {
	return "replica_set"
}

func (r *replicaset) pullItems() error {
	list, err := r.client.getReplicaSets().List(context.Background(), metaV1ListOption)
	if err != nil {
		return fmt.Errorf("failed to get replicasets resource: %w", err)
	}
	r.items = list.Items
	return nil
}

func (r *replicaset) metric(election bool) (inputsMeas, error) {
	if err := r.pullItems(); err != nil {
		return nil, err
	}
	var res inputsMeas

	for _, item := range r.items {
		met := &replicasetMetric{
			tags: map[string]string{
				"replica_set": item.Name,
				"namespace":   defaultNamespace(item.Namespace),
			},
			fields: map[string]interface{}{
				"replicas_desired":       *item.Spec.Replicas,
				"fully_labeled_replicas": item.Status.FullyLabeledReplicas,
				"replicas_ready":         item.Status.ReadyReplicas,
				"replicas":               item.Status.Replicas,
			},
			election: election,
		}

		for _, ref := range item.OwnerReferences {
			if ref.Kind == "Deployment" {
				met.tags["deployment"] = ref.Name
				break
			}
		}

		met.tags.append(r.extraTags)
		res = append(res, met)
	}

	count, _ := r.count()
	for ns, c := range count {
		met := &replicasetMetric{
			tags:     map[string]string{"namespace": ns},
			fields:   map[string]interface{}{"count": c},
			election: election,
		}
		met.tags.append(r.extraTags)
		res = append(res, met)
	}

	return res, nil
}

func (r *replicaset) object(election bool) (inputsMeas, error) {
	if err := r.pullItems(); err != nil {
		return nil, err
	}
	var res inputsMeas

	for _, item := range r.items {
		obj := &replicasetObject{
			tags: map[string]string{
				"name":             fmt.Sprintf("%v", item.UID),
				"replica_set_name": item.Name,
				"namespace":        defaultNamespace(item.Namespace),
			},
			fields: map[string]interface{}{
				"age":       int64(time.Since(item.CreationTimestamp.Time).Seconds()),
				"ready":     item.Status.ReadyReplicas,
				"available": item.Status.AvailableReplicas,
			},
			election: election,
		}

		for _, ref := range item.OwnerReferences {
			if ref.Kind == "Deployment" {
				obj.tags["deployment"] = ref.Name
				break
			}
		}

		if y, err := yaml.Marshal(item); err != nil {
			l.Debugf("failed to get object yaml %s, namespace %s, name %s, ignored", err.Error(), item.Namespace, item.Name)
		} else {
			obj.fields["yaml"] = string(y)
		}

		obj.tags.append(r.extraTags)

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)
		obj.fields.delete("annotations")
		obj.fields.delete("yaml")

		res = append(res, obj)
	}

	return res, nil
}

func (r *replicaset) count() (map[string]int, error) {
	if err := r.pullItems(); err != nil {
		return nil, err
	}

	m := make(map[string]int)
	for _, item := range r.items {
		m[defaultNamespace(item.Namespace)]++
	}
	if len(m) == 0 {
		m["default"] = 0
	}

	return m, nil
}

type replicasetMetric struct {
	tags     tagsType
	fields   fieldsType
	election bool
}

func (r *replicasetMetric) LineProto() (*point.Point, error) {
	return point.NewPoint("kube_replicaset", r.tags, r.fields, point.MOptElectionV2(r.election))
}

//nolint:lll
func (*replicasetMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_replicaset",
		Desc: "Kubernetes replicaset 指标数据",
		Type: "metric",
		Tags: map[string]interface{}{
			"replica_set": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":   inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"deployment":  inputs.NewTagInfo("The name of the deployment which the object belongs to."),
		},
		Fields: map[string]interface{}{
			"count":                  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of replicasets"},
			"replicas_desired":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Replicas is the number of desired replicas."},
			"fully_labeled_replicas": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of fully labeled replicas per ReplicaSet."},
			"replicas_ready":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of ready replicas for this replica set."},
			"replicas":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Replicas is the most recently oberved number of replicas."},
		},
	}
}

type replicasetObject struct {
	tags     tagsType
	fields   fieldsType
	election bool
}

func (r *replicasetObject) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes_replica_sets", r.tags, r.fields, point.OOptElectionV2(r.election))
}

//nolint:lll
func (*replicasetObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_replica_sets",
		Desc: "Kubernetes replicaset 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":             inputs.NewTagInfo("UID"),
			"replica_set_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"deployment":       inputs.NewTagInfo("The name of the deployment which the object belongs to."),
		},
		Fields: map[string]interface{}{
			"age":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"ready":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of ready replicas for this replica set."},
			"available": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of available replicas (ready for at least minReadySeconds) for this replica set."},
			"message":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerK8sResourceMetric(func(c k8sClientX, m map[string]string) k8sResourceMetricInterface {
		return newReplicaset(c, m)
	})
	registerK8sResourceObject(func(c k8sClientX, m map[string]string) k8sResourceObjectInterface {
		return newReplicaset(c, m)
	})
	registerMeasurement(&replicasetMetric{})
	registerMeasurement(&replicasetObject{})
}

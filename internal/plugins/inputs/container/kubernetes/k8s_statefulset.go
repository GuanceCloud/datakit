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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	apiappsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	statefulsetMetricMeasurement = "kube_statefulset"
	statefulsetObjectMeasurement = "kubernetes_statefulsets"
)

//nolint:gochecknoinits
func init() {
	registerResource("statefulset", true, false, newStatefulset)
	registerMeasurements(&statefulsetMetric{}, &statefulsetObject{})
}

type statefulset struct {
	client    k8sClient
	continued string
}

func newStatefulset(client k8sClient) resource {
	return &statefulset{client: client}
}

func (s *statefulset) hasNext() bool { return s.continued != "" }

func (s *statefulset) getMetadata(ctx context.Context, ns, fieldSelector string) (metadata, error) {
	opt := metav1.ListOptions{
		Limit:         queryLimit,
		Continue:      s.continued,
		FieldSelector: fieldSelector,
	}

	list, err := s.client.GetStatefulSets(ns).List(ctx, opt)
	if err != nil {
		return nil, err
	}

	s.continued = list.Continue
	return &statefulsetMetadata{list}, nil
}

type statefulsetMetadata struct {
	list *apiappsv1.StatefulSetList
}

func (m *statefulsetMetadata) transformMetric() pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		met := typed.NewPointKV(statefulsetMetricMeasurement)

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("statefulset", item.Name)
		met.SetTag("namespace", item.Namespace)

		met.SetField("replicas", item.Status.Replicas)
		met.SetField("replicas_updated", item.Status.UpdatedReplicas)
		met.SetField("replicas_ready", item.Status.ReadyReplicas)
		met.SetField("replicas_current", item.Status.CurrentReplicas)
		met.SetField("replicas_available", item.Status.AvailableReplicas)

		if item.Spec.Replicas != nil {
			met.SetField("replicas_desired", *item.Spec.Replicas)
		}

		met.SetCustomerTags(item.Labels, getGlobalCustomerKeys())
		res = append(res, met)
	}

	return res
}

func (m *statefulsetMetadata) transformObject() pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		obj := typed.NewPointKV(statefulsetObjectMeasurement)

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("uid", fmt.Sprintf("%v", item.UID))
		obj.SetTag("statefulset_name", item.Name)
		obj.SetTag("namespace", item.Namespace)

		obj.SetField("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		obj.SetField("replicas", item.Status.Replicas)
		obj.SetField("replicas_updated", item.Status.UpdatedReplicas)
		obj.SetField("replicas_ready", item.Status.ReadyReplicas)
		obj.SetField("replicas_current", item.Status.CurrentReplicas)
		obj.SetField("replicas_available", item.Status.AvailableReplicas)

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

type statefulsetMetric struct{}

//nolint:lll
func (*statefulsetMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: statefulsetMetricMeasurement,
		Desc: "The metric of the Kubernetes StatefulSet.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":              inputs.NewTagInfo("The UID of StatefulSet."),
			"statefulset":      inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"replicas":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of Pods created by the StatefulSet controller."},
			"replicas_desired":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The desired number of replicas of the given Template."},
			"replicas_ready":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods created for this StatefulSet with a Ready Condition."},
			"replicas_current":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of Pods created by the StatefulSet controller from the StatefulSet version indicated by currentRevision."},
			"replicas_updated":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of Pods created by the StatefulSet controller from the StatefulSet version indicated by updateRevision."},
			"replicas_available": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of available pods (ready for at least minReadySeconds) targeted by this StatefulSet."},
		},
	}
}

type statefulsetObject struct{}

//nolint:lll
func (*statefulsetObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: statefulsetObjectMeasurement,
		Desc: "The object of the Kubernetes StatefulSet.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":             inputs.NewTagInfo("The UID of StatefulSet."),
			"uid":              inputs.NewTagInfo("The UID of StatefulSet."),
			"statefulset_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"age":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"replicas":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of Pods created by the StatefulSet controller."},
			"replicas_desired":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The desired number of replicas of the given Template."},
			"replicas_ready":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods created for this StatefulSet with a Ready Condition."},
			"replicas_current":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of Pods created by the StatefulSet controller from the StatefulSet version indicated by currentRevision."},
			"replicas_updated":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of Pods created by the StatefulSet controller from the StatefulSet version indicated by updateRevision."},
			"replicas_available": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of available pods (ready for at least minReadySeconds) targeted by this StatefulSet."},
			"message":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}

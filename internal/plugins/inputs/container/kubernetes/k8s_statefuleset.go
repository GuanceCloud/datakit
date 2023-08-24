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
	registerMetricResource("statefulset", gatherStatefulsetMetric)
	registerObjectResource("statefulset", gatherStatefulsetObject)
	registerMeasurement(&statefulsetMetric{})
	registerMeasurement(&statefulsetObject{})
}

func gatherStatefulsetMetric(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetStatefulSets().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeStatefulsetMetric(list), nil
}

func composeStatefulsetMetric(list *apiappsv1.StatefulSetList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		met := typed.NewPointKV()

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
		res = append(res, &statefulsetMetric{met})
	}

	return res
}

func gatherStatefulsetObject(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetStatefulSets().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeStatefulsetObject(list), nil
}

func composeStatefulsetObject(list *apiappsv1.StatefulSetList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		obj := typed.NewPointKV()

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
		res = append(res, &statefulsetObject{obj})
	}

	return res
}

type statefulsetMetric struct{ typed.PointKV }

func (d *statefulsetMetric) namespace() string { return d.GetTag("namespace") }

func (d *statefulsetMetric) addExtraTags(m map[string]string) { d.SetTags(m) }

func (d *statefulsetMetric) LineProto() (*point.Point, error) {
	return point.NewPoint("kube_statefulset", d.Tags(), d.Fields(), metricOpt)
}

//nolint:lll
func (*statefulsetMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_statefulset",
		Desc: "The metric of the Kubernetes StatefulSet.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":         inputs.NewTagInfo("The UID of StatefulSet."),
			"statefulset": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":   inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
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

type statefulsetObject struct{ typed.PointKV }

func (d *statefulsetObject) namespace() string { return d.GetTag("namespace") }

func (d *statefulsetObject) addExtraTags(m map[string]string) { d.SetTags(m) }

func (d *statefulsetObject) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes_statefulsets", d.Tags(), d.Fields(), objectOpt)
}

//nolint:lll
func (*statefulsetObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_statefulsets",
		Desc: "The object of the Kubernetes StatefulSet.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":             inputs.NewTagInfo("The UID of StatefulSet."),
			"uid":              inputs.NewTagInfo("The UID of StatefulSet."),
			"statefulset_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
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

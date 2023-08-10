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
	registerMetricResource("daemonset", gatherDaemonsetMetric)
	registerObjectResource("daemonset", gatherDaemonsetObject)
	registerMeasurement(&daemonsetMetric{})
	registerMeasurement(&daemonsetObject{})
}

func gatherDaemonsetMetric(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetDaemonSets().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeDaemonsetMetric(list), nil
}

func composeDaemonsetMetric(list *apiappsv1.DaemonSetList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		met := typed.NewPointKV()

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("daemonset", item.Name)
		met.SetTag("namespace", item.Namespace)

		met.SetField("scheduled", item.Status.CurrentNumberScheduled)
		met.SetField("desired", item.Status.DesiredNumberScheduled)
		met.SetField("misscheduled", item.Status.NumberMisscheduled)
		met.SetField("ready", item.Status.NumberReady)
		met.SetField("updated", item.Status.UpdatedNumberScheduled)
		met.SetField("daemons_unavailable", item.Status.NumberUnavailable)

		res = append(res, &daemonsetMetric{met})
	}

	return res
}

func gatherDaemonsetObject(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetDaemonSets().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeDaemonsetObject(list), nil
}

func composeDaemonsetObject(list *apiappsv1.DaemonSetList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		obj := typed.NewPointKV()

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("uid", fmt.Sprintf("%v", item.UID))
		obj.SetTag("daemonset_name", item.Name)
		obj.SetTag("namespace", item.Namespace)

		obj.SetField("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		obj.SetField("scheduled", item.Status.CurrentNumberScheduled)
		obj.SetField("desired", item.Status.DesiredNumberScheduled)
		obj.SetField("misscheduled", item.Status.NumberMisscheduled)
		obj.SetField("ready", item.Status.NumberReady)
		obj.SetField("updated", item.Status.UpdatedNumberScheduled)
		obj.SetField("daemons_unavailable", item.Status.NumberUnavailable)

		if y, err := yaml.Marshal(item); err == nil {
			obj.SetField("yaml", string(y))
		}

		obj.SetFields(transLabels(item.Labels))
		obj.SetField("annotations", typed.MapToJSON(item.Annotations))
		obj.SetField("message", typed.TrimString(obj.String(), maxMessageLength))
		obj.DeleteField("annotations")
		obj.DeleteField("yaml")

		res = append(res, &daemonsetObject{obj})
	}

	return res
}

type daemonsetMetric struct{ typed.PointKV }

func (d *daemonsetMetric) namespace() string { return d.GetTag("namespace") }

func (d *daemonsetMetric) addExtraTags(m map[string]string) { d.SetTags(m) }

func (d *daemonsetMetric) LineProto() (*point.Point, error) {
	return point.NewPoint("kube_daemonset", d.Tags(), d.Fields(), metricOpt)
}

//nolint:lll
func (*daemonsetMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_daemonset",
		Desc: "The metric of the Kubernetes DaemonSet.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":       inputs.NewTagInfo("The UID of DaemonSet."),
			"daemonset": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace": inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"scheduled":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running at least one daemon pod and are supposed to run the daemon pod."},
			"desired":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that should be running the daemon pod (including nodes correctly running the daemon pod)."},
			"misscheduled":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running the daemon pod, but are not supposed to run the daemon pod."},
			"ready":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready."},
			"updated":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that are running updated daemon pod."},
			"daemons_unavailable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have none of the daemon pod running and available (ready for at least spec.minReadySeconds)."},
		},
	}
}

type daemonsetObject struct{ typed.PointKV }

func (d *daemonsetObject) namespace() string { return d.GetTag("namespace") }

func (d *daemonsetObject) addExtraTags(m map[string]string) { d.SetTags(m) }

func (d *daemonsetObject) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes_daemonset", d.Tags(), d.Fields(), objectOpt)
}

//nolint:lll
func (*daemonsetObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_daemonset",
		Desc: "The object of the Kubernetes DaemonSet.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":           inputs.NewTagInfo("The UID of DaemonSet."),
			"uid":            inputs.NewTagInfo("The UID of DaemonSet."),
			"daemonset_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":      inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"age":                 &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"scheduled":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running at least one daemon pod and are supposed to run the daemon pod."},
			"desired":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that should be running the daemon pod (including nodes correctly running the daemon pod)."},
			"misscheduled":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running the daemon pod, but are not supposed to run the daemon pod."},
			"ready":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready."},
			"updated":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that are running updated daemon pod."},
			"daemons_unavailable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have none of the daemon pod running and available (ready for at least spec.minReadySeconds)."},
			"message":             &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}

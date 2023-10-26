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
	"sigs.k8s.io/yaml"

	apiappsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	daemonsetMetricMeasurement = "kube_daemonset"
	daemonsetObjectMeasurement = "kubernetes_daemonset"
)

//nolint:gochecknoinits
func init() {
	registerResource("daemonset", true, newDaemonset)
	registerMeasurements(&daemonsetMetric{}, &daemonsetObject{})
}

type daemonset struct {
	client    k8sClient
	continued string
}

func newDaemonset(client k8sClient) resource {
	return &daemonset{client: client}
}

func (d *daemonset) hasNext() bool { return d.continued != "" }

func (d *daemonset) getMetadata(ctx context.Context, ns string) (metadata, error) {
	opt := metav1.ListOptions{
		Limit:    queryLimit,
		Continue: d.continued,
	}

	list, err := d.client.GetDaemonSets(ns).List(ctx, opt)
	if err != nil {
		return nil, err
	}

	d.continued = list.Continue
	return &daemonsetMetadata{list}, nil
}

type daemonsetMetadata struct {
	list *apiappsv1.DaemonSetList
}

func (m *daemonsetMetadata) transformMetric() pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		met := typed.NewPointKV(daemonsetMetricMeasurement)

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("daemonset", item.Name)
		met.SetTag("namespace", item.Namespace)

		met.SetField("desired", item.Status.DesiredNumberScheduled)
		met.SetField("scheduled", item.Status.CurrentNumberScheduled)
		met.SetField("misscheduled", item.Status.NumberMisscheduled)
		met.SetField("ready", item.Status.NumberReady)
		met.SetField("updated", item.Status.UpdatedNumberScheduled)
		met.SetField("daemons_available", item.Status.NumberAvailable)
		met.SetField("daemons_unavailable", item.Status.NumberUnavailable)

		met.SetCustomerTags(item.Labels, getGlobalCustomerKeys())
		res = append(res, met)
	}

	return res
}

func (m *daemonsetMetadata) transformObject() pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		obj := typed.NewPointKV(daemonsetObjectMeasurement)

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("uid", fmt.Sprintf("%v", item.UID))
		obj.SetTag("daemonset_name", item.Name)
		obj.SetTag("namespace", item.Namespace)

		obj.SetField("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		obj.SetField("desired", item.Status.DesiredNumberScheduled)
		obj.SetField("scheduled", item.Status.CurrentNumberScheduled)
		obj.SetField("misscheduled", item.Status.NumberMisscheduled)
		obj.SetField("ready", item.Status.NumberReady)
		obj.SetField("updated", item.Status.UpdatedNumberScheduled)
		obj.SetField("daemons_available", item.Status.NumberAvailable)
		obj.SetField("daemons_unavailable", item.Status.NumberUnavailable)

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

type daemonsetMetric struct{}

//nolint:lll
func (*daemonsetMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: daemonsetMetricMeasurement,
		Desc: "The metric of the Kubernetes DaemonSet.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":       inputs.NewTagInfo("The UID of DaemonSet."),
			"daemonset": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace": inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"desired":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that should be running the daemon pod (including nodes correctly running the daemon pod)."},
			"scheduled":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running at least one daemon pod and are supposed to run the daemon pod."},
			"misscheduled":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running the daemon pod, but are not supposed to run the daemon pod."},
			"ready":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready."},
			"updated":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that are running updated daemon pod."},
			"daemons_available":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and available (ready for at least spec.minReadySeconds)."},
			"daemons_unavailable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have none of the daemon pod running and available (ready for at least spec.minReadySeconds)."},
		},
	}
}

type daemonsetObject struct{ typed.PointKV }

//nolint:lll
func (*daemonsetObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: daemonsetObjectMeasurement,
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
			"desired":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that should be running the daemon pod (including nodes correctly running the daemon pod)."},
			"scheduled":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running at least one daemon pod and are supposed to run the daemon pod."},
			"misscheduled":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running the daemon pod, but are not supposed to run the daemon pod."},
			"ready":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready."},
			"updated":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that are running updated daemon pod."},
			"daemons_available":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and available (ready for at least spec.minReadySeconds)."},
			"daemons_unavailable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have none of the daemon pod running and available (ready for at least spec.minReadySeconds)."},
			"message":             &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}

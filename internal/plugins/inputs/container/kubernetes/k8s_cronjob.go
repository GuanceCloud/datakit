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

	apibatchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	cronjobMetricMeasurement = "kube_cronjob"
	cronjobObjectMeasurement = "kubernetes_cron_jobs"
)

//nolint:gochecknoinits
func init() {
	registerResource("cronjob", true, newCronjob)
	registerMeasurements(&cronjobMetric{}, &cronjobObject{})
}

type cronjob struct {
	client    k8sClient
	continued string
}

func newCronjob(client k8sClient) resource {
	return &cronjob{client: client}
}

func (c *cronjob) hasNext() bool { return c.continued != "" }

func (c *cronjob) getMetadata(ctx context.Context, ns string) (metadata, error) {
	opt := metav1.ListOptions{
		Limit:    queryLimit,
		Continue: c.continued,
	}

	list, err := c.client.GetCronJobs(ns).List(ctx, opt)
	if err != nil {
		return nil, err
	}

	c.continued = list.Continue
	return &cronjobMetadata{list}, nil
}

type cronjobMetadata struct {
	list *apibatchv1.CronJobList
}

func (m *cronjobMetadata) transformMetric() pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		met := typed.NewPointKV(cronjobMetricMeasurement)

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("cronjob", item.Name)
		met.SetTag("namespace", item.Namespace)

		if item.Spec.Suspend != nil {
			met.SetField("spec_suspend", *item.Spec.Suspend)
		}

		met.SetCustomerTags(item.Labels, getGlobalCustomerKeys())
		res = append(res, met)
	}

	return res
}

func (m *cronjobMetadata) transformObject() pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		obj := typed.NewPointKV(cronjobObjectMeasurement)

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("uid", fmt.Sprintf("%v", item.UID))
		obj.SetTag("cron_job_name", item.Name)
		obj.SetTag("namespace", item.Namespace)

		obj.SetField("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		obj.SetField("schedule", item.Spec.Schedule)
		obj.SetField("active_jobs", len(item.Status.Active))
		obj.SetField("suspend", false)

		if item.Spec.Suspend != nil {
			obj.SetField("suspend", *item.Spec.Suspend)
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

type cronjobMetric struct{}

//nolint:lll
func (*cronjobMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: cronjobMetricMeasurement,
		Desc: "The metric of the Kubernetes CronJob.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":       inputs.NewTagInfo("The UID of CronJob."),
			"cronjob":   inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace": inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"spec_suspend": &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "This flag tells the controller to suspend subsequent executions."},
		},
	}
}

type cronjobObject struct{}

//nolint:lll
func (*cronjobObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: cronjobObjectMeasurement,
		Desc: "The object of the Kubernetes CronJob.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":          inputs.NewTagInfo("The UID of CronJob."),
			"uid":           inputs.NewTagInfo("The UID of CronJob."),
			"cron_job_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":     inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"age":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"schedule":    &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: `The schedule in Cron format, see [doc](https://en.wikipedia.org/wiki/Cron){:target="_blank"}`},
			"active_jobs": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pointers to currently running jobs."},
			"suspend":     &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "This flag tells the controller to suspend subsequent executions, it does not apply to already started executions."},
			"message":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}

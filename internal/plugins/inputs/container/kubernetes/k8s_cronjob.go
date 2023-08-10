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

	apibatchv1 "k8s.io/api/batch/v1"
)

//nolint:gochecknoinits
func init() {
	registerMetricResource("cronjob", gatherCronjobMetric)
	registerObjectResource("cronjob", gatherCronjobObject)
	registerMeasurement(&cronjobMetric{})
	registerMeasurement(&cronjobObject{})
}

func gatherCronjobMetric(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetCronJobs().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeCronjobMetric(list), nil
}

func composeCronjobMetric(list *apibatchv1.CronJobList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		met := typed.NewPointKV()

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("cronjob", item.Name)
		met.SetTag("namespace", item.Namespace)

		if item.Spec.Suspend != nil {
			met.SetField("spec_suspend", *item.Spec.Suspend)
		}

		res = append(res, &cronjobMetric{met})
	}

	return res
}

func gatherCronjobObject(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetCronJobs().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeCronjobObject(list), nil
}

func composeCronjobObject(list *apibatchv1.CronJobList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		obj := typed.NewPointKV()

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

		res = append(res, &cronjobObject{obj})
	}

	return res
}

type cronjobMetric struct{ typed.PointKV }

func (c *cronjobMetric) namespace() string { return c.GetTag("namespace") }

func (c *cronjobMetric) addExtraTags(m map[string]string) { c.SetTags(m) }

func (c *cronjobMetric) LineProto() (*point.Point, error) {
	return point.NewPoint("kube_cronjob", c.Tags(), c.Fields(), metricOpt)
}

//nolint:lll
func (*cronjobMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_cronjob",
		Desc: "The metric of the Kubernetes CronJob.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":       inputs.NewTagInfo("The UID of cronjob."),
			"cronjob":   inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace": inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"spec_suspend": &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "This flag tells the controller to suspend subsequent executions."},
		},
	}
}

type cronjobObject struct{ typed.PointKV }

func (c *cronjobObject) namespace() string { return c.GetTag("namespace") }

func (c *cronjobObject) addExtraTags(m map[string]string) { c.SetTags(m) }

func (c *cronjobObject) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes_cron_jobs", c.Tags(), c.Fields(), objectOpt)
}

//nolint:lll
func (*cronjobObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_cron_jobs",
		Desc: "The object of the Kubernetes CronJob.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":          inputs.NewTagInfo("The UID of cronjob."),
			"uid":           inputs.NewTagInfo("The UID of cronjob."),
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

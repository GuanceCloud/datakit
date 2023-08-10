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
	registerMetricResource("job", gatherJobMetric)
	registerObjectResource("job", gatherJobObject)
	registerMeasurement(&jobMetric{})
	registerMeasurement(&jobObject{})
}

func gatherJobMetric(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetJobs().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeJobMetric(list), nil
}

func composeJobMetric(list *apibatchv1.JobList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		met := typed.NewPointKV()

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("job", item.Name)
		met.SetTag("namespace", item.Namespace)

		met.SetField("active", item.Status.Active)
		met.SetField("failed", item.Status.Failed)
		met.SetField("succeeded", item.Status.Succeeded)

		met.SetField("completion_succeeded", 0)
		met.SetField("completion_failed", 0)

		var succeeded, failed int
		for _, condition := range item.Status.Conditions {
			switch condition.Type {
			case apibatchv1.JobFailed:
				failed++
			case apibatchv1.JobComplete:
				succeeded++
			case apibatchv1.JobSuspended, apibatchv1.AlphaNoCompatGuaranteeJobFailureTarget:
				// nil
			}
		}

		met.SetField("completion_succeeded", succeeded)
		met.SetField("completion_failed", failed)

		res = append(res, &jobMetric{met})
	}

	return res
}

func gatherJobObject(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetJobs().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeJobObject(list), nil
}

func composeJobObject(list *apibatchv1.JobList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		obj := typed.NewPointKV()

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("uid", fmt.Sprintf("%v", item.UID))
		obj.SetTag("job_name", item.Name)
		obj.SetTag("namespace", item.Namespace)

		obj.SetField("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		obj.SetField("active", item.Status.Active)
		obj.SetField("succeeded", item.Status.Succeeded)
		obj.SetField("failed", item.Status.Failed)
		obj.SetField("parallelism", 0)
		obj.SetField("completions", 0)
		obj.SetField("active_deadline", 0)
		obj.SetField("backoff_limit", 0)

		// 因为原数据类型（例如 item.Spec.Parallelism）就是 int32，所以此处也用 int32
		if item.Spec.Parallelism != nil {
			obj.SetField("parallelism", *item.Spec.Parallelism)
		}
		if item.Spec.Completions != nil {
			obj.SetField("completions", *item.Spec.Completions)
		}
		if item.Spec.ActiveDeadlineSeconds != nil {
			obj.SetField("active_deadline", *item.Spec.ActiveDeadlineSeconds)
		}
		if item.Spec.BackoffLimit != nil {
			obj.SetField("backoff_limit", *item.Spec.BackoffLimit)
		}

		if y, err := yaml.Marshal(item); err == nil {
			obj.SetField("yaml", string(y))
		}

		obj.SetFields(transLabels(item.Labels))
		obj.SetField("annotations", typed.MapToJSON(item.Annotations))
		obj.SetField("message", typed.TrimString(obj.String(), maxMessageLength))
		obj.DeleteField("annotations")
		obj.DeleteField("yaml")

		res = append(res, &jobObject{obj})
	}

	return res
}

type jobMetric struct{ typed.PointKV }

func (j *jobMetric) namespace() string { return j.GetTag("namespace") }

func (j *jobMetric) addExtraTags(m map[string]string) { j.SetTags(m) }

func (j *jobMetric) LineProto() (*point.Point, error) {
	return point.NewPoint("kube_job", j.Tags(), j.Fields(), metricOpt)
}

//nolint:lll
func (*jobMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_job",
		Desc: "The metric of the Kubernetes Job.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":       inputs.NewTagInfo("The UID of job."),
			"job":       inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace": inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"active":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of actively running pods."},
			"failed":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods which reached phase Failed."},
			"succeeded":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods which reached phase Succeeded."},
			"completion_succeeded": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The job has completed its execution."},
			"completion_failed":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The job has failed its execution."},
		},
	}
}

type jobObject struct{ typed.PointKV }

func (j *jobObject) namespace() string { return j.GetTag("namespace") }

func (j *jobObject) addExtraTags(m map[string]string) { j.SetTags(m) }

func (j *jobObject) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes_jobs", j.Tags(), j.Fields(), objectOpt)
}

//nolint:lll
func (*jobObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_jobs",
		Desc: "The object of the Kubernetes Job.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":      inputs.NewTagInfo("The UID of job."),
			"uid":       inputs.NewTagInfo("The UID of job."),
			"job_name":  inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace": inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"age":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"active":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of actively running pods."},
			"succeeded":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods which reached phase Succeeded."},
			"failed":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods which reached phase Failed."},
			"completions":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Specifies the desired number of successfully finished pods the job should be run with."},
			"parallelism":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Specifies the maximum desired number of pods the job should run at any given time."},
			"backoff_limit":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Specifies the number of retries before marking this job failed."},
			"active_deadline": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Specifies the duration in seconds relative to the startTime that the job may be active before the system tries to terminate it"},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}

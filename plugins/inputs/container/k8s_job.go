package container

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/batch/v1"
)

const k8sJobName = "kubernetes_jobs"

func gatherJob(client k8sClientX, extraTags map[string]string) (k8sResourceStats, error) {
	list, err := client.getJobs().List(context.Background(), metaV1ListOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs resource: %w", err)
	}

	if len(list.Items) == 0 {
		return nil, nil
	}
	return exportJob(list.Items, extraTags), nil
}

func exportJob(items []v1.Job, extraTags tagsType) k8sResourceStats {
	res := newK8sResourceStats()

	for _, item := range items {
		obj := newJob()
		obj.tags["name"] = fmt.Sprintf("%v", item.UID)
		obj.tags["job_name"] = item.Name

		obj.tags.addValueIfNotEmpty("cluster_name", item.ClusterName)
		obj.tags.addValueIfNotEmpty("namespace", defaultNamespace(item.Namespace))
		obj.tags.append(extraTags)

		obj.fields["age"] = int64(time.Since(item.CreationTimestamp.Time).Seconds())
		obj.fields["active"] = item.Status.Active
		obj.fields["succeeded"] = item.Status.Succeeded
		obj.fields["failed"] = item.Status.Failed

		// 因为原数据类型（例如 item.Spec.Parallelism）就是 int32，所以此处也用 int32
		if item.Spec.Parallelism != nil {
			obj.fields["parallelism"] = *item.Spec.Parallelism
		} else {
			obj.fields["parallelism"] = int32(0)
		}

		if item.Spec.Completions != nil {
			obj.fields["completions"] = *item.Spec.Completions
		} else {
			obj.fields["completions"] = int32(0)
		}

		if item.Spec.ActiveDeadlineSeconds != nil {
			obj.fields["active_deadline"] = *item.Spec.ActiveDeadlineSeconds
		} else {
			obj.fields["active_deadline"] = int64(0)
		}

		if item.Spec.BackoffLimit != nil {
			obj.fields["backoff_limit"] = *item.Spec.BackoffLimit
		} else {
			obj.fields["backoff_limit"] = int32(0)
		}

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)

		obj.time = time.Now()
		res.set(defaultNamespace(item.Namespace), obj)
	}
	return res
}

type job struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func newJob() *job {
	return &job{
		tags:   make(tagsType),
		fields: make(fieldsType),
	}
}

func (j *job) LineProto() (*io.Point, error) {
	return io.NewPoint(k8sJobName, j.tags, j.fields, &io.PointOption{Time: j.time, Category: datakit.Object})
}

//nolint:lll
func (*job) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: k8sJobName,
		Desc: "Kubernetes job 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("UID"),
			"job_name":     inputs.NewTagInfo("Name must be unique within a namespace."),
			"cluster_name": inputs.NewTagInfo("The name of the cluster which the object belongs to."),
			"namespace":    inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"age":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"active":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of actively running pods."},
			"succeeded":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods which reached phase Succeeded."},
			"failed":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods which reached phase Failed."},
			"completions":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Specifies the desired number of successfully finished pods the job should be run with."},
			"parallelism":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Specifies the maximum desired number of pods the job should run at any given time."},
			"backoff_limit":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Specifies the number of retries before marking this job failed."},
			"active_deadline": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Specifies the duration in seconds relative to the startTime that the job may be active before the system tries to terminate it"},
			"annotations":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "kubernetes annotations"},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},

			// TODO:
			// "pod_statuses":           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "duration":               &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&job{})
}

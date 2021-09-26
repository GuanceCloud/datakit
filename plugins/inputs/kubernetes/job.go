package kubernetes

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	batchv1 "k8s.io/api/batch/v1"
)

const kubernetesJobName = "kubernetes_jobs"

type job struct {
	client interface {
		getJobs() (*batchv1.JobList, error)
	}
	tags map[string]string
}

func (j *job) Gather() {
	start := time.Now()
	var pts []*io.Point

	list, err := j.client.getJobs()
	if err != nil {
		l.Errorf("failed of get jobs resource: %s", err)
		return
	}

	for _, obj := range list.Items {
		tags := map[string]string{
			"name":         fmt.Sprintf("%v", obj.UID),
			"job_name":     obj.Name,
			"cluster_name": obj.ClusterName,
			"namespace":    obj.Namespace,
		}
		for k, v := range j.tags {
			tags[k] = v
		}

		fields := map[string]interface{}{
			"age":       int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"active":    obj.Status.Active,
			"succeeded": obj.Status.Succeeded,
			"failed":    obj.Status.Failed,
		}

		if obj.Spec.Parallelism != nil {
			fields["parallelism"] = *obj.Spec.Parallelism
		} else {
			fields["parallelism"] = defaultInteger32Value
		}

		if obj.Spec.Completions != nil {
			fields["completions"] = *obj.Spec.Completions
		} else {
			fields["completions"] = defaultInteger32Value
		}

		if obj.Spec.ActiveDeadlineSeconds != nil {
			fields["active_deadline"] = *obj.Spec.ActiveDeadlineSeconds
		} else {
			fields["active_deadline"] = defaultInteger64Value
		}

		if obj.Spec.BackoffLimit != nil {
			fields["backoff_limit"] = *obj.Spec.BackoffLimit
		} else {
			fields["backoff_limit"] = defaultInteger32Value
		}

		addMapToFields("annotations", obj.Annotations, fields)
		addLabelToFields(obj.Labels, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesJobName, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
		} else {
			pts = append(pts, pt)
		}
	}

	if err := io.Feed(inputName, datakit.Object, pts, &io.Option{CollectCost: time.Since(start)}); err != nil {
		l.Error(err)
	}
}

func (*job) Resource() { /*empty interface*/ }

func (*job) LineProto() (*io.Point, error) { return nil, nil }

func (*job) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesJobName,
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
			//"duration":               &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}

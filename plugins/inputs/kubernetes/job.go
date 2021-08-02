package kubernetes

import (
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const kubernetesJobName = "kubernetes_jobs"

type job struct {
	client interface {
		getJobs() (*batchv1.JobList, error)
	}
	tags map[string]string
}

func (j *job) Gather() {
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
			"pod_name":     obj.ClusterName,
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
		}
		if obj.Spec.Completions != nil {
			fields["completions"] = *obj.Spec.Completions
		}
		if obj.Spec.ActiveDeadlineSeconds != nil {
			fields["active_deadline"] = *obj.Spec.ActiveDeadlineSeconds
		}
		if obj.Spec.BackoffLimit != nil {
			fields["backoff_limit"] = *obj.Spec.BackoffLimit
		}

		addJSONStringToMap("kubernetes_annotations", obj.Annotations, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesJobName, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
		} else {
			if err := io.Feed(inputName, datakit.Object, []*io.Point{pt}, nil); err != nil {
				l.Error(err)
			}
		}
	}
}

func (*job) Resource() { /*empty interface*/ }

func (*job) LineProto() (*io.Point, error) { return nil, nil }

func (*job) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesJobName,
		Desc: fmt.Sprintf("%s 对象数据", kubernetesJobName),
		Type: datakit.Object,
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("job UID"),
			"job_name":     inputs.NewTagInfo("job 名称"),
			"cluster_name": inputs.NewTagInfo("所在 cluster"),
			"namespace":    inputs.NewTagInfo("所在命名空间"),
		},
		Fields: map[string]interface{}{
			"age":                    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "存活时长，单位为秒"},
			"active":                 &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "活跃数"},
			"succeeded":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "成功数"},
			"failed":                 &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "失败数"},
			"completions":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "确定完成计数"},
			"parallelism":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "并行数量"},
			"backoff_limit":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "重试次数"},
			"active_deadline":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "活跃期限，单位为秒"},
			"kubernetes_annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "k8s annotations"},
			"message":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "详情数据"},
			// TODO:
			// "pod_statuses":           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			//"duration":               &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}

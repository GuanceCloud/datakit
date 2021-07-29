package kubernetes

import (
	"fmt"
	"time"

	batchbetav1 "k8s.io/api/batch/v1beta1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const kubernetesCronJobName = "kubernetes_cron_jobs"

type cronJob struct {
	client interface {
		getCronJobs() (*batchbetav1.CronJobList, error)
	}
}

func (c cronJob) Gather() {
	list, err := c.client.getCronJobs()
	if err != nil {
		l.Errorf("failed of get cronjobs resource: %s", err)
		return
	}

	for _, obj := range list.Items {
		tags := map[string]string{
			"name":          fmt.Sprintf("%v", obj.UID),
			"cron_job_name": obj.Name,
			"cluster_name":  obj.ClusterName,
			"namespace":     obj.Namespace,
		}
		fields := map[string]interface{}{
			"age":         int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"schedule":    obj.Spec.Schedule,
			"active_jobs": len(obj.Status.Active),
		}

		if obj.Spec.Suspend != nil {
			fields["suspend"] = *obj.Spec.Suspend
		}

		addJSONStringToMap("kubernetes_labels", obj.Labels, fields)
		addJSONStringToMap("kubernetes_annotations", obj.Annotations, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesCronJobName, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
		} else {
			if err := io.Feed(inputName, datakit.Object, []*io.Point{pt}, nil); err != nil {
				l.Error(err)
			}
		}
	}
}

func (*cronJob) LineProto() (*io.Point, error) {
	return nil, nil
}

func (*cronJob) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesCronJobName,
		Desc: kubernetesCronJobName,
		Tags: map[string]interface{}{
			"name":          inputs.NewTagInfo(""),
			"cron_job_name": inputs.NewTagInfo(""),
			"cluster_name":  inputs.NewTagInfo(""),
			"namespace":     inputs.NewTagInfo(""),
		},
		Fields: map[string]interface{}{
			"age":                    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: ""},
			"schedule":               &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"suspend":                &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: ""},
			"active_jobs":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubernetes_labels":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubernetes_annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"message":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}

package container

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1beta1 "k8s.io/api/batch/v1beta1"
)

const k8sCronJobName = "kubernetes_cron_jobs"

func gatherCronJob(client k8sClientX, extraTags map[string]string) (k8sResourceStats, error) {
	list, err := client.getCronJobs().List(context.Background(), metaV1ListOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get cronjobs resource: %w", err)
	}

	if len(list.Items) == 0 {
		return nil, nil
	}
	return exportCronJob(list.Items, extraTags), nil
}

func exportCronJob(items []v1beta1.CronJob, extraTags tagsType) k8sResourceStats {
	res := newK8sResourceStats()

	for _, item := range items {
		obj := newCronJob()
		obj.tags["name"] = fmt.Sprintf("%v", item.UID)
		obj.tags["cron_job_name"] = item.Name

		obj.tags.addValueIfNotEmpty("cluster_name", item.ClusterName)
		obj.tags.addValueIfNotEmpty("namespace", defaultNamespace(item.Namespace))
		obj.tags.append(extraTags)

		obj.fields["age"] = int64(time.Since(item.CreationTimestamp.Time).Seconds())
		obj.fields["schedule"] = item.Spec.Schedule
		obj.fields["active_jobs"] = len(item.Status.Active)

		if item.Spec.Suspend != nil {
			obj.fields["suspend"] = *item.Spec.Suspend
		} else {
			obj.fields["suspend"] = false
		}

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)

		obj.time = time.Now()
		res.set(defaultNamespace(item.Namespace), obj)
	}
	return res
}

type cronJob struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func newCronJob() *cronJob {
	return &cronJob{
		tags:   make(tagsType),
		fields: make(fieldsType),
	}
}

func (c *cronJob) LineProto() (*io.Point, error) {
	return io.NewPoint(k8sCronJobName, c.tags, c.fields, &io.PointOption{Time: c.time, Category: datakit.Object})
}

//nolint:lll
func (*cronJob) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: k8sCronJobName,
		Desc: "Kubernetes cron job 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":          inputs.NewTagInfo("UID"),
			"cron_job_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"cluster_name":  inputs.NewTagInfo("The name of the cluster which the object belongs to."),
			"namespace":     inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"age":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"schedule":    &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron"},
			"active_jobs": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pointers to currently running jobs."},
			"suspend":     &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "This flag tells the controller to suspend subsequent executions, it does not apply to already started executions."},
			"annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "kubernetes annotations"},
			"message":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&cronJob{})
}

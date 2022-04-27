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

var (
	_ k8sResourceMetricInterface = (*cronjob)(nil)
	_ k8sResourceObjectInterface = (*cronjob)(nil)
)

type cronjob struct {
	client    k8sClientX
	extraTags map[string]string
	items     []v1beta1.CronJob
}

func newCronjob(client k8sClientX, extraTags map[string]string) *cronjob {
	return &cronjob{
		client:    client,
		extraTags: extraTags,
	}
}

func (c *cronjob) name() string {
	return "cronjob"
}

func (c *cronjob) pullItems() error {
	if len(c.items) != 0 {
		return nil
	}

	list, err := c.client.getCronJobs().List(context.Background(), metaV1ListOption)
	if err != nil {
		return fmt.Errorf("failed to get cronjobs resource: %w", err)
	}

	c.items = list.Items
	return nil
}

func (c *cronjob) metric() (inputsMeas, error) {
	if err := c.pullItems(); err != nil {
		return nil, err
	}
	var res inputsMeas

	for _, item := range c.items {
		met := &cronjobMetric{
			tags: map[string]string{
				"cronjob":   item.Name,
				"namespace": defaultNamespace(item.Namespace),
			},
			fields: map[string]interface{}{
				"spec_suspend": *item.Spec.Suspend,
			},
			time: time.Now(),
		}
		// t := item.Status.LastScheduleTime
		// met.fields["duration_since_last_schedule"] = int64(time.Since(t).Seconds())

		met.tags.append(c.extraTags)
		res = append(res, met)
	}

	count, _ := c.count()
	for ns, ct := range count {
		met := &cronjobMetric{
			tags:   map[string]string{"namespace": ns},
			fields: map[string]interface{}{"count": ct},
			time:   time.Now(),
		}
		met.tags.append(c.extraTags)
		res = append(res, met)
	}

	return res, nil
}

func (c *cronjob) object() (inputsMeas, error) {
	if err := c.pullItems(); err != nil {
		return nil, err
	}
	var res inputsMeas

	for _, item := range c.items {
		obj := &cronjobObject{
			tags: map[string]string{
				"name":          fmt.Sprintf("%v", item.UID),
				"cron_job_name": item.Name,
				"cluster_name":  defaultClusterName(item.ClusterName),
				"namespace":     defaultNamespace(item.Namespace),
			},
			fields: map[string]interface{}{
				"age": int64(time.Since(item.CreationTimestamp.Time).Seconds()),

				"schedule":    item.Spec.Schedule,
				"active_jobs": len(item.Status.Active),
				"suspend":     false,
			},
			time: time.Now(),
		}
		obj.tags.append(c.extraTags)

		if item.Spec.Suspend != nil {
			obj.fields["suspend"] = *item.Spec.Suspend
		}

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)
		obj.fields.delete("annotations")

		res = append(res, obj)
	}

	return res, nil
}

func (c *cronjob) count() (map[string]int, error) {
	if err := c.pullItems(); err != nil {
		return nil, err
	}

	m := make(map[string]int)
	for _, item := range c.items {
		m[defaultNamespace(item.Namespace)]++
	}

	if len(m) == 0 {
		m["default"] = 0
	}

	return m, nil
}

type cronjobMetric struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func (c *cronjobMetric) LineProto() (*io.Point, error) {
	return io.NewPoint("kuber_cronjob", c.tags, c.fields, &io.PointOption{Time: c.time, Category: datakit.Metric})
}

//nolint:lll
func (*cronjobMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_cronjob",
		Desc: "Kubernetes cron job 指标数据",
		Type: "metric",
		Tags: map[string]interface{}{
			"cronjob":   inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace": inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"count":                        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of cronjobs"},
			"spec_suspend":                 &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "This flag tells the controller to suspend subsequent executions."},
			"duration_since_last_schedule": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "The duration since the last time the cronjob was scheduled."},
		},
	}
}

type cronjobObject struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func (c *cronjobObject) LineProto() (*io.Point, error) {
	return io.NewPoint("kubernetes_cron_jobs", c.tags, c.fields, &io.PointOption{Time: c.time, Category: datakit.Object})
}

//nolint:lll
func (*cronjobObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_cron_jobs",
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
			"message":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerK8sResourceMetric(func(c k8sClientX, m map[string]string) k8sResourceMetricInterface { return newCronjob(c, m) })
	registerK8sResourceObject(func(c k8sClientX, m map[string]string) k8sResourceObjectInterface { return newCronjob(c, m) })
	registerMeasurement(&cronjobMetric{})
	registerMeasurement(&cronjobObject{})
}

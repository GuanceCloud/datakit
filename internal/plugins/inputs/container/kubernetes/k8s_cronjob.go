// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apibatchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/informers"
)

const (
	cronjobMetricMeasurement = "kube_cronjob"
	cronjobObjectClass       = "kubernetes_cron_jobs"
)

//nolint:gochecknoinits
func init() {
	registerResource("cronjob", false, newCronjob)
}

type cronjob struct {
	client  k8sClient
	cfg     *Config
	counter map[string]int
}

func newCronjob(client k8sClient, cfg *Config) resource {
	return &cronjob{client: client, cfg: cfg, counter: make(map[string]int)}
}

func (c *cronjob) gatherMetric(ctx context.Context, timestamp int64) {
	var continued string
	for {
		list, err := c.client.GetCronJobs(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := c.buildMetricPoints(list, timestamp)
		feedMetric("k8s-cronjob-metric", c.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
	processCounter(c.cfg, "cronjob", c.counter, timestamp)
}

func (c *cronjob) gatherObject(ctx context.Context) {
	var continued string
	for {
		list, err := c.client.GetCronJobs(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := c.buildObjectPoints(list)
		feedObject("k8s-cronjob-object", c.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
}

func (*cronjob) addChangeInformer(_ informers.SharedInformerFactory) { /* nil */ }

func (c *cronjob) buildMetricPoints(list *apibatchv1.CronJobList, timestamp int64) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTimestamp(timestamp))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("cronjob", item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		if item.Spec.Suspend != nil {
			kvs = kvs.Add("spec_suspend", *item.Spec.Suspend)
		}

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, c.cfg.LabelAsTagsForMetric.All, c.cfg.LabelAsTagsForMetric.Keys)...)
		kvs = append(kvs, point.NewTags(c.cfg.ExtraTags)...)
		pt := point.NewPoint(cronjobMetricMeasurement, kvs, opts...)
		pts = append(pts, pt)

		c.counter[item.Namespace]++
	}

	return pts
}

func (c *cronjob) buildObjectPoints(list *apibatchv1.CronJobList) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultObjectOptions(), point.WithTime(ntp.Now()))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("name", string(item.UID))
		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("cron_job_name", item.Name)
		kvs = kvs.AddTag("workload_name", item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		kvs = kvs.Add("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		kvs = kvs.Add("schedule", item.Spec.Schedule)
		kvs = kvs.Add("active_jobs", len(item.Status.Active))

		if item.Spec.Suspend != nil {
			kvs = kvs.Add("suspend", *item.Spec.Suspend)
		}

		if y, err := yaml.Marshal(item); err == nil {
			kvs = kvs.Add("yaml", string(y))
		}
		kvs = kvs.Add("annotations", pointutil.MapToJSON(item.Annotations))
		kvs = append(kvs, pointutil.ConvertDFLabels(item.Labels)...)

		msg := pointutil.PointKVsToJSON(kvs)
		kvs = kvs.Add("message", pointutil.TrimString(msg, maxMessageLength))

		kvs = kvs.Del("annotations")
		kvs = kvs.Del("yaml")

		kvs = append(kvs, pointutil.ExtractSourceCodeFromAnnotations(item.Annotations)...) // add source_code
		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, c.cfg.LabelAsTagsForNonMetric.All, c.cfg.LabelAsTagsForNonMetric.Keys)...)
		kvs = append(kvs, point.NewTags(c.cfg.ExtraTags)...)
		pt := point.NewPoint(cronjobObjectClass, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

type CronjobMetric struct{}

//nolint:lll
func (*CronjobMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: cronjobMetricMeasurement,
		Desc: "The metric of the Kubernetes CronJob.",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"uid":              inputs.NewTagInfo("The UID of CronJob."),
			"cronjob":          inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"spec_suspend": &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "This flag tells the controller to suspend subsequent executions."},
		},
	}
}

type CronjobObject struct{}

//nolint:lll
func (*CronjobObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: cronjobObjectClass,
		Desc: "The object of the Kubernetes CronJob.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			"name":             inputs.NewTagInfo("The UID of CronJob."),
			"uid":              inputs.NewTagInfo("The UID of CronJob."),
			"workload_name":    inputs.NewTagInfo("The name of the workload resource."),
			"cron_job_name":    inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
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

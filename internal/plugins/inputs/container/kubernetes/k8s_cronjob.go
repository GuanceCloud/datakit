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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apibatchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const (
	cronjobMetricMeasurement = "kube_cronjob"
	cronjobObjectMeasurement = "kubernetes_cron_jobs"
	cronjobChangeSource      = "kubernetes_cron_jobs"
	cronjobChangeSourceType  = "CronJob"
)

//nolint:gochecknoinits
func init() {
	registerResource("cronjob", false, newCronjob)
	registerMeasurements(&cronjobMetric{}, &cronjobObject{})
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

	counterPts := buildPointsFromCounter("cronjob", c.counter, timestamp)
	feedMetric("k8s-counter", c.cfg.Feeder, counterPts, true)
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

func (c *cronjob) addChangeInformer(informerFactory informers.SharedInformerFactory) {
	informer := informerFactory.Batch().V1().CronJobs()
	if informer == nil {
		klog.Warn("cannot get cronjob informer")
		return
	}

	updateFunc := func(oldObj, newObj interface{}) {
		objectChangeCountVec.WithLabelValues(cronjobChangeSourceType, "update").Inc()

		oldCronjobObj, ok := oldObj.(*apibatchv1.CronJob)
		if !ok {
			klog.Warnf("converting to CronJob object failed, %v", oldObj)
			return
		}

		newCronjobObj, ok := newObj.(*apibatchv1.CronJob)
		if !ok {
			klog.Warnf("converting to CronJob object failed, %v", newObj)
			return
		}

		difftext, err := diffObject(oldCronjobObj.Spec, newCronjobObj.Spec)
		if err != nil {
			klog.Warnf("marshal failed, err: %s", err)
			return
		}

		if difftext != "" {
			objectChangeCountVec.WithLabelValues(cronjobChangeSourceType, "spec-changed").Inc()
			processChange(c.cfg.Feeder, cronjobChangeSource, cronjobChangeSourceType, difftext, newCronjobObj)
		}
	}

	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(_ interface{}) { /* skip */ },
		DeleteFunc: func(_ interface{}) { /* skip */ },
		UpdateFunc: func(oldObj, newObj interface{}) {
			updateFunc(oldObj, newObj)
		},
	})
}

func (c *cronjob) buildMetricPoints(list *apibatchv1.CronJobList, timestamp int64) []*point.Point {
	var pts []*point.Point
	opts := point.DefaultMetricOptions()

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("cronjob", item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		if item.Spec.Suspend != nil {
			kvs = kvs.AddV2("spec_suspend", *item.Spec.Suspend, false)
		}

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, c.cfg.LabelAsTagsForMetric.All, c.cfg.LabelAsTagsForMetric.Keys)...)
		kvs = append(kvs, point.NewTags(c.cfg.ExtraTags)...)
		pt := point.NewPointV2(cronjobMetricMeasurement, kvs, append(opts, point.WithTimestamp(timestamp))...)
		pts = append(pts, pt)

		c.counter[item.Namespace]++
	}

	return pts
}

func (c *cronjob) buildObjectPoints(list *apibatchv1.CronJobList) []*point.Point {
	var pts []*point.Point
	opts := point.DefaultObjectOptions()

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("name", string(item.UID))
		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("cron_job_name", item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		kvs = kvs.AddV2("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3, false)
		kvs = kvs.AddV2("schedule", item.Spec.Schedule, false)
		kvs = kvs.AddV2("active_jobs", len(item.Status.Active), false)

		if item.Spec.Suspend != nil {
			kvs = kvs.AddV2("suspend", *item.Spec.Suspend, false)
		}

		if y, err := yaml.Marshal(item); err == nil {
			kvs = kvs.AddV2("yaml", string(y), false)
		}
		kvs = kvs.AddV2("annotations", pointutil.MapToJSON(item.Annotations), false)
		kvs = append(kvs, pointutil.ConvertDFLabels(item.Labels)...)

		msg := pointutil.PointKVsToJSON(kvs)
		kvs = kvs.AddV2("message", pointutil.TrimString(msg, maxMessageLength), false)

		kvs = kvs.Del("annotations")
		kvs = kvs.Del("yaml")

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, c.cfg.LabelAsTagsForNonMetric.All, c.cfg.LabelAsTagsForNonMetric.Keys)...)
		kvs = append(kvs, point.NewTags(c.cfg.ExtraTags)...)
		pt := point.NewPointV2(cronjobObjectMeasurement, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

type cronjobMetric struct{}

//nolint:lll
func (*cronjobMetric) Info() *inputs.MeasurementInfo {
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

type cronjobObject struct{}

//nolint:lll
func (*cronjobObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: cronjobObjectMeasurement,
		Desc: "The object of the Kubernetes CronJob.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			"name":             inputs.NewTagInfo("The UID of CronJob."),
			"uid":              inputs.NewTagInfo("The UID of CronJob."),
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

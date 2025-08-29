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
	jobMetricMeasurement = "kube_job"
	jobObjectClass       = "kubernetes_jobs"
)

//nolint:gochecknoinits
func init() {
	registerResource("job", false, newJob)
}

type job struct {
	client  k8sClient
	cfg     *Config
	counter map[string]int
}

func newJob(client k8sClient, cfg *Config) resource {
	return &job{client: client, cfg: cfg, counter: make(map[string]int)}
}

func (j *job) gatherMetric(ctx context.Context, timestamp int64) {
	if !j.cfg.EnableCollectJob {
		return
	}

	var continued string
	for {
		list, err := j.client.GetJobs(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := j.buildMetricPoints(list, timestamp)
		feedMetric("k8s-job-metric", j.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}

	processCounter(j.cfg, "job", j.counter, timestamp)
}

func (j *job) gatherObject(ctx context.Context) {
	if !j.cfg.EnableCollectJob {
		return
	}

	var continued string
	for {
		list, err := j.client.GetJobs(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := j.buildObjectPoints(list)
		feedObject("k8s-job-object", j.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
}

func (*job) addChangeInformer(_ informers.SharedInformerFactory) { /* nil */ }

func (j *job) buildMetricPoints(list *apibatchv1.JobList, timestamp int64) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTimestamp(timestamp))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("job", item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		kvs = kvs.Add("active", item.Status.Active)
		kvs = kvs.Add("failed", item.Status.Failed)
		kvs = kvs.Add("succeeded", item.Status.Succeeded)

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
		kvs = kvs.Add("completion_succeeded", succeeded)
		kvs = kvs.Add("completion_failed", failed)

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, j.cfg.LabelAsTagsForMetric.All, j.cfg.LabelAsTagsForMetric.Keys)...)
		kvs = append(kvs, point.NewTags(j.cfg.ExtraTags)...)
		pt := point.NewPoint(jobMetricMeasurement, kvs, opts...)
		pts = append(pts, pt)

		j.counter[item.Namespace]++
	}

	return pts
}

func (j *job) buildObjectPoints(list *apibatchv1.JobList) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultObjectOptions(), point.WithTime(ntp.Now()))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("name", string(item.UID))
		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("job_name", item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		kvs = kvs.Add("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		kvs = kvs.Add("active", item.Status.Active)
		kvs = kvs.Add("succeeded", item.Status.Succeeded)
		kvs = kvs.Add("failed", item.Status.Failed)

		if item.Spec.Parallelism != nil {
			kvs = kvs.Add("parallelism", *item.Spec.Parallelism)
		}
		if item.Spec.Completions != nil {
			kvs = kvs.Add("completions", *item.Spec.Completions)
		}
		if item.Spec.ActiveDeadlineSeconds != nil {
			kvs = kvs.Add("active_deadline", *item.Spec.ActiveDeadlineSeconds)
		}
		if item.Spec.BackoffLimit != nil {
			kvs = kvs.Add("backoff_limit", *item.Spec.BackoffLimit)
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

		if item.Spec.Selector != nil {
			kvs = append(kvs, point.NewTags(item.Spec.Selector.MatchLabels)...)
		}

		kvs = append(kvs, pointutil.ExtractSourceCodeFromAnnotations(item.Annotations)...) // add source_code
		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, j.cfg.LabelAsTagsForNonMetric.All, j.cfg.LabelAsTagsForNonMetric.Keys)...)
		kvs = append(kvs, point.NewTags(j.cfg.ExtraTags)...)
		pt := point.NewPoint(jobObjectClass, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

type JobMetric struct{}

//nolint:lll
func (*JobMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: jobMetricMeasurement,
		Desc: "The metric of the Kubernetes Job.",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"uid":              inputs.NewTagInfo("The UID of Job."),
			"job":              inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
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

type JobObject struct{}

//nolint:lll
func (*JobObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: jobObjectClass,
		Desc: "The object of the Kubernetes Job.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			"name":                              inputs.NewTagInfo("The UID of Job."),
			"uid":                               inputs.NewTagInfo("The UID of Job."),
			"job_name":                          inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":                         inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s":                  inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
			"&lt;ALL-SELECTOR-MATCH-LABELS&gt;": inputs.NewTagInfo("Represents the selector.matchLabels for Kubernetes resources"),
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

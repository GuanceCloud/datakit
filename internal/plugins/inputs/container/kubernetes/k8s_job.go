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
	registerMeasurements(&jobMetric{}, &jobObject{})
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

		kvs = kvs.AddV2("active", item.Status.Active, false)
		kvs = kvs.AddV2("failed", item.Status.Failed, false)
		kvs = kvs.AddV2("succeeded", item.Status.Succeeded, false)

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
		kvs = kvs.AddV2("completion_succeeded", succeeded, false)
		kvs = kvs.AddV2("completion_failed", failed, false)

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, j.cfg.LabelAsTagsForMetric.All, j.cfg.LabelAsTagsForMetric.Keys)...)
		kvs = append(kvs, point.NewTags(j.cfg.ExtraTags)...)
		pt := point.NewPointV2(jobMetricMeasurement, kvs, opts...)
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

		kvs = kvs.AddV2("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3, false)
		kvs = kvs.AddV2("active", item.Status.Active, false)
		kvs = kvs.AddV2("succeeded", item.Status.Succeeded, false)
		kvs = kvs.AddV2("failed", item.Status.Failed, false)

		if item.Spec.Parallelism != nil {
			kvs = kvs.AddV2("parallelism", *item.Spec.Parallelism, false)
		}
		if item.Spec.Completions != nil {
			kvs = kvs.AddV2("completions", *item.Spec.Completions, false)
		}
		if item.Spec.ActiveDeadlineSeconds != nil {
			kvs = kvs.AddV2("active_deadline", *item.Spec.ActiveDeadlineSeconds, false)
		}
		if item.Spec.BackoffLimit != nil {
			kvs = kvs.AddV2("backoff_limit", *item.Spec.BackoffLimit, false)
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

		if item.Spec.Selector != nil {
			kvs = append(kvs, point.NewTags(item.Spec.Selector.MatchLabels)...)
		}

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, j.cfg.LabelAsTagsForNonMetric.All, j.cfg.LabelAsTagsForNonMetric.Keys)...)
		kvs = append(kvs, point.NewTags(j.cfg.ExtraTags)...)
		pt := point.NewPointV2(jobObjectClass, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

type jobMetric struct{}

//nolint:lll
func (*jobMetric) Info() *inputs.MeasurementInfo {
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

type jobObject struct{}

//nolint:lll
func (*jobObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: jobObjectClass,
		Desc: "The object of the Kubernetes Job.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			"name":                       inputs.NewTagInfo("The UID of Job."),
			"uid":                        inputs.NewTagInfo("The UID of Job."),
			"job_name":                   inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":                  inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s":           inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
			"<all_selector_matchlabels>": inputs.NewTagInfo("Represents the selector.matchLabels for Kubernetes resources"),
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

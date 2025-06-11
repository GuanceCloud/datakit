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

	apiappsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const (
	daemonsetType              = "DaemonSet"
	daemonsetMetricMeasurement = "kube_daemonset"
	daemonsetObjectClass       = "kubernetes_daemonset"
	daemonsetObjectResourceKey = "daemonset_name"
)

//nolint:gochecknoinits
func init() {
	registerResource("daemonset", false, newDaemonset)
	registerMeasurements(&daemonsetMetric{}, &daemonsetObject{})
}

type daemonset struct {
	client  k8sClient
	cfg     *Config
	counter map[string]int
}

func newDaemonset(client k8sClient, cfg *Config) resource {
	return &daemonset{client: client, cfg: cfg, counter: make(map[string]int)}
}

func (d *daemonset) gatherMetric(ctx context.Context, timestamp int64) {
	var continued string
	for {
		list, err := d.client.GetDaemonSets(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := d.buildMetricPoints(list, timestamp)
		feedMetric("k8s-daemonset-metric", d.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
	processCounter(d.cfg, "daemonset", d.counter, timestamp)
}

func (d *daemonset) gatherObject(ctx context.Context) {
	var continued string
	for {
		list, err := d.client.GetDaemonSets(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := d.buildObjectPoints(list)
		feedObject("k8s-daemonset-object", d.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
}

func (d *daemonset) addChangeInformer(informerFactory informers.SharedInformerFactory) {
	informer := informerFactory.Apps().V1().DaemonSets()
	if informer == nil {
		klog.Warn("cannot get daemonset informer")
		return
	}

	updateFunc := func(oldObj, newObj interface{}) {
		objectChangeCountVec.WithLabelValues(daemonsetType, "update").Inc()

		oldDaemonsetObj, ok := oldObj.(*apiappsv1.DaemonSet)
		if !ok {
			klog.Warnf("converting to DaemonSet object failed, %v", oldObj)
			return
		}

		newDaemonsetObj, ok := newObj.(*apiappsv1.DaemonSet)
		if !ok {
			klog.Warnf("converting to DaemonSet object failed, %v", newObj)
			return
		}

		difftext, err := diffObject(oldDaemonsetObj.Spec, newDaemonsetObj.Spec)
		if err != nil {
			klog.Warnf("marshal failed, err: %s", err)
			return
		}

		if difftext != "" {
			objectChangeCountVec.WithLabelValues(daemonsetType, "spec-changed").Inc()
			processChange(d.cfg, daemonsetObjectClass, daemonsetObjectResourceKey, daemonsetType, difftext, newDaemonsetObj)
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

func (d *daemonset) buildMetricPoints(list *apiappsv1.DaemonSetList, timestamp int64) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTimestamp(timestamp))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("daemonset", item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		kvs = kvs.AddV2("desired", item.Status.DesiredNumberScheduled, false)
		kvs = kvs.AddV2("scheduled", item.Status.CurrentNumberScheduled, false)
		kvs = kvs.AddV2("misscheduled", item.Status.NumberMisscheduled, false)
		kvs = kvs.AddV2("ready", item.Status.NumberReady, false)
		kvs = kvs.AddV2("updated", item.Status.UpdatedNumberScheduled, false)
		kvs = kvs.AddV2("daemons_available", item.Status.NumberAvailable, false)
		kvs = kvs.AddV2("daemons_unavailable", item.Status.NumberUnavailable, false)

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, d.cfg.LabelAsTagsForMetric.All, d.cfg.LabelAsTagsForMetric.Keys)...)
		kvs = append(kvs, point.NewTags(d.cfg.ExtraTags)...)
		pt := point.NewPointV2(daemonsetMetricMeasurement, kvs, opts...)
		pts = append(pts, pt)

		d.counter[item.Namespace]++
	}

	return pts
}

func (d *daemonset) buildObjectPoints(list *apiappsv1.DaemonSetList) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultObjectOptions(), point.WithTime(ntp.Now()))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("name", string(item.UID))
		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag(daemonsetObjectResourceKey, item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		kvs = kvs.AddV2("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3, false)
		kvs = kvs.AddV2("desired", item.Status.DesiredNumberScheduled, false)
		kvs = kvs.AddV2("scheduled", item.Status.CurrentNumberScheduled, false)
		kvs = kvs.AddV2("misscheduled", item.Status.NumberMisscheduled, false)
		kvs = kvs.AddV2("ready", item.Status.NumberReady, false)
		kvs = kvs.AddV2("updated", item.Status.UpdatedNumberScheduled, false)
		kvs = kvs.AddV2("daemons_available", item.Status.NumberAvailable, false)
		kvs = kvs.AddV2("daemons_unavailable", item.Status.NumberUnavailable, false)

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

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, d.cfg.LabelAsTagsForNonMetric.All, d.cfg.LabelAsTagsForNonMetric.Keys)...)
		kvs = append(kvs, point.NewTags(d.cfg.ExtraTags)...)
		pt := point.NewPointV2(daemonsetObjectClass, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

type daemonsetMetric struct{}

//nolint:lll
func (*daemonsetMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: daemonsetMetricMeasurement,
		Desc: "The metric of the Kubernetes DaemonSet.",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"uid":              inputs.NewTagInfo("The UID of DaemonSet."),
			"daemonset":        inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"desired":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that should be running the daemon pod (including nodes correctly running the daemon pod)."},
			"scheduled":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running at least one daemon pod and are supposed to run the daemon pod."},
			"misscheduled":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running the daemon pod, but are not supposed to run the daemon pod."},
			"ready":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready."},
			"updated":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that are running updated daemon pod."},
			"daemons_available":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and available (ready for at least spec.minReadySeconds)."},
			"daemons_unavailable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have none of the daemon pod running and available (ready for at least spec.minReadySeconds)."},
		},
	}
}

type daemonsetObject struct{}

//nolint:lll
func (*daemonsetObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: daemonsetObjectClass,
		Desc: "The object of the Kubernetes DaemonSet.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			"name":                       inputs.NewTagInfo("The UID of DaemonSet."),
			"uid":                        inputs.NewTagInfo("The UID of DaemonSet."),
			"daemonset_name":             inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":                  inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s":           inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
			"<all_selector_matchlabels>": inputs.NewTagInfo("Represents the selector.matchLabels for Kubernetes resources"),
		},
		Fields: map[string]interface{}{
			"age":                 &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"desired":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that should be running the daemon pod (including nodes correctly running the daemon pod)."},
			"scheduled":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running at least one daemon pod and are supposed to run the daemon pod."},
			"misscheduled":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running the daemon pod, but are not supposed to run the daemon pod."},
			"ready":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready."},
			"updated":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that are running updated daemon pod."},
			"daemons_available":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and available (ready for at least spec.minReadySeconds)."},
			"daemons_unavailable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have none of the daemon pod running and available (ready for at least spec.minReadySeconds)."},
			"message":             &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	apiappsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/yaml"
)

const (
	statefulsetType              = "StatefulSet"
	statefulsetMetricMeasurement = "kube_statefulset"
	statefulsetObjectClass       = "kubernetes_statefulsets"
	statefulsetObjectResourceKey = "statefulset_name"
)

//nolint:gochecknoinits
func init() {
	registerResource("statefulset", false, newStatefulset)
	registerMeasurements(&statefulsetMetric{}, &statefulsetObject{})
}

type statefulset struct {
	client  k8sClient
	cfg     *Config
	counter map[string]int
}

func newStatefulset(client k8sClient, cfg *Config) resource {
	return &statefulset{client: client, cfg: cfg, counter: make(map[string]int)}
}

func (s *statefulset) gatherMetric(ctx context.Context, timestamp int64) {
	var continued string
	for {
		list, err := s.client.GetStatefulSets(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := s.buildMetricPoints(list, timestamp)
		feedMetric("k8s-statefulset-metric", s.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
	processCounter(s.cfg, "statefulset", s.counter, timestamp)
}

func (s *statefulset) gatherObject(ctx context.Context) {
	var continued string
	for {
		list, err := s.client.GetStatefulSets(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := s.buildObjectPoints(list)
		feedObject("k8s-statefulset-object", s.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
}

func (s *statefulset) addChangeInformer(informerFactory informers.SharedInformerFactory) {
	informer := informerFactory.Apps().V1().StatefulSets()
	if informer == nil {
		klog.Warn("cannot get statefulset informer")
		return
	}

	addFunc := func(newObj interface{}) {
		obj, ok := newObj.(*apiappsv1.StatefulSet)
		if !ok {
			klog.Warnf("converting to StatefulSet object failed, %v", newObj)
			return
		}
		if obj.CreationTimestamp.After(controllerStartTime) {
			diffs := createNoChangedFieldDiffs(changes.StatefulSetCreate, obj.Namespace, statefulsetType, obj.Name)
			objectChangeCountVec.WithLabelValues(statefulsetType, "create").Inc()
			processChange(s.cfg, statefulsetObjectClass, statefulsetObjectResourceKey, diffs, obj)
		}
	}

	deleteFunc := func(oldObj interface{}) {
		obj, ok := oldObj.(*apiappsv1.StatefulSet)
		if !ok {
			klog.Warnf("converting to StatefulSet object failed, %v", oldObj)
			return
		}

		diffs := createNoChangedFieldDiffs(changes.StatefulSetDelete, obj.Namespace, statefulsetType, obj.Name)
		objectChangeCountVec.WithLabelValues(statefulsetType, "delete").Inc()
		processChange(s.cfg, statefulsetObjectClass, statefulsetObjectResourceKey, diffs, obj)
	}

	updateFunc := func(oldObj, newObj interface{}) {
		objectChangeCountVec.WithLabelValues(statefulsetType, "update").Inc()

		oldStatefulSetObj, ok := oldObj.(*apiappsv1.StatefulSet)
		if !ok {
			klog.Warnf("converting to StatefulSet object failed, %v", oldObj)
			return
		}

		newStatefulSetObj, ok := newObj.(*apiappsv1.StatefulSet)
		if !ok {
			klog.Warnf("converting to StatefulSet object failed, %v", newObj)
			return
		}

		diffs := compareStatefulSet(oldStatefulSetObj, newStatefulSetObj)
		if len(diffs) != 0 {
			objectChangeCountVec.WithLabelValues(statefulsetType, "spec-changed").Inc()
			processChange(s.cfg, statefulsetObjectClass, statefulsetObjectResourceKey, diffs, newStatefulSetObj)
		}
	}

	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(newObj interface{}) {
			addFunc(newObj)
		},
		DeleteFunc: func(oldObj interface{}) {
			deleteFunc(oldObj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			updateFunc(oldObj, newObj)
		},
	})
}

func (s *statefulset) buildMetricPoints(list *apiappsv1.StatefulSetList, timestamp int64) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTimestamp(timestamp))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("statefulset", item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		kvs = kvs.AddV2("replicas", item.Status.Replicas, false)
		kvs = kvs.AddV2("replicas_updated", item.Status.UpdatedReplicas, false)
		kvs = kvs.AddV2("replicas_ready", item.Status.ReadyReplicas, false)
		kvs = kvs.AddV2("replicas_current", item.Status.CurrentReplicas, false)
		kvs = kvs.AddV2("replicas_available", item.Status.AvailableReplicas, false)

		if item.Spec.Replicas != nil {
			kvs = kvs.AddV2("replicas_desired", *item.Spec.Replicas, false)
		}

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, s.cfg.LabelAsTagsForMetric.All, s.cfg.LabelAsTagsForMetric.Keys)...)
		kvs = append(kvs, point.NewTags(s.cfg.ExtraTags)...)
		pt := point.NewPointV2(statefulsetMetricMeasurement, kvs, opts...)
		pts = append(pts, pt)

		s.counter[item.Namespace]++
	}

	return pts
}

func (s *statefulset) buildObjectPoints(list *apiappsv1.StatefulSetList) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultObjectOptions(), point.WithTime(ntp.Now()))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("name", string(item.UID))
		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag(statefulsetObjectResourceKey, item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		kvs = kvs.AddV2("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3, false)
		kvs = kvs.AddV2("replicas", item.Status.Replicas, false)
		kvs = kvs.AddV2("replicas_updated", item.Status.UpdatedReplicas, false)
		kvs = kvs.AddV2("replicas_ready", item.Status.ReadyReplicas, false)
		kvs = kvs.AddV2("replicas_current", item.Status.CurrentReplicas, false)
		kvs = kvs.AddV2("replicas_available", item.Status.AvailableReplicas, false)

		if item.Spec.Replicas != nil {
			kvs = kvs.AddV2("replicas_desired", *item.Spec.Replicas, false)
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

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, s.cfg.LabelAsTagsForNonMetric.All, s.cfg.LabelAsTagsForNonMetric.Keys)...)
		kvs = append(kvs, point.NewTags(s.cfg.ExtraTags)...)
		pt := point.NewPointV2(statefulsetObjectClass, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

func compareStatefulSet(oldVal, newVal *apiappsv1.StatefulSet) []FieldDiff {
	res := comparePodTemplate(&(oldVal.Spec.Template), &(newVal.Spec.Template))
	res = append(res, compareLabels(changes.StatefulSetLabels, &(oldVal.ObjectMeta), &(newVal.ObjectMeta))...)
	res = append(res, compareAnnotations(changes.StatefulSetAnnotations, &(oldVal.ObjectMeta), &(newVal.ObjectMeta))...)

	if oldVal.Spec.Replicas != nil && newVal.Spec.Replicas != nil &&
		*oldVal.Spec.Replicas != *newVal.Spec.Replicas {
		oldReplicas := strconv.Itoa(int(*oldVal.Spec.Replicas))
		newReplicas := strconv.Itoa(int(*newVal.Spec.Replicas))
		res = append(res, FieldDiff{
			ChangeID: changes.StatefulSetReplicas,
			OldValue: oldReplicas,
			NewValue: newReplicas,
			DiffText: formatAsDiffLines("replicas", oldReplicas, newReplicas),
		})
	}

	fillOwnerInfoForDiffs(res, newVal.Namespace, statefulsetType, newVal.Name)
	return res
}

type statefulsetMetric struct{}

//nolint:lll
func (*statefulsetMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: statefulsetMetricMeasurement,
		Desc: "The metric of the Kubernetes StatefulSet.",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"uid":              inputs.NewTagInfo("The UID of StatefulSet."),
			"statefulset":      inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"replicas":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of Pods created by the StatefulSet controller."},
			"replicas_desired":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The desired number of replicas of the given Template."},
			"replicas_ready":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods created for this StatefulSet with a Ready Condition."},
			"replicas_current":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of Pods created by the StatefulSet controller from the StatefulSet version indicated by currentRevision."},
			"replicas_updated":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of Pods created by the StatefulSet controller from the StatefulSet version indicated by updateRevision."},
			"replicas_available": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of available pods (ready for at least minReadySeconds) targeted by this StatefulSet."},
		},
	}
}

type statefulsetObject struct{}

//nolint:lll
func (*statefulsetObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: statefulsetObjectClass,
		Desc: "The object of the Kubernetes StatefulSet.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			"name":                              inputs.NewTagInfo("The UID of StatefulSet."),
			"uid":                               inputs.NewTagInfo("The UID of StatefulSet."),
			"statefulset_name":                  inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":                         inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s":                  inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
			"&lt;ALL-SELECTOR-MATCH-LABELS&gt;": inputs.NewTagInfo("Represents the selector.matchLabels for Kubernetes resources"),
		},
		Fields: map[string]interface{}{
			"age":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"replicas":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of Pods created by the StatefulSet controller."},
			"replicas_desired":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The desired number of replicas of the given Template."},
			"replicas_ready":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods created for this StatefulSet with a Ready Condition."},
			"replicas_current":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of Pods created by the StatefulSet controller from the StatefulSet version indicated by currentRevision."},
			"replicas_updated":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of Pods created by the StatefulSet controller from the StatefulSet version indicated by updateRevision."},
			"replicas_available": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of available pods (ready for at least minReadySeconds) targeted by this StatefulSet."},
			"message":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}

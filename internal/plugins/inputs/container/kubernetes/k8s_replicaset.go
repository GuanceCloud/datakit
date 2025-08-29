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
)

const (
	replicasetMetricMeasurement = "kube_replicaset"
	replicasetObjectClass       = "kubernetes_replica_sets"
)

//nolint:gochecknoinits
func init() {
	registerResource("replicaset", false, newReplicaset)
}

type replicaset struct {
	client  k8sClient
	cfg     *Config
	counter map[string]int
}

func newReplicaset(client k8sClient, cfg *Config) resource {
	return &replicaset{client: client, cfg: cfg, counter: make(map[string]int)}
}

func (r *replicaset) gatherMetric(ctx context.Context, timestamp int64) {
	var continued string
	for {
		list, err := r.client.GetReplicaSets(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := r.buildMetricPoints(list, timestamp)
		feedMetric("k8s-replicaset-metric", r.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
	processCounter(r.cfg, "replicaset", r.counter, timestamp)
}

func (r *replicaset) gatherObject(ctx context.Context) {
	var continued string
	for {
		list, err := r.client.GetReplicaSets(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := r.buildObjectPoints(list)
		feedObject("k8s-replicaset-Object", r.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
}

func (*replicaset) addChangeInformer(_ informers.SharedInformerFactory) { /* nil */ }

func (r *replicaset) buildMetricPoints(list *apiappsv1.ReplicaSetList, timestamp int64) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTimestamp(timestamp))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("replicaset", item.Name)
		kvs = kvs.AddTag("replica_set", item.Name) // Deprecated
		kvs = kvs.AddTag("namespace", item.Namespace)

		kvs = kvs.Add("replicas", item.Status.Replicas)
		kvs = kvs.Add("replicas_ready", item.Status.ReadyReplicas)
		kvs = kvs.Add("replicas_available", item.Status.AvailableReplicas)
		kvs = kvs.Add("fully_labeled_replicas", item.Status.FullyLabeledReplicas)

		if item.Spec.Replicas != nil {
			kvs = kvs.Add("replicas_desired", *item.Spec.Replicas)
		}

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, r.cfg.LabelAsTagsForMetric.All, r.cfg.LabelAsTagsForMetric.Keys)...)
		kvs = append(kvs, point.NewTags(r.cfg.ExtraTags)...)
		pt := point.NewPoint(replicasetMetricMeasurement, kvs, opts...)
		pts = append(pts, pt)

		r.counter[item.Namespace]++
	}

	return pts
}

func (r *replicaset) buildObjectPoints(list *apiappsv1.ReplicaSetList) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultObjectOptions(), point.WithTime(ntp.Now()))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("name", string(item.UID))
		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("replicaset_name", item.Name)
		kvs = kvs.AddTag("replica_set_name", item.Name) // Deprecated
		kvs = kvs.AddTag("namespace", item.Namespace)

		if len(item.OwnerReferences) != 0 {
			switch item.OwnerReferences[0].Kind {
			case "Deployment":
				kvs = kvs.AddTag("deployment", item.OwnerReferences[0].Name)
			case "StatefulSet":
				kvs = kvs.AddTag("statefulset", item.OwnerReferences[0].Name)
			default:
				// nil
			}
		}

		kvs = kvs.Add("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		kvs = kvs.Add("replicas", item.Status.Replicas)
		kvs = kvs.Add("replicas_ready", item.Status.ReadyReplicas)
		kvs = kvs.Add("replicas_available", item.Status.AvailableReplicas)
		kvs = kvs.Add("ready", item.Status.ReadyReplicas)         // Deprecated
		kvs = kvs.Add("available", item.Status.AvailableReplicas) // Deprecated

		if item.Spec.Replicas != nil {
			kvs = kvs.Add("replicas_desired", *item.Spec.Replicas)
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
		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, r.cfg.LabelAsTagsForNonMetric.All, r.cfg.LabelAsTagsForNonMetric.Keys)...)
		kvs = append(kvs, point.NewTags(r.cfg.ExtraTags)...)
		pt := point.NewPoint(replicasetObjectClass, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

type ReplicasetMetric struct{}

//nolint:lll
func (*ReplicasetMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: replicasetMetricMeasurement,
		Desc: "The metric of the Kubernetes ReplicaSet.",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"uid":              inputs.NewTagInfo("The UID of ReplicaSet."),
			"replicaset":       inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"replicas":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The most recently observed number of replicas."},
			"replicas_desired":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of desired replicas."},
			"replicas_ready":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of ready replicas for this replica set."},
			"replicas_available":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of available replicas (ready for at least minReadySeconds) for this replica set."},
			"fully_labeled_replicas": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of fully labeled replicas per ReplicaSet."},
		},
	}
}

type ReplicasetObject struct{}

//nolint:lll
func (*ReplicasetObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: replicasetObjectClass,
		Desc: "The object of the Kubernetes ReplicaSet.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			"name":                              inputs.NewTagInfo("The UID of ReplicaSet."),
			"uid":                               inputs.NewTagInfo("The UID of ReplicaSet."),
			"replicaset_name":                   inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":                         inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"deployment":                        inputs.NewTagInfo("The name of the Deployment which the object belongs to."),
			"statefulset":                       inputs.NewTagInfo("The name of the StatefulSet which the object belongs to."),
			"cluster_name_k8s":                  inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
			"&lt;ALL-SELECTOR-MATCH-LABELS&gt;": inputs.NewTagInfo("Represents the selector.matchLabels for Kubernetes resources"),
		},
		Fields: map[string]interface{}{
			"age":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"replicas":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The most recently observed number of replicas."},
			"replicas_desired":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of desired replicas."},
			"replicas_ready":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of ready replicas for this replica set."},
			"replicas_available": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of available replicas (ready for at least minReadySeconds) for this replica set."},
			"message":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
			"ready":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of ready replicas for this replica set. (Deprecated)"},
			"available":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of available replicas (ready for at least minReadySeconds) for this replica set. (Deprecated)"},
		},
	}
}

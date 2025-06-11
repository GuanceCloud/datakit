// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/yaml"
)

const (
	nodeType              = "Node"
	nodeMetricMeasurement = "kube_node"
	nodeObjectClass       = "kubernetes_nodes"
	nodeObjectResourceKey = "node_name"
)

//nolint:gochecknoinits
func init() {
	registerResource("node", false, newNode)
	registerMeasurements(&nodeMetric{}, &nodeObject{})
}

type node struct {
	client  k8sClient
	cfg     *Config
	counter map[string]int
}

func newNode(client k8sClient, cfg *Config) resource {
	return &node{client: client, cfg: cfg, counter: make(map[string]int)}
}

func (n *node) gatherMetric(ctx context.Context, timestamp int64) {
	var continued string
	for {
		list, err := n.client.GetNodes().List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := n.buildMetricPoints(list, timestamp)
		feedMetric("k8s-node-metric", n.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
	processCounter(n.cfg, "node", n.counter, timestamp)
}

func (n *node) gatherObject(ctx context.Context) {
	var continued string
	for {
		list, err := n.client.GetNodes().List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := n.buildObjectPoints(list)
		feedObject("k8s-node-object", n.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
}

func (n *node) addChangeInformer(informerFactory informers.SharedInformerFactory) {
	informer := informerFactory.Core().V1().Nodes()
	if informer == nil {
		klog.Warn("cannot get node informer")
		return
	}

	updateFunc := func(oldObj, newObj interface{}) {
		objectChangeCountVec.WithLabelValues(nodeType, "update").Inc()

		oldNodeObj, ok := oldObj.(*apicorev1.Node)
		if !ok {
			klog.Warnf("converting to Node object failed, %v", oldObj)
			return
		}

		newNodeObj, ok := newObj.(*apicorev1.Node)
		if !ok {
			klog.Warnf("converting to Node object failed, %v", newObj)
			return
		}

		difftext, err := diffObject(oldNodeObj.Spec, newNodeObj.Spec)
		if err != nil {
			klog.Warnf("marshal failed, err: %s", err)
			return
		}

		if difftext != "" {
			objectChangeCountVec.WithLabelValues(nodeType, "spec-changed").Inc()
			processChange(n.cfg, nodeObjectClass, nodeObjectResourceKey, nodeType, difftext, newNodeObj)
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

func (n *node) buildMetricPoints(list *apicorev1.NodeList, timestamp int64) []*point.Point {
	var pts []*point.Point
	opts := point.DefaultMetricOptions()

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("node", config.RenameNode(item.Name))

		t := item.Status.Allocatable["cpu"]
		kvs = kvs.AddV2("cpu_allocatable", t.AsApproximateFloat64(), false)

		m := item.Status.Allocatable["memory"]
		kvs = kvs.AddV2("memory_allocatable", m.AsApproximateFloat64(), false)

		p := item.Status.Allocatable["pods"]
		kvs = kvs.AddV2("pods_allocatable", p.AsApproximateFloat64(), false)

		e := item.Status.Allocatable["ephemeral-storage"]
		kvs = kvs.AddV2("ephemeral_storage_allocatable", e.AsApproximateFloat64(), false)

		t2 := item.Status.Capacity["cpu"]
		kvs = kvs.AddV2("cpu_capacity", t2.AsApproximateFloat64(), false)

		m2 := item.Status.Capacity["memory"]
		kvs = kvs.AddV2("memory_capacity", m2.AsApproximateFloat64(), false)

		p2 := item.Status.Capacity["pods"]
		kvs = kvs.AddV2("pods_capacity", p2.AsApproximateFloat64(), false)

		e2 := item.Status.Capacity["ephemeral-storage"]
		kvs = kvs.AddV2("ephemeral_storage_capacity", e2.AsApproximateFloat64(), false)

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, n.cfg.LabelAsTagsForMetric.All, n.cfg.LabelAsTagsForMetric.Keys)...)
		kvs = append(kvs, point.NewTags(n.cfg.ExtraTags)...)
		pt := point.NewPointV2(nodeMetricMeasurement, kvs, append(opts, point.WithTimestamp(timestamp))...)
		pts = append(pts, pt)
	}

	n.counter[""] += len(list.Items)

	return pts
}

func (n *node) buildObjectPoints(list *apicorev1.NodeList) []*point.Point {
	var pts []*point.Point
	opts := point.DefaultObjectOptions()

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("name", string(item.UID))
		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag(nodeObjectResourceKey, config.RenameNode(item.Name))
		for _, condition := range item.Status.Conditions {
			if condition.Reason == "KubeletReady" {
				kvs = kvs.AddTag("status", string(condition.Type))
				break
			}
		}

		if _, ok := item.Labels["node-role.kubernetes.io/master"]; ok {
			kvs = kvs.AddTag("role", "master")
		} else {
			kvs = kvs.AddTag("role", "node")
		}

		for _, address := range item.Status.Addresses {
			if address.Type == apicorev1.NodeInternalIP {
				kvs = kvs.AddTag("internal_ip", address.Address)
				break
			}
		}

		kvs = kvs.AddV2("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3, false)
		kvs = kvs.AddV2("kubelet_version", item.Status.NodeInfo.KubeletVersion, false)

		if len(item.Spec.Taints) != 0 {
			if taints, err := json.Marshal(item.Spec.Taints); err == nil {
				kvs = kvs.AddV2("taints", string(taints), false)
			}
		}

		if item.Spec.Unschedulable {
			kvs = kvs.AddV2("unschedulable", "yes", false)
		} else {
			kvs = kvs.AddV2("unschedulable", "no", false)
		}

		for _, condition := range item.Status.Conditions {
			if condition.Type == apicorev1.NodeReady {
				kvs = kvs.AddV2("node_ready", strings.ToLower(string(condition.Status)), false)
				break
			}
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

		if n.cfg.EnableExtractK8sLabelAsTagsV1 {
			kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, true /*all labels*/, nil)...)
		} else {
			kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, n.cfg.LabelAsTagsForNonMetric.All, n.cfg.LabelAsTagsForNonMetric.Keys)...)
		}
		kvs = append(kvs, point.NewTags(n.cfg.ExtraTags)...)
		pt := point.NewPointV2(nodeObjectClass, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

type nodeMetric struct{}

//nolint:lll
func (*nodeMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: nodeMetricMeasurement,
		Desc: "The metric of the Kubernetes Node.",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"uid":              inputs.NewTagInfo("The UID of Node."),
			"node":             inputs.NewTagInfo("Name must be unique within a namespace"),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"cpu_allocatable":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The allocatable CPU of a node that is available for scheduling."},
			"memory_allocatable":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The allocatable memory of a node that is available for scheduling."},
			"pods_allocatable":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The allocatable pods of a node that is available for scheduling."},
			"ephemeral_storage_allocatable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The allocatable ephemeral-storage of a node that is available for scheduling."},
			"cpu_capacity":                  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The CPU capacity of a node."},
			"memory_capacity":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The memory capacity of a node."},
			"pods_capacity":                 &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The pods capacity of a node."},
			"ephemeral_storage_capacity":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The ephemeral-storage capacity of a node."},
		},
	}
}

type nodeObject struct{}

//nolint:lll
func (*nodeObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: nodeObjectClass,
		Desc: "The object of the Kubernetes Node.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			"name":             inputs.NewTagInfo("The UID of Node."),
			"uid":              inputs.NewTagInfo("The UID of Node."),
			"node_name":        inputs.NewTagInfo("Name must be unique within a namespace."),
			"internal_ip":      inputs.NewTagInfo("Node internal IP"),
			"role":             inputs.NewTagInfo("Node role. (master/node)"),
			"status":           inputs.NewTagInfo("NodePhase is the recently observed lifecycle phase of the node. (Pending/Running/Terminated)"),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"age":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)."},
			"kubelet_version": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Kubelet Version reported by the node."},
			"node_ready":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "NodeReady means kubelet is healthy and ready to accept pods (true/false/unknown)."},
			"taints":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Node's taints."},
			"unschedulable":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Unschedulable controls node schedulability of new pods (yes/no)."},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details."},
		},
	}
}

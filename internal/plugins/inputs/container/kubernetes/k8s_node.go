// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	nodeMetricMeasurement = "kube_node"
	nodeObjectMeasurement = "kubernetes_nodes"
)

//nolint:gochecknoinits
func init() {
	registerResource("node", false, false, newNode)
	registerMeasurements(&nodeMetric{}, &nodeObject{})
}

type node struct {
	client    k8sClient
	continued string
	counter   map[string]int
}

func newNode(client k8sClient) resource {
	return &node{client: client, counter: make(map[string]int)}
}

func (n *node) count() []pointV2 { return buildCountPoints("node", n.counter) }

func (n *node) hasNext() bool { return n.continued != "" }

func (n *node) getMetadata(ctx context.Context, _, fieldSelector string) (metadata, error) {
	opt := metav1.ListOptions{
		Limit:         queryLimit,
		Continue:      n.continued,
		FieldSelector: fieldSelector,
	}

	list, err := n.client.GetNodes().List(ctx, opt)
	if err != nil {
		return nil, err
	}

	n.continued = list.Continue
	return &nodeMetadata{n, list}, nil
}

type nodeMetadata struct {
	parent *node
	list   *apicorev1.NodeList
}

func (m *nodeMetadata) newMetric(conf *Config) pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		met := typed.NewPointKV(nodeMetricMeasurement)

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("node", item.Name)
		// "resource", "unit"

		t := item.Status.Allocatable["cpu"]
		met.SetField("cpu_allocatable", t.AsApproximateFloat64())

		m := item.Status.Allocatable["memory"]
		met.SetField("memory_allocatable", m.AsApproximateFloat64())

		p := item.Status.Allocatable["pods"]
		met.SetField("pods_allocatable", p.AsApproximateFloat64())

		e := item.Status.Allocatable["ephemeral-storage"]
		met.SetField("ephemeral_storage_allocatable", e.AsApproximateFloat64())

		t2 := item.Status.Capacity["cpu"]
		met.SetField("cpu_capacity", t2.AsApproximateFloat64())

		m2 := item.Status.Capacity["memory"]
		met.SetField("memory_capacity", m2.AsApproximateFloat64())

		p2 := item.Status.Capacity["pods"]
		met.SetField("pods_capacity", p2.AsApproximateFloat64())

		e2 := item.Status.Capacity["ephemeral-storage"]
		met.SetField("ephemeral_storage_capacity", e2.AsApproximateFloat64())

		met.SetLabelAsTags(item.Labels, conf.LabelAsTagsForMetric.All, conf.LabelAsTagsForMetric.Keys)
		res = append(res, met)
	}

	m.parent.counter[""] += len(m.list.Items)

	return res
}

func (m *nodeMetadata) newObject(conf *Config) pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		obj := typed.NewPointKV(nodeObjectMeasurement)

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("uid", fmt.Sprintf("%v", item.UID))
		obj.SetTag("node_name", item.Name)
		obj.SetTag("status", fmt.Sprintf("%v", item.Status.Phase))

		obj.SetTag("role", "node")
		if _, ok := item.Labels["node-role.kubernetes.io/master"]; ok {
			obj.SetTag("role", "master")
		}

		for _, address := range item.Status.Addresses {
			if address.Type == apicorev1.NodeInternalIP {
				obj.SetTag("internal_ip", address.Address)
				break
			}
		}

		obj.SetField("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		obj.SetField("kubelet_version", item.Status.NodeInfo.KubeletVersion)

		if y, err := yaml.Marshal(item); err == nil {
			obj.SetField("yaml", string(y))
		}

		if item.Spec.Unschedulable {
			obj.SetField("unschedulable", "yes")
		} else {
			obj.SetField("unschedulable", "no")
		}

		for _, condition := range item.Status.Conditions {
			if condition.Type == apicorev1.NodeReady {
				obj.SetField("node_ready", strings.ToLower(string(condition.Status)))
				break
			}
		}

		obj.SetFields(transLabels(item.Labels))
		obj.SetField("annotations", typed.MapToJSON(item.Annotations))
		obj.SetField("message", typed.TrimString(obj.String(), maxMessageLength))
		obj.DeleteField("annotations")
		obj.DeleteField("yaml")

		if conf.EnableExtractK8sLabelAsTagsV1 {
			obj.SetLabelAsTags(item.Labels, true /*all labels*/, nil)
		} else {
			obj.SetLabelAsTags(item.Labels, conf.LabelAsTagsForNonMetric.All, conf.LabelAsTagsForNonMetric.Keys)
		}
		res = append(res, obj)
	}

	return res
}

type nodeMetric struct{}

//nolint:lll
func (*nodeMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: nodeMetricMeasurement,
		Desc: "The metric of the Kubernetes Node.",
		Type: "metric",
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
		Name: nodeObjectMeasurement,
		Desc: "The object of the Kubernetes Node.",
		Type: "object",
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
			"age":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"kubelet_version": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Kubelet Version reported by the node."},
			"node_ready":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "NodeReady means kubelet is healthy and ready to accept pods (true/false/unknown)"},
			"unschedulable":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Unschedulable controls node schedulability of new pods (yes/no)."},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}

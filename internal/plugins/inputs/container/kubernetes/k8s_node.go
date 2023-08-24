// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apicorev1 "k8s.io/api/core/v1"
)

//nolint:gochecknoinits
func init() {
	registerMetricResource("node", gatherNodeMetric)
	registerObjectResource("node", gatherNodeObject)
	registerMeasurement(&nodeMetric{})
	registerMeasurement(&nodeObject{})
}

func gatherNodeMetric(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetNodes().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeNodeMetric(list), nil
}

func composeNodeMetric(list *apicorev1.NodeList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		met := typed.NewPointKV()

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("node", item.Name)
		met.SetTag("namespace", item.Namespace)
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

		met.SetCustomerTags(item.Labels, getGlobalCustomerKeys())
		res = append(res, &nodeMetric{met})
	}

	return res
}

func gatherNodeObject(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetNodes().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeNodeObject(list), nil
}

func composeNodeObject(list *apicorev1.NodeList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		obj := typed.NewPointKV()

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("uid", fmt.Sprintf("%v", item.UID))
		obj.SetTag("node_name", item.Name)
		obj.SetTag("namespace", item.Namespace)
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

		obj.SetFields(transLabels(item.Labels))
		obj.SetField("annotations", typed.MapToJSON(item.Annotations))
		obj.SetField("message", typed.TrimString(obj.String(), maxMessageLength))
		obj.DeleteField("annotations")
		obj.DeleteField("yaml")

		obj.SetCustomerTags(item.Labels, getGlobalCustomerKeys())
		res = append(res, &nodeObject{obj})
	}

	return res
}

type nodeMetric struct{ typed.PointKV }

func (n *nodeMetric) namespace() string { return n.GetTag("namespace") }

func (n *nodeMetric) addExtraTags(m map[string]string) { n.SetTags(m) }

func (n *nodeMetric) LineProto() (*point.Point, error) {
	return point.NewPoint("kube_node", n.Tags(), n.Fields(), metricOpt)
}

//nolint:lll
func (*nodeMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_node",
		Desc: "The metric of the Kubernetes Node.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":  inputs.NewTagInfo("The UID of Node."),
			"node": inputs.NewTagInfo("Name must be unique within a namespace"),
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

type nodeObject struct{ typed.PointKV }

func (n *nodeObject) namespace() string { return n.GetTag("namespace") }

func (n *nodeObject) addExtraTags(m map[string]string) { n.SetTags(m) }

func (n *nodeObject) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes_nodes", n.Tags(), n.Fields(), objectOpt)
}

//nolint:lll
func (*nodeObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_nodes",
		Desc: "The object of the Kubernetes Node.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":        inputs.NewTagInfo("The UID of Node."),
			"uid":         inputs.NewTagInfo("The UID of Node."),
			"node_name":   inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":   inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"internal_ip": inputs.NewTagInfo("Node internal IP"),
			"role":        inputs.NewTagInfo("Node role. (master/node)"),
			"status":      inputs.NewTagInfo("NodePhase is the recently observed lifecycle phase of the node. (Pending/Running/Terminated)"),
		},
		Fields: map[string]interface{}{
			"age":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"kubelet_version": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Kubelet Version reported by the node."},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}

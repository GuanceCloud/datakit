// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/core/v1"
)

var (
	_ k8sResourceMetricInterface = (*node)(nil)
	_ k8sResourceObjectInterface = (*node)(nil)
)

type node struct {
	client    k8sClientX
	extraTags map[string]string
	items     []v1.Node
}

func newNode(client k8sClientX, extraTags map[string]string) *node {
	return &node{
		client:    client,
		extraTags: extraTags,
	}
}

func (n *node) name() string {
	return "node"
}

func (n *node) pullItems() error {
	if len(n.items) != 0 {
		return nil
	}

	list, err := n.client.getNodes().List(context.Background(), metaV1ListOption)
	if err != nil {
		return fmt.Errorf("failed to get nodes resource: %w", err)
	}

	n.items = list.Items
	return nil
}

func (n *node) metric() (inputsMeas, error) {
	if err := n.pullItems(); err != nil {
		return nil, err
	}
	var res inputsMeas

	for _, item := range n.items {
		met := &nodeMetric{
			tags: map[string]string{
				"node":      item.Name,
				"node_name": item.Name,
				// "resource"
				// "unit"
			},
			fields: map[string]interface{}{},
		}
		// t := item.Status.LastScheduleTime
		// met.fields["node.age"] = int64(time.Since(*t).Seconds())

		c := item.Status.Allocatable["cpu"]
		met.fields["cpu_allocatable"] = c.AsApproximateFloat64()

		m := item.Status.Allocatable["memory"]
		met.fields["memory_allocatable"] = m.AsApproximateFloat64()

		p := item.Status.Allocatable["pods"]
		met.fields["pods_allocatable"] = p.AsApproximateFloat64()

		e := item.Status.Allocatable["ephemeral-storage"]
		met.fields["ephemeral_storage_allocatable"] = e.AsApproximateFloat64()

		c2 := item.Status.Capacity["cpu"]
		met.fields["cpu_capacity"] = c2.AsApproximateFloat64()

		m2 := item.Status.Capacity["memory"]
		met.fields["memory_capacity"] = m2.AsApproximateFloat64()

		p2 := item.Status.Capacity["pods"]
		met.fields["pods_capacity"] = p2.AsApproximateFloat64()
		// node.by_condition
		// node.status

		met.tags.append(n.extraTags)
		res = append(res, met)
	}

	count, _ := n.count()
	for ns, c := range count {
		met := &nodeMetric{
			tags:   map[string]string{"namespace": ns},
			fields: map[string]interface{}{"count": c},
		}
		met.tags.append(n.extraTags)
		res = append(res, met)
	}

	return res, nil
}

func (n *node) object() (inputsMeas, error) {
	if err := n.pullItems(); err != nil {
		return nil, err
	}
	var res inputsMeas

	for _, item := range n.items {
		obj := &nodeObject{
			tags: map[string]string{
				"name":         fmt.Sprintf("%v", item.UID),
				"node_name":    item.Name,
				"status":       fmt.Sprintf("%v", item.Status.Phase),
				"cluster_name": defaultClusterName(item.ClusterName),
				"namespace":    defaultNamespace(item.Namespace),
			},
			fields: map[string]interface{}{
				"age":             int64(time.Since(item.CreationTimestamp.Time).Seconds()),
				"kubelet_version": item.Status.NodeInfo.KubeletVersion,
			},
		}

		if _, ok := item.Labels["node-role.kubernetes.io/master"]; ok {
			obj.tags["role"] = "master"
		} else {
			obj.tags["role"] = "node"
		}

		for _, address := range item.Status.Addresses {
			if address.Type == v1.NodeInternalIP {
				obj.tags["internal_ip"] = address.Address
				obj.tags["node_ip"] = address.Address // depercated
				break
			}
		}
		obj.tags.append(n.extraTags)

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)
		obj.fields.delete("annotations")

		res = append(res, obj)
	}

	return res, nil
}

func (n *node) count() (map[string]int, error) {
	if err := n.pullItems(); err != nil {
		return nil, err
	}

	m := make(map[string]int)
	for _, item := range n.items {
		m[defaultNamespace(item.Namespace)]++
	}

	if len(m) == 0 {
		m["default"] = 0
	}

	return m, nil
}

type nodeMetric struct {
	tags   tagsType
	fields fieldsType
}

func (n *nodeMetric) LineProto() (*point.Point, error) {
	return point.NewPoint("kube_node", n.tags, n.fields, point.MOptElection())
}

//nolint:lll
func (*nodeMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_node",
		Desc: "Kubernetes Node 指标数据",
		Type: "metric",
		Tags: map[string]interface{}{
			"node":      inputs.NewTagInfo("Name must be unique within a namespace. (depercated)"),
			"node_name": inputs.NewTagInfo("Name must be unique within a namespace."),
		},
		Fields: map[string]interface{}{
			"count":                         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of nodes"},
			"age":                           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "The time in seconds since the creation of the node"},
			"cpu_allocatable":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The allocatable CPU of a node that is available for scheduling."},
			"memory_allocatable":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The allocatable memory of a node that is available for scheduling."},
			"pods_allocatable":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The allocatable pods of a node that is available for scheduling."},
			"ephemeral_storage_allocatable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The allocatable ephemeral-storage of a node that is available for scheduling."},
			"cpu_capacity":                  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The CPU capacity of a node."},
			"memory_capacity":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The memory capacity of a node."},
			"pods_capacity":                 &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The pods capacity of a node."},
		},
	}
}

type nodeObject struct {
	tags   tagsType
	fields fieldsType
}

func (n *nodeObject) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes_nodes", n.tags, n.fields, point.OOptElection())
}

//nolint:lll
func (*nodeObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_nodes",
		Desc: "Kubernetes node 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("UID"),
			"node_name":    inputs.NewTagInfo("Name must be unique within a namespace."),
			"node_ip":      inputs.NewTagInfo("Node IP (depercated)"),
			"internal_ip":  inputs.NewTagInfo("Node internal IP"),
			"role":         inputs.NewTagInfo("Node role. (master/node)"),
			"cluster_name": inputs.NewTagInfo("The name of the cluster which the object belongs to."),
			"namespace":    inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"status":       inputs.NewTagInfo("NodePhase is the recently observed lifecycle phase of the node. (Pending/Running/Terminated)"),
		},
		Fields: map[string]interface{}{
			"age":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"kubelet_version": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Kubelet Version reported by the node."},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerK8sResourceMetric(func(c k8sClientX, m map[string]string) k8sResourceMetricInterface { return newNode(c, m) })
	registerK8sResourceObject(func(c k8sClientX, m map[string]string) k8sResourceObjectInterface { return newNode(c, m) })
	registerMeasurement(&nodeMetric{})
	registerMeasurement(&nodeObject{})
}

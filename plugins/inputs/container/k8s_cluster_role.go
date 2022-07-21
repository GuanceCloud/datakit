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
	v1 "k8s.io/api/rbac/v1"
)

var (
	_ k8sResourceObjectInterface = (*clusterRole)(nil)
	_ k8sResourceMetricInterface = (*clusterRole)(nil)
)

type clusterRole struct {
	client    k8sClientX
	extraTags map[string]string
	items     []v1.ClusterRole
}

func newClusterRole(client k8sClientX, extraTags map[string]string) *clusterRole {
	return &clusterRole{
		client:    client,
		extraTags: extraTags,
	}
}

func (c *clusterRole) name() string {
	return "cluster_role"
}

func (c *clusterRole) pullItems() error {
	if len(c.items) != 0 {
		return nil
	}

	list, err := c.client.getClusterRoles().List(context.Background(), metaV1ListOption)
	if err != nil {
		return fmt.Errorf("failed to get clusterRoles resource: %w", err)
	}

	c.items = list.Items
	return nil
}

func (c *clusterRole) metric() (inputsMeas, error) {
	return nil, nil
}

func (c *clusterRole) count() (map[string]int, error) {
	if err := c.pullItems(); err != nil {
		return nil, err
	}

	m := make(map[string]int)
	for _, item := range c.items {
		m[defaultNamespace(item.Namespace)]++
	}

	if len(m) == 0 {
		m["default"] = 0
	}

	return m, nil
}

func (c *clusterRole) object() (inputsMeas, error) {
	if err := c.pullItems(); err != nil {
		return nil, err
	}
	var res inputsMeas

	for _, item := range c.items {
		obj := &clusterRoleObject{
			tags: map[string]string{
				"name":              fmt.Sprintf("%v", item.UID),
				"cluster_role_name": item.Name,
				"cluster_name":      defaultClusterName(item.ClusterName),
				"namespace":         defaultNamespace(item.Namespace),
			},
			fields: map[string]interface{}{
				"age":         int64(time.Since(item.CreationTimestamp.Time).Seconds()),
				"create_time": item.CreationTimestamp.Time.Unix() / int64(time.Millisecond),
			},
		}

		obj.tags.append(c.extraTags)

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)
		obj.fields.delete("annotations")

		res = append(res, obj)
	}

	return res, nil
}

type clusterRoleObject struct {
	tags   tagsType
	fields fieldsType
}

func (c *clusterRoleObject) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes_cluster_roles", c.tags, c.fields, inputs.OptElectionObject)
}

//nolint:lll
func (*clusterRoleObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_cluster_roles",
		Desc: "Kubernetes cluster role 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":              inputs.NewTagInfo("UID"),
			"cluster_role_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"cluster_name":      inputs.NewTagInfo("The name of the cluster which the object belongs to."),
		},
		Fields: map[string]interface{}{
			"age":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"create_time": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampSec, Desc: "CreationTimestamp is a timestamp representing the server time when this object was created.(milliseconds)"},
			"message":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerK8sResourceMetric(func(c k8sClientX, m map[string]string) k8sResourceMetricInterface { return newClusterRole(c, m) })
	registerK8sResourceObject(func(c k8sClientX, m map[string]string) k8sResourceObjectInterface { return newClusterRole(c, m) })
	registerMeasurement(&clusterRoleObject{})
}

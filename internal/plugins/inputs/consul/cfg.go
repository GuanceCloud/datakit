// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package consul

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type ConsulMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *ConsulMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOptElection())
}

func (m *ConsulMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "consul",
		Fields: map[string]interface{}{
			"raft_leader":                  newCountFieldInfo("raft 集群中 leader 数量"),
			"raft_peers":                   newCountFieldInfo("raft 集群中 peer 数量"),
			"serf_lan_members":             newCountFieldInfo("集群中成员数量"),
			"catalog_services":             newCountFieldInfo("集群中服务数量"),
			"catalog_service_node_healthy": newUnknownFieldInfo("该服务在该结点上是否健康"),
			"health_node_status":           newUnknownFieldInfo("结点的健康检查状态"),
			"serf_lan_member_status":       newUnknownFieldInfo("集群里成员的状态。其中 1 表示 Alive/2 表示 Leaving/3 表示 Left/4 表示 Failed"),
		},
		Tags: map[string]interface{}{
			"host":         inputs.NewTagInfo("主机名称"),
			"node":         inputs.NewTagInfo("结点名称"),
			"service_id":   inputs.NewTagInfo("服务 ID"),
			"service_name": inputs.NewTagInfo("服务名称"),
			"status":       inputs.NewTagInfo("状态。status 有 critical/maintenance/passing/warning 四种"),
			"member":       inputs.NewTagInfo("成员名称"),
		},
	}
}

type HostMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *HostMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOptElection())
}

func (m *HostMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "consul_host",
		Fields: map[string]interface{}{
			"raft_leader":      newCountFieldInfo("raft 集群中 leader 数量"),
			"raft_peers":       newCountFieldInfo("raft 集群中 peer 数量"),
			"serf_lan_members": newCountFieldInfo("集群中成员数量"),
			"catalog_services": newCountFieldInfo("集群中服务数量"),
		},
		Tags: map[string]interface{}{
			"host": inputs.NewTagInfo("主机名称"),
		},
	}
}

type ServiceMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *ServiceMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOptElection())
}

func (m *ServiceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "consul_service",
		Fields: map[string]interface{}{
			"catalog_service_node_healthy": newUnknownFieldInfo("该服务在该结点上是否健康"),
		},
		Tags: map[string]interface{}{
			"host":         inputs.NewTagInfo("主机名称"),
			"node":         inputs.NewTagInfo("结点名称"),
			"service_id":   inputs.NewTagInfo("服务 id"),
			"service_name": inputs.NewTagInfo("服务名称"),
		},
	}
}

type HealthMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *HealthMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOptElection())
}

func (m *HealthMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "consul_health",
		Fields: map[string]interface{}{
			"health_node_status": newUnknownFieldInfo("结点的健康检查状态"),
		},
		Tags: map[string]interface{}{
			"host":   inputs.NewTagInfo("主机名称"),
			"node":   inputs.NewTagInfo("结点名称"),
			"status": inputs.NewTagInfo("状态，status 有 critical/maintenance/passing/warning 四种"),
		},
	}
}

type MemberMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *MemberMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOptElection())
}

func (m *MemberMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "consul_member",
		Fields: map[string]interface{}{
			"serf_lan_member_status": newUnknownFieldInfo("集群里成员的状态，其中 1 表示 Alive，2 表示 Leaving，3 表示 Left，4 表示 Failed"),
		},
		Tags: map[string]interface{}{
			"host":   inputs.NewTagInfo("主机名称"),
			"member": inputs.NewTagInfo("成员名称"),
		},
	}
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func newUnknownFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.UnknownUnit,
		Desc:     desc,
	}
}

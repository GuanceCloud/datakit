// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package consul

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type HostMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *HostMeasurement) LineProto() (*io.Point, error) {
	return io.NewPoint(m.name, m.tags, m.fields, inputs.OptElectionMetric)
}

func (m *HostMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "consul_host",
		Fields: map[string]interface{}{
			"raft_leader":      newCountFieldInfo("raft集群中leader数量"),
			"raft_peers":       newCountFieldInfo("raft集群中peer数量"),
			"serf_lan_members": newCountFieldInfo("集群中成员数量"),
			"catalog_service":  newCountFieldInfo("集群中服务数量"),
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

func (m *ServiceMeasurement) LineProto() (*io.Point, error) {
	return io.NewPoint(m.name, m.tags, m.fields, inputs.OptElectionMetric)
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
			"service_id":   inputs.NewTagInfo("服务id"),
			"service_name": inputs.NewTagInfo("服务名称"),
		},
	}
}

type HealthMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *HealthMeasurement) LineProto() (*io.Point, error) {
	return io.NewPoint(m.name, m.tags, m.fields, inputs.OptElectionMetric)
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
			"status": inputs.NewTagInfo("状态，status有critical, maintenance, passing,warning四种"),
		},
	}
}

type MemberMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *MemberMeasurement) LineProto() (*io.Point, error) {
	return io.NewPoint(m.name, m.tags, m.fields, inputs.OptElectionMetric)
}

func (m *MemberMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "consul_member",
		Fields: map[string]interface{}{
			"serf_lan_member_status": newUnknownFieldInfo("集群里成员的状态，其中1表示Alive，2表示Leaving，3表示Left，4表示Failed"),
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

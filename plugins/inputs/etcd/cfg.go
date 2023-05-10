// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package etcd

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type ServerMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

type NetworkMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *ServerMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOptElection())
}

func (m *NetworkMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOptElection())
}

func (m *NetworkMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "etcd_network",
		Type: "metric",
		Fields: map[string]interface{}{
			"network_client_grpc_sent_bytes_total":     newCountFieldInfo("发送到 grpc 客户端的总字节数"),
			"network_client_grpc_received_bytes_total": newCountFieldInfo("接收到 grpc 客户端的总字节数"),
		},
		Tags: map[string]interface{}{
			"host": inputs.NewTagInfo("主机名称"),
		},
	}
}

func (m *ServerMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "etcd_server",
		Type: "metric",
		Fields: map[string]interface{}{
			"server_leader_changes_seen_total": newCountFieldInfo("解释到的领导者变更次数"),
			"server_proposals_committed_total": newCountFieldInfo("提交的共识提案总数"),
			"server_proposals_applied_total":   newCountFieldInfo("已应用的共识提案总数"),
			"server_proposals_pending":         newCountFieldInfo("当前待处理提案的数量"),
			"server_proposals_failed_total":    newCountFieldInfo("看到的失败提案总数"),
		},
		Tags: map[string]interface{}{
			"host":              inputs.NewTagInfo("主机名称"),
			"server_has_leader": inputs.NewTagInfo("领导者是否存在。1 是存在。0 是不存在"),
		},
	}
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

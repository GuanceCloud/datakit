// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll,goconst,funlen // Measurement metadata contains intentionally long repeated literals.
package consul

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

// Info from github.com/prometheus/consul_exporter.
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "consul",
		Cat:    point.Metric,
		Desc:   "Consul metrics scraped from Prometheus consul_exporter output and normalized into the consul measurement.",
		DescZh: "从 Prometheus consul_exporter 输出采集并标准化到 consul 指标集的 Consul 指标，涵盖 Raft、Serf、服务目录、健康检查和 KV 数据。",
		//nolint:lll
		Fields: map[string]interface{}{
			"up":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Whether the consul_exporter successfully queried Consul during the last scrape."},
			"raft_peers":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of Consul servers participating as peers in the Raft cluster."},
			"raft_leader":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Whether this Consul node reports that the Raft cluster currently has a leader."},
			"serf_lan_members":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of members in the Consul Serf LAN pool."},
			"serf_lan_member_status":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Numeric Serf LAN member status code for the tagged member. 1=alive, 2=leaving, 3=left, 4=failed."},
			"serf_wan_member_status":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Numeric Serf WAN member status code for the tagged member. 1=alive, 2=leaving, 3=left, 4=failed."},
			"catalog_services":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of services registered in the Consul catalog."},
			"service_tag":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Whether the tagged Consul service tag exists for the service_id, node, and tag label set."},
			"catalog_service_node_healthy": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Whether the tagged Consul service is healthy on the tagged node."},
			"health_node_status":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Whether the tagged Consul node health check is currently in the tagged status."},
			"health_service_status":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Whether the tagged Consul service health check is currently in the tagged status."},
			"service_checks":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Whether the tagged Consul service is linked with the tagged check_id and check_name."},
			"catalog_kv":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Numeric value stored at the tagged Consul key in the key/value catalog. Non-numeric keys are omitted."},
		},
		Tags: map[string]interface{}{
			"host":         inputs.NewTagInfo("Host name associated with the scraped Consul exporter target."),
			"check":        inputs.NewTagInfo("Consul health check label emitted by consul_exporter."),
			"check_id":     inputs.NewTagInfo("Consul health check ID."),
			"check_name":   inputs.NewTagInfo("Consul health check name."),
			"node":         inputs.NewTagInfo("Consul node name."),
			"tag":          inputs.NewTagInfo("Consul service tag value."),
			"key":          inputs.NewTagInfo("Consul KV key."),
			"service_id":   inputs.NewTagInfo("Consul service ID."),
			"service_name": inputs.NewTagInfo("Consul service name."),
			"status":       inputs.NewTagInfo("Consul health status label, such as critical, maintenance, passing, or warning."),
			"member":       inputs.NewTagInfo("Consul Serf member name."),
			"instance":     inputs.NewTagInfo("Scraped exporter instance endpoint."),
		},
	}
}

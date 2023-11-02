// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package consul

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

// Info from github.com/prometheus/consul_exporter.
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "consul",
		Type: "metric",
		//nolint:lll
		Fields: map[string]interface{}{
			"up":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Was the last query of Consul successful."},
			"raft_peers":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "How many peers (servers) are in the Raft cluster."},
			"raft_leader":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Does Raft cluster have a leader (according to this node)."},
			"serf_lan_members":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "How many members are in the cluster."},
			"serf_lan_member_status":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Status of member in the cluster. 1=Alive, 2=Leaving, 3=Left, 4=Failed."},
			"serf_wan_member_status":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "SStatus of member in the wan cluster. 1=Alive, 2=Leaving, 3=Left, 4=Failed."},
			"catalog_services":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "How many services are in the cluster."},
			"service_tag":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tags of a service."},
			"catalog_service_node_healthy": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Is this service healthy on this node?"},
			"health_node_status":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Status of health checks associated with a node."},
			"health_service_status":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Status of health checks associated with a service."},
			"service_checks":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Link the service id and check name if available."},
			"catalog_kv":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "The values for selected keys in Consul's key/value catalog. Keys with non-numeric values are omitted."},
		},
		Tags: map[string]interface{}{
			"host":         inputs.NewTagInfo("Host name."),
			"check":        inputs.NewTagInfo("Check."),
			"check_id":     inputs.NewTagInfo("Check id."),
			"check_name":   inputs.NewTagInfo("Check name."),
			"node":         inputs.NewTagInfo("Node name."),
			"tag":          inputs.NewTagInfo("Tag."),
			"key":          inputs.NewTagInfo("Key."),
			"service_id":   inputs.NewTagInfo("Service id."),
			"service_name": inputs.NewTagInfo("Service name."),
			"status":       inputs.NewTagInfo("Status: critical, maintenance, passing, warning."),
			"member":       inputs.NewTagInfo("Member name."),
			"instance":     inputs.NewTagInfo("Instance endpoint."),
		},
	}
}

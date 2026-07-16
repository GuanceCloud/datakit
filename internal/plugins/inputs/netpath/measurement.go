// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type netpathMeasurement struct{}

//nolint:lll
func (*netpathMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricName,
		Cat:    point.Network,
		Desc:   "Network path probe results from static targets or dynamic local traffic candidates, including TCP/UDP/ICMP latency and traceroute hops.",
		DescZh: "网络路径探测结果，来源于静态目标或本机流量动态候选目标，包含 TCP/UDP/ICMP 时延和 traceroute 跳点信息。",
		Tags: map[string]interface{}{
			"task_name":            &inputs.TagInfo{Desc: "Probe task name."},
			"task_source":          &inputs.TagInfo{Desc: "Task source: local, server, or dynamic."},
			"origin":               &inputs.TagInfo{Desc: "Original source name, such as `config` or `ebpf_netflow`."},
			"run_type":             &inputs.TagInfo{Desc: "Run type: scheduled, on_demand, or dynamic."},
			"protocol":             &inputs.TagInfo{Desc: "Probe protocol."},
			"src_ip":               &inputs.TagInfo{Desc: "Source IP observed by the traffic source, aligned with netflow."},
			"src_port":             &inputs.TagInfo{Desc: "Source port observed by the traffic source; * when unavailable, aligned with netflow."},
			"dst_ip":               &inputs.TagInfo{Desc: "Original destination IP observed by the traffic source, aligned with netflow."},
			"dst_port":             &inputs.TagInfo{Desc: "Original destination port; * when unavailable, aligned with netflow."},
			"dst_domain":           &inputs.TagInfo{Desc: "Configured or discovered destination domain."},
			"dst_nat_ip":           &inputs.TagInfo{Desc: "Translated destination IP used by the probe when DNAT is present, aligned with netflow."},
			"dst_nat_port":         &inputs.TagInfo{Desc: "Translated destination port used by the probe when DNAT is present, aligned with netflow."},
			"traceroute_protocol":  &inputs.TagInfo{Desc: "Protocol used to generate traceroute probes."},
			"traceroute_status":    &inputs.TagInfo{Desc: "Traceroute status, such as reached, partial, or failed."},
			"traceroute_fail_type": &inputs.TagInfo{Desc: "Normalized traceroute failure type, such as timeout, dns_error, permission, protocol_unsupported, target_unreachable, or runner_error."},
			"e2e_status":           &inputs.TagInfo{Desc: "End-to-end status: reached, partial, unknown, or failed."},
			"namespace":            &inputs.TagInfo{Desc: "Source namespace."},
			"source_container_id":  &inputs.TagInfo{Desc: "Source container ID."},
			"source_host":          &inputs.TagInfo{Desc: "Source host that discovered the candidate."},
			"source_process":       &inputs.TagInfo{Desc: "Source process name."},
			"source_service":       &inputs.TagInfo{Desc: "Source service name."},
			"src_cloud_provider":   &inputs.TagInfo{Desc: "Cloud provider associated with the source endpoint when available."},
			"dst_cloud_provider":   &inputs.TagInfo{Desc: "Cloud provider associated with the destination endpoint when available."},
			"probe_source_ip":      &inputs.TagInfo{Desc: "Source IP selected by the route used for the active probe."},
			"probe_gateway_ip":     &inputs.TagInfo{Desc: "Next-hop gateway selected by the route used for the active probe."},
			"probe_interface":      &inputs.TagInfo{Desc: "Outbound interface selected by the route used for the active probe."},
			"probe_interface_mac":  &inputs.TagInfo{Desc: "MAC address of the outbound interface used for the active probe."},
			"probe_netns":          &inputs.TagInfo{Desc: "Network namespace of the active probe."},
			"netns":                &inputs.TagInfo{Desc: "Linux network namespace."},
		},
		Fields: map[string]interface{}{
			"e2e_dest_ip": &inputs.FieldInfo{
				DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
				Desc: "IPv4 address used by the end-to-end probes.",
			},
			"e2e_queries": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Configured number of independent end-to-end probes.",
			},
			"e2e_packets_sent": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of end-to-end probes sent.",
			},
			"e2e_packets_received": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of recognizable destination responses received.",
			},
			"e2e_unknown": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of probes with ambiguous outcomes, such as a silent UDP application.",
			},
			"e2e_probe_loss_percent": &inputs.FieldInfo{
				DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent,
				Desc: "Percentage of determinate probes without a recognizable response; ambiguous UDP silence is excluded.",
			},
			"e2e_rtt_avg": &inputs.FieldInfo{
				DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationUS,
				Desc: "Average end-to-end round-trip time across received responses.",
			},
			"e2e_rtt_min": &inputs.FieldInfo{
				DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationUS,
				Desc: "Minimum end-to-end round-trip time.",
			},
			"e2e_rtt_max": &inputs.FieldInfo{
				DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationUS,
				Desc: "Maximum end-to-end round-trip time.",
			},
			"e2e_rtt_variation_samples": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of consecutive successful probe pairs used for RTT variation.",
			},
			"e2e_rtt_variation_avg": &inputs.FieldInfo{
				DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationUS,
				Desc: "Average absolute RTT difference between consecutive successful probes.",
			},
			"e2e_rtt_variation_max": &inputs.FieldInfo{
				DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationUS,
				Desc: "Maximum absolute RTT difference between consecutive successful probes.",
			},
			"e2e_tcp_connection_refused": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of TCP probes answered with connection refusal; these still prove endpoint reachability.",
			},
			"e2e_fail_reason": &inputs.FieldInfo{
				DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
				Desc: "End-to-end probe setup or execution failure reason.",
			},
			"source_pid": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "Source process ID reported by the candidate source.",
			},
			"test_run_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "DataKit-generated probe run ID.",
			},
			"dst_reverse_dns": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Reverse DNS name of destination IP when reverse_dns is enabled.",
			},
			"duration": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "Probe execution duration.",
			},
			"scheduled_at": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.TimestampUS,
				Desc:     "Scheduled execution time in Unix microseconds.",
			},
			"started_at": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.TimestampUS,
				Desc:     "Probe start time in Unix microseconds.",
			},
			"hop_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of traceroute hops.",
			},
			"traceroute_fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Traceroute setup or execution failure reason.",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Normalized traceroute JSON, or a failure message when no path is available.",
			},
			"max_ttl": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Effective traceroute max TTL after protocol limits.",
			},
			"traceroute_queries": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Configured complete traceroute run count.",
			},
		},
	}
}

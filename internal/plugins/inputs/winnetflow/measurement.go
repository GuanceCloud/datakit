// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package winnetflow

import (
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type netflowMeasurement struct{}

//nolint:lll
func (*netflowMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricName,
		Cat:    point.Network,
		Desc:   "Per-process L4 network flow metrics collected from Windows ETW, schema compatible with ebpf-net/netflow.",
		DescZh: "通过 Windows ETW 采集的按进程四层网络流量指标，字段与 ebpf-net/netflow 兼容。",
		Tags: map[string]interface{}{
			"family":         inputs.NewTagInfo("IP family: IPv4 or IPv6."),
			"direction":      inputs.NewTagInfo("Flow direction: incoming or outgoing."),
			"transport":      inputs.NewTagInfo("Transport protocol: tcp or udp."),
			"src_ip":         inputs.NewTagInfo("Source IP address."),
			"dst_ip":         inputs.NewTagInfo("Destination IP address."),
			"src_port":       inputs.NewTagInfo("Source port."),
			"dst_port":       inputs.NewTagInfo("Destination port."),
			"pid":            inputs.NewTagInfo("Process ID owning the connection."),
			"process_name":   inputs.NewTagInfo("Process name owning the connection."),
			"src_ip_type":    inputs.NewTagInfo("Source IP type: private, loopback, multicast or other."),
			"dst_ip_type":    inputs.NewTagInfo("Destination IP type: private, loopback, multicast or other."),
			"conn_side":      inputs.NewTagInfo("Local connection role: client or server."),
			"client_ip":      inputs.NewTagInfo("Client endpoint IP address."),
			"client_port":    inputs.NewTagInfo("Client endpoint port."),
			"client_ip_type": inputs.NewTagInfo("Client IP type: private, loopback, multicast or other."),
			"server_ip":      inputs.NewTagInfo("Server endpoint IP address."),
			"server_port":    inputs.NewTagInfo("Server endpoint port."),
			"server_ip_type": inputs.NewTagInfo("Server IP type: private, loopback, multicast or other."),
			"dst_nat_ip":     inputs.NewTagInfo("Destination NAT IP; N/A because Windows ETW does not expose NAT translation."),
			"dst_nat_port":   inputs.NewTagInfo("Destination NAT port; N/A because Windows ETW does not expose NAT translation."),
		},
		Fields: map[string]interface{}{
			// Values are per-interval sums (the aggregator resets each flush),
			// so they are gauges, not monotonic counters.
			"bytes_read":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Bytes read on this flow during the interval."},
			"bytes_written":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Bytes written on this flow during the interval."},
			"packets_read":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Packets read on this flow during the interval (approximate on Windows)."},
			"packets_written":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Packets/messages written during the interval; always 0 for TCP because TCPIP send events do not expose a packet count, and approximated by message count for UDP."},
			"client_sent":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Bytes sent by the client during the interval."},
			"server_sent":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Bytes sent by the server during the interval."},
			"retransmits":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "TCP retransmissions on this flow during the interval."},
			"rtt":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Average smoothed TCP RTT in microseconds."},
			"rtt_var":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Average TCP RTT variance in microseconds."},
			"tcp_closed":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "TCP connections closed during the interval."},
			"tcp_established":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "TCP connections established during the interval."},
			"tcp_connect_attempts": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "TCP active connect attempts during the interval."},
			"tcp_connect_failures": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "TCP active connect failures during the interval."},
			"tcp_close_wait":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "TCP transitions into CLOSE_WAIT state during the interval."},
			"tcp_last_ack":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "TCP transitions into LAST_ACK state during the interval."},
			"tcp_time_wait":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "TCP transitions into TIME_WAIT state during the interval."},
		},
	}
}

type httpflowMeasurement struct{}

//nolint:lll
func (*httpflowMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   httpflowMetricName,
		Cat:    point.Network,
		Desc:   "Aggregated HTTP request metrics collected from the Windows HTTP.sys ETW provider, schema compatible with ebpf-net/httpflow.",
		DescZh: "通过 Windows HTTP.sys ETW Provider 采集的 HTTP 请求指标，字段与 ebpf-net/httpflow 兼容。",
		Tags: map[string]interface{}{
			"family":         inputs.NewTagInfo("IP family: IPv4 or IPv6."),
			"direction":      inputs.NewTagInfo("Flow direction, always incoming for the HTTP.sys server side."),
			"transport":      inputs.NewTagInfo("Transport protocol, always tcp."),
			"src_ip":         inputs.NewTagInfo("Server IP address (local endpoint)."),
			"dst_ip":         inputs.NewTagInfo("Client IP address (remote endpoint)."),
			"src_port":       inputs.NewTagInfo("Server port."),
			"dst_port":       inputs.NewTagInfo("Client port."),
			"pid":            inputs.NewTagInfo("Server process ID reported by HTTP.sys on the response event."),
			"process_name":   inputs.NewTagInfo("Server process name resolved from the HTTP.sys response event PID (for example, w3wp.exe)."),
			"src_ip_type":    inputs.NewTagInfo("Server source IP type: private, loopback, multicast or other."),
			"dst_ip_type":    inputs.NewTagInfo("Client destination IP type: private, loopback, multicast or other."),
			"conn_side":      inputs.NewTagInfo("Local connection role, always server for HTTP.sys."),
			"client_ip":      inputs.NewTagInfo("Client endpoint IP address."),
			"client_port":    inputs.NewTagInfo("Client endpoint port."),
			"client_ip_type": inputs.NewTagInfo("Client IP type: private, loopback, multicast or other."),
			"server_ip":      inputs.NewTagInfo("Server endpoint IP address."),
			"server_port":    inputs.NewTagInfo("Server endpoint port."),
			"server_ip_type": inputs.NewTagInfo("Server IP type: private, loopback, multicast or other."),
			"dst_nat_ip":     inputs.NewTagInfo("Destination NAT IP; N/A because Windows ETW does not expose NAT translation."),
			"dst_nat_port":   inputs.NewTagInfo("Destination NAT port; N/A because Windows ETW does not expose NAT translation."),
		},
		Fields: map[string]interface{}{
			// Values are per-interval aggregates; non-monotonic by design.
			"method":        &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "HTTP method (GET/POST/...)."},
			"http_version":  &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "HTTP version; empty because HTTP.sys events do not carry it."},
			"path":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Request path."},
			"status_code":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "HTTP response status code."},
			"latency":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS, Desc: "Average request processing time (receive to send complete) in nanoseconds."},
			"bytes_read":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Bytes read on this flow during the interval (always 0; HTTP.sys does not expose request body sizes)."},
			"bytes_written": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Bytes written on this flow during the interval (populated for cache-served responses)."},
			"client_sent":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Bytes sent by the HTTP client; always 0 because HTTP.sys does not expose request body sizes."},
			"server_sent":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Bytes sent by the HTTP server during the interval."},
			"truncated":     &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Request path reached the configured length limit and was truncated."},
			"count":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of HTTP requests in this group during the interval."},
		},
	}
}

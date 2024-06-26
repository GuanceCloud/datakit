// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ebpf

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type ConnStatsM struct{}

//nolint:lll
func (m *ConnStatsM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "netflow",
		Type: point.Network.String(),
		Tags: map[string]interface{}{
			"host":       inputs.TagInfo{Desc: "System hostname"},
			"dst_ip":     inputs.TagInfo{Desc: "Destination IP address"},
			"dst_domain": inputs.TagInfo{Desc: "Destination domain"},
			"dst_port":   inputs.TagInfo{Desc: "Destination port"},
			"dst_nat_ip": inputs.TagInfo{Desc: "For data containing the `outging` tag, " +
				"this value is the ip after the DNAT operation"},
			"dst_nat_port": inputs.TagInfo{Desc: "For data containing the `outging` tag, " +
				"this value is the port after the DNAT operation"},
			"dst_ip_type":             inputs.TagInfo{Desc: "Destination IP type. (other/private/multicast)"},
			"src_ip":                  inputs.TagInfo{Desc: "Source IP"},
			"src_port":                inputs.TagInfo{Desc: "Source port"},
			"src_ip_type":             inputs.TagInfo{Desc: "Source IP type. (other/private/multicast)"},
			"src_k8s_pod_name":        inputs.TagInfo{Desc: "Source K8s pod name"},
			"src_k8s_deployment_name": inputs.TagInfo{Desc: "Source K8s deployment name"},
			"src_k8s_service_name":    inputs.TagInfo{Desc: "Source K8s service name"},
			"src_k8s_namespace":       inputs.TagInfo{Desc: "Source K8s namespace"},
			"dst_k8s_pod_name":        inputs.TagInfo{Desc: "Destination K8s pod name"},
			"dst_k8s_deployment_name": inputs.TagInfo{Desc: "Destination K8s deployment name"},
			"dst_k8s_service_name":    inputs.TagInfo{Desc: "Destination K8s service name"},
			"dst_k8s_namespace":       inputs.TagInfo{Desc: "Destination K8s namespace"},
			"pid":                     inputs.TagInfo{Desc: "Process identification number"},
			"process_name":            inputs.TagInfo{Desc: "Process name"},
			"transport":               inputs.TagInfo{Desc: "Transport layer protocol. (udp/tcp)"},
			"family":                  inputs.TagInfo{Desc: "Network layer protocol. (IPv4/IPv6)"},
			"direction":               inputs.TagInfo{Desc: "Use the source (src_ip:src_port) as a frame of reference to identify the connection initiator. (incoming/outgoing)"},
			"source":                  inputs.TagInfo{Desc: "Fixed value: `netflow`."},
			"sub_source":              inputs.TagInfo{Desc: "Some specific connection classifications, such as the sub_source value for Kubernetes network traffic is K8s"},
		},
		Fields: map[string]interface{}{
			"bytes_read":      newFInfInt("The number of bytes read", inputs.SizeByte),
			"bytes_written":   newFInfInt("The number of bytes written", inputs.SizeByte),
			"retransmits":     newFInfInt("The number of retransmissions", inputs.NCount),
			"rtt":             newFInfInt("TCP Latency", inputs.DurationUS),
			"rtt_var":         newFInfInt("TCP Jitter", inputs.DurationUS),
			"tcp_closed":      newFInfInt("The number of TCP connection closed", inputs.NCount),
			"tcp_established": newFInfInt("The number of TCP connection established", inputs.NCount),
		},
	}
}

type HTTPFlowM struct{}

//nolint:lll
func (m *HTTPFlowM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "httpflow",
		Type: point.Network.String(),
		Tags: map[string]interface{}{
			"host":        inputs.TagInfo{Desc: "System hostname"},
			"dst_ip":      inputs.TagInfo{Desc: "Destination IP address"},
			"dst_domain":  inputs.TagInfo{Desc: "Destination domain"},
			"dst_port":    inputs.TagInfo{Desc: "Destination port"},
			"dst_ip_type": inputs.TagInfo{Desc: "Destination IP type. (other/private/multicast)"},
			"dst_nat_ip": inputs.TagInfo{Desc: "For data containing the `outging` tag, " +
				"this value is the ip after the DNAT operation"},
			"dst_nat_port": inputs.TagInfo{Desc: "For data containing the `outging` tag, " +
				"this value is the port after the DNAT operation"},
			"src_ip":                  inputs.TagInfo{Desc: "Source IP"},
			"src_port":                inputs.TagInfo{Desc: "Source port"},
			"src_ip_type":             inputs.TagInfo{Desc: "Source IP type. (other/private/multicast)"},
			"src_k8s_pod_name":        inputs.TagInfo{Desc: "Source K8s pod name"},
			"src_k8s_deployment_name": inputs.TagInfo{Desc: "Source K8s deployment name"},
			"src_k8s_service_name":    inputs.TagInfo{Desc: "Source K8s service name"},
			"src_k8s_namespace":       inputs.TagInfo{Desc: "Source K8s namespace"},
			"dst_k8s_pod_name":        inputs.TagInfo{Desc: "Destination K8s pod name"},
			"dst_k8s_deployment_name": inputs.TagInfo{Desc: "Destination K8s deployment name"},
			"dst_k8s_service_name":    inputs.TagInfo{Desc: "Destination K8s service name"},
			"dst_k8s_namespace":       inputs.TagInfo{Desc: "Destination K8s namespace"},
			"pid":                     inputs.TagInfo{Desc: "Process identification number"},
			"process_name":            inputs.TagInfo{Desc: "Process name"},
			"transport":               inputs.TagInfo{Desc: "Transport layer protocol. (udp/tcp)"},
			"family":                  inputs.TagInfo{Desc: "Network layer protocol. (IPv4/IPv6)"},
			"direction":               inputs.TagInfo{Desc: "Use the source (src_ip:src_port) as a frame of reference to identify the connection initiator. (incoming/outgoing)"},
			"source":                  inputs.TagInfo{Desc: "Fixed value: `httpflow`."},
			"sub_source":              inputs.TagInfo{Desc: "Some specific connection classifications, such as the sub_source value for Kubernetes network traffic is K8s"},
		},
		Fields: map[string]interface{}{
			"truncated":     newFInfBool("The length of the request path has reached the upper limit of the number of bytes collected, and the request path may be truncated", inputs.UnknownUnit),
			"path":          newFString("Request path"),
			"status_code":   newFInfInt("Http status codes", inputs.UnknownUnit),
			"method":        newFString("GET/POST/..."),
			"latency":       newFInfInt("TTFB", inputs.DurationNS),
			"http_version":  newFString("1.1 / 1.0 ..."),
			"count":         newFInfInt("The total number of HTTP requests in a collection cycle", inputs.UnknownUnit),
			"bytes_read":    newFInfInt("The number of bytes read", inputs.SizeByte),
			"bytes_written": newFInfInt("The number of bytes written", inputs.SizeByte),
		},
	}
}

type DNSStatsM struct{}

//nolint:lll
func (m *DNSStatsM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "dnsflow",
		Type: point.Network.String(),
		Tags: map[string]interface{}{
			"host":                    inputs.TagInfo{Desc: "System hostname"},
			"dst_ip":                  inputs.TagInfo{Desc: "Destination IP address"},
			"dst_domain":              inputs.TagInfo{Desc: "Destination domain"},
			"dst_port":                inputs.TagInfo{Desc: "Destination port"},
			"dst_ip_type":             inputs.TagInfo{Desc: "Destination IP type. (other/private/multicast)"},
			"src_ip":                  inputs.TagInfo{Desc: "Source IP"},
			"src_port":                inputs.TagInfo{Desc: "Source port"},
			"src_ip_type":             inputs.TagInfo{Desc: "Source IP type. (other/private/multicast)"},
			"src_k8s_pod_name":        inputs.TagInfo{Desc: "Source K8s pod name"},
			"src_k8s_deployment_name": inputs.TagInfo{Desc: "Source K8s deployment name"},
			"src_k8s_service_name":    inputs.TagInfo{Desc: "Source K8s service name"},
			"src_k8s_namespace":       inputs.TagInfo{Desc: "Source K8s namespace"},
			"dst_k8s_pod_name":        inputs.TagInfo{Desc: "Destination K8s pod name"},
			"dst_k8s_deployment_name": inputs.TagInfo{Desc: "Destination K8s deployment name"},
			"dst_k8s_service_name":    inputs.TagInfo{Desc: "Destination K8s service name"},
			"dst_k8s_namespace":       inputs.TagInfo{Desc: "Destination K8s namespace"},
			// "pid":                     inputs.TagInfo{Desc: "Process identification number"},
			// "process_name":            inputs.TagInfo{Desc: "Process name"},
			"transport": inputs.TagInfo{Desc: "Transport layer protocol. (udp/tcp)"},
			"family":    inputs.TagInfo{Desc: "Network layer protocol. (IPv4/IPv6)"},
			"direction": inputs.TagInfo{Desc: "Use the source (src_ip:src_port) as a frame of reference to identify the connection initiator. (incoming/outgoing)"},
			"source":    inputs.TagInfo{Desc: "Fixed value: `dnsflow`."},
			"sub_source": inputs.TagInfo{Desc: "Some specific connection classifications, " +
				"such as the sub_source value for Kubernetes network traffic is K8s"},
		},
		Fields: map[string]interface{}{
			"rcode": newFInfInt("DNS response code: 0 - `NoError`, 1 - `FormErr`, 2 - `ServFail`, "+
				"3 - NXDomain, 4 - NotImp, 5 - Refused, ...; A value of -1 means the request timed out", inputs.UnknownUnit),
			"count":       newFInfInt("The number of DNS requests in a collection cycle", inputs.UnknownUnit),
			"latency":     newFInfInt("Average response time for DNS requests", inputs.DurationNS),
			"latency_max": newFInfInt("Maximum response time for DNS requests", inputs.DurationNS),
		},
	}
}

type BashM struct{}

//nolint:lll
func (m *BashM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "bash",
		Type: point.Logging.String(),
		Tags: map[string]interface{}{
			"host": inputs.TagInfo{Desc: "host name"},
		},
		Fields: map[string]interface{}{
			"pid":     newFString("Process identification number"),
			"user":    newFString("The user who executes the bash command"),
			"cmd":     newFString("Command"),
			"message": newFString("The bash execution record generated by the collector"),
		},
	}
}

type BPFL4Log struct{}

//nolint:lll
func (bpfl4 *BPFL4Log) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Type: point.Logging.String(),
		Name: "bpf_net_l4_log",
		Tags: map[string]interface{}{
			"host":          inputs.TagInfo{Desc: "Host name"},
			"src_ip":        inputs.TagInfo{Desc: "The IP address of the collected local network interface"},
			"src_port":      inputs.TagInfo{Desc: "Local port"},
			"dst_ip":        inputs.TagInfo{Desc: "The IP address of the foreign network interface"},
			"dst_port":      inputs.TagInfo{Desc: "Foreign port"},
			"l4_proto":      inputs.TagInfo{Desc: "Transport protocol"},
			"l7_proto":      inputs.TagInfo{Desc: "Application protocol"},
			"nic_mac":       inputs.TagInfo{Desc: "MAC address of the collected network interface"},
			"nic_name":      inputs.TagInfo{Desc: "name of the collected network interface"},
			"netns":         inputs.TagInfo{Desc: "Network namespace, format: `NS(<device id>:<inode number>)`"},
			"vni_id":        inputs.TagInfo{Desc: "Virtual Network Identifier"},
			"vxlan_packet":  inputs.TagInfo{Desc: "Whether it is a VXLAN packet"},
			"inner_traceid": inputs.TagInfo{Desc: "Correlate the layer 4 and layer 7 network log data of a TCP connection on the collected network interface"},
			"host_network":  inputs.TagInfo{Desc: "Whether the network log data is collected on the host network"},
			"virtual_nic":   inputs.TagInfo{Desc: "Whether the network log data is collected on the virtual network interface"},
			"direction":     inputs.TagInfo{Desc: "Use the source (src_ip:src_port) as a frame of reference to identify the connection initiator. (incoming/outgoing)"},

			"k8s_namespace":      inputs.TagInfo{Desc: "Kubernetes namespace"},
			"k8s_pod_name":       inputs.TagInfo{Desc: "Kubernetes pod name"},
			"k8s_container_name": inputs.TagInfo{Desc: "Kubernetes container name"},

			"sub_source":              inputs.TagInfo{Desc: "Some specific connection classifications, such as the sub_source value for Kubernetes network traffic is K8s"},
			"src_k8s_pod_name":        inputs.TagInfo{Desc: "Source K8s pod name"},
			"src_k8s_deployment_name": inputs.TagInfo{Desc: "Source K8s deployment name"},
			"src_k8s_service_name":    inputs.TagInfo{Desc: "Source K8s service name"},
			"src_k8s_namespace":       inputs.TagInfo{Desc: "Source K8s namespace"},
			"dst_k8s_pod_name":        inputs.TagInfo{Desc: "Destination K8s pod name"},
			"dst_k8s_deployment_name": inputs.TagInfo{Desc: "Destination K8s deployment name"},
			"dst_k8s_service_name":    inputs.TagInfo{Desc: "Destination K8s service name"},
			"dst_k8s_namespace":       inputs.TagInfo{Desc: "Destination K8s namespace"},
		},
		Fields: map[string]interface{}{
			"chunk_id":        newFInfInt("A connection may be divided into several chunks for upload based on time interval or TCP segment number", inputs.UnknownUnit),
			"tx_seq_min":      newFInfInt("The minimum value of the TCP sequence number of the data packet sent by the network interface, which is a 32-bit unsigned integer", inputs.UnknownUnit),
			"tx_seq_max":      newFInfInt("The maximum value of the TCP sequence number of the data packet sent by the network interface, which is a 32-bit unsigned integer", inputs.UnknownUnit),
			"rx_seq_min":      newFInfInt("The minimum value of the TCP sequence number of the data packet received by the network interface, which is a 32-bit unsigned integer", inputs.UnknownUnit),
			"rx_seq_max":      newFInfInt("The maximum value of the TCP sequence number of the data packet received by the network interface, which is a 32-bit unsigned integer", inputs.UnknownUnit),
			"tx_bytes":        newFInfInt("The number of bytes sent by the network interface", inputs.SizeByte),
			"rx_bytes":        newFInfInt("The number of bytes received by the network interface", inputs.SizeByte),
			"tx_packets":      newFInfInt("The number of packets sent by the network interface", inputs.UnknownUnit),
			"rx_packets":      newFInfInt("The number of packets received by the network interface", inputs.UnknownUnit),
			"tx_retrans":      newFInfInt("The number of retransmitted packets sent by the network interface", inputs.UnknownUnit),
			"rx_retrans":      newFInfInt("The number of retransmitted packets received by the network interface", inputs.UnknownUnit),
			"tcp_syn_retrans": newFInfInt("The number of retransmitted SYN packets sent by the network interface", inputs.UnknownUnit),
		},
	}
}

type BPFL7Log struct{}

//nolint:lll
func (bpfl7 *BPFL7Log) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Type: point.Logging.String(),
		Name: "bpf_net_l7_log",
		Tags: map[string]interface{}{
			"host":          inputs.TagInfo{Desc: "Host name"},
			"src_ip":        inputs.TagInfo{Desc: "The IP address of the collected local network interface"},
			"src_port":      inputs.TagInfo{Desc: "Local port"},
			"dst_ip":        inputs.TagInfo{Desc: "The IP address of the foreign network interface"},
			"dst_port":      inputs.TagInfo{Desc: "Foreign port"},
			"l4_proto":      inputs.TagInfo{Desc: "Transport protocol"},
			"l7_proto":      inputs.TagInfo{Desc: "Application protocol"},
			"nic_mac":       inputs.TagInfo{Desc: "MAC address of the collected network interface"},
			"nic_name":      inputs.TagInfo{Desc: "name of the collected network interface"},
			"netns":         inputs.TagInfo{Desc: "Network namespace, format: `NS(<device id>:<inode number>)`"},
			"vni_id":        inputs.TagInfo{Desc: "Virtual Network Identifier"},
			"vxlan_packet":  inputs.TagInfo{Desc: "Whether it is a VXLAN packet"},
			"inner_traceid": inputs.TagInfo{Desc: "Correlate the layer 4 and layer 7 network log data of a TCP connection on the collected network interface"},
			"host_network":  inputs.TagInfo{Desc: "Whether the network log data is collected on the host network"},
			"virtual_nic":   inputs.TagInfo{Desc: "Whether the network log data is collected on the virtual network interface"},
			"direction":     inputs.TagInfo{Desc: "Use the source (src_ip:src_port) as a frame of reference to identify the connection initiator. (incoming/outgoing)"},

			"k8s_namespace":      inputs.TagInfo{Desc: "Kubernetes namespace"},
			"k8s_pod_name":       inputs.TagInfo{Desc: "Kubernetes pod name"},
			"k8s_container_name": inputs.TagInfo{Desc: "Kubernetes container name"},

			"sub_source":              inputs.TagInfo{Desc: "Some specific connection classifications, such as the sub_source value for Kubernetes network traffic is K8s"},
			"src_k8s_pod_name":        inputs.TagInfo{Desc: "Source K8s pod name"},
			"src_k8s_deployment_name": inputs.TagInfo{Desc: "Source K8s deployment name"},
			"src_k8s_service_name":    inputs.TagInfo{Desc: "Source K8s service name"},
			"src_k8s_namespace":       inputs.TagInfo{Desc: "Source K8s namespace"},
			"dst_k8s_pod_name":        inputs.TagInfo{Desc: "Destination K8s pod name"},
			"dst_k8s_deployment_name": inputs.TagInfo{Desc: "Destination K8s deployment name"},
			"dst_k8s_service_name":    inputs.TagInfo{Desc: "Destination K8s service name"},
			"dst_k8s_namespace":       inputs.TagInfo{Desc: "Destination K8s namespace"},

			"trace_id":  inputs.TagInfo{Desc: "APM trace id"},
			"parent_id": inputs.TagInfo{Desc: "The span id of the APM span corresponding to this network request"},

			"l7_traceid": inputs.TagInfo{Desc: "Correlate the layer 7 network log data of a TCP connection on the all collected network interface"},
		},
		Fields: map[string]interface{}{
			"tx_seq": newFInfInt("The tcp sequence number of the request/response first byte sent by the network interface", inputs.UnknownUnit),
			"rx_seq": newFInfInt("The tcp sequence number of the request/response first byte received by the network interface", inputs.UnknownUnit),

			"http_status_code": newFInfInt("HTTP status code", inputs.UnknownUnit),
			"http_method":      newFString("HTTP method"),
			"http_path":        newFString("HTTP path"),
		},
	}
}

type EBPFTrace struct{}

func (bpftrace *EBPFTrace) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Type: point.Tracing.String(),
		Name: "dketrace",
		Tags: map[string]any{
			"service": inputs.TagInfo{Desc: "Service name"},

			"host":     inputs.TagInfo{Desc: "System hostname"},
			"dst_ip":   inputs.TagInfo{Desc: "Destination IP address"},
			"dst_port": inputs.TagInfo{Desc: "Destination port"},
			"src_ip":   inputs.TagInfo{Desc: "Source IP"},
			"src_port": inputs.TagInfo{Desc: "Source port"},
		},
		Fields: map[string]any{
			"span_id":   newFString("APM span id, generated by the `ebpftrace` collector"),
			"trace_id":  newFString("APM trace id, can choose between existing app trace id and ebpf generation,set by the `ebpftrace` collector"),
			"parent_id": newFString("APM parent span id, set by the `ebpftrace` collector"),

			"ebpf_trace_id":  newFString("eBPF trace id, generated by the `ebpftrace` collector"),
			"ebpf_parent_id": newFString("eBPF parent span id, generated by the `ebpftrace` collector"),
			"app_trace_id":   newFString("Trace id carried by the application in the request"),
			"app_parent_id":  newFString("Parent span id carried by the application in the request"),

			"source_type": newFString("Source type, value is `ebpf`"),

			"pid":          newFString("Process identification number"),
			"process_name": newFString("Process name"),
			"thread_name":  newFString("Thread name"),

			"duration": newFInfInt("Duration", inputs.DurationUS),
			"start":    newFInfInt("Start time", inputs.TimestampUS),

			"span_type": newFString("Span type"),

			"bytes_written": newFInfInt("Bytes written", inputs.SizeByte),
			"bytes_read":    newFInfInt("Bytes read", inputs.SizeByte),

			"http_route":       newFString("HTTP route"),
			"http_method":      newFString("HTTP method"),
			"http_status_code": newFString("HTTP status code"),

			"grpc_status_code": newFString("gRPC status code"),

			"mysql_status_code": newFInfInt("MySQL request status code", inputs.UnknownUnit),
			"mysql_err_msg":     newFString("MySQL error message"),
			"resource_type":     newFString("Redis resource type"),
			"err_msg":           newFString("Redis error message"),
			"status_msg":        newFString("Redis status message"),
			"operation":         newFString("Operation"),
			"status":            newFString("Status"),
		},
	}
}

func newFString(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.UnknownType,
		DataType: inputs.String,
		Unit:     inputs.UnknownUnit,
		Desc:     desc,
	}
}

func newFInfBool(desc, unit string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Bool,
		Unit:     unit,
		Desc:     desc,
	}
}

func newFInfInt(desc, unit string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     unit,
		Desc:     desc,
	}
}

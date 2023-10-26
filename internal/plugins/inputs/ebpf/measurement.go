// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ebpf

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type measurement struct{}

//nolint:lll
func (m *measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}

type ConnStatsM measurement

//nolint:lll
func (m *ConnStatsM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "netflow",
		Tags: map[string]interface{}{
			"host":       inputs.TagInfo{Desc: "System hostname."},
			"dst_ip":     inputs.TagInfo{Desc: "Destination IP address."},
			"dst_domain": inputs.TagInfo{Desc: "Destination domain."},
			"dst_port":   inputs.TagInfo{Desc: "Destination port."},
			"dst_nat_ip": inputs.TagInfo{Desc: "For data containing the `outging` tag, " +
				"this value is the ip after the DNAT operation."},
			"dst_nat_port": inputs.TagInfo{Desc: "For data containing the `outging` tag, " +
				"this value is the port after the DNAT operation."},
			"dst_ip_type":             inputs.TagInfo{Desc: "Destination IP type. (other/private/multicast)"},
			"src_ip":                  inputs.TagInfo{Desc: "Source IP."},
			"src_port":                inputs.TagInfo{Desc: "Source port."},
			"src_ip_type":             inputs.TagInfo{Desc: "Source IP type. (other/private/multicast)"},
			"src_k8s_pod_name":        inputs.TagInfo{Desc: "Source K8s pod name."},
			"src_k8s_deployment_name": inputs.TagInfo{Desc: "Source K8s deployment name."},
			"src_k8s_service_name":    inputs.TagInfo{Desc: "Source K8s service name."},
			"src_k8s_namespace":       inputs.TagInfo{Desc: "Source K8s namespace."},
			"dst_k8s_pod_name":        inputs.TagInfo{Desc: "Destination K8s pod name."},
			"dst_k8s_deployment_name": inputs.TagInfo{Desc: "Destination K8s deployment name."},
			"dst_k8s_service_name":    inputs.TagInfo{Desc: "Destination K8s service name."},
			"dst_k8s_namespace":       inputs.TagInfo{Desc: "Destination K8s namespace."},
			"pid":                     inputs.TagInfo{Desc: "Process identification number."},
			"process_name":            inputs.TagInfo{Desc: "Process name."},
			"transport":               inputs.TagInfo{Desc: "Transport layer protocol. (udp/tcp)"},
			"family":                  inputs.TagInfo{Desc: "Network layer protocol. (IPv4/IPv6)"},
			"direction":               inputs.TagInfo{Desc: "Use the source as a frame of reference to identify the connection initiator. (incoming/outgoing)"},
			"source":                  inputs.TagInfo{Desc: "Fixed value: `netflow`."},
			"sub_source":              inputs.TagInfo{Desc: "Some specific connection classifications, such as the sub_source value for Kubernetes network traffic is K8s."},
		},
		Fields: map[string]interface{}{
			"bytes_read":      newFInfInt("The number of bytes read.", inputs.SizeByte),
			"bytes_written":   newFInfInt("The number of bytes written.", inputs.SizeByte),
			"retransmits":     newFInfInt("The number of retransmissions.", inputs.NCount),
			"rtt":             newFInfInt("TCP Latency.", inputs.DurationUS),
			"rtt_var":         newFInfInt("TCP Jitter.", inputs.DurationUS),
			"tcp_closed":      newFInfInt("The number of TCP connection closed.", inputs.NCount),
			"tcp_established": newFInfInt("The number of TCP connection established.", inputs.NCount),
		},
	}
}

type HTTPFlowM measurement

//nolint:lll
func (m *HTTPFlowM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "httpflow",
		Tags: map[string]interface{}{
			"host":        inputs.TagInfo{Desc: "System hostname."},
			"dst_ip":      inputs.TagInfo{Desc: "Destination IP address."},
			"dst_domain":  inputs.TagInfo{Desc: "Destination domain."},
			"dst_port":    inputs.TagInfo{Desc: "Destination port."},
			"dst_ip_type": inputs.TagInfo{Desc: "Destination IP type. (other/private/multicast)"},
			"dst_nat_ip": inputs.TagInfo{Desc: "For data containing the `outging` tag, " +
				"this value is the ip after the DNAT operation."},
			"dst_nat_port": inputs.TagInfo{Desc: "For data containing the `outging` tag, " +
				"this value is the port after the DNAT operation."},
			"src_ip":                  inputs.TagInfo{Desc: "Source IP."},
			"src_port":                inputs.TagInfo{Desc: "Source port."},
			"src_ip_type":             inputs.TagInfo{Desc: "Source IP type. (other/private/multicast)"},
			"src_k8s_pod_name":        inputs.TagInfo{Desc: "Source K8s pod name."},
			"src_k8s_deployment_name": inputs.TagInfo{Desc: "Source K8s deployment name."},
			"src_k8s_service_name":    inputs.TagInfo{Desc: "Source K8s service name."},
			"src_k8s_namespace":       inputs.TagInfo{Desc: "Source K8s namespace."},
			"dst_k8s_pod_name":        inputs.TagInfo{Desc: "Destination K8s pod name."},
			"dst_k8s_deployment_name": inputs.TagInfo{Desc: "Destination K8s deployment name."},
			"dst_k8s_service_name":    inputs.TagInfo{Desc: "Destination K8s service name."},
			"dst_k8s_namespace":       inputs.TagInfo{Desc: "Destination K8s namespace."},
			"pid":                     inputs.TagInfo{Desc: "Process identification number."},
			"process_name":            inputs.TagInfo{Desc: "Process name."},
			"transport":               inputs.TagInfo{Desc: "Transport layer protocol. (udp/tcp)"},
			"family":                  inputs.TagInfo{Desc: "Network layer protocol. (IPv4/IPv6)"},
			"direction":               inputs.TagInfo{Desc: "Use the source as a frame of reference to identify the connection initiator. (incoming/outgoing)"},
			"source":                  inputs.TagInfo{Desc: "Fixed value: `httpflow`."},
			"sub_source":              inputs.TagInfo{Desc: "Some specific connection classifications, such as the sub_source value for Kubernetes network traffic is K8s."},
		},
		Fields: map[string]interface{}{
			"truncated":     newFInfBool("The length of the request path has reached the upper limit of the number of bytes collected, and the request path may be truncated.", inputs.UnknownUnit),
			"path":          newFString("Request path."),
			"status_code":   newFInfInt("Http status codes.", inputs.UnknownUnit),
			"method":        newFString("GET/POST/..."),
			"latency":       newFInfInt("TTFB.", inputs.DurationNS),
			"http_version":  newFString("1.1 / 1.0 ..."),
			"count":         newFInfInt("The total number of HTTP requests in a collection cycle.", inputs.UnknownUnit),
			"bytes_read":    newFInfInt("The number of bytes read.", inputs.SizeByte),
			"bytes_written": newFInfInt("The number of bytes written.", inputs.SizeByte),
		},
	}
}

type DNSStatsM measurement

func (m *DNSStatsM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "dnsflow",
		Tags: map[string]interface{}{
			"host":                    inputs.TagInfo{Desc: "System hostname."},
			"dst_ip":                  inputs.TagInfo{Desc: "Destination IP address."},
			"dst_domain":              inputs.TagInfo{Desc: "Destination domain."},
			"dst_port":                inputs.TagInfo{Desc: "Destination port."},
			"dst_ip_type":             inputs.TagInfo{Desc: "Destination IP type. (other/private/multicast)"},
			"src_ip":                  inputs.TagInfo{Desc: "Source IP."},
			"src_port":                inputs.TagInfo{Desc: "Source port."},
			"src_ip_type":             inputs.TagInfo{Desc: "Source IP type. (other/private/multicast)"},
			"src_k8s_pod_name":        inputs.TagInfo{Desc: "Source K8s pod name."},
			"src_k8s_deployment_name": inputs.TagInfo{Desc: "Source K8s deployment name."},
			"src_k8s_service_name":    inputs.TagInfo{Desc: "Source K8s service name."},
			"src_k8s_namespace":       inputs.TagInfo{Desc: "Source K8s namespace."},
			"dst_k8s_pod_name":        inputs.TagInfo{Desc: "Destination K8s pod name."},
			"dst_k8s_deployment_name": inputs.TagInfo{Desc: "Destination K8s deployment name."},
			"dst_k8s_service_name":    inputs.TagInfo{Desc: "Destination K8s service name."},
			"dst_k8s_namespace":       inputs.TagInfo{Desc: "Destination K8s namespace."},
			// "pid":                     inputs.TagInfo{Desc: "Process identification number."},
			// "process_name":            inputs.TagInfo{Desc: "Process name."},
			"transport": inputs.TagInfo{Desc: "Transport layer protocol. (udp/tcp)"},
			"family":    inputs.TagInfo{Desc: "Network layer protocol. (IPv4/IPv6)"},
			"direction": inputs.TagInfo{Desc: "Use the source as a frame of reference to identify the connection initiator. (incoming/outgoing)"},
			"source":    inputs.TagInfo{Desc: "Fixed value: `dnsflow`."},
			"sub_source": inputs.TagInfo{Desc: "Some specific connection classifications, " +
				"such as the sub_source value for Kubernetes network traffic is K8s."},
		},
		Fields: map[string]interface{}{
			"rcode": newFInfInt("DNS response code: 0 - `NoError`, 1 - `FormErr`, 2 - `ServFail`, "+
				"3 - NXDomain, 4 - NotImp, 5 - Refused, ...; A value of -1 means the request timed out.", inputs.UnknownUnit),
			"count":       newFInfInt("The number of DNS requests in a collection cycle.", inputs.UnknownUnit),
			"latency":     newFInfInt("Average response time for DNS requests.", inputs.DurationNS),
			"latency_max": newFInfInt("Maximum response time for DNS requests.", inputs.DurationNS),
		},
	}
}

type BashM measurement

func (m *BashM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "bash",
		Tags: map[string]interface{}{
			"host":   inputs.TagInfo{Desc: "host name"},
			"source": inputs.TagInfo{Desc: "Fixed value: bash"},
		},
		Fields: map[string]interface{}{
			"pid":     newFString("Process identification number."),
			"user":    newFString("The user who executes the bash command."),
			"cmd":     newFString("Command."),
			"message": newFString("The bash execution record generated by the collector"),
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

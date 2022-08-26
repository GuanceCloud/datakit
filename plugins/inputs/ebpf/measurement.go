// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ebpf

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *measurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.OOpt())
}

//nolint:lll
func (m *measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}

type ConnStatsM measurement

func (m *ConnStatsM) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.OOpt())
}

//nolint:lll
func (m *ConnStatsM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "netflow",
		Tags: map[string]interface{}{
			"host":                    inputs.TagInfo{Desc: "主机名"},
			"dst_ip":                  inputs.TagInfo{Desc: "目标 IP"},
			"dst_domain":              inputs.TagInfo{Desc: "目标域名"},
			"dst_port":                inputs.TagInfo{Desc: "目标端口"},
			"dst_ip_type":             inputs.TagInfo{Desc: "目标 IP 类型 (other/private/multicast)"},
			"src_ip":                  inputs.TagInfo{Desc: "源 IP"},
			"src_port":                inputs.TagInfo{Desc: "源端口, 临时端口(32768 ~ 60999)聚合后的值为 `*`"},
			"src_ip_type":             inputs.TagInfo{Desc: "源 IP 类型 (other/private/multicast)"},
			"src_k8s_pod_name":        inputs.TagInfo{Desc: "源 IP 所属 k8s 的 pod name"},
			"src_k8s_deployment_name": inputs.TagInfo{Desc: "源 IP 所属 k8s 的 deployment name"},
			"src_k8s_service_name":    inputs.TagInfo{Desc: "源 IP 所属 k8s 的 service name"},
			"src_k8s_namespace":       inputs.TagInfo{Desc: "源 IP 所在 k8s 的 namespace"},
			"dst_k8s_pod_name":        inputs.TagInfo{Desc: "目标 IP 所属 k8s 的 pod name"},
			"dst_k8s_deployment_name": inputs.TagInfo{Desc: "目标 IP 所属 k8s 的 deployment name"},
			"dst_k8s_service_name": inputs.TagInfo{
				Desc: "目标 IP 所属 service, 如果是 dst_ip 是 cluster(service) ip 则 dst_k8s_pod_name 值为 `N/A`",
			},
			"dst_k8s_namespace": inputs.TagInfo{Desc: "目标 IP 所在 k8s 的 namespace"},
			"pid":               inputs.TagInfo{Desc: "进程号"},
			"process_name":      inputs.TagInfo{Desc: "进程名"},
			"transport":         inputs.TagInfo{Desc: "传输协议 (udp/tcp)"},
			"family":            inputs.TagInfo{Desc: "TCP/IP 协议族 (IPv4/IPv6)"},
			"direction":         inputs.TagInfo{Desc: "传输方向 (incoming/outgoing)"},
			"source":            inputs.TagInfo{Desc: "固定值: netflow"},
			"sub_source":        inputs.TagInfo{Desc: "用于 netflow 的部分特定连接分类，如 Kubernetes 流量的值为 K8s"},
		},
		Fields: map[string]interface{}{
			"bytes_read":      newFInfInt("读取字节数", inputs.SizeByte),
			"bytes_written":   newFInfInt("写入字节数", inputs.SizeByte),
			"retransmits":     newFInfInt("重传次数", inputs.NCount),
			"rtt":             newFInfInt("TCP Latency", inputs.DurationUS),
			"rtt_var":         newFInfInt("TCP Jitter", inputs.DurationUS),
			"tcp_closed":      newFInfInt("TCP 关闭次数", inputs.NCount),
			"tcp_established": newFInfInt("TCP 建立连接次数", inputs.NCount),
		},
	}
}

type HTTPFlowM measurement

func (m *HTTPFlowM) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

//nolint:lll
func (m *HTTPFlowM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "httpflow",
		Tags: map[string]interface{}{
			"host":                    inputs.TagInfo{Desc: "主机名"},
			"dst_ip":                  inputs.TagInfo{Desc: "目标 IP"},
			"dst_port":                inputs.TagInfo{Desc: "目标端口"},
			"dst_ip_type":             inputs.TagInfo{Desc: "目标 IP 类型 (other/private/multicast)"},
			"src_ip":                  inputs.TagInfo{Desc: "源 IP"},
			"src_port":                inputs.TagInfo{Desc: "源端口, 临时端口(32768 ~ 60999)聚合后的值为 `*`"},
			"src_ip_type":             inputs.TagInfo{Desc: "源 IP 类型 (other/private/multicast)"},
			"src_k8s_pod_name":        inputs.TagInfo{Desc: "源 IP 所属 k8s 的 pod name"},
			"src_k8s_deployment_name": inputs.TagInfo{Desc: "源 IP 所属 k8s 的 deployment name"},
			"src_k8s_service_name":    inputs.TagInfo{Desc: "源 IP 所属 k8s 的 service name"},
			"src_k8s_namespace":       inputs.TagInfo{Desc: "源 IP 所在 k8s 的 namespace"},
			"dst_k8s_pod_name":        inputs.TagInfo{Desc: "目标 IP 所属 k8s 的 pod name"},
			"dst_k8s_deployment_name": inputs.TagInfo{Desc: "目标 IP 所属 k8s 的 deployment name"},
			"dst_k8s_service_name": inputs.TagInfo{
				Desc: "目标 IP 所属 service, 如果是 dst_ip 是 cluster(service) ip 则 dst_k8s_pod_name 值为 `N/A`",
			},
			"dst_k8s_namespace": inputs.TagInfo{Desc: "目标 IP 所在 k8s 的 namespace"},
			// "pid":               inputs.TagInfo{Desc: "进程号"},
			"transport":  inputs.TagInfo{Desc: "传输协议 (udp/tcp)"},
			"family":     inputs.TagInfo{Desc: "TCP/IP 协议族 (IPv4/IPv6)"},
			"direction":  inputs.TagInfo{Desc: "传输方向 (incoming/outgoing)"},
			"source":     inputs.TagInfo{Desc: "固定值: httpflow"},
			"sub_source": inputs.TagInfo{Desc: "用于 httpflow 的部分特定连接分类，如 Kubernetes 流量的值为 K8s"},
		},
		Fields: map[string]interface{}{
			"path":         newFString("请求路径"),
			"status_code":  newFInfInt("http 状态码，如 200, 301, 404 ...", inputs.UnknownUnit),
			"method":       newFString("GET/POST/..."),
			"latency":      newFInfInt("ttfb", inputs.DurationNS),
			"http_version": newFString("1.1 / 1.0 ..."),
		},
	}
}

type DNSStatsM measurement

func (m *DNSStatsM) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

func (m *DNSStatsM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "dnsflow",
		Tags: map[string]interface{}{
			"host":      inputs.TagInfo{Desc: "host name"},
			"src_ip":    inputs.TagInfo{Desc: "DNS client address"},
			"src_port":  inputs.TagInfo{Desc: "DNS client port"},
			"dst_ip":    inputs.TagInfo{Desc: "DNS server address"},
			"dst_port":  inputs.TagInfo{Desc: "DNS server port"},
			"transport": inputs.TagInfo{Desc: "传输协议 (udp/tcp)"},
			"family":    inputs.TagInfo{Desc: "TCP/IP 协议族 (IPv4/IPv6)"},
			"source":    inputs.TagInfo{Desc: "固定值: dnsflow"},
		},
		Fields: map[string]interface{}{
			"timeout": newFInfBool("DNS 请求超时", inputs.UnknownUnit),
			"rcode": newFInfInt("DNS 响应码: 0 - NoError, 1 - FormErr, 2 - ServFail, "+
				"3 - NXDomain, 4 - NotImp, 5 - Refused, ...", inputs.UnknownUnit),
			"resp_time": newFInfInt("DNS 请求的响应时间间隔", inputs.DurationNS),
		},
	}
}

type BashM measurement

func (m *BashM) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

func (m *BashM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "bash",
		Tags: map[string]interface{}{
			"host":   inputs.TagInfo{Desc: "host name"},
			"source": inputs.TagInfo{Desc: "固定值: bash"},
		},
		Fields: map[string]interface{}{
			"pid":     newFString("bash 进程的 pid"),
			"user":    newFString("执行 bash 命令的用户"),
			"cmd":     newFString("bash 命令"),
			"message": newFString("单条 bash 执行记录"),
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

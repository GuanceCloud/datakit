// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package class

import (
	"fmt"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
)

// NetworkInterfacePerformance 表示网络接口性能数据.
// nolint:stylecheck
type Win32_PerfFormattedData_Tcpip_NetworkInterface struct {
	Name                     string `wmi:"Name"`
	Caption                  string `wmi:"Caption"`
	Description              string `wmi:"Description"`
	BytesReceivedPersec      uint64 `wmi:"BytesReceivedPersec"`
	BytesSentPersec          uint64 `wmi:"BytesSentPersec"`
	BytesTotalPersec         uint64 `wmi:"BytesTotalPersec"`
	CurrentBandwidth         uint64 `wmi:"CurrentBandwidth"`
	OutputQueueLength        uint64 `wmi:"OutputQueueLength"`
	PacketsOutboundDiscarded uint64 `wmi:"PacketsOutboundDiscarded"`
	PacketsOutboundErrors    uint64 `wmi:"PacketsOutboundErrors"`
	PacketsPersec            uint64 `wmi:"PacketsPersec"`
	PacketsReceivedDiscarded uint64 `wmi:"PacketsReceivedDiscarded"`
	PacketsReceivedErrors    uint64 `wmi:"PacketsReceivedErrors"`
	PacketsReceivedPersec    uint64 `wmi:"PacketsReceivedPersec"`
	PacketsSentPersec        uint64 `wmi:"PacketsSentPersec"`
}

// String 方法提供结构体的格式化输出.
func (n *Win32_PerfFormattedData_Tcpip_NetworkInterface) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("网络接口名称: %s\n", n.Name))
	builder.WriteString(fmt.Sprintf("网络接口名称 Caption: %s\n", n.Caption))
	builder.WriteString(fmt.Sprintf("网络接口名称 Description: %s\n", n.Description))
	builder.WriteString(fmt.Sprintf("每秒接收字节数: %d\n", n.BytesReceivedPersec))
	builder.WriteString(fmt.Sprintf("每秒发送字节数: %d\n", n.BytesSentPersec))
	builder.WriteString(fmt.Sprintf("每秒总字节数: %d\n", n.BytesTotalPersec))
	builder.WriteString(fmt.Sprintf("当前带宽: %d bps\n", n.CurrentBandwidth))
	builder.WriteString(fmt.Sprintf("输出队列长度: %d\n", n.OutputQueueLength))
	builder.WriteString(fmt.Sprintf("丢弃的出站数据包: %d\n", n.PacketsOutboundDiscarded))
	builder.WriteString(fmt.Sprintf("出站错误数据包: %d\n", n.PacketsOutboundErrors))
	builder.WriteString(fmt.Sprintf("每秒数据包数: %d\n", n.PacketsPersec))
	builder.WriteString(fmt.Sprintf("丢弃的接收数据包: %d\n", n.PacketsReceivedDiscarded))
	builder.WriteString(fmt.Sprintf("接收错误数据包: %d\n", n.PacketsReceivedErrors))
	builder.WriteString(fmt.Sprintf("每秒接收数据包数: %d\n", n.PacketsReceivedPersec))
	builder.WriteString(fmt.Sprintf("每秒发送数据包数: %d\n", n.PacketsSentPersec))

	return builder.String()
}

func (n *Win32_PerfFormattedData_Tcpip_NetworkInterface) ToPoint(host string) *point.Point {
	opts := point.DefaultMetricOptions()
	var kvs point.KVs
	kvs = kvs.AddTag("host", host).
		AddTag("interface", n.Name).
		AddV2("bytes_recv/sec", n.BytesReceivedPersec, false).
		AddV2("bytes_sent/sec", n.BytesSentPersec, false).
		AddV2("drop_in", n.PacketsReceivedDiscarded, false).
		AddV2("drop_out", n.PacketsOutboundDiscarded, false).
		AddV2("err_out", n.PacketsOutboundErrors, false).
		AddV2("packets_recv/sec", n.PacketsReceivedPersec, false).
		AddV2("packets_sent/sec", n.PacketsSentPersec, false)

	return point.NewPointV2("net", kvs, opts...)
}

// 网卡的基本信息.
type Win32_NetworkAdapterConfiguration struct {
	Description          string
	MACAddress           string
	IPAddress            []string
	IPSubnet             []string
	DefaultIPGateway     []string
	DNSServerSearchOrder []string
	DHCPEnabled          bool
}

func (c *Win32_NetworkAdapterConfiguration) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("网卡名称: %s\n", c.Description))
	builder.WriteString(fmt.Sprintf("MAC地址: %s\n", c.MACAddress))

	if len(c.IPAddress) > 0 {
		builder.WriteString("IP地址:\n")
		for i, ip := range c.IPAddress {
			builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, ip))
		}
	} else {
		builder.WriteString("IP地址: 无\n")
	}

	if len(c.IPSubnet) > 0 {
		builder.WriteString("子网掩码:\n")
		for i, mask := range c.IPSubnet {
			builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, mask))
		}
	} else {
		builder.WriteString("子网掩码: 无\n")
	}

	if len(c.DefaultIPGateway) > 0 {
		builder.WriteString("默认网关:\n")
		for i, gw := range c.DefaultIPGateway {
			builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, gw))
		}
	} else {
		builder.WriteString("默认网关: 无\n")
	}

	if len(c.DNSServerSearchOrder) > 0 {
		builder.WriteString("DNS服务器:\n")
		for i, dns := range c.DNSServerSearchOrder {
			builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, dns))
		}
	} else {
		builder.WriteString("DNS服务器: 无\n")
	}

	builder.WriteString(fmt.Sprintf("DHCP启用: %v", c.DHCPEnabled))

	return builder.String()
}

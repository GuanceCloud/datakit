package netebpf

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}

type ConnStatsM measurement

func (m *ConnStatsM) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *ConnStatsM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "netflow",
		Tags: map[string]interface{}{
			"host":        inputs.TagInfo{Desc: "主机名"},
			"dst_ip":      inputs.TagInfo{Desc: "目标 IP"},
			"dst_domain":  inputs.TagInfo{Desc: "目标域名"},
			"dst_port":    inputs.TagInfo{Desc: "目标端口"},
			"dst_ip_type": inputs.TagInfo{Desc: "目标 IP 类型 (other/private/multicast)"},
			"src_ip":      inputs.TagInfo{Desc: "源 IP"},
			"src_port":    inputs.TagInfo{Desc: "源端口, 临时端口(32768 ~ 60999)聚合后的值为 `*`"},
			"src_ip_type": inputs.TagInfo{Desc: "源 IP 类型 (other/private/multicast)"},
			"pid":         inputs.TagInfo{Desc: "进程号"},
			"transport":   inputs.TagInfo{Desc: "传输协议 (udp/tcp)"},
			"family":      inputs.TagInfo{Desc: "TCP/IP 协议族 (IPv4/IPv6)"},
			"direction":   inputs.TagInfo{Desc: "传输方向 (incoming/outgoing)"},
			"source":      inputs.TagInfo{Desc: "固定值: netflow"},
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

func newFInfInt(desc, unit string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     unit,
		Desc:     desc,
	}
}

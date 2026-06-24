// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package net

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

// https://tools.ietf.org/html/rfc1213#page-48
// https://www.kernel.org/doc/html/latest/networking/snmp_counter.html
// https://sourceforge.net/p/net-tools/code/ci/master/tree/statistics.c#l178
// nolint:lll
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricName,
		Cat:    point.Metric,
		Desc:   "Network interface counters from gopsutil and Linux TCP/UDP protocol counters from `/proc/net/snmp`; `*/sec` fields are derived per-second rates.",
		DescZh: "通过 gopsutil 采集网络接口计数，并从 Linux `/proc/net/snmp` 采集 TCP/UDP 协议计数；`*/sec` 字段为派生的每秒速率。",
		Fields: map[string]interface{}{
			"bytes_sent":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Cumulative bytes sent by the interface."},
			"bytes_sent/sec":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.BytesPerSec, Desc: "The number of bytes sent by the interface per second."},
			"bytes_recv":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Cumulative bytes received by the interface."},
			"bytes_recv/sec":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.BytesPerSec, Desc: "The number of bytes received by the interface per second."},
			"packets_sent":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative packets sent by the interface."},
			"packets_sent/sec": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of packets sent by the interface per second."},
			"packets_recv":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative packets received by the interface."},
			"packets_recv/sec": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of packets received by the interface per second."},
			"err_in":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative receive errors detected by the interface."},
			"err_out":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative transmit errors detected by the interface."},
			"drop_in":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative received packets dropped by the interface."},
			"drop_out":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative transmitted packets dropped by the interface."},

			// linux only
			"tcp_insegs":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative TCP segments received by the TCP layer. Linux only"},
			"tcp_insegs/sec":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of packets received by the TCP layer per second. Linux only"},
			"tcp_outsegs":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative TCP segments sent by the TCP layer. Linux only"},
			"tcp_outsegs/sec":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of packets sent by the TCP layer per second. Linux only"},
			"tcp_activeopens":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative TCP active opens, where TCP sends SYN and enters SYN-SENT. Linux only"},
			"tcp_passiveopens":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative TCP passive opens, where TCP receives SYN, replies SYN+ACK, and enters SYN-RCVD. Linux only"},
			"tcp_estabresets":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative TCP connections that transitioned directly to CLOSED from ESTABLISHED or CLOSE-WAIT. Linux only"},
			"tcp_attemptfails":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative failed TCP connection attempts as defined by TCP-MIB AttemptFails. Linux only"},
			"tcp_outrsts":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative TCP segments sent containing the RST flag. Linux only"},
			"tcp_retranssegs":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative TCP segments retransmitted with one or more previously transmitted octets. Linux only"},
			"tcp_inerrs":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative incoming TCP segments received in error. Linux only"},
			"tcp_incsumerrors":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative incoming TCP segments with checksum errors. Linux only"},
			"tcp_rtoalgorithm":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Algorithm identifier used to determine retransmission timeout values. Linux only"},
			"tcp_rtomin":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The minimum value permitted by a TCP implementation for the retransmission timeout, measured in milliseconds. Linux only"},
			"tcp_rtomax":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The maximum value permitted by a TCP implementation for the retransmission timeout, measured in milliseconds. Linux only"},
			"tcp_maxconn":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The limit on the total number of TCP connections the entity can support. Linux only"},
			"tcp_currestab":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of TCP connections for which the current state is either ESTABLISHED or CLOSE-WAIT. Linux only"},
			"udp_incsumerrors":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative incoming UDP datagrams with checksum errors. Linux only"},
			"udp_indatagrams":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative UDP datagrams delivered to UDP users. Linux only"},
			"udp_indatagrams/sec":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of UDP datagram delivered to UDP users per second. Linux only"},
			"udp_outdatagrams":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative UDP datagrams sent from this entity. Linux only"},
			"udp_outdatagrams/sec": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of UDP datagram sent from this entity per second. Linux only"},
			"udp_rcvbuferrors":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative UDP receive buffer errors. Linux only"},
			"udp_noports":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative UDP packets received for unknown ports. Linux only"},
			"udp_sndbuferrors":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative UDP send buffer errors. Linux only"},
			"udp_inerrors":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative UDP packet receive errors. Linux only"},
			"udp_memerrors":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative UDP memory errors. Linux only"},
			"udp_ignoredmulti":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative ignored UDP multicast packets. Linux only"},
		},
		Tags: map[string]interface{}{
			"host":      &inputs.TagInfo{Desc: "System hostname."},
			"interface": &inputs.TagInfo{Desc: "Network interface name."},
		},
	}
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package net

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

// https://tools.ietf.org/html/rfc1213#page-48
// https://www.kernel.org/doc/html/latest/networking/snmp_counter.html
// https://sourceforge.net/p/net-tools/code/ci/master/tree/statistics.c#l178
// nolint:lll
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Type: "metric",
		Fields: map[string]interface{}{
			"bytes_sent":       newFieldsInfoIByte("The number of bytes sent by the interface."),
			"bytes_sent/sec":   newFieldsInfoIBytePerSec("The number of bytes sent by the interface per second."),
			"bytes_recv":       newFieldsInfoIByte("The number of bytes received by the interface."),
			"bytes_recv/sec":   newFieldsInfoIBytePerSec("The number of bytes received by the interface per second."),
			"packets_sent":     newFieldsInfoCount("The number of packets sent by the interface."),
			"packets_sent/sec": newFieldsInfoCountPerSec("The number of packets sent by the interface per second."),
			"packets_recv":     newFieldsInfoCount("The number of packets received by the interface."),
			"packets_recv/sec": newFieldsInfoCountPerSec("The number of packets received by the interface per second."),
			"err_in":           newFieldsInfoCount("The number of receive errors detected by the interface."),
			"err_out":          newFieldsInfoCount("The number of transmit errors detected by the interface."),
			"drop_in":          newFieldsInfoCount("The number of received packets dropped by the interface."),
			"drop_out":         newFieldsInfoCount("The number of transmitted packets dropped by the interface."),
			// linux only
			"tcp_insegs":           newFieldsInfoCount("The number of packets received by the TCP layer."),
			"tcp_insegs/sec":       newFieldsInfoCountPerSec("The number of packets received by the TCP layer per second."),
			"tcp_outsegs":          newFieldsInfoCount("The number of packets sent by the TCP layer."),
			"tcp_outsegs/sec":      newFieldsInfoCountPerSec("The number of packets sent by the TCP layer per second."),
			"tcp_activeopens":      newFieldsInfoCount("It means the TCP layer sends a SYN, and come into the SYN-SENT state."),
			"tcp_passiveopens":     newFieldsInfoCount("It means the TCP layer receives a SYN, replies a SYN+ACK, come into the SYN-RCVD state."),
			"tcp_estabresets":      newFieldsInfoCount("The number of times TCP connections have made a direct transition to the CLOSED state from either the ESTABLISHED state or the CLOSE-WAIT state."),
			"tcp_attemptfails":     newFieldsInfoCount("The number of times TCP connections have made a direct transition to the CLOSED state from either the SYN-SENT state or the SYN-RCVD state, plus the number of times TCP connections have made a direct transition to the LISTEN state from the SYN-RCVD state."),
			"tcp_outrsts":          newFieldsInfoCount("The number of TCP segments sent containing the RST flag."),
			"tcp_retranssegs":      newFieldsInfoCount("The total number of segments re-transmitted - that is, the number of TCP segments transmitted containing one or more previously transmitted octets."),
			"tcp_inerrs":           newFieldsInfoCount("The number of incoming TCP segments in error."),
			"tcp_incsumerrors":     newFieldsInfoCount("The number of incoming TCP segments in checksum error."),
			"tcp_rtoalgorithm":     newFieldsInfoCount("The algorithm used to determine the timeout value used for retransmitting unacknowledged octets."),
			"tcp_rtomin":           newFieldsInfoMS("The minimum value permitted by a TCP implementation for the retransmission timeout, measured in milliseconds."),
			"tcp_rtomax":           newFieldsInfoMS("The maximum value permitted by a TCP implementation for the retransmission timeout, measured in milliseconds."),
			"tcp_maxconn":          newFieldsInfoCount("The limit on the total number of TCP connections the entity can support."),
			"tcp_currestab":        newFieldsInfoCount("The number of TCP connections for which the current state is either ESTABLISHED or CLOSE-WAIT."),
			"udp_incsumerrors":     newFieldsInfoCount("The number of incoming UDP datagram in checksum error.s"),
			"udp_indatagrams":      newFieldsInfoCount("The number of UDP datagram delivered to UDP users."),
			"udp_indatagrams/sec":  newFieldsInfoCountPerSec("The number of UDP datagram delivered to UDP users per second."),
			"udp_outdatagrams":     newFieldsInfoCount("The number of UDP datagram sent from this entity."),
			"udp_outdatagrams/sec": newFieldsInfoCountPerSec("The number of UDP datagram sent from this entity per second."),
			"udp_rcvbuferrors":     newFieldsInfoCount("The number of receive buffer errors."),
			"udp_noports":          newFieldsInfoCount("The number of packets to unknown port received."),
			"udp_sndbuferrors":     newFieldsInfoCount("The number of send buffer errors."),
			"udp_inerrors":         newFieldsInfoCount("The number of packet receive errors."),
			"udp_memerrors":        newFieldsInfoCount("The number of memory errors."),
			"udp_ignoredmulti":     newFieldsInfoCount(inputs.TODO),
		},
		Tags: map[string]interface{}{
			"host":      &inputs.TagInfo{Desc: "System hostname."},
			"interface": &inputs.TagInfo{Desc: "Network interface name."},
		},
	}
}

func newFieldsInfoIByte(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

func newFieldsInfoIBytePerSec(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.BytesPerSec,
		Desc:     desc,
	}
}

func newFieldsInfoCount(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func newFieldsInfoCountPerSec(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func newFieldsInfoMS(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.DurationMS,
		Desc:     desc,
	}
}

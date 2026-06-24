// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netstat

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type (
	docMeasurement  struct{}
	portMeasurement struct{}
)

// Info , reflected in the document
//
//nolint:lll
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricName,
		Cat:    point.Metric,
		Desc:   "Host network socket state metrics, reporting the current number of TCP sockets in each TCP state and UDP sockets grouped by IP version.",
		DescZh: "主机网络 socket 状态指标，按 IP 版本上报当前各 TCP 状态连接数和 UDP socket 数。",
		Fields: map[string]interface{}{
			"tcp_established": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of TCP sockets in ESTABLISHED state.",
				Taggedby: []string{"ip_version"},
			},
			"tcp_syn_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of TCP sockets in SYN_SENT state after sending a connection request.",
				Taggedby: []string{
					"ip_version",
				},
			},
			"tcp_syn_recv": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of TCP sockets in SYN_RECV state after receiving a connection request and sending an acknowledgement.",
				Taggedby: []string{
					"ip_version",
				},
			},
			"tcp_fin_wait1": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of TCP sockets in FIN_WAIT1 state after the local endpoint has requested connection termination.",
				Taggedby: []string{
					"ip_version",
				},
			},
			"tcp_fin_wait2": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of TCP sockets in FIN_WAIT2 state while waiting for the remote endpoint to terminate the connection.",
				Taggedby: []string{
					"ip_version",
				},
			},
			"tcp_time_wait": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of TCP sockets in TIME_WAIT state after connection termination.",
				Taggedby: []string{
					"ip_version",
				},
			},
			"tcp_close": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of TCP sockets in CLOSE state.",
				Taggedby: []string{"ip_version"},
			},
			"tcp_close_wait": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of TCP sockets in CLOSE_WAIT state after the remote endpoint has requested termination.",
				Taggedby: []string{
					"ip_version",
				},
			},
			"tcp_last_ack": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of TCP sockets in LAST_ACK state while waiting for acknowledgement of the final termination segment.",
				Taggedby: []string{
					"ip_version",
				},
			},
			"tcp_listen": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of TCP sockets in LISTEN state waiting for incoming connection requests.",
				Taggedby: []string{
					"ip_version",
				},
			},
			"tcp_closing": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of TCP sockets in CLOSING state while both endpoints are closing the connection.",
				Taggedby: []string{
					"ip_version",
				},
			},
			"tcp_none": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of TCP sockets whose state was reported as NONE by the operating system.",
				Taggedby: []string{
					"ip_version",
				},
			},
			"udp_socket": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of UDP sockets.",
				Taggedby: []string{
					"ip_version",
				},
			},
		},

		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "Host name."},
			"addr_port":  &inputs.TagInfo{Desc: "Addr and port. Optional."},
			"ip_version": &inputs.TagInfo{Desc: "IP version, 4 for IPV4, 6 for IPV6, unknown for others"},
		},
	}
}

// Info , reflected in the document
//
//nolint:lll
func (*portMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePort,
		Cat:  point.Metric,
		Desc: "Configured port-level network socket state metrics, reporting current " +
			"TCP state counts and UDP socket counts for each matched local address and port.",
		DescZh: "配置端口级别的网络 socket 状态指标，按匹配的本地地址和端口上报当前 TCP 状态连接数和 UDP socket 数。",
		Fields: map[string]interface{}{
			"tcp_established": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of matched TCP sockets in ESTABLISHED state.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
			"tcp_syn_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of matched TCP sockets in SYN_SENT state after sending a connection request.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
			"tcp_syn_recv": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of matched TCP sockets in SYN_RECV state after receiving a connection request and sending an acknowledgement.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
			"tcp_fin_wait1": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of matched TCP sockets in FIN_WAIT1 state after the local endpoint has requested connection termination.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
			"tcp_fin_wait2": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of matched TCP sockets in FIN_WAIT2 state while waiting for the remote endpoint to terminate the connection.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
			"tcp_time_wait": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of matched TCP sockets in TIME_WAIT state after connection termination.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
			"tcp_close": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of matched TCP sockets in CLOSE state.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
			"tcp_close_wait": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of matched TCP sockets in CLOSE_WAIT state after the remote endpoint has requested termination.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
			"tcp_last_ack": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of matched TCP sockets in LAST_ACK state while waiting for acknowledgement of the final termination segment.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
			"tcp_listen": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of matched TCP sockets in LISTEN state waiting for incoming connection requests.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
			"tcp_closing": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of matched TCP sockets in CLOSING state while both endpoints are closing the connection.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
			"tcp_none": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of matched TCP sockets whose state was reported as NONE by the operating system.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
			"udp_socket": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current number of matched UDP sockets.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
			"pid": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc: "Process ID associated with the last matched socket in this address/port group. " +
					"This field is omitted from the aggregate `netstat` measurement.",
				Taggedby: []string{"addr_port", "ip_version"},
			},
		},

		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "Host name."},
			"addr_port":  &inputs.TagInfo{Desc: "Matched local address and port, or configured port value when the input is configured without an address."},
			"ip_version": &inputs.TagInfo{Desc: "IP version, 4 for IPV4, 6 for IPV6, unknown for others"},
		},
	}
}

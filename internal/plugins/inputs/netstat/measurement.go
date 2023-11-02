// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netstat

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

// Info , reflected in the document
//
//nolint:lll
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Type: "metric",
		Fields: map[string]interface{}{
			"tcp_established": newFieldInfoC("ESTABLISHED : The number of TCP state be open connection, data received to be delivered to the user. "),
			"tcp_syn_sent":    newFieldInfoC("SYN_SENT : The number of TCP state be waiting for a machine connection request after sending a connecting request."),
			"tcp_syn_recv":    newFieldInfoC("SYN_RECV : The number of TCP state be waiting for confirmation of connection acknowledgement after both sender and receiver has sent / received connection request."),
			"tcp_fin_wait1":   newFieldInfoC("FIN_WAIT1 : The number of TCP state be waiting for a connection termination request from remote TCP host or acknowledgment of connection termination request sent previously."),
			"tcp_fin_wait2":   newFieldInfoC("FIN_WAIT2 : The number of TCP state be waiting for connection termination request from remote TCP host."),
			"tcp_time_wait":   newFieldInfoC("TIME_WAIT : The number of TCP state be waiting sufficient time to pass to ensure remote TCP host received acknowledgement of its request for connection termination."),
			"tcp_close":       newFieldInfoC("CLOSE : The number of TCP state be waiting for a connection termination request acknowledgement from remote TCP host."),
			"tcp_close_wait":  newFieldInfoC("CLOSE_WAIT : The number of TCP state be waiting for a connection termination request from local user."),
			"tcp_last_ack":    newFieldInfoC("LAST_ACK : The number of TCP state be waiting for connection termination request acknowledgement previously sent to remote TCP host including its acknowledgement of connection termination request."),
			"tcp_listen":      newFieldInfoC("LISTEN : The number of TCP state be waiting for a connection request from any remote TCP host."),
			"tcp_closing":     newFieldInfoC("CLOSING : The number of TCP state be waiting for a connection termination request acknowledgement from remote TCP host."),
			"tcp_none":        newFieldInfoC("NONE"),
			"udp_socket":      newFieldInfoC("UDP : The number of UDP connection."),
			"pid":             newFieldInfoC("PID. Optional."),
		},

		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "Host name"},
			"addr_port":  &inputs.TagInfo{Desc: "Addr and port. Optional."},
			"ip_version": &inputs.TagInfo{Desc: "IP version, 4 for IPV4, 6 for IPV6, unknown for others"},
		},
	}
}

// NewFieldInfoC new count field.
func newFieldInfoC(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

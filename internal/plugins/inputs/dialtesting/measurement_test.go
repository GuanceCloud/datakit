// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestMeasurementInfo(t *testing.T) {
	cases := []struct {
		name      string
		meas      inputs.Measurement
		wantName  string
		tagKeys   []string
		fieldKeys []string
	}{
		{
			name:     "http",
			meas:     &httpMeasurement{},
			wantName: "http_dial_testing",
			tagKeys:  []string{"url", "method", "node_name", LabelDF},
			fieldKeys: []string{
				"status_code", "message", "response_time", "seq_number", "config_vars",
			},
		},
		{
			name:     "tcp",
			meas:     &tcpMeasurement{},
			wantName: "tcp_dial_testing",
			tagKeys:  []string{"dest_host", "dest_port", "proto", LabelDF},
			fieldKeys: []string{
				"message", "traceroute", "response_time", "success", "config_vars",
			},
		},
		{
			name:     "icmp",
			meas:     &icmpMeasurement{},
			wantName: "icmp_dial_testing",
			tagKeys:  []string{"dest_host", "proto", "node_name", LabelDF},
			fieldKeys: []string{
				"average_round_trip_time", "packet_loss_percent", "packets_sent", "success", "config_vars",
			},
		},
		{
			name:     "websocket",
			meas:     &websocketMeasurement{},
			wantName: "websocket_dial_testing",
			tagKeys:  []string{"url", "proto", "node_name", LabelDF},
			fieldKeys: []string{
				"response_message", "sent_message", "response_time", "success", "config_vars",
			},
		},
		{
			name:     "multi",
			meas:     &multiMeasurement{},
			wantName: "multi_dial_testing",
			tagKeys:  []string{"name", "node_name", "status", LabelDF},
			fieldKeys: []string{
				"last_step", "steps", "response_time", "success", "config_vars",
			},
		},
		{
			name:     "grpc",
			meas:     &grpcMeasurement{},
			wantName: "grpc_dial_testing",
			tagKeys:  []string{"server", "dest_host", "method", LabelDF},
			fieldKeys: []string{
				"message", "response_time", "success", "seq_number", "config_vars",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			info := tc.meas.Info()
			if !assert.NotNil(t, info) {
				return
			}

			assert.Equal(t, tc.wantName, info.Name)
			assert.Equal(t, point.DialTesting, info.Cat)

			for _, key := range tc.tagKeys {
				assert.Contains(t, info.Tags, key)
				assert.IsType(t, &inputs.TagInfo{}, info.Tags[key])
			}

			for _, key := range tc.fieldKeys {
				assert.Contains(t, info.Fields, key)
				assert.IsType(t, &inputs.FieldInfo{}, info.Fields[key])
			}
		})
	}
}

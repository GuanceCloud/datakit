// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package custom testing.
package custom

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func TestCustom_Process(t *testing.T) {
	type fields struct {
		SpiltBody       bool
		logTopicsMap    map[string]string
		metricTopicsMap map[string]string
		SpiltTopicsMap  map[string]bool
		feeder          *io.MockedFeeder
	}
	type args struct {
		msg *sarama.ConsumerMessage
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "log",
			fields: fields{
				SpiltBody:    false,
				logTopicsMap: map[string]string{"apm": "apm.p"},
				feeder:       io.NewMockedFeeder(),
			},
			args: args{
				msg: &sarama.ConsumerMessage{Topic: "apm", Value: []byte("this is msg body")},
			},
		},
		{
			name: "metric",
			fields: fields{
				SpiltBody:       false,
				metricTopicsMap: map[string]string{"apm": "apm.p"},
				feeder:          io.NewMockedFeeder(),
			},
			args: args{
				msg: &sarama.ConsumerMessage{Topic: "apm", Value: []byte("this is msg body")},
			},
		},
		{
			name: "spilt_json_body_true",
			fields: fields{
				SpiltBody:    true,
				logTopicsMap: map[string]string{"apm": "apm.p"},
				feeder:       io.NewMockedFeeder(),
			},
			args: args{
				msg: &sarama.ConsumerMessage{
					Topic: "apm",
					Value: []byte(`[{"index":"1","message":"log msg"},{"index":"1","message":"log msg"}]`),
				},
			},
		},
		{
			name: "spilt_json_body_single",
			fields: fields{
				SpiltBody:      true,
				logTopicsMap:   map[string]string{"apm": "apm.p"},
				feeder:         io.NewMockedFeeder(),
				SpiltTopicsMap: map[string]bool{"apm": true},
			},
			args: args{
				msg: &sarama.ConsumerMessage{
					Topic: "apm",
					Value: []byte(`[{"index":"1","message":"log msg"},{"index":"1","message":"log msg"}]`),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mq := &Custom{
				SpiltBody:       tt.fields.SpiltBody,
				LogTopicsMap:    tt.fields.logTopicsMap,
				MetricTopicsMap: tt.fields.metricTopicsMap,
				SpiltTopicsMap:  tt.fields.SpiltTopicsMap,
				feeder:          tt.fields.feeder,
			}
			mq.Process(tt.args.msg)
			ps, err := tt.fields.feeder.AnyPoints(time.Second)
			if err != nil {
				t.Errorf("feeder anyPoints err:%v", err)
				return
			}
			if len(ps) == 0 {
				t.Logf("ps len =0")
				return
			}
			pt := ps[0]
			t.Logf("%v", pt)
			if pt.Get("message") == nil {
				t.Errorf("not has tag: [message]")
			}

			assert.NotEmptyf(t, pt.GetTag("type"), "tag [type] missing")
		})
	}
}

type KafkaMessage struct {
	Topic string `json:"topic"`
	Value []byte `json:"value"`
}

func TestMashellJson(t *testing.T) {
	messages := []KafkaMessage{
		{Topic: "bfySpan", Value: []byte("value1")},
		{Topic: "bfySpan", Value: []byte("value2")},
		{Topic: "bfySpan", Value: []byte("value3")},
		{Topic: "bfySpan", Value: []byte("value4")},
	}

	// 将多个消息打包为HTTP请求的主体
	jsonData, err := json.MarshalIndent(messages, "", "	")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	fmt.Println(string(jsonData))
}

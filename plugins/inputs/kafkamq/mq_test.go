// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kafkamq  mq
package kafkamq

import (
	"reflect"
	"testing"

	"github.com/Shopify/sarama"
)

func TestInput_getKafkaVersion(t *testing.T) {
	type fields struct {
		KafkaVersion string
	}
	tests := []struct {
		name   string
		fields fields
		want   sarama.KafkaVersion
	}{
		{name: "case1", fields: fields{KafkaVersion: "2.0.0"}, want: sarama.V2_0_0_0},
		{name: "case2", fields: fields{KafkaVersion: "2.12-2.8.1"}, want: sarama.V1_0_0_0},
		{name: "case3", fields: fields{KafkaVersion: "0.8.2.0"}, want: sarama.V0_8_2_0},
		{name: "case4", fields: fields{KafkaVersion: "0.7.0"}, want: sarama.V1_0_0_0},
		{name: "case4", fields: fields{KafkaVersion: "2.8.0"}, want: sarama.V2_8_0_0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Input{
				KafkaVersion: tt.fields.KafkaVersion,
			}
			if got := getKafkaVersion(i.KafkaVersion); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getKafkaVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kafkamq  mq
package kafkamq

import (
	"reflect"
	"testing"

	"github.com/IBM/sarama"
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
		{name: "case2", fields: fields{KafkaVersion: "2.12-2.8.1"}, want: sarama.V2_1_0_0},
		{name: "case3", fields: fields{KafkaVersion: "0.8.2.0"}, want: sarama.V0_8_2_0},
		{name: "case4", fields: fields{KafkaVersion: "0.7.0"}, want: sarama.V2_1_0_0},
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

func Test_newSaramaConfig(t *testing.T) {
	type args struct {
		opts []option
	}
	tests := []struct {
		name string
		args args
		want *sarama.Config
	}{
		{
			name: "case_offset_new",
			args: args{opts: []option{withOffset(-1), withAssignors(""), withVersion("2.0.0")}},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newSaramaConfig(tt.args.opts...)
			if got.Version != sarama.V2_0_0_0 {
				t.Errorf("with version error")
			}
			if got.Consumer.Offsets.Initial != sarama.OffsetNewest {
				t.Errorf("offset must be: newest")
			}
		})
	}
}

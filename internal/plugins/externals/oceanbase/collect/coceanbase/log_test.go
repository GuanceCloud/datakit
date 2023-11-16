// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package coceanbase

import (
	"reflect"
	"testing"
)

func Test_normalizeResultArray(t *testing.T) {
	type args struct {
		in []map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want []map[string]interface{}
	}{
		{
			name: "normal",
			args: args{
				in: []map[string]interface{}{
					{
						"aaa": []uint8("bbb"),
						"ccc": []uint8("ddd"),
					},
					{
						"aaaa": []uint8("bbbb"),
						"cccc": []uint8("dddd"),
						"eeee": []uint8("ffff"),
					},
				},
			},
			want: []map[string]interface{}{
				{
					"aaa": "bbb",
					"ccc": "ddd",
				},
				{
					"aaaa": "bbbb",
					"cccc": "dddd",
					"eeee": "ffff",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizeResultArray(tt.args.in)
			if !reflect.DeepEqual(tt.args.in, tt.want) {
				t.Errorf("normalizeResultArray() = %v, want %v", tt.args.in, tt.want)
			}
		})
	}
}

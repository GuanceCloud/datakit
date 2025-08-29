// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package trace

import (
	"reflect"
	"testing"
)

func TestAddLabels(t *testing.T) {
	type args struct {
		labels []string
		tags   []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "case01", args: args{labels: []string{"a", "b"}, tags: []string{"c", "d"}}, want: []string{"a", "b", "c", "d"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddLabels(tt.args.labels, tt.args.tags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDelLabels(t *testing.T) {
	type args struct {
		labels []string
		tags   []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "case01", args: args{labels: []string{"a", "b", "c", "d"}, tags: []string{"c"}}, want: []string{"a", "b", "d"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DelLabels(tt.args.labels, tt.args.tags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

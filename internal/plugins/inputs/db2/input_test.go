// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package db2

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/external"
)

func TestNeedElectionFlag(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "check_election",
			args: args{name: inputName},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := external.NeedElectionFlag(tt.args.name); got != tt.want {
				t.Errorf("NeedElectionFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

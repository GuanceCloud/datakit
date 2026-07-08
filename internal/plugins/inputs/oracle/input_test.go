// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"database/sql"
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

func TestParseResourceLimitValue(t *testing.T) {
	tests := []struct {
		name  string
		value sql.NullString
		want  int64
		ok    bool
	}{
		{
			name:  "numeric value with leading spaces",
			value: sql.NullString{String: "       472", Valid: true},
			want:  472,
			ok:    true,
		},
		{
			name:  "unlimited value",
			value: sql.NullString{String: " UNLIMITED", Valid: true},
			ok:    false,
		},
		{
			name:  "null value",
			value: sql.NullString{},
			ok:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseResourceLimitValue(tt.value)
			if ok != tt.ok {
				t.Fatalf("parseResourceLimitValue() ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("parseResourceLimitValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

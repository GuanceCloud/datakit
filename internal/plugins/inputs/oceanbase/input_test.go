// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oceanbase

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestObfuscateSQL(t *testing.T) {
	cases := []struct {
		name string
		sql  string
		want string
	}{
		{
			name: "obfuscate_sql",
			sql: `select * 
			from t1 
			where a=1`,
			want: "select * from t1 where a = ?",
		},
		{
			name: "obfuscate_sql_with_comment",
			sql: `
			select * 
			from t1
			-- comment
			where a=1
			`,
			want: "select * from t1 where a = ?",
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.want, obfuscateSQL(tc.sql))
	}
}

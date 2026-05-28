// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oceanbase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/external"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"
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
		assert.Equal(t, tc.want, util.ObfuscateSQL(tc.sql))
	}
}

func TestOBVersionGreaterOrEqualThan(t *testing.T) {
	cases := []struct {
		name      string
		version   string
		target    string
		wantMatch bool
	}{
		{
			name:      "v4 matches",
			version:   "4.2.1.11",
			target:    "4.0.0",
			wantMatch: true,
		},
		{
			name:      "v3 does not match",
			version:   "3.2.4.3",
			target:    "4.0.0",
			wantMatch: false,
		},
		{
			name:      "empty version",
			version:   "",
			target:    "4.0.0",
			wantMatch: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := &Input{obVersion: tc.version}
			assert.Equal(t, tc.wantMatch, ipt.isOBVersionGreaterOrEqualThan(tc.target))
		})
	}
}

func TestVersionedSQL(t *testing.T) {
	v3 := &Input{obVersion: "3.2.4.3"}
	assert.Equal(t, sqlTenantNamesV3, v3.tenantNamesSQL())
	assert.Equal(t, SQLPlanCacheV3, v3.planCacheSQL())
	assert.Equal(t, SQLClogV3, v3.clogSQL())

	v4 := &Input{obVersion: "4.2.1.11"}
	assert.Equal(t, sqlTenantNamesV4, v4.tenantNamesSQL())
	assert.Equal(t, SQLPlanCacheV4, v4.planCacheSQL())
	assert.Equal(t, SQLClogV4, v4.clogSQL())
}

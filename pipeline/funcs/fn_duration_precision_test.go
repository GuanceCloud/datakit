// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestDurationPrecision(t *testing.T) {
	cases := []struct {
		name, pl, in string
		outkey       string
		expect       interface{}
		fail         bool
	}{
		{
			name: "cast int",
			in:   `{"ts":12345}`,
			pl: `
	json(_, ts)
	cast(ts, "int")
	duration_precision(ts, "ms", "ns")
	`,
			outkey: "ts",
			expect: int64(12345000000),
			fail:   false,
		},
		{
			name: "cast int",
			in:   `{"ts":12345000}`,
			pl: `
	json(_, ts)
	cast(ts, "int")
	duration_precision(ts, "ms", "s")
	`,
			outkey: "ts",
			expect: int64(12345),
			fail:   false,
		},
		{
			name: "cast int",
			in:   `{"ts":12345000}`,
			pl: `
	json(_, ts)
	cast(ts, "int")
	duration_precision(ts, "s", "s")
	`,
			outkey: "ts",
			expect: int64(12345000),
			fail:   false,
		},
		{
			name: "cast int",
			in:   `{"ts":12345000}`,
			pl: `
	json(_, ts)
	cast(ts, "int")
	duration_precision(ts, "ns", "us")
	`,
			outkey: "ts",
			expect: int64(12345),
			fail:   false,
		},
	}

	for idx, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.pl)
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}
			ret, err := runner.Run("test", map[string]string{},
				map[string]interface{}{
					"message": tc.in,
				}, "message", time.Now())
			if err != nil || ret.Error != nil {
				t.Error(err, " ", ret.Error)
				return
			}
			t.Log(ret)
			v := ret.Fields[tc.outkey]
			tu.Equals(t, tc.expect, v)
			t.Logf("[%d] PASS", idx)
		})
	}
}

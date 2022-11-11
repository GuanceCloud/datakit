// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"
)

func TestSample(t *testing.T) {
	cases := []struct {
		name, pl, in string
		fail         bool
		outKey       string
	}{
		{
			name: "fifty-fifty",
			in:   `dummy input`,
			pl: `
if sample(0.5) {
	add_key(hello, "world")
}
			`,
			fail:   false,
			outKey: "hello",
		},
		{
			name: "definite",
			in:   `dummy input`,
			pl: `
if sample(1) {
	add_key(hello, "world")
}
			`,
			fail:   false,
			outKey: "hello",
		},
		{
			name: "expression as arg",
			in:   `dummy input`,
			pl: `
if sample(2 * 0.1) {
	add_key(hello, "world")
}
			`,
			fail:   false,
			outKey: "hello",
		},
		{
			name: "negative probability",
			in:   `dummy input`,
			pl: `
sample(-0.5)
			`,
			fail: true,
		},
		{
			name: "probability out of range",
			in:   `dummy input`,
			pl: `
sample(2)
			`,
			fail: true,
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
			_, _, f, _, _, err := runScript(runner,
				"test", nil, map[string]any{"message": tc.in}, time.Now())
			if err == nil {
				if v, has := f[tc.outKey]; has {
					t.Logf("k/v pair `%s = %s` has been added to output", tc.outKey, v)
				}
				t.Logf("[%d] PASS", idx)
			} else {
				t.Error(err)
			}
		})
	}
}

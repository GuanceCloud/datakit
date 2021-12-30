package funcs

import (
	"testing"

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

			if err := runner.Run(tc.in); err != nil {
				t.Error(err)
				return
			}
			t.Log(runner.Result())
			v, _ := runner.GetContent(tc.outkey)
			tu.Equals(t, tc.expect, v)
			t.Logf("[%d] PASS", idx)
		})
	}
}

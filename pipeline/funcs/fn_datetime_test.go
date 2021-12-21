package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestDateTime(t *testing.T) {
	// local timezone: utc+0800
	cst := time.FixedZone("CST", 8*3600)
	time.Local = cst

	cases := []struct {
		name, pl, in string
		outkey       string
		expect       interface{}
		fail         bool
	}{
		{
			name: "ANSIC s",
			in:   `{"a":{"timestamp": "1638253518", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 's', 'ANSIC')
	`,
			outkey: "a.timestamp",
			expect: "Tue Nov 30 14:25:18 2021",
			fail:   false,
		},
		{
			name: "ANSIC ms",
			in:   `{"a":{"timestamp": "1638253518000", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ms', 'ANSIC')
	`,
			outkey: "a.timestamp",
			expect: "Tue Nov 30 14:25:18 2021",
			fail:   false,
		},
		{
			name: "UnixDate s",
			in:   `{"a":{"timestamp": "1638253518", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 's', 'UnixDate')
	`,
			outkey: "a.timestamp",
			expect: "Tue Nov 30 14:25:18 CST 2021",
			fail:   false,
		},
		{
			name: "UnixDate ms",
			in:   `{"a":{"timestamp": "1638253518999", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ms', 'UnixDate')
	`,
			outkey: "a.timestamp",
			expect: "Tue Nov 30 14:25:18 CST 2021",
			fail:   false,
		},
		{
			name: "RubyDate ms",
			in:   `{"a":{"timestamp": "1638253518999", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ms', 'RubyDate')
	`,
			outkey: "a.timestamp",
			expect: "Tue Nov 30 14:25:18 +0800 2021",
			fail:   false,
		},
		{
			name: "RubyDate s",
			in:   `{"a":{"timestamp": "1638253518", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 's', 'RubyDate')
	`,
			outkey: "a.timestamp",
			expect: "Tue Nov 30 14:25:18 +0800 2021",
			fail:   false,
		},
		{
			name: "RFC822 ms",
			in:   `{"a":{"timestamp": "1638253518999", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ms', 'RFC822')
	`,
			outkey: "a.timestamp",
			expect: "30 Nov 21 14:25 CST",
			fail:   false,
		},
		{
			name: "RFC822 s",
			in:   `{"a":{"timestamp": "1638253518", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 's', 'RFC822')
	`,
			outkey: "a.timestamp",
			expect: "30 Nov 21 14:25 CST",
			fail:   false,
		},
		{
			name: "RFC822Z ms",
			in:   `{"a":{"timestamp": "1638253518999", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ms', 'RFC822Z')
	`,
			outkey: "a.timestamp",
			expect: "30 Nov 21 14:25 +0800",
			fail:   false,
		},
		{
			name: "RFC822Z s",
			in:   `{"a":{"timestamp": "1638253518", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 's', 'RFC822Z')
	`,
			outkey: "a.timestamp",
			expect: "30 Nov 21 14:25 +0800",
			fail:   false,
		},
		{
			name: "RFC850 ms",
			in:   `{"a":{"timestamp": "1638253518999", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ms', 'RFC850')
	`,
			outkey: "a.timestamp",
			expect: "Tuesday, 30-Nov-21 14:25:18 CST",
			fail:   false,
		},
		{
			name: "RFC850 s",
			in:   `{"a":{"timestamp": "1638253518", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 's', 'RFC850')
	`,
			outkey: "a.timestamp",
			expect: "Tuesday, 30-Nov-21 14:25:18 CST",
			fail:   false,
		},
		{
			name: "RFC1123 ms",
			in:   `{"a":{"timestamp": "1638253518999", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ms', 'RFC1123')
	`,
			outkey: "a.timestamp",
			expect: "Tue, 30 Nov 2021 14:25:18 CST",
			fail:   false,
		},
		{
			name: "RFC1123 s",
			in:   `{"a":{"timestamp": "1638253518", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 's', 'RFC1123')
	`,
			outkey: "a.timestamp",
			expect: "Tue, 30 Nov 2021 14:25:18 CST",
			fail:   false,
		},
		{
			name: "RFC1123Z ms",
			in:   `{"a":{"timestamp": "1638253518999", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ms', 'RFC1123Z')
	`,
			outkey: "a.timestamp",
			expect: "Tue, 30 Nov 2021 14:25:18 +0800",
			fail:   false,
		},
		{
			name: "RFC1123Z s",
			in:   `{"a":{"timestamp": "1638253518", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 's', 'RFC1123Z')
	`,
			outkey: "a.timestamp",
			expect: "Tue, 30 Nov 2021 14:25:18 +0800",
			fail:   false,
		},
		{
			name: "RFC3339 s",
			in:   `{"a":{"timestamp": "1610960605", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 's', 'RFC3339')
	`,
			outkey: "a.timestamp",
			expect: "2021-01-18T17:03:25+08:00",
			fail:   false,
		},
		{
			name: "RFC3339 ms",
			in:   `{"a":{"timestamp": "1610960605000", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ms', 'RFC3339')
	`,
			outkey: "a.timestamp",
			expect: "2021-01-18T17:03:25+08:00",
			fail:   false,
		},
		{
			name: "RFC3339Nano s",
			in:   `{"a":{"timestamp": "1610960605", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 's', 'RFC3339Nano')
	`,
			outkey: "a.timestamp",
			expect: "2021-01-18T17:03:25+08:00",
			fail:   false,
		},
		{
			name: "RFC3339Nano ms",
			in:   `{"a":{"timestamp": "1610960605001", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ms', 'RFC3339Nano')
	`,
			outkey: "a.timestamp",
			expect: "2021-01-18T17:03:25.001+08:00",
			fail:   false,
		},
		{
			name: "Kitchen ms",
			in:   `{"a":{"timestamp": "1610960605001", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ms', 'Kitchen')
	`,
			outkey: "a.timestamp",
			expect: "5:03PM",
			fail:   true,
		},
		{
			name: "Kitchen s",
			in:   `{"a":{"timestamp": "1610960605", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 's', 'Kitchen')
	`,
			outkey: "a.timestamp",
			expect: "5:03PM",
			fail:   true,
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

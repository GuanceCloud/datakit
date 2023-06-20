// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	tu "github.com/GuanceCloud/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
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
		},
		{
			name: "udef_ms",
			in:   `{"a":{"timestamp": "1610960605000", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ms', '%Y-%m-%d')
	`,
			outkey: "a.timestamp",
			expect: "2021-01-18",
		},
		{
			name: "udef_us",
			in:   `{"a":{"timestamp": "1610960605000000", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'us', '%Y-%m-%d %H:%M:%S')
	`,
			outkey: "a.timestamp",
			expect: "2021-01-18 17:03:25",
		},
		{
			name: "udef_ns",
			in:   `{"a":{"timestamp": "1610960605000000000", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ns', '%Y-%m-%d %H:%M:%S')
	`,
			outkey: "a.timestamp",
			expect: "2021-01-18 17:03:25",
		},
		{
			name: "udef_ns_tz_1",
			in:   `{"a":{"timestamp": "1610960605000000000", "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ns', '%Y-%m-%d %H:%M:%S', tz="Asia/Tokyo")
	`,
			outkey: "a.timestamp",
			expect: "2021-01-18 18:03:25",
		},
		{
			name: "udef_ns_tz_2",
			in:   `{"a":{"timestamp": 1610960605000000000, "second":2},"age":47}`,
			pl: `
	json(_, a.timestamp)
	datetime(a.timestamp, 'ns', fmt='%Y-%m-%d %H:%M:%S', tz="UTC")
	`,
			outkey: "a.timestamp",
			expect: "2021-01-18 09:03:25",
		},
	}

	for idx, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.pl)
			if tc.fail && err == nil {
				t.Error("unknown error")
			}
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}
			pt := ptinput.NewPlPoint(
				point.Logging, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			if errR != nil {
				t.Fatal(errR)
			}

			if tc.fail {
				t.Logf("[%d]expect error: %s", idx, err)
			}
			v, _, _ := pt.Get(tc.outkey)
			tu.Equals(t, tc.expect, v)
			t.Logf("[%d] PASS", idx)
		})
	}
}

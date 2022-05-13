// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const Hour8 = int64(8 * time.Hour)

func TestAdjustTimezone(t *testing.T) {
	// local timezone: utc+0800
	cst := time.FixedZone("CST", 8*3600)
	time.Local = cst

	tn := time.Now().Add(time.Minute * 10)
	cases := []struct {
		name, pl, in string
		outkey       string
		expect       interface{}
		fail         bool
	}{
		{
			name: "time fmt: ANSIC",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format(time.ANSIC)),
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() - int64(tn.Nanosecond()) - Hour8,
			fail:   false,
		},
		{
			name: "time fmt: ANSIC",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format(time.ANSIC)),
			pl: `
			json(_, time)
			default_time(time)
			adjust_timezone(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() - int64(tn.Nanosecond()),
			fail:   false,
		},
		{
			name: "nginx log datetime, 02/Jan/2006:15:04:05 -0700",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format("02/Jan/2006:15:04:05 -0700")),
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() - int64(tn.Nanosecond()),
			fail:   false,
		},
		{
			name: "[auto] nginx log datetime, 02/Jan/2006:15:04:05 -0700",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format("02/Jan/2006:15:04:05 -0700")),
			pl: `
			json(_, time)
			default_time(time)
			adjust_timezone(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() - int64(tn.Nanosecond()),
			fail:   false,
		},
		{
			name: "redis log datetime, 02 Jan 2006 15:04:05.000",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format("02 Jan 2006 15:04:05.000")),
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: tn.UnixNano()/1000000*1000000 - Hour8,
			fail:   false,
		},
		{
			name: "[auto] redis log datetime, 02 Jan 2006 15:04:05.000",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format("02 Jan 2006 15:04:05.000")),
			pl: `
			json(_, time)
			default_time(time)
			adjust_timezone(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() / 1000000 * 1000000,
			fail:   false,
		},
		{
			name: "mysql log datetime, 060102 15:04:05",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format("060102 15:04:05")),
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() - int64(tn.Nanosecond()) - Hour8,
			fail:   false,
		},
		{
			name: "[auto] mysql log datetime, 060102 15:04:05",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format("060102 15:04:05")),
			pl: `
			json(_, time)
			default_time(time)
			adjust_timezone(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() - int64(tn.Nanosecond()),
			fail:   false,
		},
		{
			name: "gin log datetime, 2006/01/02 - 15:04:05",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format("2006/01/02 - 15:04:05")),
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() - int64(tn.Nanosecond()) - Hour8,
			fail:   false,
		},
		{
			name: "[auto] gin log datetime, 2006/01/02 - 15:04:05",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format("2006/01/02 - 15:04:05")),
			pl: `
			json(_, time)
			default_time(time)
			adjust_timezone(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() - int64(tn.Nanosecond()),
			fail:   false,
		},
		{
			name: "apache log datetime, Mon Jan 2 15:04:05.000000 2006",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format("Mon Jan 2 15:04:05.000000 2006")),
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: tn.UnixNano()/1000*1000 - Hour8,
			fail:   false,
		},
		{
			name: "[auto] apache log datetime, Mon Jan 2 15:04:05.000000 2006",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format("Mon Jan 2 15:04:05.000000 2006")),
			pl: `
			json(_, time)
			default_time(time)
			adjust_timezone(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() / 1000 * 1000,
			fail:   false,
		},
		{
			name: "1 postgresql log datetime, 2006-01-02 15:04:05.000 UTC",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format("2006-01-02 15:04:05.000 UTC")),
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: tn.UnixNano()/1000000*1000000 - Hour8,
			fail:   false,
		},
		{
			name: "2 [auto] postgresql log datetime, 2006-01-02 15:04:05.000 UTC",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format("2006-01-02 15:04:05.000 UTC")),
			pl: `
			json(_, time)
			default_time(time)
			adjust_timezone(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() / 1000000 * 1000000,
			fail:   false,
		},
		/* remove temporary
		{
			name: "3 postgresql log datetime, 2006-01-02 15:04:05.000 UTC",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Add(9*time.Hour).Format("2006-01-02 15:04:05.000 UTC")),
			pl: `
			json(_, time)
			default_time(time, "Asia/Tokyo")
		`, // utc +0900
			outkey: "time",
			expect: tn.UnixNano() / 1000000 * 1000000,
			fail:   false,
		}, */
		{
			name: "4 postgresql log datetime, 2006-01-02 15:04:05.000 UTC",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Add(-time.Duration(Hour8)).Format("2006-01-02 15:04:05.000 UTC")),
			pl: `
			json(_, time)
			default_time(time)
		`, // utc -0800
			outkey: "time",
			expect: tn.UnixNano()/1000000*1000000 - 2*Hour8,
			fail:   false,
		},
		{
			name: "5 [auto] postgresql log datetime, 2006-01-02 15:04:05.000 UTC",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Add(-8*time.Hour).Format("2006-01-02 15:04:05.000 UTC")),
			pl: `
			json(_, time)
			default_time(time)
			adjust_timezone(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() / 1000000 * 1000000,
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
			pt, _ := io.MakePoint("test", map[string]string{},
				map[string]interface{}{
					"message": tc.in,
				}, time.Now())
			ret, err := runner.Run(pt)
			tu.Equals(t, nil, err)
			t.Log(ret)
			v := ret.Fields[tc.outkey]
			tu.Equals(t, nil, err)
			tu.Equals(t, tc.expect, v)
			t.Logf("[%d] PASS", idx)
		})
	}
}

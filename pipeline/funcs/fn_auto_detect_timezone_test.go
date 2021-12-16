package funcs

import (
	"fmt"
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

const Hour8 = 8 * timeHourNanosec

func TestAutoDetectTimezone(t *testing.T) {
	// local timezone: utc+0800
	cst := time.FixedZone("CST", 8*3600)
	time.Local = cst

	tn := time.Now()
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
			auto_detect_timezone(time)
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
			auto_detect_timezone(time)
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
			auto_detect_timezone(time)
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
			auto_detect_timezone(time)
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
			auto_detect_timezone(time)
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
			auto_detect_timezone(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() / 1000 * 1000,
			fail:   false,
		},
		{
			name: "postgresql log datetime, 2006-01-02 15:04:05.000 UTC",
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
			name: "[auto] postgresql log datetime, 2006-01-02 15:04:05.000 UTC",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Format("2006-01-02 15:04:05.000 UTC")),
			pl: `
			json(_, time)
			default_time(time)
			auto_detect_timezone(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() / 1000000 * 1000000,
			fail:   false,
		},
		{
			name: "postgresql log datetime, 2006-01-02 15:04:05.000 UTC",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Add(-8*time.Hour).Format("2006-01-02 15:04:05.000 UTC")),
			pl: `
			json(_, time)
			default_time(time, "America/Los_Angeles")
		`, // utc -0800
			outkey: "time",
			expect: tn.UnixNano() / 1000000 * 1000000,
			fail:   false,
		},
		{
			name: "postgresql log datetime, 2006-01-02 15:04:05.000 UTC",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Add(-8*time.Hour).Format("2006-01-02 15:04:05.000 UTC")),
			pl: `
			json(_, time)
			default_time(time)
		`, // utc -0800
			outkey: "time",
			expect: tn.UnixNano()/1000000*1000000 - 2*Hour8,
			fail:   false,
		},
		{
			name: "[auto] postgresql log datetime, 2006-01-02 15:04:05.000 UTC",
			in:   fmt.Sprintf(`{"time":"%s"}`, tn.UTC().Add(-8*time.Hour).Format("2006-01-02 15:04:05.000 UTC")),
			pl: `
			json(_, time)
			default_time(time)
			auto_detect_timezone(time)
		`,
			outkey: "time",
			expect: tn.UnixNano() / 1000000 * 1000000,
			fail:   false,
		},
		// {
		// 	name: "10 Dec 2021 03:49:20.937",
		// 	in: `
		// 	{
		// 		"time":"10 Dec 2021 03:49:20.937",
		// 		"second":2,
		// 		"third":"abc",
		// 		"forth":true
		// 	}
		// 	`,
		// 	pl: `
		// 	json(_, time)
		// 	default_time(time)
		// 	auto_detect_timezone(time)
		// `,
		// 	outkey: "time",
		// 	expect: int64(1639108160937000000),
		// 	fail:   false,
		// },
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

			err = runner.Run(tc.in)
			tu.Equals(t, nil, err)
			t.Log(runner.Result())

			v, err := runner.GetContent(tc.outkey)
			tu.Equals(t, nil, err)
			tu.Equals(t, tc.expect, v)

			t.Logf("[%d] PASS", idx)
		})
	}
}

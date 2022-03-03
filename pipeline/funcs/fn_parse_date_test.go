package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestParseDate(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name     string
		pl, in   string
		outKey   string
		expected int64
		fail     bool
	}{
		{
			name: "normal",
			pl: `
grok(_, "%{INT:hour}:%{INT:min}:%{INT:sec}\\.%{INT:msec}")
parse_date("time", "", "", "", hour, min, sec, msec, "", "", "+8")
`,
			in:     "16:40:03.290",
			outKey: "time",
			expected: time.Date(now.Year(), now.Month(), now.Day(),
				16, 40, 0o3, 290*1000*1000, time.FixedZone("UTC+8", 8*60*60)).UnixNano(),
			fail: false,
		},

		{
			name: "normal",
			pl: `
grok(_, "%{INT:hour}:%{INT:min}:%{INT:sec}\\.%{INT:msec}")
parse_date(key="time", h=hour, M=min, s=sec, ms=msec, zone="+8")
`,
			in:     "16:40:03.290",
			outKey: "time",
			expected: time.Date(now.Year(), now.Month(), now.Day(),
				16, 40, 0o3, 290*1000*1000, time.FixedZone("UTC+8", 8*60*60)).UnixNano(),
			fail: false,
		},

		{
			name: "normal",
			pl: `
grok(_, "%{INT:year}-%{INT:month}-%{INT:day} %{INT:hour}:%{INT:min}:%{INT:sec}\\.%{INT:msec}")
parse_date(key="time", y=year, m=month, d=day, h=hour, M=min, s=sec, ms=msec, zone="+8")
`,
			in:     "2020-12-12 16:40:03.290",
			outKey: "time",
			expected: time.Date(2020, 12, 12,
				16, 40, 0o3, 290*1000*1000, time.FixedZone("UTC+8", 8*60*60)).UnixNano(),
			fail: false,
		},

		{
			name: "normal",
			pl: `
grok(_, "%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day} %{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}")
parse_date(key="time", y=year, m=month, d=day, h=hour, M=min, s=sec, zone=tz)
`,
			in:     "Mon Sep  6 16:40:03 CST 2021",
			outKey: "time",
			expected: time.Date(2021, 9, 6,
				16, 40, 0o3, 0, time.FixedZone("UTC+8", 8*60*60)).UnixNano(),
			fail: false,
		},

		{
			name: "normal",
			pl: `
grok(_, "%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day}\\s+%{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}")
parse_date(key="time", y=year, m=month, d=day, h=hour, M=min, s=sec, zone=tz)
`,
			in:     "Mon Sep  6 16:40:03 CST 2021",
			outKey: "time",
			expected: time.Date(2021, 9, 6,
				16, 40, 0o3, 0, time.FixedZone("UTC+8", 8*60*60)).UnixNano(),
			fail: false,
		},

		{
			name: "partial year",
			pl: `
grok(_, "%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day}\\s+%{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}")
parse_date(key="time", y=year, m=month, d=day, h=hour, M=min, s=sec, zone=tz)
`,
			in:     "Mon Sep  6 16:40:03 CST 00",
			outKey: "time",
			expected: time.Date(2000, 9, 6,
				16, 40, 0o3, 0, time.FixedZone("UTC+8", 8*60*60)).UnixNano(),
		},

		{
			name: "partial year",
			pl: `
grok(_, "%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day}\\s+%{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}")
parse_date(key="time", y=year, m=month, d=day, h=hour, M=min, s=sec, zone=tz)
`,
			in:     "Mon Sep  6 16:40:03 CST 09",
			outKey: "time",
			expected: time.Date(2009, 9, 6,
				16, 40, 0o3, 0, time.FixedZone("UTC+8", 8*60*60)).UnixNano(),
		},

		{
			name: "us",
			pl: `
grok(_, "%{INT:hour}:%{INT:min}:%{INT:sec}\\.%{INT:us}")
parse_date("time", "", "", "", hour, min, sec, "", us, "", "+8")
`,
			in:     "16:40:03.290290",
			outKey: "time",
			expected: time.Date(now.Year(), now.Month(), now.Day(),
				16, 40, 0o3, 290290*1000, time.FixedZone("UTC+8", 8*60*60)).UnixNano(),
		},

		{
			name: "ns",
			pl: `
grok(_, "%{INT:hour}:%{INT:min}:%{INT:sec}\\.%{INT:ns}")
parse_date("time", "", "", "", hour, min, sec, "", "", ns, "+8")
`,
			in:     "16:40:03.290290330",
			outKey: "time",
			expected: time.Date(now.Year(), now.Month(), now.Day(),
				16, 40, 0o3, 290290330, time.FixedZone("UTC+8", 8*60*60)).UnixNano(),
		},

		{
			name: "ns for kwargs",
			pl: `
grok(_, "%{INT:hour}:%{INT:min}:%{INT:sec}\\.%{INT:ns}")
parse_date(key="time", h=hour, M=min, s=sec, ns=ns, zone="CST")
`,
			in:     "16:40:03.290290330",
			outKey: "time",
			expected: time.Date(now.Year(), now.Month(), now.Day(),
				16, 40, 0o3, 290290330, time.FixedZone("UTC+8", 8*60*60)).UnixNano(),
		},

		{
			name: "missed second: use 0",
			pl: `
grok(_, "%{INT:hour}:%{INT:min}")
parse_date(key="time", h=hour, M=min, zone="CST")`,
			in:     "16:40",
			outKey: "time",
			expected: time.Date(now.Year(), now.Month(), now.Day(),
				16, 40, 0o0, 0, time.FixedZone("UTC+8", 8*60*60)).UnixNano(),
		},

		{
			name: "normal",
			pl: `
grok(_, "%{INT:hour}:%{INT:min}:%{INT:sec}\\.%{INT:msec}")
parse_date("time", "", "", "", hour, min, sec, msec, "", "", "")
`,
			in:     "16:40:03.290",
			outKey: "time",
			expected: time.Date(now.Year(), now.Month(), now.Day(),
				16, 40, 0o3, 290*1000*1000, time.FixedZone("UTC", 0)).UnixNano(),
			fail: false,
		},

		{
			name: "invalid hour",
			pl: `
grok(_, "%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day}\\s+%{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}")
parse_date(key="time", y=year, m=month, d=day, h=hour, M=min, s=sec, zone=tz)
`,
			in:   "Mon Sep  6 25:40:03 CST 2021",
			fail: true,
		},

		{
			name: "invalid hour",
			pl: `
grok(_, "%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day}\\s+%{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}")
parse_date(key="time", y=year, m=month, d=day, h=hour, M=min, s=sec, zone=tz)
`,
			in:   "Mon Sep  6 -2:40:03 CST 2021",
			fail: true,
		},

		{
			name: "invalid minute",
			pl: `
grok(_, "%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day}\\s+%{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}")
parse_date(key="time", y=year, m=month, d=day, h=hour, M=min, s=sec, zone=tz)
`,
			in:   "Mon Sep  6 12:61:03 CST 2021",
			fail: true,
		},

		{
			name: "invalid minute",
			pl: `
grok(_, "%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day}\\s+%{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}")
parse_date(key="time", y=year, m=month, d=day, h=hour, M=min, s=sec, zone=tz)
`,
			in:   "Mon Sep  6 12:60:03 CST 2021",
			fail: true,
		},

		{
			name: "invalid second",
			pl: `
grok(_, "%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day}\\s+%{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}")
parse_date(key="time", y=year, m=month, d=day, h=hour, M=min, s=sec, zone=tz)
`,
			in:   "Mon Sep  6 12:11:61 CST 2021",
			fail: true,
		},

		{
			name: "invalid minute",
			pl: `
grok(_, "%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day}\\s+%{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}")
parse_date(key="time", y=year, m=month, d=day, h=hour, M=min, s=sec, zone=tz)
`,
			in:   "Mon Sep  6 12:12:-1 CST 2021",
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

			err = runner.Run(tc.in)
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Error(err)
				}
			} else {
				t.Log(runner.Result())
				v, _ := runner.GetContent(tc.outKey)
				tu.Equals(t, tc.expected, v)
				t.Logf("[%d] PASS", idx)
			}
		})
	}
}

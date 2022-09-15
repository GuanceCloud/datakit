// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/engine"
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
		expect       time.Time
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
			expect: time.Unix(0, tn.UnixNano()-int64(tn.Nanosecond())-Hour8),
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
			expect: time.Unix(0, tn.UnixNano()-int64(tn.Nanosecond())-int64(time.Hour)),
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
			_, _, f, tn, _, err := engine.RunScript(runner, "test", nil, map[string]interface{}{
				"message": tc.in,
			}, time.Now())

			tu.Equals(t, nil, err)
			t.Log(f)
			var v interface{}
			if tc.outkey != "time" {
				v = f[tc.outkey]
			} else {
				v = tn
			}
			tu.Equals(t, nil, err)
			assert.Equal(t, tc.expect, v)
			t.Logf("[%d] PASS", idx)
		})
	}
}

func TestDetectTimezone(t *testing.T) {
	// logTS, nowTS, minuteAllow, expTS
	tn := time.Date(2022, 6, 1, 10, 11, 12, int(time.Millisecond), time.Local)

	cases := []struct {
		name          string
		logTS         int64
		nowTS         int64
		durationAllow int64
		expTS         int64
		neq           int64
	}{
		{
			name: "id 1, hour -2, minute -2, sec -2, nanosec -15",
			// log: 2022-6-1 10:11:12.001
			// now: 2022-6-1 12:11:14.014
			// exp: 2022-6-1 12:11:12.001
			logTS:         time.Date(2022, 6, 1, 10, 11, 12, int(time.Millisecond), time.Local).UnixNano(),
			nowTS:         time.Date(2022, 6, 1, 12, 11, 14, int(time.Millisecond)*15, time.Local).UnixNano(),
			expTS:         time.Date(2022, 6, 1, 12, 11, 12, int(time.Millisecond), time.Local).UnixNano(),
			durationAllow: defaultMinuteDelta * int64(time.Minute),
		},
		{
			name: "id 2, hour -2, minute 0, sec 0, nanosec +5",
			// log: 2022-6-1 10:11:14.02
			// now: 2022-6-1 12:11:14.014
			// exp: 2022-7-5 12:11:14.02
			logTS:         time.Date(2022, 6, 1, 10, 11, 14, int(time.Millisecond)*20, time.Local).UnixNano(),
			nowTS:         time.Date(2022, 6, 1, 12, 11, 14, int(time.Millisecond)*15, time.Local).UnixNano(),
			expTS:         time.Date(2022, 6, 1, 12, 11, 14, int(time.Millisecond)*20, time.Local).UnixNano(),
			durationAllow: defaultMinuteDelta * int64(time.Minute),
		},
		{
			name:          "id 3, hour -2, minute +2, sec +1",
			logTS:         tn.Add(time.Minute*2 + time.Second - 2*time.Hour).UnixNano(),
			nowTS:         tn.UnixNano(),
			expTS:         tn.Add(time.Minute*2 + time.Second - time.Hour).UnixNano(),
			durationAllow: defaultMinuteDelta * int64(time.Minute),
		},
		{
			name:          "id 3, hour -1, minute -58, sec 0",
			logTS:         time.Date(2022, 6, 1, 11, 1, 1, 0, time.Local).UnixNano(),
			nowTS:         time.Date(2022, 6, 1, 12, 59, 1, 0, time.Local).UnixNano(),
			expTS:         time.Date(2022, 6, 1, 13, 1, 1, 0, time.Local).UnixNano(),
			durationAllow: defaultMinuteDelta * int64(time.Minute),
		},
	}
	for _, v := range cases {
		tsAct := detectTimezone(v.logTS, v.nowTS, v.durationAllow)

		if v.neq != 0 {
			tsAct += v.neq
		}
		assert.Equal(t, time.Unix(0, v.expTS),
			time.Unix(0, tsAct),
			fmt.Sprintf(v.name))
	}
}

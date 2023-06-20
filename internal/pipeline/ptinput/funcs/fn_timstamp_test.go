// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"math"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
)

func TestTimestamp(t *testing.T) {
	cases := []struct {
		name, pl, in string
		outkey       string
		expect       interface{}
		fail         bool
	}{
		{
			name:   "timestamp_default",
			in:     `time_now`,
			pl:     `add_key(time1, timestamp())`,
			outkey: "time1",
			expect: time.Now().UnixNano(),
			fail:   false,
		},
		{
			name:   "time_now_ns",
			in:     `time_now`,
			pl:     `add_key(time1, timestamp("ns"))`,
			outkey: "time1",
			expect: time.Now().UnixNano(),
			fail:   false,
		},
		{
			name:   "time_now_us",
			in:     `time_now`,
			pl:     `add_key(time1, timestamp("us")* 1000)`,
			outkey: "time1",
			expect: time.Now().UnixNano(),
			fail:   false,
		},
		{
			name:   "time_now_ms",
			in:     `time_now`,
			pl:     `add_key(time1, timestamp("ms")*1000000)`,
			outkey: "time1",
			expect: time.Now().UnixNano(),
			fail:   false,
		},
		{
			name:   "time_now_s",
			in:     `time_now`,
			pl:     `add_key(time1, timestamp("s")*1000000000)`,
			outkey: "time1",
			expect: time.Now().UnixNano(),
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

			pt := ptinput.NewPlPoint(
				point.Logging, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			if errR != nil {
				t.Fatal(errR.Error())
			}

			v, _, _ := pt.Get(tc.outkey)
			ts_act, ok := v.(int64)
			if !ok {
				t.Fatalf("type of %s is not int64", tc.outkey)
			}
			switch expect := tc.expect.(type) {
			case int64:
				assert.GreaterOrEqual(t, 1e9, math.Abs(float64(expect-ts_act)))
			default:
				t.Fatal("undefined action")
			}
			t.Logf("[%d] PASS", idx)
		})
	}
}

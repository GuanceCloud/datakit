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

func TestParseInt(t *testing.T) {
	cases := []struct {
		name, pl, in string
		expected     interface{}
		fail         bool
		outkey       string
	}{
		{
			name: "format",
			pl: `a = 17
			b = format_int(a, 16)
			if b != "11" {
				add_key(abc, b)
				exit()
			}
			c = parse_int(b, 16)
			if c != a {
				add_key(abc, c)
				exit()
			} else {
				add_key(abc, "ok")
			}
			`,
			in:       `test`,
			expected: "ok",
			outkey:   "abc",
		},
		{
			name: "parse",
			pl: `a = "11" # 0x11
 			b = parse_int(a, 16)
			if b != 17 {
				add_key(abc, b)
				exit()
			}
			c = format_int(b, 16)
			if c != a {
				add_key(abc, c)
				exit()
			} else {
				add_key(abc, "ok")
			}
			`,
			in:       `test`,
			expected: "ok",
			outkey:   "abc",
		},
		{
			name: "spanid-format",
			pl: `a = 7665324064912355185
			b = format_int(a, 16)
			if b != "6a60b39fd95aaf71" {
				add_key(abc, b)
				exit()
			}
			c = parse_int(b, 16)
			if c != a {
				add_key(abc, c)
				exit()
			} else {
				add_key(abc, "ok")
			}
			`,
			in:       `test`,
			expected: "ok",
			outkey:   "abc",
		},
		{
			name: "spanid-format-all-in-one",
			pl: `a = "7665324064912355185"
			b = format_int(parse_int(a, 10), 16)
			if b != "6a60b39fd95aaf71" {
				add_key(abc, b)
				exit()
			} else {
				add_key(abc, "ok")
			}
			`,
			in:       `test`,
			expected: "ok",
			outkey:   "abc",
		},
		{
			name: "parse",
			pl: `a = "6a60b39fd95aaf71" 
 			b = parse_int(a, 16)
			if b != 7665324064912355185 {
				add_key(abc, b)
				exit()
			}
			c = format_int(b, 16)
			if c != a {
				add_key(abc, c)
			} else {
				add_key(abc, "ok")
			}
			`,
			in:       `test`,
			expected: "ok",
			outkey:   "abc",
		},
		{
			name: "parsex",
			pl: `a = "0x6a60b39fd95aaf71" 
 			b = parse_int(a, 0)
			if b != 7665324064912355185 {
				add_key(abc, b)
				exit()
			}
			c = format_int(b, 16)
			if "0x"+c != a {
				add_key(abc, c)
			} else {
				add_key(abc, "ok")
			}
			`,
			in:       `test`,
			expected: "ok",
			outkey:   "abc",
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
			// tu.Equals(t, nil, err)
			tu.Equals(t, tc.expected, v)
			t.Logf("[%d] PASS", idx)
		})
	}
}

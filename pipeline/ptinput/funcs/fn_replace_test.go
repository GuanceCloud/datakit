// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ptinput"
)

func TestReplace(t *testing.T) {
	cases := []struct {
		name     string
		pl, in   string
		outKey   string
		expected interface{}
		fail     bool
	}{
		{
			name: `normal`,
			pl: "json(_, `str`)\n" +
				"replace(`str`, \"(1[0-9]{2})[0-9]{4}([0-9]{4})\", \"$1****$2\")",
			in:       `{"str": "13789123014"}`,
			outKey:   "str",
			fail:     false,
			expected: "137****3014",
		},

		{
			name: `normal`,
			pl: "json(_, `str`)\n" +
				"replace(`str`, \"([a-z]*) \\\\w*\", \"$1 ***\")",
			in:       `{"str": "zhang san"}`,
			outKey:   "str",
			expected: "zhang ***",
			fail:     false,
		},

		{
			name: `normal`,
			pl: "json(_, `str`)\n" +
				"replace(`str`, \"([1-9]{4})[0-9]{10}([0-9]{4})\", \"$1**********$2\")",
			in:       `{"str": "362201200005302565"}`,
			outKey:   "str",
			expected: "3622**********2565",
			fail:     false,
		},

		{
			name: `normal`,
			pl: "json(_, `str`)\n" +
				"replace(`str`, '([\u4e00-\u9fa5])[\u4e00-\u9fa5]([\u4e00-\u9fa5])', \"$1＊$2\")",
			in:       `{"str": "小阿卡"}`,
			outKey:   "str",
			expected: "小＊卡",
			fail:     false,
		},
		{
			name: `normal`,
			pl: "json(_, `str`)\n" +
				"replace(str1, '([\u4e00-\u9fa5])[\u4e00-\u9fa5]([\u4e00-\u9fa5])', \"$1＊$2\")",
			in:       `{"str": "小阿卡"}`,
			outKey:   "str",
			expected: "小阿卡",
			fail:     false,
		},
		{
			name: `not enough args`,
			pl: "json(_, `str`)\n" +
				"replace(`str`, '([\u4e00-\u9fa5])[\u4e00-\u9fa5]([\u4e00-\u9fa5])')",
			in:   `{"str": "小阿卡"}`,
			fail: true,
		},
		{
			name: `invalid arg type`,
			pl: "json(_, `str`)\n" +
				"replace(`str`, 2, \"$1＊$2\")",
			in:   `{"str": "小阿卡"}`,
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

			pt := ptinput.GetPoint()
			ptinput.InitPt(pt, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			if errR != nil {
				ptinput.PutPoint(pt)
				t.Fatal(errR)
			}

			if v, ok := pt.Fields[tc.outKey]; !ok {
				if !tc.fail {
					t.Errorf("[%d]expect error: %s", idx, errR.Error())
				}
			} else {
				tu.Equals(t, tc.expected, v)
				t.Logf("[%d] PASS", idx)
			}

			t.Logf("[%d] PASS", idx)
			ptinput.PutPoint(pt)
		})
	}
}

package funcs

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
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
			name:     `normal`,
			pl:       `json(_, str) replace(str, "(1[0-9]{2})[0-9]{4}([0-9]{4})", "$1****$2")`,
			in:       `{"str": "13789123014"}`,
			outKey:   "str",
			fail:     false,
			expected: "137****3014",
		},

		{
			name:     `normal`,
			pl:       `json(_, str) replace(str, "([a-z]*) \\w*", "$1 ***")`,
			in:       `{"str": "zhang san"}`,
			outKey:   "str",
			expected: "zhang ***",
			fail:     false,
		},

		{
			name:     `normal`,
			pl:       `json(_, str) replace(str, "([1-9]{4})[0-9]{10}([0-9]{4})", "$1**********$2")`,
			in:       `{"str": "362201200005302565"}`,
			outKey:   "str",
			expected: "3622**********2565",
			fail:     false,
		},

		{
			name:     `normal`,
			pl:       `json(_, str) replace(str, '([\u4e00-\u9fa5])[\u4e00-\u9fa5]([\u4e00-\u9fa5])', "$1＊$2")`,
			in:       `{"str": "小阿卡"}`,
			outKey:   "str",
			expected: "小＊卡",
			fail:     false,
		},

		{
			name:     `normal`,
			pl:       `json(_, str) replace(str1, '([\u4e00-\u9fa5])[\u4e00-\u9fa5]([\u4e00-\u9fa5])', "$1＊$2")`,
			in:       `{"str": "小阿卡"}`,
			outKey:   "str",
			expected: "小阿卡",
			fail:     false,
		},

		{
			name: `not enough args`,
			pl:   `json(_, str) replace(str, '([\u4e00-\u9fa5])[\u4e00-\u9fa5]([\u4e00-\u9fa5])')`,
			in:   `{"str": "小阿卡"}`,
			fail: true,
		},

		{
			name: `invalid arg type`,
			pl:   `json(_, str) replace(str, 2, "$1＊$2")`,
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

package funcs

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

type funcCase struct {
	name     string
	data     string
	script   string
	expected interface{}
	key      string
}

func TestDecode(t *testing.T) {
	testCase := []*funcCase{
		{
			data:   "他没测试哎",
			script: `decode(_,"gbk")`,
			key:    "changed",
		},
	}
	for idx, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			decode, _ := NewDecoder("gbk")
			runner, err := NewTestingRunner(tc.script)
			tu.Equals(t, nil, err)

			err = runner.Run(tc.data)
			tu.Equals(t, nil, err)

			r, err := runner.GetContentStr(tc.key)
			res, _ := decode.decoder.String(tc.data)
			tu.Equals(t, nil, err)
			tu.Equals(t, res, r)

			t.Logf("[%d] PASS", idx)
		})
	}
}

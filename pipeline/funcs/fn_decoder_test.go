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
			data:     "ЀЀЀЀЀЀЀЀЀЀЀЀЀЀЀഀ\u0A00Ԁ؀܀ࠀऀ\u0A00\u0B00ऀఀഀ\u0E00ༀऀကᄀሀᄀऀ܀ጀ᐀ᔀᘀᄀᜀ᠀ᤀᨀᬀᰀᴀḀ ἀ ℀",
			script:   `decode(_,"gbk")`,
			expected: "wwwwwwwww",
			key:      "changed",
		},
	}

	for idx, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.script)
			tu.Equals(t, nil, err)

			err = runner.Run(tc.data)
			tu.Equals(t, nil, err)

			r, err := runner.GetContentStr(tc.key)
			tu.Equals(t, nil, err)
			tu.Equals(t, tc.expected, r)

			t.Logf("[%d] PASS", idx)
		})
	}
}

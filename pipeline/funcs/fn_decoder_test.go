// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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
			data:   "jjjjj世界",
			script: `decode(_,"gbk")`,
			key:    "_",
		},
	}
	for idx, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.script)
			tu.Equals(t, nil, err)

			err = runner.Run(tc.data)
			tu.Equals(t, nil, err)

			r, err := runner.Data.GetContentStr(tc.key)
			res := isUtf8([]byte(r))
			tu.Equals(t, nil, err)
			tu.Equals(t, res, true)

			t.Logf("[%d] PASS", idx)
		})
	}
}

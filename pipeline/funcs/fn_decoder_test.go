// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

type funcCase struct {
	name     string
	in       string
	script   string
	expected interface{}
	key      string
}

func TestDecode(t *testing.T) {
	testCase := []*funcCase{
		{
			in:     "他没测试哎",
			script: `decode(_,"gbk")`,
			key:    "changed",
		},
	}
	for idx, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			decode, _ := NewDecoder("gbk")
			runner, err := NewTestingRunner(tc.script)
			tu.Equals(t, nil, err)
			ret, err := runner.Run("test", map[string]string{},
				map[string]interface{}{
					"message": tc.in,
				}, time.Now())
			tu.Equals(t, nil, err)
			tu.Equals(t, nil, ret.Error)

			r := ret.Fields[tc.key]
			res, _ := decode.decoder.String(tc.in)
			tu.Equals(t, nil, err)
			tu.Equals(t, res, r)

			t.Logf("[%d] PASS", idx)
		})
	}
}

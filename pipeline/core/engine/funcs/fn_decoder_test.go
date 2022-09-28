// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type funcCase struct {
	name     string
	in       string
	script   string
	expected interface{}
	key      string
}

func TestDecode(t *testing.T) {
	data := []string{"测试一下", "不知道", "测试一下123456", "哈哈哈哈哈", "-汪98阿萨德离开家"}
	decode_data_slice := make([]string, 10)

	for idx, cont := range data {
		decode_data, _ := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(cont))
		decode_data_slice[idx] = string(decode_data)
	}

	testCase := []*funcCase{
		{
			in:     decode_data_slice[0],
			script: `decode(_,"gbk")`,
			key:    "message",
		},
		{
			in:     decode_data_slice[1],
			script: `decode(_,"gbk")`,
			key:    "message",
		},
		{
			in:     decode_data_slice[2],
			script: `decode(_,"gbk")`,
			key:    "message",
		},
		{
			in:     decode_data_slice[3],
			script: `decode(_,"gbk")`,
			key:    "message",
		},
		{
			in:     decode_data_slice[4],
			script: `decode(_,"gbk")`,
			key:    "message",
		},
	}
	for idx, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.script)
			tu.Equals(t, nil, err)

			_, _, f, _, _, err := runScript(runner, "test", nil, map[string]interface{}{
				"message": tc.in,
			}, time.Now())

			tu.Equals(t, nil, err)

			tu.Equals(t, nil, err)
			tu.Equals(t, data[idx], f[tc.key])

			t.Logf("[%d] PASS", idx)
		})
	}
}

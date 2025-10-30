// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLogConfigs(t *testing.T) {
	cases := []struct {
		in        string
		parseFail bool
	}{
		{
			in: "[{\"disable\":true}]",
		},
		// fail
		{
			in:        "[{]",
			parseFail: true,
		},
		{
			in: "[{\"disable\":false,\"source\":\"testing-source\",\"service\":\"testing-service\",\"pipeline\":\"test.p\"}]",
		},
		{
			in: "[{\"tags\":{\"some_tag\":\"some_value\"}}]",
		},
		{
			in: "[{\"multiline_match\":\"^\\\\[[0-9]{4}\"}]",
		},
		// fail，multiline_match 需要 4 条斜线转义等于 1 根
		{
			in:        "[{\"multiline_match\":\"^\\d{4}-\\d{2}\"}]",
			parseFail: true,
		},
		{
			in: "[{\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]",
		},
		{
			in: "[{\"disable\":false,\"source\":\"testing-source\",\"service\":\"testing-service\",\"pipeline\":\"test.p\",\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\", \"tags\":{\"some_tag\":\"some_value\"}}]",
		},
		// many config
		{
			in: "[{\"disable\":false, \"path\":\"/var/log/app1.log\"}, {\"disable\":true, \"path\":\"/var/log/app2.log\"}]",
		},
	}

	for idx, tc := range cases {
		// 直接测试 newLogConfigs 函数
		defaults := &loggingDefaults{
			extraTags: make(map[string]string),
			setLabelAsTags: func(labels map[string]string) map[string]string {
				return make(map[string]string)
			},
		}
		info := &containerLogInfo{
			runtime: "docker",
			logPath: "/var/log/container.log",
		}

		configs, err := newLogConfigs(defaults, info, tc.in)
		if tc.parseFail && assert.Error(t, err) {
			t.Logf("[%d][OK   ] %s\n", idx, err)
			continue
		}
		if !assert.NoError(t, err) {
			t.Logf("[%d][ERROR] %s\n", idx, err)
			continue
		}

		// 基本验证：确保解析成功且返回了配置
		assert.NotNil(t, configs)
		t.Logf("[%d][OK   ] parsed %d configs\n", idx, len(configs))
	}
}

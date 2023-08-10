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
		in                     string
		out                    logConfigs
		parseFail, contentFail bool
	}{
		{
			in: "[{\"disable\":true}]",
			out: logConfigs{
				&logConfig{
					Disable: true,
				},
			},
		},
		// fail
		{
			in:        "[{]",
			parseFail: true,
		},
		{
			in: "[{\"disable\":false,\"source\":\"testing-source\",\"service\":\"testing-service\",\"pipeline\":\"test.p\"}]",
			out: logConfigs{
				&logConfig{
					Disable:  false,
					Source:   "testing-source",
					Service:  "testing-service",
					Pipeline: "test.p",
				},
			},
		},
		{
			in: "[{\"tags\":{\"some_tag\":\"some_value\"}}]",
			out: logConfigs{
				&logConfig{
					Tags: map[string]string{"some_tag": "some_value"},
				},
			},
		},
		{
			in: "[{\"multiline_match\":\"^\\\\[[0-9]{4}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline: `^\[[0-9]{4}`,
				},
			},
		},
		{
			in: "[{\"multiline_match\":\"^\\\\[[0-9]{4}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline: "^\\[[0-9]{4}",
				},
			},
		},
		// fail，multiline_match 需要 4 条斜线转义等于 1 根
		{
			in:        "[{\"multiline_match\":\"^\\d{4}-\\d{2}\"}]",
			parseFail: true,
		},
		{
			in: "[{\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline: "^\\d{4}-\\d{2}",
				},
			},
		},
		{
			// 等于上一条测试
			in: "[{\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline: `^\d{4}-\d{2}`,
				},
			},
		},
		// 解析通过，内容错误
		{
			in: "[{\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline: `^\\d{4}-\\d{2}`,
				},
			},
			contentFail: true,
		},
		{
			// 8 条斜线等于实际 2 条
			in: "[{\"multiline_match\":\"^\\\\\\\\d{4}-\\\\\\\\d{2}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline: "^\\\\d{4}-\\\\d{2}",
				},
			},
		},
		{
			// 等同上一条测试
			in: "[{\"multiline_match\":\"^\\\\\\\\d{4}-\\\\\\\\d{2}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline: `^\\d{4}-\\d{2}`,
				},
			},
		},
		// 解析通过，内容错误
		{
			in: "[{\"tags\":{\"some_tag\":\"some_value\"}}]",
			out: logConfigs{
				&logConfig{
					Tags: map[string]string{"some_tag": "some_value_11"},
				},
			},
			contentFail: true,
		},
		{
			in: "[{\"disable\":false,\"source\":\"testing-source\",\"service\":\"testing-service\",\"pipeline\":\"test.p\",\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\", \"tags\":{\"some_tag\":\"some_value\"}}]",
			out: logConfigs{
				&logConfig{
					Disable:   false,
					Source:    "testing-source",
					Service:   "testing-service",
					Pipeline:  "test.p",
					Multiline: "^\\d{4}-\\d{2}",
					Tags:      map[string]string{"some_tag": "some_value"},
				},
			},
		},
		// many config
		{
			in: "[{\"disable\":false}, {\"disable\":true}]",
			out: logConfigs{
				&logConfig{
					Disable: false,
				},
				&logConfig{
					Disable: true,
				},
			},
		},
	}

	for idx, tc := range cases {
		info := &logInstance{
			configStr: tc.in,
		}

		err := info.parseLogConfigs()
		if tc.parseFail && assert.Error(t, err) {
			t.Logf("[%d][OK   ] %s\n", idx, err)
			continue
		}
		if !assert.NoError(t, err) {
			t.Logf("[%d][ERROR] %s\n", idx, err)
			continue
		}

		res := info.configs

		if tc.contentFail {
			assert.Equal(t, tc.contentFail, assert.NotEqual(t, tc.out, res))
		} else {
			assert.Equal(t, !tc.contentFail, assert.Equal(t, tc.out, res))
		}

		t.Logf("[%d][OK   ] %v\n", idx, tc)
	}
}

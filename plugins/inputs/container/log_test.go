package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLogConfig(t *testing.T) {
	cases := []struct {
		in                     string
		out                    *containerLogConfig
		parseFail, contentFail bool
	}{
		{
			in: "[{\"disable\":true}]",
			out: &containerLogConfig{
				Disable: true,
			},
		},
		// fail
		{
			in:        "[{]",
			parseFail: true,
		},
		{
			in: "[{\"disable\":false,\"source\":\"testing-source\",\"service\":\"testing-service\",\"pipeline\":\"test.p\"}]",
			out: &containerLogConfig{
				Disable:  false,
				Source:   "testing-source",
				Service:  "testing-service",
				Pipeline: "test.p",
			},
		},
		{
			in: "[{\"only_images\":[\"image:<your_image_regexp>\"]}]",
			out: &containerLogConfig{
				OnlyImages: []string{"image:<your_image_regexp>"},
			},
		},
		{
			in: "[{\"tags\":{\"some_tag\":\"some_value\"}}]",
			out: &containerLogConfig{
				Tags: map[string]string{"some_tag": "some_value"},
			},
		},
		{
			in: "[{\"multiline_match\":\"^\\\\[[0-9]{4}\"}]",
			out: &containerLogConfig{
				Multiline: `^\[[0-9]{4}`,
			},
		},
		{
			in: "[{\"multiline_match\":\"^\\\\[[0-9]{4}\"}]",
			out: &containerLogConfig{
				Multiline: "^\\[[0-9]{4}",
			},
		},
		// fail，multiline_match 需要 4 条斜线转义等于 1 根
		{
			in:        "[{\"multiline_match\":\"^\\d{4}-\\d{2}\"}]",
			parseFail: true,
		},
		{
			in: "[{\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]",
			out: &containerLogConfig{
				Multiline: "^\\d{4}-\\d{2}",
			},
		},
		{
			// 等于上一条测试
			in: "[{\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]",
			out: &containerLogConfig{
				Multiline: `^\d{4}-\d{2}`,
			},
		},
		// 解析通过，内容错误
		{
			in: "[{\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]",
			out: &containerLogConfig{
				Multiline: `^\\d{4}-\\d{2}`,
			},
			contentFail: true,
		},
		{
			// 8 条斜线等于实际 2 条
			in: "[{\"multiline_match\":\"^\\\\\\\\d{4}-\\\\\\\\d{2}\"}]",
			out: &containerLogConfig{
				Multiline: "^\\\\d{4}-\\\\d{2}",
			},
		},
		{
			// 等同上一条测试
			in: "[{\"multiline_match\":\"^\\\\\\\\d{4}-\\\\\\\\d{2}\"}]",
			out: &containerLogConfig{
				Multiline: `^\\d{4}-\\d{2}`,
			},
		},
		// 解析通过，内容错误
		{
			in: "[{\"tags\":{\"some_tag\":\"some_value\"}}]",
			out: &containerLogConfig{
				Tags: map[string]string{"some_tag": "some_value_11"},
			},
			contentFail: true,
		},
		{
			in: "[{\"disable\":false,\"source\":\"testing-source\",\"service\":\"testing-service\",\"pipeline\":\"test.p\",\"only_images\":[\"image:<your_image_regexp>\"],\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\", \"tags\":{\"some_tag\":\"some_value\"}}]",
			out: &containerLogConfig{
				Disable:    false,
				Source:     "testing-source",
				Service:    "testing-service",
				Pipeline:   "test.p",
				Multiline:  "^\\d{4}-\\d{2}",
				OnlyImages: []string{"image:<your_image_regexp>"},
				Tags:       map[string]string{"some_tag": "some_value"},
			},
		},
	}

	for idx, tc := range cases {
		c, err := parseContainerLogConfig(tc.in)
		if tc.parseFail && assert.Error(t, err) {
			t.Logf("[%d][OK   ] %s\n", idx, err)
			continue
		}
		if !assert.NoError(t, err) {
			t.Logf("[%d][ERROR] %s\n", idx, err)
			continue
		}

		if tc.contentFail {
			assert.Equal(t, tc.contentFail, assert.NotEqual(t, c, tc.out))
		} else {
			assert.Equal(t, !tc.contentFail, assert.Equal(t, c, tc.out))
		}

		t.Logf("[%d][OK   ] %v\n", idx, tc)
	}
}

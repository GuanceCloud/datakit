// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
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

		configs, _, err := newLogConfigs(defaults, info, tc.in)
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

func TestSetAutoMultiline(t *testing.T) {
	t.Run("disabled-without-explicit-multiline", func(t *testing.T) {
		cfg := &logConfig{}
		defaults := &loggingDefaults{
			enableMultiline:            false,
			autoMultilineExtraPatterns: []string{`^EXTRA`},
		}
		cfg.setAutoMultiline(defaults)
		assert.Empty(t, cfg.multilinePattern)
	})

	t.Run("enabled-builds-independent-slice", func(t *testing.T) {
		extra := []string{`^EXTRA`}
		cfg := &logConfig{}
		defaults := &loggingDefaults{
			enableMultiline:            true,
			autoMultilineExtraPatterns: extra,
		}
		cfg.setAutoMultiline(defaults)
		assert.Empty(t, cfg.multilinePattern)
		assert.True(t, cfg.autoMultiline)
		assert.Equal(t, []string{`^EXTRA`}, cfg.extraPatterns)

		cfg.extraPatterns[0] = `^CHANGED`
		assert.Equal(t, `^EXTRA`, defaults.autoMultilineExtraPatterns[0])
	})
}

func TestDuplicateLogConfigPathUsesLaterConfig(t *testing.T) {
	defaults := &loggingDefaults{
		extraTags: make(map[string]string),
		setLabelAsTags: func(labels map[string]string) map[string]string {
			return make(map[string]string)
		},
	}
	info := &containerLogInfo{
		containerName: "test-container",
		runtime:       runtime.DockerRuntime,
		logPath:       "/var/log/container.log",
	}

	configs, _, err := newLogConfigs(defaults, info, `[
		{"source":"first","service":"first-service","pipeline":"first.p"},
		{"source":"second","service":"second-service","pipeline":"second.p"}
	]`)
	require.NoError(t, err)
	require.Len(t, configs, 1)
	assert.Equal(t, "/var/log/container.log", configs[0].Path)
	assert.Equal(t, "second", configs[0].Source)
	assert.Equal(t, "second-service", configs[0].Service)
	assert.Equal(t, "second.p", configs[0].Pipeline)
}

func TestDuplicateFileLogConfigPathUsesLaterConfig(t *testing.T) {
	defaults := &loggingDefaults{
		extraTags: make(map[string]string),
		setLabelAsTags: func(labels map[string]string) map[string]string {
			return make(map[string]string)
		},
	}
	info := &containerLogInfo{
		containerName: "test-container",
		runtime:       runtime.DockerRuntime,
		logPath:       "/var/log/container.log",
		mounts: runtime.Mounts{
			{Destination: "/var/log/app", Source: "/host/var/log/app"},
		},
	}

	configs, _, err := newLogConfigs(defaults, info, `[
		{"type":"file","path":"/var/log/app/app.log","source":"first","service":"first-service"},
		{"type":"file","path":"/var/log/app/app.log","source":"second","service":"second-service"}
	]`)
	require.NoError(t, err)
	require.Len(t, configs, 1)
	assert.Equal(t, "/var/log/app/app.log", configs[0].Path)
	assert.Equal(t, "/host/var/log/app/app.log", configs[0].hostFilePath)
	assert.Equal(t, "second", configs[0].Source)
	assert.Equal(t, "second-service", configs[0].Service)
}

func TestDuplicateLogConfigPathUsesLaterDisabledConfig(t *testing.T) {
	defaults := &loggingDefaults{
		extraTags: make(map[string]string),
		setLabelAsTags: func(labels map[string]string) map[string]string {
			return make(map[string]string)
		},
	}
	info := &containerLogInfo{
		containerName: "test-container",
		runtime:       runtime.DockerRuntime,
		logPath:       "/var/log/container.log",
	}

	configs, _, err := newLogConfigs(defaults, info, `[
		{"path":"/var/log/container.log","source":"first"},
		{"path":"/var/log/container.log","disable":true}
	]`)
	require.NoError(t, err)
	require.Len(t, configs, 1)
	assert.Equal(t, "/var/log/container.log", configs[0].Path)
	assert.True(t, configs[0].Disable)
}

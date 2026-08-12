// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestExportHelpers(t *testing.T) {
	ipt := defaultInput()

	t.Run("dashboard translations", func(t *testing.T) {
		zh := ipt.Dashboard(inputs.I18nZh)
		en := ipt.Dashboard(inputs.I18nEn)

		assert.Equal(t, "概览", zh["group_overview"])
		assert.Equal(t, "拨测任务", zh["group_tasks"])
		assert.Equal(t, "Overview", en["group_overview"])
		assert.Equal(t, "Tasks", en["group_tasks"])
		assert.Nil(t, ipt.Dashboard(inputs.I18n(-1)))
	})

	t.Run("monitor translations", func(t *testing.T) {
		assert.Empty(t, ipt.Monitor(inputs.I18nZh))
		assert.Empty(t, ipt.Monitor(inputs.I18nEn))
		assert.Nil(t, ipt.Monitor(inputs.I18n(-1)))
	})

	t.Run("env docs are prefixed and complete", func(t *testing.T) {
		envs := ipt.GetENVDoc()
		if !assert.Len(t, envs, 17) {
			return
		}

		assert.Equal(t, "ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK", envs[0].ENVName)
		assert.Equal(t, "disable_internal_network_task", envs[0].ConfField)
		assert.Equal(t, "ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST", envs[1].ENVName)
		assert.Equal(t, "ENV_INPUT_DIALTESTING_ENABLE_DEBUG_API", envs[2].ENVName)
		assert.Equal(t, "`-`", envs[2].ConfField)
		assert.Equal(t, "ENV_INPUT_DIALTESTING_ELECTION", envs[3].ENVName)
		assert.Equal(t, "ENV_INPUT_DIALTESTING_BROWSER_ENABLED", envs[4].ENVName)
		assert.Equal(t, "browser.enabled", envs[4].ConfField)
		assert.Equal(t, "ENV_INPUT_DIALTESTING_BROWSER_ENGINE", envs[5].ENVName)
		assert.Equal(t, "browser.engine", envs[5].ConfField)
		assert.Equal(t, "ENV_INPUT_DIALTESTING_BROWSER_ENGINE_PATH", envs[6].ENVName)
		assert.Equal(t, "browser.engine_path", envs[6].ConfField)
		assert.Equal(t, "ENV_INPUT_DIALTESTING_BROWSER_CA_CERT_FILE", envs[7].ENVName)
		assert.Equal(t, "browser.ca_cert_file", envs[7].ConfField)
		assert.Equal(t, "ENV_INPUT_DIALTESTING_BROWSER_CA_CERT_DIR", envs[8].ENVName)
		assert.Equal(t, "browser.ca_cert_dir", envs[8].ConfField)
		assert.Equal(t, "ENV_INPUT_DIALTESTING_BROWSER_PROXY_URL", envs[9].ENVName)
		assert.Equal(t, "browser.proxy_url", envs[9].ConfField)
		assert.Equal(t, "ENV_INPUT_DIALTESTING_BROWSER_MAX_CONCURRENCY", envs[10].ENVName)
		assert.Equal(t, "browser.max_concurrency", envs[10].ConfField)

		assert.Equal(t, "ENV_DIALTESTING_DEBUG_MAX_CONCURRENT_RUNS", envs[11].ENVName)
		assert.Equal(t, "Int", envs[11].Type)
		assert.Equal(t, "`100`", envs[11].Default)
		assert.Equal(t, "`-`", envs[11].ConfField)
		assert.Equal(t, "ENV_DIALTESTING_DEBUG_MAX_QUEUED_RUNS", envs[12].ENVName)
		assert.Equal(t, "ENV_DIALTESTING_DEBUG_MAX_QUEUE_WAIT", envs[13].ENVName)
		assert.Equal(t, "Duration", envs[13].Type)
		assert.Equal(t, "`3m`", envs[13].Default)
		assert.Equal(t, "ENV_DIALTESTING_DEBUG_RESULT_TTL", envs[14].ENVName)
		assert.Equal(t, "ENV_DIALTESTING_DEBUG_MAX_LONG_POLL_WAIT", envs[15].ENVName)
		assert.Equal(t, "ENV_DIALTESTING_DEBUG_MAX_RETAINED_TERMINAL_RUNS", envs[16].ENVName)
		assert.Equal(t, "`2000`", envs[16].Default)
	})
}

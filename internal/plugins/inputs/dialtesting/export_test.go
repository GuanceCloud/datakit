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
		if !assert.Len(t, envs, 4) {
			return
		}

		assert.Equal(t, "ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK", envs[0].ENVName)
		assert.Equal(t, "disable_internal_network_task", envs[0].ConfField)
		assert.Equal(t, "ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST", envs[1].ENVName)
		assert.Equal(t, "ENV_INPUT_DIALTESTING_ENABLE_DEBUG_API", envs[2].ENVName)
		assert.Equal(t, "ENV_INPUT_DIALTESTING_ELECTION", envs[3].ENVName)
	})
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dk

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestSetup(t *T.T) {
	t.Run("replace", func(t *T.T) {
		i := def()
		i.setup("1.2.3.4:4321")
		assert.Equal(t, "http://1.2.3.4:4321/metrics", i.url)
	})
}

func TestReadenv(t *T.T) {
	t.Run("-", func(t *T.T) {
		i := def()

		i.ReadEnv(map[string]string{
			"ENV_INPUT_DK_ENABLE_ALL_METRICS": "on",
		})

		assert.Nil(t, i.MetricFilter)

		i = def()

		i.ReadEnv(map[string]string{
			"ENV_INPUT_DK_ADD_METRICS": `["a", "b"]`,
		})

		assert.Contains(t, i.MetricFilter, "a")
		assert.Contains(t, i.MetricFilter, "b")

		i = def()

		i.ReadEnv(map[string]string{
			"ENV_INPUT_DK_ONLY_METRICS": `["a", "b"]`,
		})

		assert.Len(t, i.MetricFilter, 2)
		assert.Contains(t, i.MetricFilter, "a")
		assert.Contains(t, i.MetricFilter, "b")
	})
}

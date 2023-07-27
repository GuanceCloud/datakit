// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package export

import (
	T "testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestIntegration(t *T.T) {
	t.Run("export-all", func(t *T.T) {
		i := NewIntegration(WithTopDir("test-export-all"),
			WithI18n([]inputs.I18n{inputs.I18nZh, inputs.I18nEn}))

		assert.NoError(t, i.Export())
	})
}

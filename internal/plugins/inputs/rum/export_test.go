// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestInput_Dashboard(t *testing.T) {
	ipt := defaultInput()
	assert.NotNil(t, ipt.Dashboard(inputs.I18nZh))
	assert.NotNil(t, ipt.Dashboard(inputs.I18nEn))
	assert.Nil(t, ipt.Dashboard(inputs.I18n(-1)))
}

func TestInput_Monitor(t *testing.T) {
	ipt := defaultInput()
	assert.NotNil(t, ipt.Monitor(inputs.I18nZh))
	assert.NotNil(t, ipt.Monitor(inputs.I18nEn))
	assert.Nil(t, ipt.Monitor(inputs.I18n(-1)))
}

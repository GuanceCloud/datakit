// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeInputMeasurementVersion(t *T.T) {
	assert.Equal(t, "v1", NormalizeInputMeasurementVersion(" V1 "))
	assert.Equal(t, "v2", NormalizeInputMeasurementVersion("v2"))
	assert.Equal(t, "v2", NormalizeInputMeasurementVersion(""))
	assert.Equal(t, "v2", NormalizeInputMeasurementVersion("latest"))
}

func TestIsOverrideMeasurement(t *T.T) {
	oldCfg := Cfg
	defer func() {
		Cfg = oldCfg
	}()

	Cfg = DefaultConfig()

	Cfg.MeasurementVersion = "v2"
	assert.True(t, IsOverrideMeasurement("v1"))

	Cfg.MeasurementVersion = "v1"
	assert.False(t, IsOverrideMeasurement("v2"))

	Cfg.MeasurementVersion = ""
	assert.True(t, IsOverrideMeasurement("v2"))
	assert.True(t, IsOverrideMeasurement(""))
	assert.True(t, IsOverrideMeasurement("latest"))
	assert.False(t, IsOverrideMeasurement("v1"))
}

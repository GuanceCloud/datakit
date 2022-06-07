// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestRunPl(t *testing.T) {
}

func TestSCriptName(t *testing.T) {
	pt, err := NewPoint("m_name", map[string]string{"service": "svc_name"}, map[string]interface{}{"message@json": "a"}, &PointOption{
		Category: datakit.Logging,
	})
	assert.Equal(t, nil, err)

	name, ok := scriptName(datakit.Tracing, pt, nil)
	assert.Equal(t, true, ok)
	assert.Equal(t, "svc_name.p", name)

	name, ok = scriptName(datakit.Tracing, pt, map[string]string{"c": "d"})
	assert.Equal(t, true, ok)
	assert.Equal(t, "svc_name.p", name)

	_, ok = scriptName(datakit.Tracing, pt, map[string]string{"svc_name": "-"})
	assert.Equal(t, false, ok)

	name, ok = scriptName(datakit.Tracing, pt, map[string]string{"svc_name": "def.p"})
	assert.Equal(t, true, ok)
	assert.Equal(t, "def.p", name)

	pt2, err := NewPoint("m_name", map[string]string{}, map[string]interface{}{"message@json": "a"}, &PointOption{
		Category: datakit.Logging,
	})
	assert.Equal(t, nil, err)
	_, ok = scriptName(datakit.Tracing, pt2, map[string]string{"m_name": "def.p"})
	assert.Equal(t, false, ok)

	name, ok = scriptName(datakit.Metric, pt, map[string]string{"abc": "def.p"})
	assert.Equal(t, true, ok)
	assert.Equal(t, "m_name.p", name)

	name, ok = scriptName(datakit.Metric, pt, map[string]string{"m_name": "def.p"})
	assert.Equal(t, true, ok)
	assert.Equal(t, "def.p", name)

	_, ok = scriptName(datakit.Metric, pt, map[string]string{"m_name": "-"})
	assert.Equal(t, false, ok)

	_, ok = scriptName(datakit.Metric, nil, map[string]string{"m_name": "-"})
	assert.Equal(t, false, ok)
}

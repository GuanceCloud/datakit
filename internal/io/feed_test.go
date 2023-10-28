// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	T "testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func Test_forceBlocking(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		assert.Nil(t, forceBlocking(point.Metric, "some", nil))

		opt := forceBlocking(point.Logging, "some", nil)
		assert.True(t, opt.Blocking)

		opt = forceBlocking(point.RUM, "some", nil)
		assert.True(t, opt.Blocking)
	})
}

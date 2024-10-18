// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithOptions(t *testing.T) {
	t.Run("with-service", func(t *testing.T) {
		opt := defaultOption()
		WithSource("testing-source")(opt)
		WithService("testing-service")(opt)

		res := map[string]string{"service": "testing-service"}
		assert.Equal(t, opt.extraTags, res)
	})

	t.Run("with-default-service", func(t *testing.T) {
		opt := defaultOption()
		WithSource("testing-source")(opt)
		WithService("")(opt)

		res := map[string]string{"service": "testing-source"}
		assert.Equal(t, opt.extraTags, res)
	})

	t.Run("with-default-service", func(t *testing.T) {
		opt := defaultOption()
		WithService("")(opt)
		WithSource("testing-source")(opt)

		res := map[string]string{"service": "default"}
		assert.Equal(t, opt.extraTags, res)
	})

	t.Run("with-non-service", func(t *testing.T) {
		opt := defaultOption()
		WithSource("testing-source")(opt)

		res := map[string]string{"service": "default"}
		assert.Equal(t, opt.extraTags, res)
	})
}

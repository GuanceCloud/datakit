// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package testutils

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPort(t *T.T) {
	t.Run("base-100", func(t *T.T) {
		for i := 0; i < 10; i++ {
			get := RandPort("tcp")
			assert.True(t, get > baseOffset)
			t.Logf("get: %d -> %d", i, get)
		}

		for _, i := range []int{1234, 6379, 3306, 2375} {
			get := RandPort("tcp")
			assert.True(t, get > baseOffset)
			t.Logf("get: %d -> %d", i, get)
		}
	})

	t.Run("larger-than-base", func(t *T.T) {
		for i := baseOffset; i < baseOffset+10; i++ {
			get := RandPort("tcp")
			assert.True(t, get > baseOffset)
			t.Logf("get: %d -> %d", i, get)
		}
	})

	t.Run("larger-than-max", func(t *T.T) {
		for i := maxPort; i < maxPort+10; i++ {
			get := RandPort("tcp")
			assert.True(t, get > baseOffset)
			t.Logf("get: %d -> %d", i, get)
		}
	})
}

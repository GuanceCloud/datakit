// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package plval

import (
	"net"
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPrivateIP(t *T.T) {
	t.Run("192.168", func(t *T.T) {
		assert.True(t, isPrivateIP(net.ParseIP("192.168.1.0")))
	})

	t.Run("192.167", func(t *T.T) {
		assert.False(t, isPrivateIP(net.ParseIP("192.167.1.0")))
	})

	t.Run("10", func(t *T.T) {
		assert.True(t, isPrivateIP(net.ParseIP("10.168.1.0")))
	})

	t.Run("172.16", func(t *T.T) {
		assert.True(t, isPrivateIP(net.ParseIP("172.16.1.0")))
	})

	t.Run("172.15", func(t *T.T) {
		assert.False(t, isPrivateIP(net.ParseIP("172.15.1.0")))
	})

	t.Run("172.32", func(t *T.T) {
		assert.False(t, isPrivateIP(net.ParseIP("172.32.1.0")))
	})

	t.Run("loopback", func(t *T.T) {
		assert.True(t, isPrivateIP(net.ParseIP("127.0.0.1")))
	})
}

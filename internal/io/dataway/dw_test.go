// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDWInit(t *T.T) {
	t.Run("basic", func(t *T.T) {
		dw := NewDefaultDataway()
		urls := []string{
			"https://host.com?token=tkn_11111111111111111111",
			"https://host.com?token=tkn_22222222222222222222",
		}

		require.NoError(t, dw.Init(WithURLs(urls...)))

		assert.Len(t, dw.eps, 2)
		assert.Equal(t, dw.HTTPTimeout, time.Second*30)
		assert.Equal(t, dw.MaxIdleConnsPerHost, 64)
	})

	t.Run("invalid-timeout", func(t *T.T) {
		dw := NewDefaultDataway()
		urls := []string{
			"https://host.com?token=tkn_11111111111111111111",
			"https://host.com?token=tkn_22222222222222222222",
		}
		dw.HTTPTimeout = -30 * time.Second

		require.NoError(t, dw.Init(WithURLs(urls...)))
		assert.Equal(t, 30*time.Second, dw.HTTPTimeout)
	})

	t.Run("invalid-max-retry-count", func(t *T.T) {
		dw := NewDefaultDataway()
		urls := []string{
			"https://host.com?token=tkn_11111111111111111111",
			"https://host.com?token=tkn_22222222222222222222",
		}
		dw.HTTPTimeout = -30 * time.Second
		dw.MaxRetryCount = 100

		require.NoError(t, dw.Init(WithURLs(urls...)))
		assert.Equal(t, 30*time.Second, dw.HTTPTimeout)
		assert.Equal(t, 10, dw.MaxRetryCount)

		dw.MaxRetryCount = -1
		require.NoError(t, dw.Init(WithURLs(urls...)))
		assert.Equal(t, 1, dw.MaxRetryCount)

		dw.MaxRetryCount = 0
		require.NoError(t, dw.Init(WithURLs(urls...)))
		assert.Equal(t, 1, dw.MaxRetryCount)
	})
}

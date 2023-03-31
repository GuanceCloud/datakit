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
		dw := &Dataway{
			URLs: []string{
				"https://host.com?token=tkn_11111111111111111111",
				"https://host.com?token=tkn_22222222222222222222",
			},
		}

		require.NoError(t, dw.doInit())

		assert.Len(t, dw.eps, 2)
		assert.Equal(t, dw.httpTimeout, time.Second*30)
		assert.Equal(t, dw.MaxIdleConnsPerHost, 64)
	})

	t.Run("invalid-timeout", func(t *T.T) {
		dw := &Dataway{
			URLs: []string{
				"https://host.com?token=tkn_11111111111111111111",
				"https://host.com?token=tkn_22222222222222222222",
			},
			HTTPTimeout: "invalid-duration",
		}

		err := dw.doInit()
		require.Error(t, err)
		t.Logf("doInit: %s", err)
	})
}

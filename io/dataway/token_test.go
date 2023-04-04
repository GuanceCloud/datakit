// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckToken(t *T.T) {
	t.Run("ok", func(t *T.T) {
		assert.NoError(t, CheckToken("token_00000000000000000000000000000011"))
		assert.NoError(t, CheckToken("tkn_00000000000000000000000000000011"))
		assert.NoError(t, CheckToken("tokn_000000000000000000001111"))

		assert.Error(t, CheckToken("token_0000000000000000000000000000001"))
		assert.Error(t, CheckToken("tkn_0000000000000000000000000000001"))
		assert.Error(t, CheckToken("tokn_0000000000000000000011"))
	})

	t.Run("check-token-error-msg", func(t *T.T) {
		err := CheckToken("token_0000000000000000000000000000001")
		t.Logf("%s", err)

		err = CheckToken("tkn_0000000000000000000000000000001")
		t.Logf("%s", err)

		err = CheckToken("tokn_0000000000000000000011")
		t.Logf("%s", err)
	})
}

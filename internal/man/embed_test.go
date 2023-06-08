// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package man

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *T.T) {
	t.Skip()
	t.Run("list-all", func(t *T.T) {
		dirs, err := AllDocs.ReadDir("docs/zh/")
		assert.NoError(t, err)

		t.Logf("get %d dirs", len(dirs))
		for _, x := range dirs {
			t.Logf("%s", x.Name())
		}
	})
}

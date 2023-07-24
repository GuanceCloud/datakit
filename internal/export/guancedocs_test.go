// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package export

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestGuanceDocs(t *T.T) {
	t.Run("export-all", func(t *T.T) {
		dir := t.Name()
		gd := NewGuanceDodcs(
			WithTopDir(dir),
		)
		assert.NoError(t, gd.Export())
	})
}

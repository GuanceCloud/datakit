package io

import (
	T "testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func TestPointConvert(t *T.T) {
	t.Run("basic", func(t *T.T) {
		pt := point.NewPointV2([]byte(`abc`), point.NewKVs(map[string]any{"abc": 123}))
		pts := ptConvert(pt)

		assert.Equal(t, 1, len(pts))
	})
}

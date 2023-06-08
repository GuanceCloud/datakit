// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	T "testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

func TestPointHash(t *T.T) {
	t.Run("basic", func(t *T.T) {
		name := "abc"
		tags := map[string]string{
			"t1": "v1",
			"t2": "v2",
		}

		fs := map[string]any{
			"f1": 123,
			"f2": 3.14,
			"f3": false,
			"f5": "foo bar",
		}

		pt1 := dkpt.MustNewPoint(name, tags, fs,
			&dkpt.PointOption{
				Category: datakit.Logging, // logging can accept string field on dkpt
			})
		pt2, err := point.NewPoint(name, tags, fs)
		assert.NoError(t, err)

		h1 := lineprotoHash(pt1)
		h2 := pointHash(pt2)

		assert.Equal(t, h1, h2)

		t.Logf("hash: %d", h1)
	})
}

func BenchmarkPointHash(b *T.B) {
	name := "abc"
	tags := map[string]string{
		"t1": "v1",
		"t2": "v2",
	}

	fs := map[string]any{
		"f1": 123,
		"f2": 3.14,
		"f3": false,
		"f5": "foo bar",
	}

	lppt := dkpt.MustNewPoint(name, tags, fs,
		&dkpt.PointOption{
			Category: datakit.Logging, // logging can accept string field on dkpt
		})

	pt, err := point.NewPoint(name, tags, fs)
	assert.NoError(b, err)

	b.Run("lineproto", func(b *T.B) {
		for i := 0; i < b.N; i++ {
			lineprotoHash(lppt)
		}
	})

	b.Run("point", func(b *T.B) {
		for i := 0; i < b.N; i++ {
			pointHash(pt)
		}
	})

	// benchmark Result:
	// =======
	// goos: darwin
	// goarch: arm64
	// pkg: gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/point
	// BenchmarkPointHash
	// BenchmarkPointHash/lineproto
	// BenchmarkPointHash/lineproto-10         	 1310508	       910.9 ns/op	     813 B/op	      20 allocs/op
	// BenchmarkPointHash/point
	// BenchmarkPointHash/point-10             	 1524051	       787.1 ns/op	     648 B/op	      23 allocs/op
	// PASS
	// ok  	gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/point	4.261s
}

package io

import (
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

func TestNPoints(t *T.T) {
	f := NewMockedFeeder()

	n := 10

	t.Run("wait-forever", func(t *T.T) {
		go func() {
			pt, _ := point.NewPoint(t.Name(), nil, map[string]any{"abc": 123}, nil)
			pts := []*point.Point{pt}

			for i := 0; i < n; i++ {
				assert.NoError(t, f.Feed(t.Name(), datakit.Metric, pts))
				time.Sleep(time.Millisecond * 10)
			}
			return
		}()

		pts, err := f.NPoints(n)
		assert.NoError(t, err)
		assert.Equal(t, n, len(pts))

		for _, pt := range pts {
			t.Logf("%s", pt.String())
		}
	})

	t.Run("wait-10ms-and-timeout", func(t *T.T) {
		go func() {
			pt, _ := point.NewPoint(t.Name(), nil, map[string]any{"abc": 123}, nil)
			pts := []*point.Point{pt}

			for i := 0; i < n; i++ {
				assert.NoError(t, f.Feed(t.Name(), datakit.Metric, pts))
				time.Sleep(time.Millisecond * 10)
			}
			return
		}()

		_, err := f.NPoints(n, time.Millisecond*10)
		assert.Error(t, err)
		t.Logf("got expected error: %s", err.Error())
		return
	})

	t.Run("wait-1s", func(t *T.T) {
		go func() {
			pt, _ := point.NewPoint(t.Name(), nil, map[string]any{"abc": 123}, nil)
			pts := []*point.Point{pt}

			for i := 0; i < n; i++ {
				assert.NoError(t, f.Feed(t.Name(), datakit.Metric, pts))
				time.Sleep(time.Millisecond * 10)
			}
			return
		}()

		pts, err := f.NPoints(n, time.Second)
		assert.NoError(t, err)
		assert.Equal(t, n, len(pts))

		for _, pt := range pts {
			t.Logf("%s", pt.String())
		}
	})

	t.Run("feed-busy", func(t *T.T) {
		pt, _ := point.NewPoint(t.Name(), nil, map[string]any{"abc": 123}, nil)
		pts := []*point.Point{pt}

		for i := 0; i < chanCap; i++ {
			assert.NoError(t, f.Feed(t.Name(), datakit.Metric, pts))
		}

		err := f.Feed(t.Name(), datakit.Metric, pts)
		assert.Error(t, err)
		t.Logf("got expect error: %s", err)

		return
	})
}

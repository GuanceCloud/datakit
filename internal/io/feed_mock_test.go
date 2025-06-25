// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"sync"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func TestNPoints(t *T.T) {
	f := NewMockedFeeder()

	n := 10

	t.Run("wait-forever", func(t *T.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()

			pt, _ := point.NewPoint(t.Name(), nil, map[string]any{"abc": 123})
			pts := []*point.Point{pt}

			for i := 0; i < n; i++ {
				assert.NoError(t, f.Feed(point.Metric, pts, WithSource(t.Name())))
				time.Sleep(time.Millisecond * 10)
			}
		}()

		pts, err := f.NPoints(n)
		assert.NoError(t, err)
		assert.Equal(t, n, len(pts))

		for _, pt := range pts {
			t.Logf("%s", pt.LineProto())
		}

		wg.Wait()
	})

	t.Run("wait-10ms-and-timeout", func(t *T.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			pt, _ := point.NewPoint(t.Name(), nil, map[string]any{"abc": 123})
			pts := []*point.Point{pt}

			for i := 0; i < n; i++ {
				assert.NoError(t, f.Feed(point.Metric, pts, WithSource(t.Name())))
				time.Sleep(time.Millisecond * 10)
			}
		}()

		_, err := f.NPoints(n, time.Millisecond*10)
		assert.Error(t, err)
		t.Logf("got expected error: %s", err.Error())

		wg.Wait()
	})

	t.Run("wait-1s", func(t *T.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			pt, _ := point.NewPoint(t.Name(), nil, map[string]any{"abc": 123})
			pts := []*point.Point{pt}

			for i := 0; i < n; i++ {
				assert.NoError(t, f.Feed(point.Metric, pts, WithSource(t.Name())))
				time.Sleep(time.Millisecond * 10)
			}
		}()

		pts, err := f.NPoints(n, time.Second)
		assert.NoError(t, err)
		assert.Equal(t, n, len(pts))

		for _, pt := range pts {
			t.Logf("%s", pt.LineProto())
		}

		wg.Wait()
	})

	t.Run("feed-busy", func(t *T.T) {
		pt, _ := point.NewPoint(t.Name(), nil, map[string]any{"abc": 123})
		pts := []*point.Point{pt}

		// cleanup chans
		for {
			select {
			case <-f.ch:
			default:
				goto out
			}
		}
	out:

		for i := 0; i < chanCap; i++ {
			assert.NoError(t, f.Feed(point.Metric, pts, WithSource(t.Name())), "feed err on %dth", i)
		}

		err := f.Feed(point.Metric, pts, WithSource(t.Name()))
		assert.Error(t, err)
		t.Logf("got expect error: %s", err)
	})
}

func TestFeedOptions(t *T.T) {
	t.Run("basic", func(t *T.T) {
		t.Skip("TODO")
	})
}

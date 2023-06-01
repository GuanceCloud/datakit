// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package goroutine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TaskOk(ctx context.Context) error {
	return nil
}

func TestNormal(t *testing.T) {
	t.Run("with no error", func(t *testing.T) {
		g := Group{}
		g.Go(func(ctx context.Context) error {
			return nil
		})
		g.Go(func(ctx context.Context) error {
			return nil
		})
		g.Go(func(ctx context.Context) error {
			return nil
		})

		err := g.Wait()

		assert.NoError(t, err)
	})

	t.Run("run with errors", func(t *testing.T) {
		g := Group{}
		g.Go(func(ctx context.Context) error {
			return errors.New("err")
		})
		err := g.Wait()
		assert.Error(t, err)
	})

	t.Run("panic catch", func(t *testing.T) {
		isPanic := false
		panicCb := func(buf []byte) bool {
			isPanic = true
			return true
		}
		g := Group{
			panicCb:    panicCb,
			panicTimes: 6,
		}
		count := 0
		g.Go(func(ctx context.Context) error {
			count++
			panic("panic error")
		})

		err := g.Wait()

		assert.True(t, isPanic)
		assert.Equal(t, 6, count)
		assert.Error(t, err)
	})

	t.Run("group with cancel", func(t *testing.T) {
		g := WithCancel(context.Background())

		isFinished := false

		g.Go(func(ctx context.Context) error {
			timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			for {
				select {
				case <-ctx.Done():
					isFinished = true
					return nil
				case <-timeoutCtx.Done():
					return nil
				}
			}
		})
		g.cancel()
		err := g.Wait()
		assert.NoError(t, err)
		assert.True(t, isFinished)
	})
	type ctxKey string

	t.Run("group with context", func(t *testing.T) {
		var key ctxKey = "name"
		g := WithContext(context.WithValue(context.Background(), key, "demo"))

		val := ""
		g.Go(func(ctx context.Context) error {
			ctxVal := ctx.Value(key)
			v, ok := ctxVal.(string)
			assert.True(t, ok)
			val = v
			return nil
		})

		assert.NoError(t, g.Wait())
		assert.Equal(t, val, "demo")
	})
}

func TestGOMAXPROCS(t *testing.T) {
	sleep1s := func(ctx context.Context) error {
		time.Sleep(time.Second)
		return nil
	}

	g := Group{}
	now := time.Now()
	g.GOMAXPROCS(2)
	g.Go(sleep1s)
	g.Go(sleep1s)
	g.Go(sleep1s)
	g.Go(sleep1s)

	err := g.Wait()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, time.Since(now).Milliseconds(), int64(2000))
}

/* failed
func TestStat(t *testing.T) {
	g := NewGroup(Option{})

	g.Go(func(ctx context.Context) error {
		time.Sleep(1 * time.Second)
		return errors.New("xxxxxx")
	})
	g.Go(func(ctx context.Context) error {
		time.Sleep(1 * time.Second)
		return nil
	})
	g.Go(func(ctx context.Context) error {
		time.Sleep(1 * time.Second)
		return nil
	})

	err := g.Wait()
	assert.Error(t, err)

	for k, v := range stat {
		assert.Equal(t, k, "default")
		assert.Equal(t, int64(3), v.Total)
		assert.Equal(t, int64(1), v.ErrCount)
		assert.Greater(t, v.CostTime, int64(3000000))
	}
} */

func TestNestGroup(t *testing.T) {
	sleepTime := 1 * time.Second
	startTime := time.Now()
	g := NewGroup(Option{})
	g.Go(func(ctx context.Context) error {
		gChild := NewGroup(Option{})
		gChild.Go(func(ctx context.Context) error {
			time.Sleep(sleepTime)
			return nil
		})
		return gChild.Wait()
	})

	err := g.Wait()
	assert.NoError(t, err)
	assert.Greater(t, time.Since(startTime), sleepTime)
}

func TestMultiGroup(t *testing.T) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(goroutineGroups, goroutineCostVec, goroutineStoppedVec, goroutineCounterVec)

	sleep5s := func(ctx context.Context) error {
		tick := time.NewTicker(5 * time.Second)
		defer tick.Stop()

		for {
			mfs, err := reg.Gather()
			require.NoError(t, err)
			t.Logf("get metrics: %s\n====================================================\n\n",
				metrics.MetricFamily2Text(mfs))

			select {
			case <-tick.C:
				return nil
			default:
				time.Sleep(time.Second)
			}
		}
	}
	sleep5s2 := func(ctx context.Context) error {
		time.Sleep(5 * time.Second)
		return nil
	}

	g1 := NewGroup(Option{Name: "inputs_jolokia"})
	g2 := NewGroup(Option{Name: "inputs_jolokia"})
	g3 := NewGroup(Option{Name: "inputs_jolokia"})

	g1.Go(sleep5s)
	g2.Go(sleep5s2)
	g3.Go(sleep5s2)

	g1.Wait()
	g2.Wait()
	g3.Wait()
}

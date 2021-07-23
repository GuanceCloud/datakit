package goroutine

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TaskOk(ctx context.Context) error {
	return nil
}

func TestNormal(t *testing.T) {
	t.Run("with no error", func(t *testing.T) {
		g := Group{}
		g.Go(func(ctx context.Context) error {
			fmt.Println("go 1")
			return nil
		})
		g.Go(func(ctx context.Context) error {
			fmt.Println("go 2")
			return nil
		})
		g.Go(func(ctx context.Context) error {
			fmt.Println("go 3")
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
		panicCb := func(buf []byte) {
			panicMsg := string(buf)
			fmt.Println(panicMsg)
			isPanic = true
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
}

func TestGOMAXPROCS(t *testing.T) {
	var sleep1s = func(ctx context.Context) error {
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
	assert.Greater(t, time.Since(now).Milliseconds(), int64(2000))
}

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

	g.Wait()

	for k, v := range stat {
		assert.Equal(t, k, "default")
		assert.Equal(t, int64(3), v.Total)
		assert.Equal(t, int64(1), v.ErrCount)
		assert.Greater(t, v.CostTime, int64(3000000))
	}

}

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

package goroutine

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
type Group struct {
	chs []func(ctx context.Context) error

	name string
	err  error
	ctx  context.Context

	panicCb  func([]byte) bool          // callback when panic
	postCb   func(error, time.Duration) // job finish callback
	beforeCb func()                     // job before callback

	panicTimeout time.Duration // time duration between panicCb call
	ch           chan func(ctx context.Context) error
	cancel       func()
	wg           sync.WaitGroup

	errOnce    sync.Once
	workerOnce sync.Once
	panicTimes int8 // max panic times
}

// WithContext create a Group.
// given function from Go will receive this context,.
func WithContext(ctx context.Context) *Group {
	return &Group{ctx: ctx}
}

// WithCancel create a new Group and an associated Context derived from ctx.
//
// given function from Go will receive context derived from this ctx,
// The derived Context is canceled the first time a function passed to Go
// returns a non-nil error or the first time Wait returns, whichever occurs
// first.
func WithCancel(ctx context.Context) *Group {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{ctx: ctx, cancel: cancel}
}

func (g *Group) do(f func(ctx context.Context) error) {
	if g.beforeCb != nil {
		g.beforeCb()
	}
	ctx := g.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	panicTimes := g.panicTimes - 1

	var err error
	var run func()

	var startTime time.Time
	var endTime time.Time
	var costTime time.Duration

	run = func() {
		defer func() {
			endTime = time.Now()

			if r := recover(); r != nil {
				isPanicRetry := true
				buf := make([]byte, 1024) //nolint:gomnd
				buf = buf[:runtime.Stack(buf, false)]
				if g.panicCb != nil {
					isPanicRetry = g.panicCb(buf)
				}

				if isPanicRetry && panicTimes > 0 {
					panicTimes--
					if g.panicTimeout > 0 {
						time.Sleep(g.panicTimeout)
					}
					run()
					return
				}
				err = fmt.Errorf("goroutine: panic recovered: %s", r)
			}
			if err != nil {
				g.errOnce.Do(func() {
					g.err = err
					if g.cancel != nil {
						g.cancel()
					}
				})
			}
			costTime = endTime.Sub(startTime)
			if g.postCb != nil {
				g.postCb(err, costTime)
			}
			g.wg.Done()
		}()
		startTime = time.Now()
		err = f(ctx)
	}

	run()
}

// GOMAXPROCS set max goroutine to work.
func (g *Group) GOMAXPROCS(n int) {
	if n <= 0 {
		panic("goroutine: GOMAXPROCS must great than 0")
	}

	g.workerOnce.Do(func() {
		g.ch = make(chan func(context.Context) error, n)
		for i := 0; i < n; i++ {
			go func() {
				for f := range g.ch {
					g.do(f)
				}
			}()
		}
	})
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *Group) Go(f func(ctx context.Context) error) {
	g.wg.Add(1)
	if g.ch != nil {
		select {
		case g.ch <- f:
		default:
			g.chs = append(g.chs, f)
		}
		return
	}
	go g.do(f)
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *Group) Wait() error {
	if g.ch != nil {
		for _, f := range g.chs {
			g.ch <- f
		}
	}
	g.wg.Wait()
	if g.ch != nil {
		close(g.ch) // let all receiver exit
	}
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}

func (g *Group) Name() string {
	return g.name
}

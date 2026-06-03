package chrome

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/dialtesting/browserdial/runner"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

// Pool keeps Chrome browser processes warm and gives each run an isolated
// browser context. A worker runs one task at a time.
type Pool struct {
	ctx     context.Context
	workers chan *poolWorker
	done    chan struct{}

	mu     sync.Mutex
	all    []*poolWorker
	closed bool
}

type poolWorker struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewPool(ctx context.Context, size int, options runner.EngineOptions) (*Pool, error) {
	if size <= 0 {
		size = 1
	}
	pool := &Pool{
		ctx:     ctx,
		workers: make(chan *poolWorker, size),
		done:    make(chan struct{}),
	}
	for i := 0; i < size; i++ {
		worker, err := newPoolWorker(ctx, options)
		if err != nil {
			pool.Close()
			return nil, err
		}
		pool.all = append(pool.all, worker)
		pool.workers <- worker
	}
	return pool, nil
}

func newPoolWorker(ctx context.Context, options runner.EngineOptions) (*poolWorker, error) {
	executable, err := resolveExecutable(options.ChromePath)
	if err != nil {
		return nil, err
	}
	width, height := viewportSize(options.ViewportWidth, options.ViewportHeight)
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, chromeAllocatorOptions(executable, width, height, options)...)
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	if err := chromedp.Run(browserCtx, chromedp.Navigate("about:blank")); err != nil {
		browserCancel()
		allocCancel()
		return nil, fmt.Errorf("start pooled chrome: %w", err)
	}
	return &poolWorker{
		ctx: browserCtx,
		cancel: func() {
			browserCancel()
			allocCancel()
		},
	}, nil
}

func (p *Pool) Factory(ctx context.Context, options runner.EngineOptions) (runner.Engine, error) {
	select {
	case <-p.ctx.Done():
		return nil, p.ctx.Err()
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-p.done:
		return nil, fmt.Errorf("chrome pool is closed")
	case worker := <-p.workers:
		select {
		case <-p.done:
			return nil, fmt.Errorf("chrome pool is closed")
		default:
		}
		browserContextID, targetID, err := worker.newIsolatedTarget(ctx)
		if err != nil {
			p.releaseWorker(worker)
			return nil, err
		}
		taskCtx, taskCancel := chromedp.NewContext(worker.ctx, chromedp.WithTargetID(targetID))
		var releaseOnce sync.Once
		release := func() {
			releaseOnce.Do(func() {
				taskCancel()
				worker.disposeBrowserContext(browserContextID)
				p.releaseWorker(worker)
			})
		}
		engine, err := newEngineFromContext(taskCtx, chromedp.Run, release)
		if err != nil {
			release()
			return nil, err
		}
		return engine, nil
	}
}

func (w *poolWorker) newIsolatedTarget(ctx context.Context) (cdp.BrowserContextID, target.ID, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	browser := chromedp.FromContext(w.ctx).Browser
	if browser == nil {
		return "", "", fmt.Errorf("pooled chrome browser is not initialized")
	}
	executor := cdp.WithExecutor(opCtx, browser)
	browserContextID, err := target.CreateBrowserContext().WithDisposeOnDetach(true).Do(executor)
	if err != nil {
		return "", "", err
	}
	targetID, err := target.CreateTarget("about:blank").
		WithBrowserContextID(browserContextID).
		WithNewWindow(true).
		Do(executor)
	if err != nil {
		_ = target.DisposeBrowserContext(browserContextID).Do(executor)
		return "", "", err
	}
	return browserContextID, targetID, nil
}

func (w *poolWorker) disposeBrowserContext(browserContextID cdp.BrowserContextID) {
	if browserContextID == "" {
		return
	}
	browser := chromedp.FromContext(w.ctx).Browser
	if browser == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = target.DisposeBrowserContext(browserContextID).Do(cdp.WithExecutor(ctx, browser))
}

func (p *Pool) releaseWorker(worker *poolWorker) {
	select {
	case p.workers <- worker:
	case <-p.ctx.Done():
	case <-p.done:
	}
}

func (p *Pool) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	close(p.done)
	workers := append([]*poolWorker(nil), p.all...)
	p.mu.Unlock()

	for _, worker := range workers {
		worker.cancel()
	}
}

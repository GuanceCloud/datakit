// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

type flusher struct {
	cat point.Category
	wal *WALQueue
	dw  *Dataway // refer to dataway instance
	idx int

	// sendBuf and marshalBuf reused during read WAL from disk-queue.
	// When read from mem-queue, these 2 buffer not used.
	sendBuf,
	marshalBuf []byte // buffer reusable during send & read HTTP body
}

// StartFlushWorkers init wal-queue on each category.
func (dw *Dataway) StartFlushWorkers() error {
	if err := dw.setupWAL(); err != nil {
		return err
	}

	worker := func(cat point.Category, n int) {
		l.Infof("start %dth workers on %q", n, cat)
		dwFlusher := datakit.G("dw-flusher/" + cat.Alias())
		for i := 0; i < n; i++ {
			dwFlusher.Go(func(_ context.Context) error {
				f := dw.newFlusher(cat)
				f.idx = i
				f.start()
				return nil
			})
		}
	}

	// start wal-queue flush workers on each category
	for cat := range dw.walq {
		n := dw.WAL.Workers
		// nolint: exhaustive
		switch cat { // some category do not need much workers
		case point.Metric,
			point.Network,
			point.Logging,
			point.Tracing,
			point.RUM:

		case point.DialTesting:
			n = 0 // dial-testing are direct to point.DynamicDWCategory
		default:
			n = 1
		}

		l.Infof("start %d flush workers on %s...", n, cat.Alias())
		worker(cat, n)
	}

	return nil
}

func (dw *Dataway) newFlusher(cat point.Category) *flusher {
	// we need extra spaces to read body and it's mata info from disk cache.
	extra := int(float64(dw.MaxRawBodySize) * .1)
	return &flusher{
		cat:        cat,
		wal:        dw.walq[cat],
		sendBuf:    make([]byte, dw.MaxRawBodySize),
		marshalBuf: make([]byte, dw.MaxRawBodySize+extra),
		dw:         dw,
	}
}

func (dw *Dataway) enqueueBody(w *writer, b *body) error {
	q := dw.walq[w.category]
	if q == nil {
		return fmt.Errorf("WAL on %s not set, should not been here", w.category)
	}

	l.Debugf("walq pub %s to %s(q: %+#v)", b, w.category.Alias(), q)

	walQueueMemLenVec.WithLabelValues(w.category.Alias()).Set(float64(len(q.mem)))

	return q.Put(b)
}

func (f *flusher) start() {
	cleanFailCacheTick := time.NewTicker(f.dw.WAL.FailCacheCleanInterval)
	defer cleanFailCacheTick.Stop()

	l.Infof("flushWorker on %s starting...", f.cat.Alias())

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Infof("dataway flush worker(%dth) on %s exit", f.idx, f.cat.Alias())
			return

		case <-cleanFailCacheTick.C:
			// try clean fail-cached data if any
			if err := f.cleanFailCache(); err != nil {
				l.Warnf("cleanFailCache: %s, ignored", err)
			}

		default: // get from WAL queue(form chan or diskcache)
			b, err := f.wal.Get(withReusableBuffer(f.sendBuf, f.marshalBuf))
			if err != nil {
				l.Warnf("Get() from wal-queue: %s, ignored", err)
			}

			if b == nil { // sleep when there is nothing to flush.
				time.Sleep(time.Second)
			} else {
				l.Debugf("walq get on %s, got body %s(from %s) payload", f.cat.Alias(), b, b.from)

				if err := f.do(b); err != nil {
					l.Warnf("do: %s, b: %s, ignored", err, b)
				}
			}
		}
	}
}

func (f *flusher) do(b *body, opts ...WriteOption) error {
	gzOn := gzipNotSet
	if f.dw.GZip {
		gzOn = gzipSet
	}

	w := getWriter(
		WithHTTPEncoding(b.enc()),
		WithGzip(gzOn),
		// cache all data into fail-cache
		WithCacheAll(true),
		WithCategory(b.cat()),
	)
	defer putWriter(w)

	return f.dw.doFlush(w, b, opts...)
}

func (dw *Dataway) doFlush(w *writer, b *body, opts ...WriteOption) error {
	for _, opt := range opts {
		opt(w)
	}

	// Append extra headers if exist.
	//
	// These headers comes from fail-cache, so we reuse them, it's import for sinked body.
	for _, h := range b.headers() {
		WithHTTPHeader(h.Key, h.Value)(w)
	}

	isGzip := "F"
	if w.cacheClean {
		isGzip = "T" // fail-cache always gzipped before HTTP POST
	}

	if dw.GZip && !w.cacheClean { // under cacheClean, all body has been gzipped during previous POST
		var (
			zstart = time.Now()
			gz     = getZipper()
		)

		isGzip = "T"
		defer putZipper(gz)

		if zbuf, err := gz.zip(b.buf()); err != nil {
			l.Errorf("gzip: %s", err.Error())
			return err
		} else {
			ncopy := copy(b.sendBuf, zbuf)
			l.Debugf("copy %d(origin: %d) zipped bytes to buf", ncopy, len(b.buf()))
			b.CacheData.Payload = b.sendBuf[:ncopy]
		}

		buildBodyCostVec.WithLabelValues(
			w.category.String(),
			w.httpEncoding.String(),
			"gzip",
		).Observe(float64(time.Since(zstart)) / float64(time.Second))
	}

	defer func() {
		// NOTE: for multiple dw.eps, here only 1 flush metric.
		walWorkerFlush.WithLabelValues(
			b.cat().Alias(),
			isGzip,
			b.from.String()).Observe(float64(len(b.buf())))

		// b always comes from pool, no matter from disk queue or mem queue.
		l.Debugf("put back %s", b)
		putBody(b)
	}()

	// drop expired packages
	if b.expired(dw.DropExpiredPackageAt) {
		l.Warnf("drop expired package %s", b.pretty())
		flushDroppedPackageVec.WithLabelValues(b.cat().String()).Inc()
		return nil
	}

	for _, ep := range dw.eps {
		if err := ep.writePointData(w, b); err != nil {
			// 4xx error do not cache data.
			if errors.Is(err, errWritePoints4XX) {
				writeDropPointsCounterVec.WithLabelValues(w.category.String(), err.Error()).Add(float64(b.npts()))
				continue // current endpoint POST 4xx ignored, but other endpoint maybe ok.
			}

			l.Errorf("writePointData: %s", err)

			// For a exist failed-cache, we do not need to re-cache it.
			// and make it fail, the diskcache will rollback and Get() the same data again.
			if w.cacheClean {
				return fmt.Errorf("clean fail-cache failed: %w", err)
			}

			//nolint:exhaustive
			switch b.cat() {
			case point.Metric, // these categories are not default cached.
				point.MetricDeprecated,
				point.Object,
				point.ObjectChange, // Deprecated.
				point.CustomObject,
				point.DynamicDWCategory:

				if !w.cacheAll {
					writeDropPointsCounterVec.WithLabelValues(w.category.String(), err.Error()).Add(float64(b.npts()))
					l.Warnf("drop %d pts on %s, not cached", b.npts, w.category)
					continue
				}

			default: // other categories are default cached.
			}

			if err := dw.dumpFailCache(b); err != nil {
				l.Errorf("dumpFailCache %v pts on %s: %s", b.npts, w.category, err)
			} else {
				l.Debugf("dumping %q to failcache ok", b)
			}
		}
	}

	return nil
}

func (dw *Dataway) dumpFailCache(b *body) error {
	if x, err := b.dump(); err != nil {
		return err
	} else {
		return dw.walFail.disk.Put(x) // directly put dumpped body to disk-queue, not mem-queue.
	}
}

func (f *flusher) cleanFailCache() error {
	return f.dw.walFail.DiskGet(func(b *body) error {
		l.Debugf("clean body %s", b)

		var ( // @b will reset within f.do(), pre-fetch it's meta for metric update.
			cat  = b.cat()
			size = len(b.buf())
		)

		if err := f.do(b, WithCacheClean(true), WithHTTPHeader("X-Fail-Cache-Retry", "1")); err != nil {
			return err
		}

		// only update metric on clean-ok
		flushFailCacheVec.WithLabelValues(cat.Alias()).Observe(float64(size))
		return nil
	}, withReusableBuffer(f.sendBuf, f.marshalBuf))
}

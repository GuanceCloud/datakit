// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"errors"
	"path/filepath"
	"time"

	"github.com/GuanceCloud/cliutils/diskcache"
	"github.com/GuanceCloud/cliutils/point"
)

var defaultRotateAt = 3 * time.Second

type Cache interface {
	// NOTE: reuse callback in diskcache to keep interface ok
	// it's better to define Get as
	//   Get() []byte
	BufGet([]byte, diskcache.Fn) error
	Put([]byte) error
	Size() int64
	Close() error
}

type WALConf struct {
	MaxCapacityGB          float64       `toml:"max_capacity_gb"`
	Path                   string        `toml:"path,omitempty"`
	NoPos                  bool          `toml:"no_pos"`
	FailCacheCleanInterval time.Duration `toml:"fail_cache_clean_interval"`

	NoDropCategories []string `toml:"no_drop_categories"`
	Workers          int      `toml:"workers"`
	MemCap           int      `toml:"mem_cap,omitempty"`
}

type WALQueue struct {
	disk   Cache
	mem    chan *body
	dw     *Dataway // back-ref to dataway configures
	noDrop bool
}

func NewWAL(dw *Dataway, c Cache) *WALQueue {
	q := &WALQueue{
		disk: c,
		dw:   dw,
	}

	if dw.WAL.Workers <= 0 {
		dw.WAL.Workers = 1 // set minimal worker, or no flush worker running.
	}

	// TIPS: set mem_cap = -1 to disable mem-queue
	if dw.WAL.MemCap != -1 && dw.WAL.Workers > 0 {
		dw.WAL.MemCap = dw.WAL.Workers
	}

	if dw.WAL.MemCap >= 0 {
		l.Infof("set wal mem queue cap to %d", dw.WAL.MemCap)
		q.mem = make(chan *body, dw.WAL.MemCap)
	}

	return q
}

// Put put a ready-to-send Dataway body to the send queue.
func (q *WALQueue) Put(b *body) error {
	select {
	case q.mem <- b: // @b will reuse by flush worker
		walPointCounterVec.WithLabelValues(b.cat().Alias(), "M").Add(float64(b.npts()))
		return nil
	default: // pass: put b into disk WAL
	}

	putStatus := ""

	l.Debugf("dump body %s to disk queue", b)

	defer func() {
		if putStatus != "" {
			walPointCounterVec.WithLabelValues(b.cat().Alias(), putStatus).Add(float64(b.npts()))
		}
		putBody(b) // b has dump to disk, do not used any more.
	}()

	if x, err := b.dump(); err != nil {
		putStatus = "drop"
		return err
	} else {
		retried := 0
	__retry:
		if err := q.disk.Put(x); err != nil {
			if errors.Is(err, diskcache.ErrCacheFull) && q.noDrop {
				time.Sleep(time.Second)
				retried++
				l.Warnf("WAL full, %d retrying...", retried)
				goto __retry
			} else {
				putStatus = "drop"
			}

			return err
		} else {
			if retried > 0 {
				l.Warnf("WAL retried %d times", retried)
				walPutRetriedVec.WithLabelValues(b.cat().Alias()).Observe(float64(retried))
			}
			// NOTE: do not set putStatus here, we'll update walPointCounterVec during Get().
			return nil
		}
	}
}

// Get fetch a ready-to-send Dataway body from the send queue.
func (q *WALQueue) Get(opts ...bodyOpt) (*body, error) {
	var b *body
	select {
	case b = <-q.mem:
		// fast path: we get body from WAL.mem
		return b, nil // NOTE: no opts are applied to @b if comes from channel
	default: // pass: then read from disk queue.
	}

	// slow path: we get body from WAL.disk
	b = getReuseBufferBody(opts...)

	defer func() {
		if len(b.buf()) == 0 { // no data read from disk
			putBody(b)
		} else {
			// Update the metric within Get,because after datakit start, there may be old
			// cached data in WAL.disk, we'd add them to current running datakit's metric.
			walPointCounterVec.WithLabelValues(b.cat().Alias(), "D").Add(float64(b.npts()))
		}
	}()

	var raw []byte
	if err := q.disk.BufGet(b.marshalBuf, func(x []byte) error {
		raw = x
		// ASAP ok on Get: we should not occupy the Get lock here, and other flush workers
		// need to read next raw body.
		return nil
	}); err != nil {
		if errors.Is(err, diskcache.ErrNoData) {
			return nil, nil
		}

		l.Errorf("BufGet: %s", err)
		return nil, err
	}

	if len(raw) == 0 { // no job available
		return nil, nil
	}

	l.Debugf("from queue get %d bytes", len(raw))

	if err := b.loadCache(raw); err != nil {
		return nil, err
	}

	b.from = walFromDisk
	b.gzon = isGzip(b.buf())

	return b, nil
}

type walBodyCallback func(*body) error

// DiskGet will fallback if callback failed.
func (q *WALQueue) DiskGet(fn walBodyCallback, opts ...bodyOpt) error {
	b := getReuseBufferBody(opts...)

	if err := q.disk.BufGet(b.marshalBuf, func(x []byte) error {
		if len(x) == 0 {
			return nil
		}

		if err := b.loadCache(x); err != nil {
			l.Warnf("load cache failed: %s, ignored", err)
			return nil
		}

		b.from = walFromDisk
		b.gzon = isGzip(b.buf())
		if err := fn(b); err != nil {
			l.Warnf("walBodyCallback: %s, we try again, ignored", err)
			return err
		} else {
			return nil
		}
	}); err != nil {
		if errors.Is(err, diskcache.ErrNoData) {
			return nil
		} else {
			return err
		}
	}

	return nil
}

func (dw *Dataway) doSetupWAL(opts ...diskcache.CacheOption) (*WALQueue, error) {
	dc, err := diskcache.Open(opts...)
	if err != nil {
		return nil, err
	}

	return NewWAL(dw, dc), nil
}

func (dw *Dataway) setupWAL() error {
	for _, cat := range point.AllCategories() {
		cacheDir := filepath.Join(dw.WAL.Path, cat.String())
		opts := []diskcache.CacheOption{
			diskcache.WithPath(cacheDir),
			diskcache.WithNoPos(dw.WAL.NoPos),
			diskcache.WithNoLock(true),            // disable .lock file checking
			diskcache.WithWakeup(defaultRotateAt), // short wakeup on WAL queue
		}

		switch cat { // nolint:exhaustive
		case point.Object, point.CustomObject:
			opts = append(opts,
				// 128MiB: object do not need too much capacity
				diskcache.WithCapacity(128*(1<<20)),
			)
		default:
			opts = append(opts,
				diskcache.WithCapacity(int64(dw.WAL.MaxCapacityGB*float64(1<<30))),
			)
		}

		if dw.isNoDropWAL(cat) {
			opts = append(opts,
				diskcache.WithNoDrop(true), // no-drop any data if cache full.
			)
		} else {
			opts = append(opts,
				diskcache.WithFILODrop(true), // drop new data if cache full, no matter normal WAL or fail-cache WAL.
			)
		}

		if wal, err := dw.doSetupWAL(opts...); err != nil {
			l.Errorf("NewWALCache %s with capacity %f GB: %s", cacheDir, dw.WAL.MaxCapacityGB, err.Error())
			return err
		} else {
			wal.noDrop = dw.isNoDropWAL(cat)
			dw.walq[cat] = wal
			l.Infof("diskcache.New on %q ok(%+#v)", cat.Alias(), wal)
		}
	}

	// setup fail-cache
	if wal, err := dw.doSetupWAL(
		diskcache.WithPath(filepath.Join(dw.WAL.Path, "fc")),
		diskcache.WithFILODrop(true), // under fail-cache, still drop data if WAL disk full(no matter which category)
		diskcache.WithNoLock(true),
		diskcache.WithNoPos(dw.WAL.NoPos),
		diskcache.WithWakeup(defaultRotateAt),
		diskcache.WithCapacity(int64(dw.WAL.MaxCapacityGB*float64(1<<30)))); err != nil {
		return err
	} else {
		dw.walFail = wal
	}
	return nil
}

func (dw *Dataway) isNoDropWAL(cat point.Category) bool {
	if dw.WAL == nil { // all categories are drop if WAL full.
		return false
	}

	for _, c := range dw.WAL.NoDropCategories {
		if c == cat.Alias() {
			return true
		}
	}

	return false
}

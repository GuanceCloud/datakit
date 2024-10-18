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
	Workers                int           `toml:"workers"`
	MemCap                 int           `toml:"mem_cap,omitempty"`
	Path                   string        `toml:"path,omitempty"`
	FailCacheCleanInterval time.Duration `toml:"fail_cache_clean_interval"`
}

type WALQueue struct {
	disk Cache
	mem  chan *body
	dw   *Dataway // back-ref to dataway configures
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
		defer putBody(b) // b has dump to disk, do not used any more.
	}()

	if x, err := b.dump(); err != nil {
		putStatus = "drop"
		return err
	} else {
		if err := q.disk.Put(x); err != nil {
			putStatus = "drop"
			return err
		} else {
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

	l.Debugf("from queue %d get bytes", len(raw))

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

func (dw *Dataway) doSetupWAL(cacheDir string) (*WALQueue, error) {
	dc, err := diskcache.Open(
		diskcache.WithPath(cacheDir),

		// drop new data if cache full, no matter normal WAL or fail-cache WAL.
		diskcache.WithFILODrop(true),

		diskcache.WithCapacity(int64(dw.WAL.MaxCapacityGB*float64(1<<30))),
		diskcache.WithWakeup(defaultRotateAt), // short wakeup on wal queue
	)
	if err != nil {
		l.Errorf("NewWALCache %s with capacity %f GB: %s", cacheDir, dw.WAL.MaxCapacityGB, err.Error())
		return nil, err
	}

	l.Infof("diskcache.New ok(%q) of %f GB", dw.WAL.Path, dw.WAL.MaxCapacityGB)

	return NewWAL(dw, dc), nil
}

func (dw *Dataway) setupWAL() error {
	for _, cat := range point.AllCategories() {
		if wal, err := dw.doSetupWAL(filepath.Join(dw.WAL.Path, cat.String())); err != nil {
			return err
		} else {
			dw.walq[cat] = wal
		}
	}

	if wal, err := dw.doSetupWAL(filepath.Join(dw.WAL.Path, "fc")); err != nil {
		return err
	} else {
		dw.walFail = wal
	}
	return nil
}

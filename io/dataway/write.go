// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"errors"

	"github.com/GuanceCloud/cliutils/diskcache"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/failcache"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	pb "google.golang.org/protobuf/proto"
)

var (
	seprator    = []byte("\n")
	MaxKodoBody = 10 * 1000 * 1000
)

type WriteOption func(w *writer)

func WithCategory(cat string) WriteOption {
	return func(w *writer) {
		w.category = cat
	}
}

func WithPoints(pts []*dkpt.Point) WriteOption {
	return func(w *writer) {
		w.pts = pts
	}
}

func WithDynamicURL(urlStr string) WriteOption {
	return func(w *writer) {
		w.dynamicURL = urlStr
	}
}

func WithFailCache(fc failcache.Cache) WriteOption {
	return func(w *writer) {
		w.fc = fc
	}
}

func WithCacheAll(on bool) WriteOption {
	return func(w *writer) {
		w.cacheAll = on
	}
}

func WithCacheClean(on bool) WriteOption {
	return func(w *writer) {
		w.cacheClean = on
	}
}

func WithGzip(on bool) WriteOption {
	return func(w *writer) {
		w.gzip = on
	}
}

type writer struct {
	category   string
	dynamicURL string

	pts                  []*dkpt.Point
	gzip                 bool
	isSinker             bool
	cacheClean, cacheAll bool

	fc failcache.Cache
}

func isGzip(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	// See: https://stackoverflow.com/a/6059342/342348
	return data[0] == 0x1f && data[1] == 0x8b
}

func (dw *Dataway) cleanCache(w *writer, data []byte) error {
	pd := &CacheData{}
	if err := pb.Unmarshal(data, pd); err != nil {
		log.Warnf("pb.Unmarshal(%d bytes -> %s): %s, ignored", len(data), w.category, err)
		return nil
	}

	cat := point.Category(pd.Category)

	WithGzip(isGzip(pd.Payload))(w) // check if bytes is gzipped
	WithCategory(cat.URL())(w)      // use category in cached data

	for _, ep := range dw.eps {
		// If some of endpoint send ok, any failed write will cause re-write on these ok ones.
		// So, do NOT configure multiple endpoint in dataway URL list.
		if err := ep.writePointData(
			&body{buf: pd.Payload}, w); err != nil {
			log.Warnf("cleanCache: %s", err)
			return err
		}
	}

	// only set metric on clean-ok
	flushFailCacheVec.WithLabelValues(cat.String()).Observe(float64(len(pd.Payload)))
	return nil
}

func (dw *Dataway) Write(opts ...WriteOption) error {
	w := getWriter()
	defer putWriter(w)

	for _, opt := range opts {
		if opt != nil {
			opt(w)
		}
	}

	if w.cacheClean {
		if w.fc == nil {
			return nil
		}

		if err := w.fc.Get(func(x []byte) error {
			if len(x) == 0 {
				return nil
			}

			log.Debugf("try flush %d bytes on %q", len(x), w.category)

			return dw.cleanCache(w, x)
		}); err != nil {
			if !errors.Is(err, diskcache.ErrEOF) {
				log.Warnf("on %s failcache.Get: %s, ignored", w.category, err)
			}
		}

		// always ok on clean-cache
		return nil
	}

	// Points in cache do not send to sinkers.
	// sink points to multiple sinkers, after sinker, not-sinked points
	// are passed to default dataway.
	if len(dw.Sinkers) > 0 {
		var remainPts []*dkpt.Point
		for _, sinker := range dw.Sinkers {
			log.Debugf("try sink %q to %s...", w.category, sinker)

			arr, err := sinker.sink(point.CatURL(w.category), w.pts)
			if err != nil {
				log.Warnf("sink %d points to %q failed, remains %d, ignored", len(w.pts), w.category, len(arr))
			}

			for _, i := range arr {
				remainPts = append(remainPts, w.pts[i])
			}
		}

		if len(remainPts) == 0 { // no point remaining
			return nil
		}

		// sending remaining points
		w.pts = remainPts
	}

	// write points to multiple endpoints
	for _, ep := range dw.eps {
		if err := ep.writePoints(w); err != nil {
			return err
		}
	}

	return nil
}

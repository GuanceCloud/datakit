// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"errors"
	"fmt"

	"github.com/GuanceCloud/cliutils/diskcache"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/failcache"
	pb "google.golang.org/protobuf/proto"
)

var MaxKodoBody = 10 * 1000 * 1000

type WriteOption func(w *writer)

func WithCategory(cat point.Category) WriteOption {
	return func(w *writer) {
		w.category = cat
	}
}

func WithPoints(points []*point.Point) WriteOption {
	return func(w *writer) {
		w.points = points
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

func WithBatchSize(n int) WriteOption {
	return func(w *writer) {
		w.batchSize = n
	}
}

func WithBatchBytesSize(n int) WriteOption {
	return func(w *writer) {
		w.batchBytesSize = n
	}
}

func WithHTTPEncoding(t point.Encoding) WriteOption {
	return func(w *writer) {
		w.httpEncoding = t
	}
}

type writer struct {
	category   point.Category
	dynamicURL string

	points []*point.Point

	// if bothe batch limit set, prefer batchBytesSize.
	batchBytesSize int // limit point pyaload bytes approximately
	batchSize      int // limit point count

	httpEncoding point.Encoding

	gzip                 bool
	cacheClean, cacheAll bool

	httpHeaders map[string]string

	fc failcache.Cache
}

func isGzip(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	// See: https://stackoverflow.com/a/6059342/342348
	return data[0] == 0x1f && data[1] == 0x8b
}

func loadCache(data []byte) (*CacheData, error) {
	pd := &CacheData{}
	if err := pb.Unmarshal(data, pd); err != nil {
		return nil, fmt.Errorf("loadCache: %w", err)
	}

	return pd, nil
}

func (dw *Dataway) cleanCache(w *writer, data []byte) error {
	pd, err := loadCache(data)
	if err != nil {
		log.Warnf("pb.Unmarshal(%d bytes -> %s): %s, ignored", len(data), w.category, err)
		return nil
	}

	cat := point.Category(pd.Category)
	enc := point.Encoding(pd.PayloadType)

	WithGzip(isGzip(pd.Payload))(w) // check if bytes is gzipped
	WithCategory(cat)(w)            // use category in cached data
	WithHTTPEncoding(enc)(w)

	for _, ep := range dw.eps {
		// If some of endpoint send ok, any failed write will cause re-write on these ok ones.
		// So, do NOT configure multiple endpoint in dataway URL list.
		if err := ep.writePointData(&body{buf: pd.Payload}, w); err != nil {
			log.Warnf("cleanCache: %s", err)
			return err
		}
	}

	// only set metric on clean-ok
	flushFailCacheVec.WithLabelValues(cat.String()).Observe(float64(len(pd.Payload)))
	return nil
}

func (dw *Dataway) doGroupPoints(ptg *ptGrouper, cat point.Category, points []*point.Point) {
	for _, pt := range points {
		// clear kvs for current pt
		ptg.kvarr = ptg.kvarr[:0]
		ptg.extKVs = ptg.extKVs[:0]

		ptg.pt = pt
		ptg.cat = cat

		tv := ptg.sinkHeaderValue(dw.globalTags, dw.GlobalCustomerKeys)

		ptg.groupedPts[tv] = append(ptg.groupedPts[tv], pt)
	}
}

func (dw *Dataway) groupPoints(ptg *ptGrouper, cat point.Category, points []*point.Point) {
	dw.doGroupPoints(ptg, cat, points)
	groupedRequestVec.WithLabelValues(cat.String()).Observe(float64(len(ptg.groupedPts)))
}

func (dw *Dataway) Write(opts ...WriteOption) error {
	w := getWriter()
	defer putWriter(w)

	for _, opt := range opts {
		if opt != nil {
			opt(w)
		}
	}

	// set content encoding(protobuf/line-protocol/json)
	WithHTTPEncoding(dw.contentEncoding)(w)

	// setup gzip on or off
	WithGzip(dw.GZip)(w)

	// set raw body size limit
	if dw.MaxRawBodySize > 0 {
		WithBatchBytesSize(dw.MaxRawBodySize)(w)
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

	// split single point array into multiple part according to
	// different X-Global-Tags.
	if dw.EnableSinker &&
		(len(dw.globalTags) > 0 || len(dw.GlobalCustomerKeys) > 0) &&
		len(dw.eps) > 0 {
		ptg := getGrouper()
		defer putGrouper(ptg)

		dw.groupPoints(ptg, w.category, w.points)

		for k, points := range ptg.groupedPts {
			w.httpHeaders[HeaderXGlobalTags] = k
			w.points = points

			// only apply to 1st dataway address
			if err := dw.eps[0].writePoints(w); err != nil {
				return err
			}
		}
	} else {
		// write points to multiple endpoints
		for _, ep := range dw.eps {
			if err := ep.writePoints(w); err != nil {
				return err
			}
		}
	}

	return nil
}

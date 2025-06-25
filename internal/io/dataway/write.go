// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"fmt"

	"github.com/GuanceCloud/cliutils/point"
)

type WriteOption func(w *writer)

func WithCategory(cat point.Category) WriteOption {
	return func(w *writer) {
		w.category = cat
	}
}

func WithHTTPHeader(k, v string) WriteOption {
	return func(w *writer) {
		w.httpHeaders[k] = v
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

func WithStorageIndex(name string) WriteOption {
	return func(w *writer) {
		w.indexName = name
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

func WithGzip(on gzipFlag) WriteOption {
	return func(w *writer) {
		w.gzip = on
	}
}

func WithGzipDuringBuildBody(on bool) WriteOption {
	return func(w *writer) {
		w.gzipDuringBuildBody = on
	}
}

func WithBatchSize(n int) WriteOption {
	return func(w *writer) {
		w.batchSize = n
	}
}

func WithHTTPEncoding(t point.Encoding) WriteOption {
	return func(w *writer) {
		w.httpEncoding = t
	}
}

func WithMaxBodyCap(x int) WriteOption {
	return func(w *writer) {
		if x > 0 {
			w.batchBytesSize = x
		}
	}
}

func WithBodyCallback(cb bodyCallback) WriteOption {
	return func(w *writer) {
		w.bcb = cb
	}
}

func WithNoWAL(on bool) WriteOption {
	return func(w *writer) {
		w.noWAL = on
	}
}

type writer struct {
	category point.Category

	indexName,
	dynamicURL string

	points []*point.Point

	// if bothe batch limit set, prefer batchBytesSize.
	batchBytesSize int // limit point pyaload bytes approximately
	batchSize      int // limit point count

	httpEncoding point.Encoding

	gzip gzipFlag
	cacheClean,
	cacheAll,
	noWAL,
	gzipDuringBuildBody bool

	httpHeaders map[string]string

	bcb bodyCallback
}

func (w *writer) reset() {
	w.category = point.UnknownCategory
	w.dynamicURL = ""
	w.indexName = ""
	w.points = w.points[:0]
	w.gzip = gzipNotSet
	w.cacheClean = false
	w.cacheAll = false
	w.noWAL = false
	w.gzipDuringBuildBody = false
	w.batchBytesSize = defaultBatchSize
	w.batchSize = 0
	w.bcb = nil

	for k := range w.httpHeaders {
		delete(w.httpHeaders, k)
	}
	w.httpEncoding = encNotSet
}

func (dw *Dataway) doGroupPoints(ptg *ptGrouper, cat point.Category, points []*point.Point) {
	for _, pt := range points {
		// clear kvs for current pt
		ptg.kvarr = ptg.kvarr[:0]
		ptg.extKVs = ptg.extKVs[:0]

		ptg.pt = pt
		ptg.cat = cat

		tv := ptg.sinkHeaderValue(dw.globalTags, dw.GlobalCustomerKeys)

		l.Debugf("add point to group %q", tv)

		ptg.groupedPts[tv] = append(ptg.groupedPts[tv], pt)
	}
}

func (dw *Dataway) groupPoints(ptg *ptGrouper,
	cat point.Category,
	points []*point.Point,
) {
	dw.doGroupPoints(ptg, cat, points)
	groupedRequestVec.WithLabelValues(cat.String()).Observe(float64(len(ptg.groupedPts)))
}

func (dw *Dataway) Write(opts ...WriteOption) error {
	gzOn := gzipNotSet
	if dw.GZip {
		gzOn = gzipSet
	}

	w := getWriter(
		// set content encoding(protobuf/line-protocol/json)
		WithHTTPEncoding(dw.contentEncoding),
		// setup gzip on or off
		WithGzip(gzOn),
		// set raw body size limit
		WithMaxBodyCap(dw.MaxRawBodySize),
	)

	defer putWriter(w)

	// Append extra wirte options from caller
	for _, opt := range opts {
		if opt != nil {
			opt(w)
		}
	}

	// apply index name to HTTP header.
	if w.indexName != "" {
		WithHTTPHeader(HeaderXStorageIndexName, w.indexName)(w)
	}

	if w.bcb == nil { // set default callback
		if w.noWAL {
			if len(dw.eps) == 0 {
				return fmt.Errorf("no endpoints on dataway, should not been here")
			}

			// NOTE: only send to 1st dataway endpoint.
			w.bcb = dw.eps[0].writePointData
		} else {
			w.bcb = dw.enqueueBody // enqueu to WAL
		}
	}

	// split single point array into multiple part according to
	// different X-Global-Tags.
	if dw.sinkEnabled() {
		l.Debugf("under sinker...")

		ptg := getGrouper()
		defer putGrouper(ptg)

		dw.groupPoints(ptg, w.category, w.points)

		for k, points := range ptg.groupedPts {
			WithHTTPHeader(HeaderXGlobalTags, k)(w)
			WithPoints(points)(w)

			if err := w.buildPointsBody(); err != nil {
				return err
			}
		}
	} else {
		if err := w.buildPointsBody(); err != nil {
			return err
		}
	}

	return nil
}

func (dw *Dataway) sinkEnabled() bool {
	return dw.EnableSinker &&
		(len(dw.globalTags) > 0 || len(dw.GlobalCustomerKeys) > 0) &&
		len(dw.eps) > 0
}

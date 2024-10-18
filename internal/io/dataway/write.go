// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bytes"
	"compress/gzip"

	"github.com/GuanceCloud/cliutils/point"
)

var MaxKodoBody = 10 * 1000 * 1000

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

func WithGzip(on int) WriteOption {
	return func(w *writer) {
		w.gzip = on
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

type gzipWriter struct {
	buf *bytes.Buffer
	w   *gzip.Writer
}

type writer struct {
	category   point.Category
	dynamicURL string

	points []*point.Point

	// if bothe batch limit set, prefer batchBytesSize.
	batchBytesSize int // limit point pyaload bytes approximately
	batchSize      int // limit point count

	httpEncoding point.Encoding

	gzip                 int
	cacheClean, cacheAll bool

	httpHeaders map[string]string

	bcb bodyCallback
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
	gzOn := 0
	if dw.GZip {
		gzOn = 1
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

	if w.bcb == nil { // set default callback
		w.bcb = dw.enqueueBody
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

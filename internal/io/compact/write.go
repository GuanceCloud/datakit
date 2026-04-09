// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package compact

import (
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
)

type WriteOption func(w *Writer)

func WithCategory(cat point.Category) WriteOption {
	return func(w *Writer) {
		w.Category = cat
	}
}

func WithHTTPHeader(k, v string) WriteOption {
	return func(w *Writer) {
		w.HTTPHeaders[k] = v
	}
}

func WithPoints(points []*point.Point) WriteOption {
	return func(w *Writer) {
		w.Points = points
	}
}

func WithDynamicURL(urlStr string) WriteOption {
	return func(w *Writer) {
		w.DynamicURL = urlStr
	}
}

func WithStorageIndex(name string) WriteOption {
	return func(w *Writer) {
		w.IndexName = name
	}
}

func WithCacheAll(on bool) WriteOption {
	return func(w *Writer) {
		w.CacheAll = on
	}
}

func WithCacheClean(on bool) WriteOption {
	return func(w *Writer) {
		w.CacheClean = on
	}
}

func WithGzip(on GzipFlag) WriteOption {
	return func(w *Writer) {
		w.Gzip = on
	}
}

func WithGzipDuringBuildBody(on bool) WriteOption {
	return func(w *Writer) {
		w.gzipDuringBuildBody = on
	}
}

func WithBatchSize(n int) WriteOption {
	return func(w *Writer) {
		w.batchSize = n
	}
}

func WithHTTPEncoding(t point.Encoding) WriteOption {
	return func(w *Writer) {
		w.HTTPEncoding = t
	}
}

func WithMaxBodyCap(x int) WriteOption {
	return func(w *Writer) {
		if x > 0 {
			w.batchBytesSize = x
		}
	}
}

func WithBodyCallback(cb bodyCallback) WriteOption {
	return func(w *Writer) {
		w.Callback = cb
	}
}

func WithNoWAL(on bool) WriteOption {
	return func(w *Writer) {
		w.NoWAL = on
	}
}

type Writer struct {
	Category point.Category

	IndexName,

	DynamicURL string

	// if bothe batch limit set, prefer batchBytesSize.
	batchBytesSize int // limit point pyaload bytes approximately
	batchSize      int // limit point count

	Callback bodyCallback

	HTTPEncoding point.Encoding

	Gzip GzipFlag
	CacheClean,
	CacheAll,
	NoWAL,
	gzipDuringBuildBody bool

	HTTPHeaders map[string]string

	Points []*point.Point
}

func (w *Writer) reset() {
	w.Category = point.UnknownCategory
	w.DynamicURL = ""
	w.IndexName = ""
	w.Points = w.Points[:0]
	w.Gzip = GzipNotSet
	w.CacheClean = false
	w.CacheAll = false
	w.NoWAL = false
	w.gzipDuringBuildBody = false
	w.batchBytesSize = defaultBatchSize
	w.batchSize = 0
	w.Callback = nil

	for k := range w.HTTPHeaders {
		delete(w.HTTPHeaders, k)
	}
	w.HTTPEncoding = encNotSet
}

func Setup() {
	l = logger.SLogger("compact")
}

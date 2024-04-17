// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	sync "sync"

	"github.com/GuanceCloud/cliutils/point"
)

var wpool sync.Pool

func getWriter(opts ...WriteOption) *writer {
	var w *writer

	if x := wpool.Get(); x == nil {
		w = &writer{
			httpHeaders:    map[string]string{},
			batchBytesSize: 1 << 20, // 1MB
			body:           &body{},
		}
	} else {
		w = x.(*writer)
	}

	for _, opt := range opts {
		if opt != nil {
			opt(w)
		}
	}

	return w
}

func putWriter(w *writer) {
	w.category = point.UnknownCategory
	w.dynamicURL = ""
	w.points = w.points[:0]
	w.gzip = false
	w.cacheClean = false
	w.cacheAll = false
	w.batchBytesSize = 1 << 20
	w.batchSize = 0
	w.fc = nil
	w.parts = 0
	w.body.reset()

	if w.zipper != nil {
		w.zipper.buf.Reset()
		w.zipper.w.Reset(w.zipper.buf)
	}

	for k := range w.httpHeaders {
		delete(w.httpHeaders, k)
	}
	w.httpEncoding = point.LineProtocol
	wpool.Put(w)
}

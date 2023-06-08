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

func getWriter() *writer {
	w := wpool.Get()
	if w == nil {
		w = &writer{}
	}

	return w.(*writer)
}

func putWriter(w *writer) {
	w.category = point.UnknownCategory
	w.dynamicURL = ""
	w.pts = w.pts[:0]
	w.gzip = false
	w.cacheClean = false
	w.cacheAll = false
	w.fc = nil
	wpool.Put(w)
}

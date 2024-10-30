// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	bytes "bytes"
	"sync"

	gzip "github.com/klauspost/compress/gzip"
)

var zippool sync.Pool

func getZipper() *gzipWriter {
	if x := zippool.Get(); x == nil {
		buf := bytes.Buffer{}
		w := gzip.NewWriter(&buf)
		return &gzipWriter{buf: &buf, w: w}
	} else {
		return x.(*gzipWriter)
	}
}

func putZipper(z *gzipWriter) {
	if z != nil {
		// reset zip buffer and the writer.
		z.buf.Reset()
		z.w.Reset(z.buf)
		zippool.Put(z)
	}
}

type gzipWriter struct {
	buf *bytes.Buffer
	w   *gzip.Writer
}

func (z *gzipWriter) zip(data []byte) ([]byte, error) {
	if _, err := z.w.Write(data); err != nil {
		return nil, err
	}

	if err := z.w.Flush(); err != nil {
		return nil, err
	}

	if err := z.w.Close(); err != nil {
		return nil, err
	}

	return z.buf.Bytes(), nil
}

func isGzip(data []byte) int8 {
	if len(data) < 2 {
		return -1
	}

	// See: https://stackoverflow.com/a/6059342/342348
	if data[0] == 0x1f && data[1] == 0x8b {
		return 1
	} else {
		return 0
	}
}

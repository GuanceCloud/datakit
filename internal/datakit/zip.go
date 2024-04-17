// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package datakit

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sync"
)

var gzReaderPool,
	gzWriterPool sync.Pool

type gzipWriter struct {
	buf *bytes.Buffer
	z   *gzip.Writer
}

type gzipReader struct {
	reader *bytes.Reader
	z      *gzip.Reader
}

func getGZipReader(data []byte) *gzipReader {
	if x := gzReaderPool.Get(); x == nil {
		reader := bytes.NewReader(data)
		z, err := gzip.NewReader(reader)
		if err != nil {
			panic(fmt.Sprintf("should not been here: %q", err.Error()))
		}

		return &gzipReader{
			z:      z,
			reader: reader,
		}
	} else {
		r := x.(*gzipReader)
		r.reader.Reset(data) // reader point to new data
		if err := r.z.Reset(r.reader); err != nil {
			return nil
		}

		return r
	}
}

func putGZipReader(r *gzipReader) {
	if r == nil {
		return
	}

	gzReaderPool.Put(r)
}

func getGZipWriter() *gzipWriter {
	if x := gzWriterPool.Get(); x == nil {
		buf := bytes.Buffer{}
		return &gzipWriter{
			buf: &buf,
			z:   gzip.NewWriter(&buf),
		}
	} else {
		return x.(*gzipWriter)
	}
}

func putGZipWriter(w *gzipWriter) {
	if w == nil {
		return
	}

	w.buf.Reset()
	w.z.Reset(w.buf)
	gzWriterPool.Put(w)
}

func GZipStrV2(str string) ([]byte, error) {
	zw := getGZipWriter()
	defer putGZipWriter(zw)

	if _, err := io.WriteString(zw.z, str); err != nil {
		return nil, err
	}

	if err := zw.z.Flush(); err != nil {
		return nil, err
	}

	if err := zw.z.Close(); err != nil {
		return nil, err
	}
	return zw.buf.Bytes(), nil
}

func GZipStr(str string) ([]byte, error) {
	var z bytes.Buffer
	zw := gzip.NewWriter(&z)
	if _, err := io.WriteString(zw, str); err != nil {
		return nil, err
	}

	if err := zw.Flush(); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return z.Bytes(), nil
}

func GZip(data []byte) ([]byte, error) {
	zw := getGZipWriter()
	defer putGZipWriter(zw)

	if _, err := zw.z.Write(data); err != nil {
		return nil, err
	}

	if err := zw.z.Flush(); err != nil {
		return nil, err
	}

	if err := zw.z.Close(); err != nil {
		return nil, err
	}

	return zw.buf.Bytes(), nil
}

func UnGZip(data []byte) ([]byte, error) {
	zr := getGZipReader(data)
	defer putGZipReader(zr)

	raw, err := io.ReadAll(zr.z)
	if err != nil {
		return nil, err
	}

	return raw, zr.z.Close()
}

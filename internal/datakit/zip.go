// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package datakit

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"sync"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
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

var deflateReaderPool,
	deflateWriterPool sync.Pool

type deflateWriter struct {
	buf *bytes.Buffer
	z   *flate.Writer
}

type deflateReader struct {
	reader *bytes.Reader
	z      io.ReadCloser
}

func getDeflateWriter() *deflateWriter {
	if x := deflateWriterPool.Get(); x == nil {
		buf := new(bytes.Buffer)
		z, _ := flate.NewWriter(buf, flate.DefaultCompression)
		return &deflateWriter{
			buf: buf,
			z:   z,
		}
	} else {
		return x.(*deflateWriter)
	}
}

func putDeflateWriter(w *deflateWriter) {
	if w == nil {
		return
	}
	w.buf.Reset()
	w.z.Reset(w.buf)
	deflateWriterPool.Put(w)
}

func getDeflateReader(data []byte) *deflateReader {
	if x := deflateReaderPool.Get(); x == nil {
		reader := bytes.NewReader(data)
		z := flate.NewReader(reader)
		return &deflateReader{
			z:      z,
			reader: reader,
		}
	} else {
		r := x.(*deflateReader)
		r.reader.Reset(data)
		if err := r.z.(flate.Resetter).Reset(r.reader, nil); err != nil {
			return nil
		}
		return r
	}
}

func putDeflateReader(r *deflateReader) {
	if r == nil {
		return
	}
	deflateReaderPool.Put(r)
}

func DeflateZip(data []byte) ([]byte, error) {
	zw := getDeflateWriter()
	defer putDeflateWriter(zw)

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

func UnDeflateZip(data []byte) ([]byte, error) {
	zr := getDeflateReader(data)
	defer putDeflateReader(zr)

	raw, err := io.ReadAll(zr.z)
	if err != nil {
		return nil, err
	}

	return raw, zr.z.Close()
}

var brReaderPool,
	brWriterPool sync.Pool

type brotliWriter struct {
	buf *bytes.Buffer
	z   *brotli.Writer
}

type brotliReader struct {
	reader *bytes.Reader
	z      *brotli.Reader
}

func getBrotliWriter() *brotliWriter {
	if x := brWriterPool.Get(); x == nil {
		buf := new(bytes.Buffer)
		z := brotli.NewWriter(buf)
		return &brotliWriter{
			buf: buf,
			z:   z,
		}
	} else {
		return x.(*brotliWriter)
	}
}

func putBrotliWriter(w *brotliWriter) {
	if w == nil {
		return
	}
	w.buf.Reset()
	w.z.Reset(w.buf)
	brWriterPool.Put(w)
}

func getBrotliReader(data []byte) *brotliReader {
	if x := brReaderPool.Get(); x == nil {
		reader := bytes.NewReader(data)
		z := brotli.NewReader(reader)
		return &brotliReader{
			z:      z,
			reader: reader,
		}
	} else {
		r := x.(*brotliReader)
		r.reader.Reset(data)
		if err := r.z.Reset(r.reader); err != nil {
			return nil
		}
		return r
	}
}

func putBrotliReader(r *brotliReader) {
	if r == nil {
		return
	}
	brReaderPool.Put(r)
}

func BrotliZip(data []byte) ([]byte, error) {
	zw := getBrotliWriter()
	defer putBrotliWriter(zw)

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

func UnBrotliZip(data []byte) ([]byte, error) {
	zr := getBrotliReader(data)
	defer putBrotliReader(zr)

	raw, err := io.ReadAll(zr.z)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

var zstdReaderPool,
	zstdWriterPool sync.Pool

type zstdWriter struct {
	buf *bytes.Buffer
	z   *zstd.Encoder
}

type zstdReader struct {
	reader *bytes.Reader
	z      *zstd.Decoder
}

func getZstdWriter() *zstdWriter {
	if x := zstdWriterPool.Get(); x == nil {
		buf := new(bytes.Buffer)
		z, _ := zstd.NewWriter(buf)
		return &zstdWriter{
			buf: buf,
			z:   z,
		}
	} else {
		return x.(*zstdWriter)
	}
}

func putZstdWriter(w *zstdWriter) {
	if w == nil {
		return
	}
	w.buf.Reset()
	w.z.Reset(w.buf)
	zstdWriterPool.Put(w)
}

func getZstdReader(data []byte) *zstdReader {
	if x := zstdReaderPool.Get(); x == nil {
		reader := bytes.NewReader(data)
		z, _ := zstd.NewReader(reader)
		return &zstdReader{
			z:      z,
			reader: reader,
		}
	} else {
		r := x.(*zstdReader)
		r.reader.Reset(data)
		if err := r.z.Reset(r.reader); err != nil {
			return nil
		}
		return r
	}
}

func putZstdReader(r *zstdReader) {
	if r == nil {
		return
	}
	zstdReaderPool.Put(r)
}

func ZstdZip(data []byte) ([]byte, error) {
	zw := getZstdWriter()
	defer putZstdWriter(zw)

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

func UnZstdZip(data []byte) ([]byte, error) {
	zr := getZstdReader(data)
	defer putZstdReader(zr)

	raw, err := io.ReadAll(zr.z)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

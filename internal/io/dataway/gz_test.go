// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	bytes "bytes"
	"compress/gzip"
	"io"
	"runtime"
	T "testing"

	"github.com/GuanceCloud/cliutils/point"
	kgzip "github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/s2"
	"github.com/klauspost/compress/snappy"
	pgzip "github.com/klauspost/pgzip"
	"github.com/stretchr/testify/assert"
)

func TestEqualGZip(t *T.T) {
	r := point.NewRander()
	pts := r.Rand(1000)

	enc := point.GetEncoder(point.WithEncEncoding(point.Protobuf))
	defer point.PutEncoder(enc)

	arr, err := enc.Encode(pts)
	assert.NoError(t, err)

	assert.Len(t, arr, 1)

	var buf bytes.Buffer

	// pgzip
	pgz := pgzip.NewWriter(&buf)
	pgz.SetConcurrency(128*(1<<10), 1)

	_, err = pgz.Write(arr[0])
	assert.NoError(t, err)
	assert.NoError(t, pgz.Close())

	pgzBytes := buf.Bytes()
	t.Logf("raw data: %d bytes, pgzip: %d", len(arr[0]), len(pgzBytes))

	assert.Equal(t, gzipSet, isGzip(pgzBytes))

	// unzip pgzBytes with go's gzip
	gzr := bytes.NewBuffer(pgzBytes)
	gogz, err := gzip.NewReader(gzr)
	assert.NoError(t, err)

	unzipBuf := bytes.NewBuffer(nil)
	_, err = io.Copy(unzipBuf, gogz)
	assert.NoError(t, err)
	assert.Equal(t, arr[0], unzipBuf.Bytes())
	assert.NoError(t, gogz.Close())

	// kgzip
	buf.Reset()
	kgz := kgzip.NewWriter(&buf)

	_, err = kgz.Write(arr[0])
	assert.NoError(t, err)
	assert.NoError(t, kgz.Close())

	kgzBytes := buf.Bytes()
	t.Logf("raw data: %d bytes, kgzip: %d", len(arr[0]), len(kgzBytes))

	assert.Equal(t, gzipSet, isGzip(kgzBytes))

	// unzip pgzBytes with go's gzip
	gzr = bytes.NewBuffer(kgzBytes)
	gogz, err = gzip.NewReader(gzr)
	assert.NoError(t, err)

	unzipBuf.Reset()
	_, err = io.Copy(unzipBuf, gogz)
	assert.NoError(t, err)
	assert.Equal(t, arr[0], unzipBuf.Bytes())
	assert.NoError(t, gogz.Close())
}

func BenchmarkPGZip(b *T.B) {
	r := point.NewRander()
	pts := r.Rand(1000)

	enc := point.GetEncoder(point.WithEncEncoding(point.Protobuf))
	defer point.PutEncoder(enc)

	arr, err := enc.Encode(pts)
	assert.NoError(b, err)

	assert.Len(b, arr, 1)

	for _, bc := range []struct {
		name        string
		block, para int
	}{
		{
			"1k-1core",
			1 << 10,
			1,
		},
		{
			"4k-1core",
			4 * (1 << 10),
			1,
		},

		{
			"8k-1core",
			8 * (1 << 10),
			1,
		},

		{
			"128k-1core",
			128 * (1 << 10),
			1,
		},

		{
			"1m-1core",
			(1 << 20),
			1,
		},

		{
			"1k",
			1 << 10,
			runtime.GOMAXPROCS(0),
		},
		{
			"4k",
			4 * (1 << 10),
			runtime.GOMAXPROCS(0),
		},

		{
			"8k",
			8 * (1 << 10),
			runtime.GOMAXPROCS(0),
		},

		{
			"128k",
			128 * (1 << 10),
			runtime.GOMAXPROCS(0),
		},

		{
			"1m",
			(1 << 20),
			runtime.GOMAXPROCS(0),
		},
	} {
		b.Run(bc.name, func(b *T.B) {
			var buf bytes.Buffer
			w := pgzip.NewWriter(&buf)
			w.SetConcurrency(bc.block, 1)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w.Write(arr[0])
				buf.Reset()
				w.Reset(&buf)
			}
		})
	}
}

func BenchmarkGZip(b *T.B) {
	r := point.NewRander()
	pts := r.Rand(1000)
	enc := point.GetEncoder(point.WithEncEncoding(point.Protobuf))
	defer point.PutEncoder(enc)

	arr, err := enc.Encode(pts)
	assert.NoError(b, err)

	assert.Len(b, arr, 1)

	b.Run("pgzip", func(b *T.B) {
		var buf bytes.Buffer
		w := pgzip.NewWriter(&buf)
		w.SetConcurrency(128*(1<<10), 1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w.Write(arr[0])
			w.Flush()
			buf.Reset()
			w.Reset(&buf)
		}
	})

	b.Run("gogzip", func(b *T.B) {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w.Write(arr[0])
			w.Flush()
			buf.Reset()
			w.Reset(&buf)
		}
	})

	b.Run("kgzip", func(b *T.B) {
		var buf bytes.Buffer
		w := kgzip.NewWriter(&buf)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w.Write(arr[0])
			w.Flush()
			buf.Reset()
			w.Reset(&buf)
		}
	})

	b.Run(`snapy`, func(b *T.B) {
		var buf bytes.Buffer
		w := snappy.NewWriter(&buf)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := w.Write(arr[0])
			assert.NoError(b, err)
			assert.NoError(b, w.Flush())
			buf.Reset()
			w.Reset(&buf)
		}
	})

	b.Run(`s2`, func(b *T.B) {
		var buf bytes.Buffer
		w := s2.NewWriter(&buf)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := w.Write(arr[0])
			assert.NoError(b, err)
			assert.NoError(b, w.Flush())
			buf.Reset()
			w.Reset(&buf)
		}
	})
}

func TestRatio(t *T.T) {
	r := point.NewRander()
	pts := r.Rand(1000)

	t.Run(`snapy`, func(t *T.T) {
		enc := point.GetEncoder(point.WithEncEncoding(point.Protobuf))
		defer point.PutEncoder(enc)

		arr, err := enc.Encode(pts)
		assert.NoError(t, err)

		assert.Len(t, arr, 1)

		var buf bytes.Buffer
		w := snappy.NewWriter(&buf)

		_, err = w.Write(arr[0])
		assert.NoError(t, err)
		assert.NoError(t, w.Flush())
		w.Reset(&buf)

		t.Logf("ratio: %.04f", float64(len(buf.Bytes()))/float64(len(arr[0])))
	})

	t.Run(`s2`, func(t *T.T) {
		enc := point.GetEncoder(point.WithEncEncoding(point.Protobuf))
		defer point.PutEncoder(enc)

		arr, err := enc.Encode(pts)
		assert.NoError(t, err)

		assert.Len(t, arr, 1)

		var buf bytes.Buffer
		w := s2.NewWriter(&buf)

		_, err = w.Write(arr[0])
		assert.NoError(t, err)
		assert.NoError(t, w.Flush())
		w.Reset(&buf)

		t.Logf("ratio: %.04f", float64(len(buf.Bytes()))/float64(len(arr[0])))
	})

	t.Run(`kgzip`, func(t *T.T) {
		enc := point.GetEncoder(point.WithEncEncoding(point.Protobuf))
		defer point.PutEncoder(enc)

		arr, err := enc.Encode(pts)
		assert.NoError(t, err)

		assert.Len(t, arr, 1)

		var buf bytes.Buffer
		w := kgzip.NewWriter(&buf)

		_, err = w.Write(arr[0])
		assert.NoError(t, err)
		assert.NoError(t, w.Flush())
		assert.NoError(t, w.Close())

		t.Logf("ratio: %.04f", float64(len(buf.Bytes()))/float64(len(arr[0])))
	})

	t.Run(`kgzip-lp`, func(t *T.T) {
		enc := point.GetEncoder(point.WithEncEncoding(point.LineProtocol))
		defer point.PutEncoder(enc)

		arr, err := enc.Encode(pts)
		assert.NoError(t, err)

		assert.Len(t, arr, 1)

		var buf bytes.Buffer
		w := kgzip.NewWriter(&buf)

		_, err = w.Write(arr[0])
		assert.NoError(t, err)
		assert.NoError(t, w.Flush())
		assert.NoError(t, w.Close())

		t.Logf("ratio: %.04f", float64(len(buf.Bytes()))/float64(len(arr[0])))
	})
}

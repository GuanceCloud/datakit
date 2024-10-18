// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/diskcache"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWALLoad(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		dw := NewDefaultDataway()

		dw.WAL.Path = t.TempDir()

		assert.NoError(t, dw.Init())
		assert.NoError(t, dw.setupWAL())

		cat := point.Logging
		pts := point.RandPoints(100)
		w := getWriter(WithPoints(pts),
			WithCategory(cat),
			WithBodyCallback(dw.enqueueBody),
			WithHTTPEncoding(dw.contentEncoding))

		w.buildPointsBody()

		b, err := dw.walq[cat].Get()
		require.NoError(t, err)
		require.NotNil(t, b)
		assert.Equal(t, walFromMem, b.from)

		defer putBody(b)

		dec := point.GetDecoder(point.WithDecEncoding(dw.contentEncoding))
		defer point.PutDecoder(dec)

		// check if body in WAL are the same as @pts
		got, err := dec.Decode(b.buf())
		assert.NoError(t, err)
		assert.Equal(t, len(pts), len(got))
	})

	t.Run(`no-mem-queue`, func(t *T.T) {
		dw := NewDefaultDataway()

		dw.WAL.Path = t.TempDir()
		dw.WAL.MemCap = -1 // disable mem-queue

		assert.NoError(t, dw.Init())
		assert.NoError(t, dw.setupWAL())

		cat := point.Logging
		pts := point.RandPoints(100)
		w := getWriter(WithPoints(pts),
			WithCategory(cat),
			WithBodyCallback(dw.enqueueBody),
			WithHTTPEncoding(dw.contentEncoding))

		w.buildPointsBody()

		dc := dw.walq[cat].disk.(*diskcache.DiskCache)
		assert.NoError(t, dc.Rotate()) // force rotate

		f := dw.newFlusher(cat)

		b, err := f.wal.Get(withReusableBuffer(f.sendBuf, f.marshalBuf))
		require.NoError(t, err)
		require.NotNil(t, b)
		assert.Equal(t, walFromDisk, b.from)

		defer putBody(b)

		dec := point.GetDecoder(point.WithDecEncoding(dw.contentEncoding))
		defer point.PutDecoder(dec)

		// check if body in WAL are the same as @pts
		got, err := dec.Decode(b.buf())
		assert.NoError(t, err)
		assert.Equal(t, len(pts), len(got))
	})

	t.Run(`no-mem-queue-auto-rotate`, func(t *T.T) {
		dw := NewDefaultDataway()

		dw.WAL.Path = t.TempDir()
		dw.WAL.MemCap = -1 // disable mem-queue

		assert.NoError(t, dw.Init())
		assert.NoError(t, dw.setupWAL())

		cat := point.Logging
		pts := point.RandPoints(100)
		w := getWriter(WithPoints(pts),
			WithCategory(cat),
			WithBodyCallback(dw.enqueueBody),
			WithHTTPEncoding(dw.contentEncoding))

		w.buildPointsBody()

		time.Sleep(time.Second * 4) // default auto rotate is 3sec

		f := dw.newFlusher(cat)

		b, err := f.wal.Get(withReusableBuffer(f.sendBuf, f.marshalBuf))
		require.NoError(t, err)
		require.NotNil(t, b)
		assert.Equal(t, walFromDisk, b.from)

		defer putBody(b)

		dec := point.GetDecoder(point.WithDecEncoding(dw.contentEncoding))
		defer point.PutDecoder(dec)

		// check if body in WAL are the same as @pts
		got, err := dec.Decode(b.buf())
		assert.NoError(t, err)
		assert.Equal(t, len(pts), len(got))
	})

	t.Run(`full-mem-queue`, func(t *T.T) {
		dw := NewDefaultDataway()

		dw.WAL.Path = t.TempDir()

		assert.NoError(t, dw.Init())
		assert.NoError(t, dw.setupWAL())

		cat := point.Logging
		pts := point.RandPoints(100)
		w := getWriter(WithPoints(pts),
			WithCategory(cat),
			WithBodyCallback(dw.enqueueBody),
			WithHTTPEncoding(dw.contentEncoding))

		w.buildPointsBody()
		w.buildPointsBody() // 2nd write will dump to disk

		time.Sleep(time.Second * 4) // default auto rotate is 3sec

		f := dw.newFlusher(cat)

		for i := 0; i < 2; i++ {
			b, err := f.wal.Get(withReusableBuffer(f.sendBuf, f.marshalBuf))
			require.NoError(t, err)
			require.NotNil(t, b)

			dec := point.GetDecoder(point.WithDecEncoding(dw.contentEncoding))
			defer point.PutDecoder(dec)

			// check if body in WAL are the same as @pts
			got, err := dec.Decode(b.buf())
			assert.NoError(t, err)
			assert.Equal(t, len(pts), len(got))
			if i == 0 { // from mem
				assert.Equal(t, walFromMem, b.from)
			} else { // from disk
				assert.Equal(t, walFromDisk, b.from)
			}

			putBody(b)
		}

		b, err := f.wal.Get(withReusableBuffer(f.sendBuf, f.marshalBuf)) // no data any more
		assert.Nil(t, b)
		assert.NoError(t, err)
	})
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"strings"
	sync "sync"
	T "testing"

	"github.com/GuanceCloud/cliutils/metrics"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	gofakeit "github.com/brianvoe/gofakeit/v6"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkipLargePoint(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		metricsReset()

		var kvsLarge, kvsSmall point.KVs
		kvsLarge = kvsLarge.Add(`large`, strings.Repeat(`x`, 1024), false, false)
		kvsSmall = kvsSmall.Add(`small`, strings.Repeat(`x`, 10), false, false)
		smallPt1 := point.NewPointV2(`small1`, kvsSmall)
		largePt := point.NewPointV2(`large`, kvsLarge)
		smallPt2 := point.NewPointV2(`small2`, kvsSmall)

		w := getWriter(WithHTTPEncoding(point.Protobuf),
			WithMaxBodyCap(1024),
			WithPoints([]*point.Point{smallPt1, largePt, smallPt2, largePt}),
			WithCategory(point.Logging),
			WithBodyCallback(func(w *writer, b *body) error {
				// TODO
				return nil
			}))
		defer putWriter(w)

		assert.NoError(t, w.buildPointsBody())

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)

		mfs, err := reg.Gather()
		require.NoError(t, err)
		t.Logf("get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_skipped_point_total`,
			`logging`)
		require.NotNil(t, m)
		assert.Equal(t, float64(2), m.GetCounter().GetValue())
	})

	t.Run(`skip-all`, func(t *T.T) {
		metricsReset()

		var kvsLarge point.KVs
		kvsLarge = kvsLarge.Add(`large`, strings.Repeat(`x`, 1024), false, false)
		largePt := point.NewPointV2(`large`, kvsLarge)

		w := getWriter(WithHTTPEncoding(point.Protobuf),
			WithMaxBodyCap(1024),
			WithPoints([]*point.Point{largePt, largePt}),
			WithCategory(point.Logging),
			WithBodyCallback(func(w *writer, b *body) error {
				// TODO
				return nil
			}))
		defer putWriter(w)

		assert.NoError(t, w.buildPointsBody())

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)

		mfs, err := reg.Gather()
		require.NoError(t, err)
		t.Logf("get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_skipped_point_total`,
			`logging`)
		require.NotNil(t, m)
		assert.Equal(t, float64(2), m.GetCounter().GetValue())
	})

	t.Run(`skip-nothing`, func(t *T.T) {
		metricsReset()

		var kvsSmall point.KVs
		kvsSmall = kvsSmall.Add(`small`, strings.Repeat(`x`, 10), false, false)
		smallPt1 := point.NewPointV2(`small1`, kvsSmall)
		smallPt2 := point.NewPointV2(`small2`, kvsSmall)

		w := getWriter(WithHTTPEncoding(point.Protobuf),
			WithMaxBodyCap(1024),
			WithPoints([]*point.Point{smallPt1, smallPt2}),
			WithCategory(point.Logging),
			WithBodyCallback(func(w *writer, b *body) error {
				// TODO
				return nil
			}))
		defer putWriter(w)

		assert.NoError(t, w.buildPointsBody())

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)

		mfs, err := reg.Gather()
		require.NoError(t, err)
		t.Logf("get metrics:\n%s", metrics.MetricFamily2Text(mfs))

		m := metrics.GetMetricOnLabels(mfs,
			`datakit_io_dataway_skipped_point_total`,
			`logging`)
		require.NotNil(t, m)
		assert.Equal(t, float64(0), m.GetCounter().GetValue())
	})
}

func TestCrashBuildBody(t *T.T) {
	t.Run("test-build-random-points", func(t *T.T) {
		type Foo struct {
			Measurement string

			TS int64

			// tags
			T1Key string `fake:"{regex:[a-zA-Z0-9_]{64}}"`
			T2Key string `fake:"{regex:[a-zA-Z0-9_]{64}}"`
			T3Key string `fake:"{regex:[a-zA-Z0-9_]{64}}"`

			T1 string
			T2 string
			T3 string

			SKey, S string `fake:"{regex:[a-zA-Z0-9]{128}}"`

			I8Key  string `fake:"{regex:[a-zA-Z0-9_]{64}}"`
			I16Key string `fake:"{regex:[a-zA-Z0-9_]{64}}"`
			I32Key string `fake:"{regex:[a-zA-Z0-9_]{64}}"`
			I64Key string `fake:"{regex:[a-zA-Z0-9_]{64}}"`
			I64    int64
			I8     int8
			I16    int16
			I32    int32

			U8Key  string `fake:"{regex:[a-zA-Z0-9_]{64}}"`
			U16Key string `fake:"{regex:[a-zA-Z0-9_]{64}}"`
			U32Key string `fake:"{regex:[a-zA-Z0-9_]{64}}"`
			U64Key string `fake:"{regex:[a-zA-Z0-9_]{64}}"`

			U8  uint8
			U16 uint16
			U32 uint32
			U64 uint64

			BKey   string `fake:"{regex:[a-zA-Z0-9_]{64}}"`
			DKey   string `fake:"{regex:[a-zA-Z0-9_]{64}}"`
			F64Key string `fake:"{regex:[a-zA-Z0-9_]{64}}"`
			F32Key string `fake:"{regex:[a-zA-Z0-9_]{64}}"`
			B      bool
			D      []byte
			F64    float64
			F32    float32
		}

		var pts []*point.Point
		for i := 0; i < 1000; i++ {
			var f Foo
			assert.NoError(t, gofakeit.Struct(&f))
			var kvs point.KVs
			kvs = kvs.AddTag("T_"+f.T1Key, f.T1)
			kvs = kvs.AddTag("T_"+f.T2Key, f.T2)
			kvs = kvs.AddTag("T_"+f.T3Key, f.T3)

			kvs = kvs.AddV2("S_"+f.SKey, f.S, true)

			kvs = kvs.AddV2("I8_"+f.I8Key, f.I8, true)
			kvs = kvs.AddV2("I16_"+f.I16Key, f.I16, true)
			kvs = kvs.AddV2("I32_"+f.I32Key, f.I32, true)
			kvs = kvs.AddV2("I64_"+f.I64Key, f.I64, true)

			kvs = kvs.AddV2("U8_"+f.U8Key, f.U8, true)
			kvs = kvs.AddV2("U16_"+f.U16Key, f.U16, true)
			kvs = kvs.AddV2("U32_"+f.U32Key, f.U32, true)
			kvs = kvs.AddV2("U64_"+f.U64Key, f.U64, true)

			kvs = kvs.AddV2("F32_"+f.F32Key, f.F32, true)
			kvs = kvs.AddV2("F64_"+f.F64Key, f.F64, true)

			kvs = kvs.AddV2("B_"+f.BKey, f.B, true)
			kvs = kvs.AddV2("D_"+f.DKey, f.D, true)

			if f.TS < 0 {
				f.TS = 0
			}

			pt := point.NewPointV2(f.Measurement, kvs, point.WithTimestamp(f.TS))
			pts = append(pts, pt)
		}

		var wg sync.WaitGroup

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)

		wg.Add(32)
		for i := 0; i < 32; i++ {
			go func(idx int) {
				defer wg.Done()
				for i := 0; i < 100; i++ {
					w := getWriter()

					WithPoints(pts)(w)
					WithMaxBodyCap(10 * 1024 * 1024)(w)
					WithHTTPEncoding(point.LineProtocol)(w)

					assert.NoError(t, w.buildPointsBody())
					putWriter(w)
				}
			}(i)
		}

		t.Logf("wait...")
		wg.Wait()

		mfs, err := reg.Gather()
		require.NoError(t, err)
		t.Logf("get metrics:\n%s", metrics.MetricFamily2Text(mfs))
	})
}

func TestBuildBody(t *T.T) {
	cases := []struct {
		name string
		enc  point.Encoding
		pts  []*point.Point
	}{
		{
			name: "short",
			enc:  point.Protobuf,
			pts: func() []*point.Point {
				var pts []*point.Point
				for i := 0; i < 32; i++ {
					pts = append(pts, point.NewPointV2("some",
						point.NewKVs(map[string]any{
							"f1": 1,
							"f2": 2,
						}), point.WithTimestamp(1)))
				}
				return pts
			}(),
		},

		{
			name: "rand-large",
			enc:  point.Protobuf,
			pts: func() []*point.Point {
				r := point.NewRander()
				return r.Rand(1000)
			}(),
		},
	}

	batchSize := 8

	for _, tc := range cases {
		t.Run("build-"+tc.name, func(t *T.T) {
			w := getWriter()
			defer putWriter(w)

			WithPoints(tc.pts)(w)
			WithBatchSize(batchSize)
			WithHTTPEncoding(tc.enc)(w)

			assert.NoError(t, w.buildPointsBody())
		})
	}

	// test body == pts
	for _, tc := range cases {
		t.Run("point-checking-"+tc.name, func(t *T.T) {
			w := getWriter()
			defer putWriter(w)

			WithPoints(tc.pts)(w)
			WithBatchSize(batchSize)(w)
			WithHTTPEncoding(tc.enc)(w)

			var arr []*body
			WithBodyCallback(func(_ *writer, b *body) error {
				arr = append(arr, b)
				return nil
			})(w)

			w.buildPointsBody()

			var (
				extractPts []*point.Point
				dec        *point.Decoder
			)

			dec = point.GetDecoder(point.WithDecEncoding(tc.enc))
			defer point.PutDecoder(dec)

			t.Logf("decoded into %d parts", len(arr))

			for _, x := range arr {
				assert.True(t, x.npts() > 0)
				assert.True(t, x.rawLen() > 0)
				assert.Equal(t, tc.enc, x.enc())

				var (
					raw = x.buf()
					err error
				)

				if x.gzon == 1 {
					raw, err = uhttp.Unzip(x.buf())
					require.NoError(t, err)
				}

				pts, err := dec.Decode(raw)
				assert.NoErrorf(t, err, "decode %q failed", raw)
				extractPts = append(extractPts, pts...)
			}

			assert.Equal(t, len(tc.pts), len(extractPts))
			for i, got := range extractPts {
				assert.Equal(t, tc.pts[i].Pretty(), got.Pretty())
			}
		})
	}

	for _, tc := range cases {
		bodyByteBatch := 10 * 1024
		t.Run("body-size-decode-"+tc.name, func(t *T.T) {
			w := getWriter()
			defer putWriter(w)

			t.Logf("enc: %s", tc.enc)

			WithPoints(tc.pts)(w)
			WithMaxBodyCap(bodyByteBatch)(w)
			WithHTTPEncoding(tc.enc)(w)

			var arr []*body
			WithBodyCallback(func(_ *writer, b *body) error {
				arr = append(arr, b)
				return nil
			})(w)

			w.buildPointsBody()

			assert.True(t, len(arr) > 0)

			var (
				extractPts []*point.Point
				dec        *point.Decoder
			)

			dec = point.GetDecoder(point.WithDecEncoding(tc.enc))
			defer point.PutDecoder(dec)

			t.Logf("decoded into %d parts(byte size: %d)", len(arr), bodyByteBatch)

			for _, x := range arr {
				assert.True(t, x.npts() > 0)
				assert.True(t, x.rawLen() > 0)
				assert.Equal(t, tc.enc, x.enc())

				var (
					raw = x.buf()
					err error
				)
				if x.gzon == 1 {
					raw, err = uhttp.Unzip(x.buf())
					if err != nil {
						assert.NoError(t, err)
					}
				}

				pts, err := dec.Decode(raw)
				assert.NoErrorf(t, err, "decode %q failed", raw)
				extractPts = append(extractPts, pts...)
			}

			assert.Equal(t, len(tc.pts), len(extractPts))

			for i, got := range extractPts {
				eq, why := tc.pts[i].EqualWithReason(got)
				assert.Truef(t, eq, why)
				assert.Equal(t, tc.pts[i].Pretty(), got.Pretty())
			}
		})
	}
}

func BenchmarkBuildBody(b *T.B) {
	r := point.NewRander(point.WithRandText(3))

	cases := []struct {
		name  string
		pts   []*point.Point
		batch int
		gz    gzipFlag
		enc   point.Encoding
	}{
		{
			name:  "1k-pts-on-json-batch256",
			pts:   r.Rand(1024),
			batch: 256,
			enc:   point.JSON,
		},

		{
			name:  "1k-pts-on-json-batch1024",
			pts:   r.Rand(1024),
			batch: 1024,
			enc:   point.JSON,
		},

		{
			name:  "1k-pts-on-line-proto-batch256",
			pts:   r.Rand(1024),
			batch: 256,
			enc:   point.LineProtocol,
		},

		{
			name:  "1k-pts-on-line-protocol-batch1024",
			pts:   r.Rand(1024),
			batch: 1024,
			enc:   point.LineProtocol,
		},

		{
			name:  "1k-pts-on-protobuf-batch256",
			pts:   r.Rand(1024),
			batch: 256,
			enc:   point.Protobuf,
		},

		{
			name:  "1k-pts-on-protobuf-batch1024",
			pts:   r.Rand(1024),
			batch: 1024,
			enc:   point.Protobuf,
		},

		{
			name:  "gz-1k-pts-on-protobuf-batch1024",
			pts:   r.Rand(1024),
			batch: 1024,
			enc:   point.Protobuf,
			gz:    1,
		},

		{
			name:  "10k-pts-on-protobuf-batch4k",
			pts:   r.Rand(10240),
			batch: 4096,
			enc:   point.Protobuf,
		},

		{
			name:  "10k-pts-on-protobuf-batch10k",
			pts:   r.Rand(10240),
			batch: 10240,
			enc:   point.Protobuf,
		},
	}

	for _, bc := range cases {
		b.ResetTimer()
		b.Run(bc.name, func(b *T.B) {
			w := getWriter(WithBodyCallback(func(w *writer, body *body) error {
				putBody(body) // release body
				return nil
			}))
			defer putWriter(w)

			WithBatchSize(bc.batch)(w)
			WithPoints(bc.pts)(w)
			WithHTTPEncoding(bc.enc)(w)
			WithGzip(bc.gz)(w)

			for i := 0; i < b.N; i++ {
				w.buildPointsBody()
			}
		})
	}
}

func TestBodyCacheData(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		cb := func(w *writer, b *body) error {
			t.Logf("send-buf: %p/%d", b.sendBuf, len(b.sendBuf))
			t.Logf("buf(): %p/%d", b.buf(), len(b.buf()))
			t.Logf("payload: %p/%d", b.CacheData.Payload, len(b.CacheData.Payload))

			assert.Equal(t, len(b.sendBuf), cap(b.sendBuf))       // b.sendBuf should always equal to it's cap
			assert.Equal(t, len(b.marshalBuf), cap(b.marshalBuf)) // b.sendBuf should always equal to it's cap

			// b.buf() should point to b.sendBuf
			assert.Equal(t, b.sendBuf[:len(b.buf())], b.buf())

			return nil
		}

		r := point.NewRander()
		pts := r.Rand(1000)

		w := getWriter(WithBodyCallback(cb))
		defer putWriter(w)
		WithPoints(pts)(w)
		WithMaxBodyCap(10 * 1024 * 1024)(w)
		WithHTTPEncoding(point.LineProtocol)(w)

		w.buildPointsBody()
	})

	t.Run(`body-dump-and-load`, func(t *T.T) {
		cb := func(w *writer, b *body) error {
			t.Logf("send-buf: %p/%d", b.sendBuf, len(b.sendBuf))
			t.Logf("buf(): %p/%d", b.buf(), len(b.buf()))
			t.Logf("payload: %p/%d", b.CacheData.Payload, len(b.CacheData.Payload))

			pb, err := b.dump()
			assert.NoError(t, err)
			t.Logf("marshal-buf: %p?/%d", b.marshalBuf, len(b.marshalBuf))
			t.Logf("pb: %p?/%d", pb, len(pb))

			assert.Equal(t, len(b.sendBuf), cap(b.sendBuf))       // b.sendBuf should always equal to it's cap
			assert.Equal(t, len(b.marshalBuf), cap(b.marshalBuf)) // b.sendBuf should always equal to it's cap

			// b.buf() should point to b.sendBuf
			assert.Equal(t, b.marshalBuf[:len(pb)], pb)

			newBody := getNewBufferBody(withNewBuffer(defaultBatchSize))
			defer putBody(newBody)

			assert.NoError(t, newBody.loadCache(pb)) // newBody load pb data
			assert.Equal(t, newBody.CacheData, b.CacheData)
			assert.Equal(t, b.buf(), newBody.buf())

			t.Logf("body: %s", newBody.pretty())

			assert.Len(t, newBody.headers(), 2)
			for _, h := range newBody.headers() {
				switch h.Key {
				case "header-1":
					assert.Equal(t, `value-1`, h.Value)
				case "header-2":
					assert.Equal(t, `value-2`, h.Value)
				default:
					assert.Truef(t, false, "should not been here")
				}
			}

			assert.Equal(t, "http://some.dynamic.url?token=tkn_xyz", b.url())

			t.Logf("buf: %q", newBody.buf()[:10])

			putBody(b)

			return nil
		}

		pts := point.RandPoints(100)
		w := getWriter(WithBodyCallback(cb),
			WithHTTPHeader("header-1", "value-1"),
			WithHTTPHeader("header-2", "value-2"),
			WithDynamicURL("http://some.dynamic.url?token=tkn_xyz"),
			WithPoints(pts),
			WithHTTPEncoding(point.LineProtocol),
		)
		defer putWriter(w)

		w.buildPointsBody()
	})
}

func TestPBMarshalSize(t *T.T) {
	t.Run(`basic-1mb-body`, func(t *T.T) {
		sendBuf := make([]byte, 1<<20)
		marshalBuf := make([]byte, 1<<20+100*(1<<10))
		b := getReuseBufferBody(withReusableBuffer(sendBuf, marshalBuf))

		b.CacheData.Category = int32(point.Metric)
		b.CacheData.PayloadType = int32(point.Protobuf)
		b.CacheData.Payload = sendBuf // we really get 1MB body to send
		b.CacheData.Pts = 10000
		b.CacheData.RawLen = (1 << 20)
		b.Headers = append(b.Headers, &HTTPHeader{Key: HeaderXGlobalTags, Value: "looooooooooooooooooooooong value"})
		b.DynURL = "https://openway.guance.com/v1/write/logging?token=tkn_11111111111111111111111"

		pbbuf, err := b.dump()
		assert.NoError(t, err)

		t.Logf("#pbbuf: %d, raised: %.8f", len(pbbuf), float64(len(pbbuf)-len(b.buf()))/float64(len(b.buf())))
	})

	t.Run(`basic-4mb-body`, func(t *T.T) {
		size := 4 * (1 << 20)
		sendBuf := make([]byte, size)
		marshalBuf := make([]byte, size+int(float64(size)*.1))
		b := getReuseBufferBody(withReusableBuffer(sendBuf, marshalBuf))

		b.CacheData.Category = int32(point.Metric)
		b.CacheData.PayloadType = int32(point.Protobuf)
		b.CacheData.Payload = sendBuf // we really get 1MB body to send
		b.CacheData.Pts = 10000
		b.CacheData.RawLen = int32(size)
		b.Headers = append(b.Headers, &HTTPHeader{Key: HeaderXGlobalTags, Value: "looooooooooooooooooooooong value"})
		b.DynURL = "https://openway.guance.com/v1/write/logging?token=tkn_11111111111111111111111"

		pbbuf, err := b.dump()
		assert.NoError(t, err)

		t.Logf("#pbbuf: %d, raised: %.8f", len(pbbuf), float64(len(pbbuf)-len(b.buf()))/float64(len(b.buf())))
	})

	t.Run(`basic-10mb-body`, func(t *T.T) {
		size := 10 * (1 << 20)
		sendBuf := make([]byte, size)
		marshalBuf := make([]byte, size+int(float64(size)*.1))
		b := getReuseBufferBody(withReusableBuffer(sendBuf, marshalBuf))

		b.CacheData.Category = int32(point.Metric)
		b.CacheData.PayloadType = int32(point.Protobuf)
		b.CacheData.Payload = sendBuf // we really get 1MB body to send
		b.CacheData.Pts = 10000
		b.CacheData.RawLen = int32(size)
		b.Headers = append(b.Headers, &HTTPHeader{Key: HeaderXGlobalTags, Value: "looooooooooooooooooooooong value"})
		b.DynURL = "https://openway.guance.com/v1/write/logging?token=tkn_11111111111111111111111"

		pbbuf, err := b.dump()
		assert.NoError(t, err)

		t.Logf("#pbbuf: %d, raised: %.8f", len(pbbuf), float64(len(pbbuf)-len(b.buf()))/float64(len(b.buf())))
	})

	t.Run(`error`, func(t *T.T) {
		size := 1 << 20
		sendBuf := make([]byte, size)
		marshalBuf := make([]byte, size)
		b := getReuseBufferBody(withReusableBuffer(sendBuf, marshalBuf))

		b.CacheData.Category = int32(point.Metric)
		b.CacheData.PayloadType = int32(point.Protobuf)
		b.CacheData.Payload = sendBuf // we really get 1MB body to send
		b.CacheData.Pts = 10000
		b.CacheData.RawLen = int32(size)
		b.Headers = append(b.Headers, &HTTPHeader{Key: HeaderXGlobalTags, Value: "looooooooooooooooooooooong value"})
		b.DynURL = "https://openway.guance.com/v1/write/logging?token=tkn_11111111111111111111111"

		_, err := b.dump()
		assert.Error(t, err)
		t.Logf("[expected] dump error: %s", err.Error())
	})
}

func BenchmarkBodyDumpAndLoad(b *T.B) {
	b.Run(`get-CacheData-size`, func(b *T.B) {
		size := 1 << 20
		sendBuf := make([]byte, size)
		marshalBuf := make([]byte, size)
		_b := getReuseBufferBody(withReusableBuffer(sendBuf, marshalBuf))

		_b.CacheData.Category = int32(point.Metric)
		_b.CacheData.PayloadType = int32(point.Protobuf)
		_b.CacheData.Payload = sendBuf // we really get 1MB body to send
		_b.CacheData.Pts = 10000
		_b.CacheData.RawLen = int32(size)
		_b.Headers = append(_b.Headers, &HTTPHeader{Key: HeaderXGlobalTags, Value: "looooooooooooooooooooooong value"})
		_b.DynURL = "https://openway.guance.com/v1/write/logging?token=tkn_11111111111111111111111"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_b.Size()
		}
	})

	b.Run(`dump-and-load`, func(b *T.B) {
		cb := func(w *writer, x *body) error {
			defer putBody(x)

			func() { // check if any error
				pb, err := x.dump()
				assert.NoError(b, err)

				newBody := getNewBufferBody(withNewBuffer(defaultBatchSize))
				assert.NoError(b, newBody.loadCache(pb))
				putBody(newBody)
			}()

			b.ResetTimer()
			// reuse @x during benchmark
			for i := 0; i < b.N; i++ {
				pb, _ := x.dump()

				newBody := getNewBufferBody(withNewBuffer(defaultBatchSize))
				newBody.loadCache(pb)
				putBody(newBody)
			}
			return nil
		}

		pts := point.RandPoints(100)
		w := getWriter(WithBodyCallback(cb),
			WithHTTPHeader("header-1", "value-1"),
			WithHTTPHeader("header-2", "value-2"),
			WithDynamicURL("http://some.dynamic.url?token=tkn_xyz"),
			WithPoints(pts),
			WithBodyCallback(cb),
		)

		defer putWriter(w)

		w.buildPointsBody()
	})
}

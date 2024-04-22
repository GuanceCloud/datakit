// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
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
					WithBatchBytesSize(10 * 1024 * 1024)(w)
					WithHTTPEncoding(point.LineProtocol)(w)

					assert.NoError(t, w.buildPointsBody(nil))
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
						})))
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

			assert.NoError(t, w.buildPointsBody(nil))

			t.Logf("get %d bodies", w.parts)
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
			cb := func(_ *writer, b *body) error {
				arr = append(arr, b)
				return nil
			}

			w.buildPointsBody(cb)

			var (
				extractPts []*point.Point
				dec        *point.Decoder
			)

			dec = point.GetDecoder(point.WithDecEncoding(tc.enc))
			defer point.PutDecoder(dec)

			t.Logf("decoded into %d parts", len(arr))

			for _, x := range arr {
				assert.True(t, x.npts > 0)
				assert.True(t, x.rawLen > 0)
				assert.Equal(t, tc.enc, x.payloadEnc)

				var (
					raw = x.buf
					err error
				)

				if x.gzon {
					raw, err = uhttp.Unzip(x.buf)
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
			WithBatchBytesSize(bodyByteBatch)(w)
			WithHTTPEncoding(tc.enc)(w)

			var arr []*body
			cb := func(_ *writer, b *body) error {
				arr = append(arr, b)
				return nil
			}

			w.buildPointsBody(cb)

			assert.True(t, len(arr) > 0)

			var (
				extractPts []*point.Point
				dec        *point.Decoder
			)

			dec = point.GetDecoder(point.WithDecEncoding(tc.enc))
			defer point.PutDecoder(dec)

			t.Logf("decoded into %d parts(byte size: %d)", len(arr), bodyByteBatch)

			for _, x := range arr {
				assert.True(t, x.npts > 0)
				assert.True(t, x.rawLen > 0)
				assert.Equal(t, tc.enc, x.payloadEnc)

				var (
					raw = x.buf
					err error
				)
				if x.gzon {
					raw, err = uhttp.Unzip(x.buf)
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
		b.Run(bc.name, func(b *T.B) {
			w := getWriter()
			defer putWriter(w)

			WithBatchSize(bc.batch)(w)
			WithPoints(bc.pts)(w)
			WithHTTPEncoding(bc.enc)(w)

			for i := 0; i < b.N; i++ {
				w.buildPointsBody(nil)
			}
		})
	}
}

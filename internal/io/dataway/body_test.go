// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	T "testing"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

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

			arr, err := w.buildPointsBody()
			assert.NoError(t, err)

			t.Logf("get %d bodies", len(arr))
			for _, b := range arr {
				t.Logf("body: %s, compress ratio: %.3f", b, float64(len(b.buf))/float64(b.rawLen))
			}
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

			arr, err := w.buildPointsBody()
			assert.NoError(t, err)

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
				assert.Equal(t, tc.enc, x.payload)

				if x.gzon {
					raw, err := uhttp.Unzip(x.buf)
					if err != nil {
						assert.NoError(t, err)
					}

					pts, err := dec.Decode(raw)
					assert.NoErrorf(t, err, "decode %q failed", raw)
					extractPts = append(extractPts, pts...)
				}
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

			arr, err := w.buildPointsBody()
			assert.NoError(t, err)

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
				assert.Equal(t, tc.enc, x.payload)

				if x.gzon {
					raw, err := uhttp.Unzip(x.buf)
					if err != nil {
						assert.NoError(t, err)
					}

					pts, err := dec.Decode(raw)
					assert.NoErrorf(t, err, "decode %q failed", raw)
					extractPts = append(extractPts, pts...)
				}
			}

			assert.Equal(t, len(tc.pts), len(extractPts))

			for i, got := range extractPts {
				assert.Equal(t, tc.pts[i].Pretty(), got.Pretty())
			}
		})
	}
}

func BenchmarkBuildBody(b *T.B) {
	cases := []struct {
		name  string
		pts   []*point.Point
		batch int
		enc   point.Encoding
	}{
		{
			name:  "1k-pts-on-json-batch256",
			pts:   point.RandPoints(1024),
			batch: 256,
			enc:   point.JSON,
		},

		{
			name:  "1k-pts-on-json-batch1024",
			pts:   point.RandPoints(1024),
			batch: 1024,
			enc:   point.JSON,
		},

		{
			name:  "1k-pts-on-line-proto-batch256",
			pts:   point.RandPoints(1024),
			batch: 256,
			enc:   point.LineProtocol,
		},

		{
			name:  "1k-pts-on-line-protocol-batch1024",
			pts:   point.RandPoints(1024),
			batch: 1024,
			enc:   point.LineProtocol,
		},

		{
			name:  "1k-pts-on-protobuf-batch256",
			pts:   point.RandPoints(1024),
			batch: 256,
			enc:   point.Protobuf,
		},

		{
			name:  "1k-pts-on-protobuf-batch1024",
			pts:   point.RandPoints(1024),
			batch: 1024,
			enc:   point.Protobuf,
		},

		{
			name:  "10k-pts-on-protobuf-batch4k",
			pts:   point.RandPoints(10240),
			batch: 4096,
			enc:   point.Protobuf,
		},

		{
			name:  "10k-pts-on-protobuf-batch10k",
			pts:   point.RandPoints(10240),
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
				_, err := w.buildPointsBody()
				assert.NoError(b, err)
			}
		})
	}
}

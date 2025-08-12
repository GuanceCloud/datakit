// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	sync "sync"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlush(t *T.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		t.Logf("body md5: %s", fmt.Sprintf("%x", md5.Sum(body)))

		if r.Header.Get("Content-Encoding") == "gzip" {
			t.Logf("is gzip body")
			unz, err := uhttp.Unzip(body)
			require.NoError(t, err)
			body = unz
		} else {
			t.Logf("not gzip body")
		}

		encoding := point.HTTPContentType(r.Header.Get("Content-Type"))
		var dec *point.Decoder

		switch encoding { // nolint: exhaustive
		case point.Protobuf:
			dec = point.GetDecoder(point.WithDecEncoding(point.Protobuf))
			defer point.PutDecoder(dec)

		case point.LineProtocol:
			dec = point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
			defer point.PutDecoder(dec)

		default: // not implemented
			t.Logf("[ERROR] unknown encoding %s", encoding)
			return
		}

		if dec != nil {
			pts, err := dec.Decode(body)
			assert.NoError(t, err)

			nwarns := 0
			for _, pt := range pts {
				if len(pt.Warns()) > 0 {
					nwarns++
				}

				t.Logf(pt.LineProto())
			}

			t.Logf("decode %d points, %d with warnnings", len(pts), nwarns)
		}

		w.WriteHeader(200)
	}))

	defer ts.Close()
	time.Sleep(time.Second)

	t.Run("conc", func(t *T.T) {
		dw := NewDefaultDataway()

		walDir, err := os.MkdirTemp("", "dw-wal")
		require.NoError(t, err)
		dw.WAL.Path = walDir

		defer os.RemoveAll(walDir) // clean up

		dw.ContentEncoding = "v2"

		assert.NoError(t, dw.Init(WithURLs(fmt.Sprintf("%s?token=tkn_xxxxxxxxxxxxxxxxxxx", ts.URL))))

		t.Logf("dataway: %s", dw)

		cat := point.Logging

		// setup WAL queue
		dw.setupWAL()

		rnd := point.NewRander(point.WithFixedTags(true), point.WithRandText(3))

		cases := []struct {
			name string
			pts  []*point.Point
		}{
			// {
			// 	`1pt`,
			// 	rnd.Rand(1),
			// },
			{
				`100pt`,
				rnd.Rand(100),
			},
		}

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)
		for _, tc := range cases {
			t.Run(tc.name, func(t *T.T) {
				wg := sync.WaitGroup{}
				nworker := 1
				njob := 10

				wg.Add(nworker)
				for i := 0; i < nworker; i++ {
					go func() {
						defer wg.Done()
						for x := 0; x < njob; x++ {
							assert.NoError(t, dw.Write(
								WithPoints(tc.pts),
								WithCategory(cat),
								WithMaxBodyCap(10*(1<<20)), // 1MB buffer
							))
						}
					}()
				}

				time.Sleep(time.Second) // wait workers ok

				f := dw.newFlusher(cat)

				for {
					b, err := f.wal.Get(withReusableBuffer(f.sendBuf, f.marshalBuf))
					assert.NoError(t, err)
					if b == nil {
						break
					}

					raw := b.buf()

					dec := point.GetDecoder(point.WithDecEncoding(point.Protobuf))
					defer point.PutDecoder(dec)

					pts, err := dec.Decode(raw)
					require.NoError(t, err)
					require.Len(t, pts, len(tc.pts))

					// each point equal
					for idx := range pts {
						require.Equal(t, pts[idx].LineProto(), tc.pts[idx].LineProto())
					}

					assert.NoError(t, f.do(b, WithCategory(cat)))
				}

				wg.Wait()

				mfs, err := reg.Gather()
				require.NoError(t, err)
				t.Logf("get metrics:\n%s", metrics.MetricFamily2Text(mfs))

				t.Cleanup(func() {
					metricsReset()
				})
			})
		}
	})

	t.Run(`basic`, func(t *T.T) {
		t.Skip()
		dw := NewDefaultDataway()
		dw.WAL.Path = os.TempDir()

		assert.NoError(t, dw.Init(WithURLs(fmt.Sprintf("%s?token=tkn_xxxxxxxxxxxxxxxxxxx", ts.URL))))

		cat := point.Metric

		// setup WAL queue
		dw.setupWAL()

		var kvs point.KVs
		kvs = kvs.Set("f1", 123).
			Set("f2", "abc")
		pt := point.NewPoint(t.Name(), kvs, point.WithTimestamp(123))

		t.Logf("pt: %s", pt.LineProto())

		require.NoError(t,
			dw.Write(WithPoints([]*point.Point{pt}), WithCategory(cat), WithMaxBodyCap(1<<20)))

		q := dw.walq[cat]
		b, err := q.Get()
		require.NoError(t, err)

		defer putBody(b)

		t.Logf("body: %s", b)

		raw, err := uhttp.Unzip(b.buf())
		require.NoError(t, err, "body chksum: %x", md5.Sum(b.buf()))
		t.Logf("raw: %q", raw)

		dec := point.GetDecoder(point.WithDecEncoding(point.Protobuf))
		defer point.PutDecoder(dec)

		pts, err := dec.Decode(raw)
		require.NoError(t, err)
		require.Len(t, pts, 1)
		require.Equal(t, pts[0].LineProto(), pt.LineProto())

		f := dw.newFlusher(cat)
		assert.NoError(t, f.do(b, WithCategory(cat)))
	})
}

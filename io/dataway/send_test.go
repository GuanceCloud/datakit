// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"testing"

	lp "github.com/GuanceCloud/cliutils/lineproto"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

func TestBuildBody(t *testing.T) {
	cases := []struct {
		name string
		pts  []*point.Point
	}{
		{
			name: "short",
			pts:  point.RandPoints(1000),
		},
	}

	maxKodoPack = uint64(8 * 1024)
	minGZSize = 1024

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dw := &DataWayDefault{}
			bodies, err := dw.buildBody(tc.pts)
			if err != nil {
				t.Error(err)
			}

			t.Logf("get %d bodies", len(bodies))
			for _, b := range bodies {
				t.Logf("body: %s, compress ratio: %.3f", b, float64(len(b.buf))/float64(b.rawBufBytes))
			}
		})
	}

	// test body === pts
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dw := &DataWayDefault{}
			bodies, err := dw.buildBody(tc.pts)
			if err != nil {
				t.Error(err)
			}

			t.Logf("get %d bodies", len(bodies))

			for _, b := range bodies {
				t.Logf("body: %s", b)

				if b.gzon {
					x, err := uhttp.Unzip(b.buf)
					if err != nil {
						assert.NoError(t, err)
					}
					b.buf = x
				}

				if pts, err := lp.ParsePoints(b.buf, nil); err != nil {
					t.Error(err)
				} else {
					begin := b.idxRange[0]
					end := b.idxRange[1]

					assert.Equal(t, len(pts), end-begin)

					for i := 0; i < len(pts); i++ {
						assert.Equal(t, tc.pts[begin+i].String(), pts[i].String(), "index %d <> %d", begin+i, i)
					}
				}
			}
		})
	}
}

func BenchmarkBuildBody(b *testing.B) {
	cases := []struct {
		name string
		pts  []*point.Point
	}{
		{
			name: "short",
			pts:  point.RandPoints(1000),
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(t *testing.B) {
			for i := 0; i < b.N; i++ {
				dw := &DataWayDefault{}
				_, err := dw.buildBody(tc.pts)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

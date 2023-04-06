// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	T "testing"

	lp "github.com/GuanceCloud/cliutils/lineproto"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

func TestBuildBody(t *T.T) {
	cases := []struct {
		name string
		pts  []*dkpt.Point
	}{
		{
			name: "short",
			pts:  dkpt.RandPoints(1000),
		},
	}

	maxBody := 8 * 1024

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			bodies, err := buildBody(tc.pts, maxBody)
			if err != nil {
				t.Error(err)
			}

			t.Logf("get %d bodies", len(bodies))
			for _, b := range bodies {
				t.Logf("body: %s, compress ratio: %.3f", b, float64(len(b.buf))/float64(b.rawLen))
			}
		})
	}

	// test body === pts
	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			bodies, err := buildBody(tc.pts, maxBody)
			if err != nil {
				t.Error(err)
			}

			var totalBodies []byte

			for _, b := range bodies {
				if b.gzon {
					x, err := uhttp.Unzip(b.buf)
					if err != nil {
						assert.NoError(t, err)
					}
					totalBodies = append(totalBodies, append(x, '\n')...)
				}
			}

			pts, err := lp.ParsePoints(totalBodies, nil)
			require.NoError(t, err)
			for i, got := range pts {
				require.Equal(t, got.String(), tc.pts[i].String())
			}
		})
	}
}

func BenchmarkBuildBody(b *T.B) {
	cases := []struct {
		name string
		pts  []*dkpt.Point
	}{
		{
			name: "short",
			pts:  dkpt.RandPoints(1000),
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(t *T.B) {
			for i := 0; i < b.N; i++ {
				_, err := buildBody(tc.pts, MaxKodoBody)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

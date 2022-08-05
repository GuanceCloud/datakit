package dataway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
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
			bodies, err := buildBody(tc.pts, true)
			if err != nil {
				t.Error(err)
			}

			t.Logf("get %d bodies", len(bodies))
			for _, b := range bodies {
				t.Logf("body: %s", b)
			}
		})
	}

	// test body === pts
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bodies, err := buildBody(tc.pts, false)
			if err != nil {
				t.Error(err)
			}

			t.Logf("get %d bodies", len(bodies))
			for _, b := range bodies {
				t.Logf("body: %s", b)

				if pts, err := lp.ParsePoints(b.buf, nil); err != nil {
					t.Error(err)
				} else {
					begin := b.idxRange[0]
					end := b.idxRange[1]

					assert.Equal(t, len(pts), end-begin)

					for i := 0; i < len(b.idxRange); i++ {
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
				_, err := buildBody(tc.pts, true)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

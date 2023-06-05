// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plmap"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
)

func TestAgg(t *testing.T) {
	cases := []struct {
		name, pl string
		in       []string
		fail     bool
		out      map[string]map[string]any
	}{
		{
			name: "",
			pl: `
				set_tag("t0", "t1111")
				set_tag("t1", "t1111")
				set_tag("t2", "t2__")
				f0 = _
				cast(f0, "int")
				agg_create("abc", on_interval="5s",on_count=0, const_tags={"a":"b"})
				agg_create("def")

				agg_metric("abc", "f1", "avg", ["t2","t0"], "f0")
				agg_metric("def", "f1", "avg", ["t1", "t2"], "f0")
				`,
			in: []string{`1`, `2`},
			out: map[string]map[string]any{
				"abc": {
					"f1": float64(1.5),
					"t0": "t1111",
				},
				"def": {
					"f1": float64(1.5),
					"t1": "t1111",
				},
			},
		},
	}

	for idx, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.pl)
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}

			ptsLi := map[string][]*point.Point{}
			fn := func(n string, d any) error {
				ptsLi[n] = append(ptsLi[n], d.([]*point.Point)...)
				return nil
			}

			buks := plmap.NewAggBuks(fn)
			for _, tcIn := range tc.in {
				pt := ptinput.GetPoint()
				pt.AggBuckets = buks
				ptinput.InitPt(pt, "test", nil, map[string]any{"message": tcIn}, time.Now())
				errR := runScript(runner, pt)
				if errR != nil {
					ptinput.PutPoint(pt)
					t.Fatal(*errR)
				}
				ptinput.PutPoint(pt)
			}

			buks.StopAllBukScanner()

			for bukName, kv := range tc.out {
				for k, v := range kv {
					if len(ptsLi[bukName]) == 0 {
						t.Fatal("no data")
					}
					pt := ptsLi[bukName][0]
					f, _ := pt.Fields()
					tags := pt.Tags()
					if _, ok := f[k]; ok {
						assert.Equal(t, v, f[k])
					} else if _, ok := tags[k]; ok {
						assert.Equal(t, v, tags[k])
					} else {
						t.Error(k)
					}
				}
			}
		})
	}
}

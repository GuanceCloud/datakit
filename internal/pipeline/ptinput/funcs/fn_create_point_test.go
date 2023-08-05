// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
)

func TestCreatePoint(t *testing.T) {
	cases := []struct {
		name, in string
		allPl    map[string]string
		outkey   string
		expect   []ptinput.PlInputPt
		fail     bool
	}{
		{
			name: "default",
			in:   `{"a": "1", "b": 2, "c": {"d" : "x1"}}`,
			allPl: map[string]string{
				"main.p": `
			d = load_json(_)
			r = {}
			for x in d["c"] {
				r[x] = d["c"][x]
			}
			r["b"] = d["b"]
			create_point("n1", {"a": d["a"]}, r)
			`,
			},
			outkey: "abc",
			expect: []ptinput.PlInputPt{
				ptinput.NewPlPoint(point.Metric, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
			},
		},
		{
			name: "default-use",
			in:   `{"a": "1", "b": 2, "c": {"d" : "x1"}}`,
			allPl: map[string]string{
				"main.p": `
			d = load_json(_)
			r = {}
			for x in d["c"] {
				r[x] = d["c"][x]
			}
			r["b"] = d["b"]
			create_point("n1", {"a": d["a"]}, r, after_use="abc.p")
			`,
				"abc.p": `add_key("aa", 1)`,
			},
			outkey: "abc",
			expect: []ptinput.PlInputPt{
				ptinput.NewPlPoint(point.Metric, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d":  "x1",
					"b":  float64(2.0),
					"aa": int64(1),
				}, time.Time{}),
			},
		},
		{
			name: "default-use",
			in:   `{"a": "1", "b": 2, "c": {"d" : "x1"}}`,
			allPl: map[string]string{
				"main.p": `
			d = load_json(_)
			r = {}
			for x in d["c"] {
				r[x] = d["c"][x]
			}
			r["b"] = d["b"]
			create_point("n1", {"a": d["a"]}, r, ts = 1, after_use="abc.p")
			`,
				"abc.p": `add_key("aa", 1)`,
			},
			outkey: "abc",
			expect: []ptinput.PlInputPt{
				ptinput.NewPlPoint(point.Metric, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d":  "x1",
					"b":  float64(2.0),
					"aa": int64(1),
				}, time.Unix(0, 1)),
			},
		},
		{
			name: "failed-use",
			in:   `{"a": "1", "b": 2, "c": {"d" : "x1"}}`,
			allPl: map[string]string{
				"main.p": `
			d = load_json(_)
			r = {}
			for x in d["c"] {
				r[x] = d["c"][x]
			}
			r["b"] = d["b"]
			create_point("n1", {"a": d["a"]}, r, after_use="absc.p")
			`,
				"abc.p": `add_key("aa", 1)`,
			},
			outkey: "abc",
			fail:   true,
			expect: []ptinput.PlInputPt{
				ptinput.NewPlPoint(point.Metric, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d":  "x1",
					"b":  float64(2.0),
					"aa": int64(1),
				}, time.Time{}),
			},
		},
		{
			name: "L",
			in:   `{"a": "1", "b": 2, "c": {"d" : "x1"}}`,
			allPl: map[string]string{
				"main.p": `
			d = load_json(_)
			r = {}
			for x in d["c"] {
				r[x] = d["c"][x]
			}
			r["b"] = d["b"]
			create_point("n1", {"a": d["a"]}, r)
			create_point("n1", {"a": d["a"]}, r, category="L")
			create_point("n1", {"a": d["a"]}, r, category="M")
			create_point("n1", {"a": d["a"]}, r, category="T")
			create_point("n1", {"a": d["a"]}, r, category="R")
			create_point("n1", {"a": d["a"]}, r, category="N")
			create_point("n1", {"a": d["a"]}, r, category="O")
			create_point("n1", {"a": d["a"]}, r, category="CO")
			create_point("n1", {"a": d["a"]}, r, category="S")
			`,
			},
			outkey: "abc",
			expect: []ptinput.PlInputPt{
				ptinput.NewPlPoint(point.Metric, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.Logging, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.Metric, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.Tracing, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.RUM, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.Network, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.Object, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.CustomObject, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.Security, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
			},
		},
		{
			name: "L",
			in:   `{"a": "1", "b": 2, "c": {"d" : "x1"}}`,
			allPl: map[string]string{
				"main.p": `
				d = load_json(_)
				r = {}
				for x in d["c"] {
					r[x] = d["c"][x]
				}
				r["b"] = d["b"]
				create_point("n1", {"a": d["a"]}, r)
				create_point("n1", {"a": d["a"]}, r, category="logging")
				create_point("n1", {"a": d["a"]}, r, category="metric")
				create_point("n1", {"a": d["a"]}, r, category="tracing")
				create_point("n1", {"a": d["a"]}, r, category="rum")
				create_point("n1", {"a": d["a"]}, r, category="network")
				create_point("n1", {"a": d["a"]}, r, category="object")
				create_point("n1", {"a": d["a"]}, r, category="custom_object")
				create_point("n1", {"a": d["a"]}, r, category="security")
				`,
			},
			outkey: "abc",
			expect: []ptinput.PlInputPt{
				ptinput.NewPlPoint(point.Metric, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.Logging, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.Metric, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.Tracing, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.RUM, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.Network, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.Object, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.CustomObject, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
				ptinput.NewPlPoint(point.Security, "n1", map[string]string{
					"a": "1",
				}, map[string]any{
					"d": "x1",
					"b": float64(2.0),
				}, time.Time{}),
			},
		},
	}

	for idx, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pls, errs := NewTestingRunner2(tc.allPl)
			if len(errs) != 0 {
				if tc.fail {
					t.Logf("[%d]expect error: %v", idx, errs)
				} else {
					t.Errorf("[%d] failed: %v", idx, errs)
				}
				return
			}
			runner, ok := pls["main.p"]
			if !ok {
				t.Fatal(ok)
			}

			pt := ptinput.NewPlPoint(
				point.Logging, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			t.Log(pt.Fields())
			if errR != nil {
				if tc.fail {
					return
				}
				t.Fatal(errR)
			}

			pts := pt.GetSubPoint()

			if len(pts) != len(tc.expect) {
				t.Fatal("len(pt)!= len(tc.expect)")
			}
			for i, pt := range pts {
				expect := tc.expect[i]
				assert.Equal(t, expect.Tags(), pt.Tags())
				assert.Equal(t, expect.Fields(), pt.Fields())
				assert.Equal(t, expect.GetPtName(), pt.GetPtName())
				assert.Equal(t, expect.Category(), pt.Category())
				if !expect.PtTime().IsZero() {
					assert.Equal(t, expect.PtTime(), pt.PtTime())
				}
			}
		})
	}
}

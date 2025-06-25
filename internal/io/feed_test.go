// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

type plcase struct {
	name string

	pts []*point.Point

	epts []*point.Point

	option *Option
}

func newpt(name string, tags map[string]string,
	fields map[string]any, tn time.Time,
) *point.Point {
	kvs := append(point.NewTags(tags), point.NewKVs(fields)...)
	return point.NewPointV2(
		name, kvs, append(point.DefaultLoggingOptions(),
			point.WithTime(tn))...,
	)
}

func TestRunpl(t *T.T) {
	pipeline.InitPipeline(nil, nil, nil, "")

	tn := time.Now()
	cases := []plcase{
		{
			name: "a_with_opt",
			pts: []*point.Point{
				newpt("a_with_opt", map[string]string{
					"tag_1": "value_1",
				}, map[string]any{
					"field_1": "value_2",
				}, tn),
			},

			epts: []*point.Point{
				newpt("a_with_opt", map[string]string{
					"tag_1": "value_1",
				}, map[string]any{
					"field_1": "value_2",
					"a":       int64(1),
				}, tn),
			},

			option: &Option{
				PlOption: &lang.LogOption{
					ScriptMap: map[string]string{
						"a_with_opt": "a.p",
					},
				},
			},
		},

		{
			name: "a_status",
			pts: []*point.Point{
				newpt("a", map[string]string{
					"tag_1": "value_1",
				}, map[string]any{
					"field_1": "value_2",
				}, tn),
			},

			epts: nil, // filtered

			option: &Option{
				PlOption: &lang.LogOption{
					ScriptMap: map[string]string{
						"a_with_opt": "a.p",
					},
					IgnoreStatus: []string{"info"},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(``, func(t *T.T) {
			m, ok := plval.GetManager()
			if !ok {
				t.Error("!ok")
			}
			m.LoadScripts(constants.NSRemote,
				map[point.Category]map[string]string{
					point.Logging: {"a.p": "add_key('a', 1)"},
				}, nil)
			if _, ok := m.QueryScript(point.Logging, "a.p"); !ok {
				t.Error("!ok")
			}

			fo := GetFeedData()
			fo.input = "a"
			fo.cat = point.Logging
			fo.pts = c.pts
			fo.plOption = c.option.PlOption
			// epts, _, _, err := beforeFeed("a", point.Logging, c.pts, c.option)
			epts, _, _, err := beforeFeed(fo)
			if err != nil {
				t.Error(err)
			}

			assert.True(t, len(c.epts) == len(epts))

			for i, pt := range epts {
				t.Log(pt.Pretty())
				t.Log(c.epts[i].Pretty())
				pt.Equal(c.epts[i])
			}
		})
	}
}

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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

type mockFeederOutputer struct {
	feeds []*feedData
}

func (m *mockFeederOutputer) Write(fd *feedData) error {
	m.feeds = append(m.feeds, fd)
	return nil
}

func (*mockFeederOutputer) WriteLastError(string, ...metrics.LastErrorOption) {}

func (*mockFeederOutputer) Reader(point.Category) <-chan *feedData { return nil }

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
	return point.NewPoint(
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

			dkio := getIO()

			fo := GetFeedData()
			fo.input = "a"
			fo.cat = point.Logging
			fo.pts = c.pts
			fo.plOption = c.option.PlOption
			epts, _, _, err := dkio.beforeFeed(fo)
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

func Test_correctPointTime(t *T.T) {
	t.Run("basic", func(t *T.T) {
		var kvs point.KVs
		kvs = kvs.Set("f1", 1)
		now := time.Unix(0, 456)

		pts := []*point.Point{
			point.NewPoint("some", kvs, point.WithTimestamp(123)),
			point.NewPoint("some", kvs, point.WithTime(now)), // no correction
		}

		after, n := correctPointTime(pts, now, 1)
		assert.Equal(t, 1, n)
		assert.Equal(t, now.UnixNano(), after[0].Time().UnixNano())
		assert.Equal(t, int64(123), after[0].Get("__orig_time").(int64))
	})
}

func TestAddMetricInputTag(t *T.T) {
	metricPt := point.NewPoint("metric", point.NewKVs(nil).Set("value", 1), point.DefaultMetricOptions()...)
	loggingPt := point.NewPoint("logging", point.NewKVs(nil).Set("message", "test"), point.DefaultLoggingOptions()...)

	feeder := new(ioFeeder)
	feeder.addMetricInputTag(&feedData{cat: point.Metric, pts: []*point.Point{metricPt}, inputName: "snmp"})
	feeder.addMetricInputTag(&feedData{cat: point.Logging, pts: []*point.Point{loggingPt}, inputName: "snmp"})

	assert.Equal(t, "dk.snmp", metricPt.GetTag(InputSourceTagKey))
	assert.Empty(t, loggingPt.GetTag(InputSourceTagKey))

	unknownPt := point.NewPoint("metric", point.NewKVs(nil).Set("value", 1), point.DefaultMetricOptions()...)
	feeder.addMetricInputTag(&feedData{cat: point.Metric, pts: []*point.Point{unknownPt}})
	assert.Equal(t, "dk."+unknownInputName, unknownPt.GetTag(InputSourceTagKey))
}

func TestPLAggFeedAddsInputTag(t *T.T) {
	oldDefIO := defIO
	output := &mockFeederOutputer{}
	defIO = getIO()
	defIO.foDataway = output
	t.Cleanup(func() { defIO = oldDefIO })

	pt := point.NewPoint("metric", point.NewKVs(nil).Set("value", 1), point.DefaultMetricOptions()...)
	assert.NoError(t, PLAggFeed(point.Metric, "aggregation", []*point.Point{pt}))

	assert.Len(t, output.feeds, 1)
	assert.Equal(t, "dk."+pipelineInputName, output.feeds[0].pts[0].GetTag(InputSourceTagKey))
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	T "testing"

	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func BenchmarkGroup(b *T.B) {
	r := point.NewRander(point.WithFixedTags(true), point.WithRandText(3))

	pts := r.Rand(1000)

	dw := &Dataway{
		URLs: []string{
			"https://fake-dataway.com?token=tkn_xxxxxxxxxx",
		},
		GlobalCustomerKeys: []string{"source"},
		EnableSinker:       true,
	}

	assert.NoError(b, dw.Init())

	b.Run("one-tpgrouper", func(b *T.B) {
		ptg := getGrouper()
		defer putGrouper(ptg)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dw.doGroupPoints(ptg, point.Logging, pts)
		}
	})

	b.Run("multiple-tpgrouper", func(b *T.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ptg := getGrouper()
			dw.doGroupPoints(ptg, point.Logging, pts)
			putGrouper(ptg)
		}
	})
}

func TestGroupPoint(t *T.T) {
	t.Run("duplicate-keys", func(t *T.T) {
		metricsReset()
		dw := &Dataway{
			URLs: []string{"https://fake-dataway.com?token=tkn_xxxxxxxxxx"},
			GlobalCustomerKeys: []string{
				"category",
			},
			EnableSinker: true,
			GZip:         true,
		}

		assert.NoError(t, dw.Init())

		pts := []*point.Point{
			point.NewPointV2("some",
				append(point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("category", "system", point.WithKVTagSet(true)))),
		}

		ptg := getGrouper()
		defer putGrouper(ptg)
		dw.groupPoints(ptg, point.Security, pts)
		res := ptg.groupedPts

		for k := range res {
			t.Logf("key: %s, pts: %d", k, len(res[k]))
		}

		assert.Len(t, res["category=system,category=security"], 1)
	})

	t.Run("customer-keys", func(t *T.T) {
		metricsReset()
		dw := &Dataway{
			URLs: []string{
				"https://fake-dataway.com?token=tkn_xxxxxxxxxx",
			},
			GlobalCustomerKeys: []string{
				"class",
				"tag2",
				"t1", "t2", "t3", "t4",
				"t5", "t6", "t7", "t8",
			},
			EnableSinker: true,
			GZip:         true,
		}

		assert.NoError(t, dw.Init())

		pts := []*point.Point{
			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
					point.NewKV("t2", "v1", point.WithKVTagSet(true)),
					point.NewKV("t3", "v1", point.WithKVTagSet(true)),
					point.NewKV("t4", "v1", point.WithKVTagSet(true)),
					point.NewKV("t5", "v1", point.WithKVTagSet(true)),
					point.NewKV("t6", "v1", point.WithKVTagSet(true)),
					point.NewKV("t7", "v1", point.WithKVTagSet(true)),
					point.NewKV("t8", "v1", point.WithKVTagSet(true)),
					point.NewKV("tag1", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
					point.NewKV("tag2", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("message", "ns1", point.WithKVTagSet(true)),
					point.NewKV("tag2", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some", point.NewKVs(map[string]any{"f1": false})), // no tags
		}

		ptg := getGrouper()
		defer putGrouper(ptg)
		dw.groupPoints(ptg, point.Object, pts)
		res := ptg.groupedPts

		for k := range res {
			t.Logf("key: %s, pts: %d", k, len(res[k]))
		}

		assert.Len(t, res["t1=v1,t2=v1,t3=v1,t4=v1,t5=v1,t6=v1,t7=v1,t8=v1,class=some"], 1)
		assert.Len(t, res["t1=v1,class=some"], 1)
		assert.Len(t, res["t1=v1,tag2=new-value,class=some"], 1)
		assert.Len(t, res["tag2=new-value,class=some"], 1)
		assert.Len(t, res["class=some"], 1)

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)
		mfs, err := reg.Gather()
		assert.NoError(t, err)
		t.Logf("metrics:\n%s", metrics.MetricFamily2Text(mfs))
	})

	t.Run("random-pts-on-logging", func(t *T.T) {
		metricsReset()
		dw := &Dataway{
			URLs: []string{
				"https://fake-dataway.com?token=tkn_xxxxxxxxxx",
			},
			GlobalCustomerKeys: []string{"source"},
			EnableSinker:       true,
			GZip:               true,
		}

		assert.NoError(t, dw.Init(WithGlobalTags(map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		})))

		r := point.NewRander(point.WithFixedTags(true), point.WithRandText(3))

		pts := r.Rand(10)

		ptg := getGrouper()
		defer putGrouper(ptg)
		dw.groupPoints(ptg, point.Logging, pts)
		res := ptg.groupedPts

		for k, arr := range res {
			assert.Len(t, arr, 1)
			t.Logf("header value: %q", k)
		}

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)
		mfs, err := reg.Gather()
		assert.NoError(t, err)
		t.Logf("metrics:\n%s", metrics.MetricFamily2Text(mfs))
	})

	t.Run("basic", func(t *T.T) {
		metricsReset()
		dw := &Dataway{
			URLs: []string{
				"https://fake-dataway.com?token=tkn_xxxxxxxxxx",
			},
			GlobalCustomerKeys: []string{"namespace", "app"},
			EnableSinker:       true,
			GZip:               true,
		}

		assert.NoError(t, dw.Init(WithGlobalTags(map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		})))

		pts := []*point.Point{
			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),

					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
					point.NewKV("tag1", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),

					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
					point.NewKV("tag2", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some", point.NewKVs(map[string]any{"f1": false})), /* no tags */
		}

		ptg := getGrouper()
		defer putGrouper(ptg)
		dw.groupPoints(ptg, point.Metric, pts)
		res := ptg.groupedPts

		for k, arr := range res {
			t.Logf("header value: %q, pts: %d", k, len(arr))
		}

		assert.Len(t, res["tag1=new-value"], 1)
		assert.Len(t, res["tag2=new-value"], 1)
		assert.Len(t, res[""], 2)

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)
		mfs, err := reg.Gather()
		assert.NoError(t, err)
		t.Logf("metrics:\n%s", metrics.MetricFamily2Text(mfs))
	})

	t.Run("no-global-tags", func(t *T.T) {
		metricsReset()
		dw := &Dataway{
			URLs: []string{
				"https://fake-dataway.com?token=tkn_xxxxxxxxxx",
			},

			EnableSinker:       true,
			GlobalCustomerKeys: []string{"namespace", "app"},
			GZip:               true,
		}

		assert.NoError(t, dw.Init())

		pts := []*point.Point{
			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
					point.NewKV("tag1", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f3": false}),
					point.NewKV("t2", "v4", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
					point.NewKV("tag2", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("namespace", "ns1", point.WithKVTagSet(true)),
					point.NewKV("tag2", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some" /* no tags */, point.NewKVs(map[string]any{"f1": false})),
		}

		ptg := getGrouper()
		defer putGrouper(ptg)

		dw.groupPoints(ptg, point.Logging, pts)
		res := ptg.groupedPts
		assert.Len(t, res["namespace=ns1"], 1)
		assert.Len(t, res[""], 4)

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)
		mfs, err := reg.Gather()
		assert.NoError(t, err)
		t.Logf("metrics:\n%s", metrics.MetricFamily2Text(mfs))
	})

	t.Run("no-global-tags-on-object", func(t *T.T) {
		metricsReset()
		dw := &Dataway{
			URLs: []string{
				"https://fake-dataway.com?token=tkn_xxxxxxxxxx",
			},
			GlobalCustomerKeys: []string{"class"},
			EnableSinker:       true,
			GZip:               true,
		}

		assert.NoError(t, dw.Init())

		pts := []*point.Point{
			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
					point.NewKV("tag1", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
					point.NewKV("tag2", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("namespace", "ns1", point.WithKVTagSet(true)),
					point.NewKV("tag2", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some", point.NewKVs(map[string]any{"f1": false})), /* no tags */
		}

		ptg := getGrouper()
		defer putGrouper(ptg)
		dw.groupPoints(ptg, point.Object, pts)
		res := ptg.groupedPts

		for k := range res {
			t.Logf("key: %s", k)
		}

		assert.Len(t, res["class=some"], 5)

		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)
		mfs, err := reg.Gather()
		assert.NoError(t, err)
		t.Logf("metrics:\n%s", metrics.MetricFamily2Text(mfs))
	})

	t.Run("no-global-tags-no-customer-tag-keys", func(t *T.T) {
		metricsReset()
		dw := &Dataway{
			URLs: []string{
				"https://fake-dataway.com?token=tkn_xxxxxxxxxx",
			},
			EnableSinker: true,
			GZip:         true,
		}

		assert.NoError(t, dw.Init())

		pts := []*point.Point{
			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
					point.NewKV("tag1", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("t1", "v1", point.WithKVTagSet(true)),
					point.NewKV("tag2", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some",
				append(
					point.NewKVs(map[string]any{"f1": false}),
					point.NewKV("namespace", "ns1", point.WithKVTagSet(true)),
					point.NewKV("tag2", "new-value", point.WithKVTagSet(true)),
				)),

			point.NewPointV2("some", point.NewKVs(map[string]any{"f1": false})), /* no tags */
		}

		ptg := getGrouper()
		defer putGrouper(ptg)

		dw.groupPoints(ptg, point.Object, pts)
		res := ptg.groupedPts
		assert.Len(t, res[""], 5)

		metricsReset()
		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)
		mfs, err := reg.Gather()
		assert.NoError(t, err)
		t.Logf("metrics:\n%s", metrics.MetricFamily2Text(mfs))
	})
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package plmap

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

func TestAggBuckets(t *testing.T) {
	ptsLi := map[string][]*dkpt.Point{}
	var fn UploadFunc = func(cat point.Category, n string, d any) error {
		ptsLi[n] = append(ptsLi[n], d.([]*dkpt.Point)...)
		return nil
	}

	buks := NewAggBuks(fn)
	buks.CreateBucket(point.Metric, "bucket_a", time.Second*5, 0, false, nil)
	buks.CreateBucket(point.Metric, "bucket_a", time.Second, 0, false, nil)
	buks.CreateBucket(point.Metric, "bucket_a", time.Second, 0, false, nil)
	buks.CreateBucket(point.Metric, "bucket_b", time.Second, 0, false, nil)

	v, ok := buks.GetBucket(point.Metric, "bucket_a")
	assert.NotEqual(t, nil, v)
	assert.Equal(t, true, ok)

	if v, ok := buks.GetBucket(point.Metric, "bucket_b"); ok {
		assert.NotEqual(t, nil, v)
		v.stopScan()
	} else {
		assert.Equal(t, true, ok)
	}
	v.AddMetric("f1", "avg", []string{"t1"}, []string{"t1_val"}, 1)
	v.AddMetric("f1", "avg", []string{"t1"}, []string{"t1_val"}, 2)
	v.AddMetric("f1", "avg", []string{"t1"}, []string{"t1_val"}, 3)

	buks.StopAllBukScanner()
	time.Sleep(time.Millisecond * 10)
}

func TestAggBuckets2(t *testing.T) {
	ptsLi := map[point.Category]map[string][]*dkpt.Point{}
	var fn UploadFunc = func(cat point.Category, n string, d any) error {
		if ptsLi[cat] == nil {
			ptsLi[cat] = map[string][]*dkpt.Point{}
		}

		ptsLi[cat][n] = append(ptsLi[cat][n], d.([]*dkpt.Point)...)
		return nil
	}

	buks := NewAggBuks(fn)
	buks.CreateBucket(point.Metric, "bucket_a", time.Second*5, 0, false, nil)
	buks.CreateBucket(point.Metric, "bucket_a", 0, 2, false, nil)
	buks.CreateBucket(point.Metric, "bucket_a", 0, 2, false, nil)
	buks.CreateBucket(point.Metric, "bucket_b", 0, 2, false, nil)
	buks.CreateBucket(point.Logging, "bucket_b", 0, 2, false, nil)

	v, ok := buks.GetBucket(point.Metric, "bucket_a")
	assert.NotEqual(t, nil, v)
	assert.Equal(t, true, ok)

	v.AddMetric("f1", "avg", []string{"t1"}, []string{"t1_val"}, 1)
	v.AddMetric("f1", "avg", []string{"t1"}, []string{"t1_val"}, 2)
	v.AddMetric("f1", "avg", []string{"t1"}, []string{"t1_val"}, 3)

	if v, ok := buks.GetBucket(point.Metric, "bucket_b"); ok {
		assert.NotEqual(t, nil, v)

		v.AddMetric("f1", "avg", []string{"t1"}, []string{"t1_val"}, 1)
		v.AddMetric("f1", "avg", []string{"t1"}, []string{"t1_val"}, 2)
		v.AddMetric("f1", "avg", []string{"t1"}, []string{"t1_val"}, 3)
		v.stopScan()
	} else {
		assert.Equal(t, true, ok)
	}

	if v, ok := buks.GetBucket(point.Logging, "bucket_b"); ok {
		assert.NotEqual(t, nil, v)

		v.AddMetric("f1", "avg", []string{"t1"}, []string{"t1_val"}, 1)
		v.AddMetric("f1", "avg", []string{"t1"}, []string{"t1_val"}, 2)
		v.AddMetric("f1", "avg", []string{"t1"}, []string{"t1_val"}, 3)
		v.stopScan()
	} else {
		assert.Equal(t, true, ok)
	}

	buks.StopAllBukScanner()

	time.Sleep(time.Millisecond * 100)
}

func TestAggMetric(t *testing.T) {
	cases := []struct {
		action string
		d      []any
		o      any
		failed bool
	}{
		{
			action: "avg",
			d:      []any{1, 2, 1, 2},
			o:      1.5,
		},
		{
			action: "sum",
			d:      []any{1, 2, 1, 2},
			o:      6.0,
		},
		{
			action: "min",
			d:      []any{1, 2, 3, 5, 1, 2, 1},
			o:      1.0,
		},
		{
			action: "max",
			d:      []any{1, 2, 1, 2},
			o:      2.0,
		},
		{
			action: "set",
			d:      []any{1, 2, 1, 2, -1},
			o:      -1.0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.action, func(t *testing.T) {
			v, ok := NewAggMetric("", tc.action)
			assert.Equal(t, !tc.failed, ok)
			if v != nil {
				for _, d := range tc.d {
					v.Append(d)
				}
			}
			assert.Equal(t, tc.o, v.Value())
		})
	}
}

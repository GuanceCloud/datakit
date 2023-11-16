// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func Test_adjustPointTime(t *T.T) {
	t.Run(`adjust`, func(t *T.T) {
		recordTime := time.Now().Round(0).Add(-time.Hour) // 1h ago
		r := point.NewRander(point.WithRandTime(recordTime))
		pts := r.Rand(2) // all points with same time

		when := time.Now().Round(0).Add(-time.Minute)

		t.Logf("before pts: %s", pts[0].Pretty())

		pts = adjustPointTime(when, pts)

		t.Logf("after pts: %s", pts[0].Pretty())

		for _, pt := range pts {
			assert.Equal(t, when, pt.Time())
		}
	})

	t.Run(`adjust-to-older-time`, func(t *T.T) {
		recordTime := time.Now().Round(0).Add(-time.Hour) // 1h ago
		r := point.NewRander(point.WithRandTime(recordTime))
		pts := r.Rand(2) // all points with same time

		when := time.Now().Round(0).Add(-2 * time.Hour)

		t.Logf("before pts: %s", pts[0].Pretty())

		pts = adjustPointTime(when, pts)

		t.Logf("after pts: %s", pts[0].Pretty())

		for _, pt := range pts {
			assert.Equal(t, when, pt.Time())
		}
	})

	t.Run(`scattered-points`, func(t *T.T) {
		when := time.Now().Round(0)

		pts := []*point.Point{
			// 2 points with interval 1h
			point.NewPointV2("p1", nil, point.WithTime(when.Add(-2*time.Hour))),
			point.NewPointV2("p2", nil, point.WithTime(when.Add(-time.Hour))),
		}

		t.Logf("before pts: %s", pts[0].Pretty())
		t.Logf("before pts: %s", pts[1].Pretty())

		pts = adjustPointTime(when, pts)

		t.Logf("after pts: %s", pts[0].Pretty())
		t.Logf("after pts: %s", pts[1].Pretty())

		assert.Equal(t, when, pts[1].Time())
		assert.Equal(t, when.Add(-time.Hour), pts[0].Time())
	})
}

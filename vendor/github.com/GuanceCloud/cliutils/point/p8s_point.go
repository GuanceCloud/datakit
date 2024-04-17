// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// mergePts merge pts when:
//
//   - they got same measurement name
//
//   - they got same tag key and tag values
//
//   - they got same time(nano-second)
//
//     NOTE: you should ensure that these time is equal, point's hash not
//     covered the time field. For prometheus metrics, these time value are
//     the same.
//
// When point.Point are logging, due to the lack of `time-series',
// we hava to merge multiple points' fields together to build a single point.
//
// For time-series data, we don't need to do this, the storage
// engine merged them automatically(grouped by time-series).
func mergePts(pts []*Point) []*Point {
	// same-hash point put together
	var res []*Point
	ptMap := map[string][]*Point{}
	for _, pt := range pts {
		hash := pt.MD5()
		ptMap[hash] = append(ptMap[hash], pt)
	}

	for _, pts := range ptMap {
		if len(pts) > 1 {
			// merge all points(with same hash) fields to the first one.
			for i := 1; i < len(pts); i++ {
				fs := pts[i].Fields()
				for _, f := range fs {
					pts[0].AddKVs(f)
				}
			}

			// keep the first point, drop all merged points.
			res = append(res, pts[0])
		}
	}

	return res
}

func doGatherPoints(reg prometheus.Gatherer) ([]*Point, error) {
	mfs, err := reg.Gather()
	if err != nil {
		return nil, err
	}

	// All gathered data should have the same timestamp, we enforce it.
	now := time.Now()

	var pts []*Point
	for _, mf := range mfs {
		arr := strings.SplitN(*mf.Name, "_", 2)

		name := arr[0]
		fieldName := arr[1]

		for _, m := range mf.Metric {
			var kvs KVs
			for _, label := range m.GetLabel() {
				kvs = append(kvs, NewKV(label.GetName(), label.GetValue(), WithKVTagSet(true)))
			}

			switch *mf.Type {
			case dto.MetricType_COUNTER:
				kvs = append(kvs, NewKV(fieldName, m.GetCounter().GetValue()))
			case dto.MetricType_SUMMARY:
				avg := uint64(m.GetSummary().GetSampleSum()) / m.GetSummary().GetSampleCount()
				kvs = append(kvs, NewKV(fieldName, avg))

			case dto.MetricType_GAUGE:
				continue // TODO
			case dto.MetricType_HISTOGRAM:
				continue // TODO
			case dto.MetricType_UNTYPED:
				continue // TODO
			case dto.MetricType_GAUGE_HISTOGRAM:
				continue // TODO
			}

			// TODO: according to specific tags, we should make them as logging.
			ts := now
			if m.TimestampMs != nil { // use metric time
				ts = time.Unix(0, int64(time.Millisecond)**m.TimestampMs)
			}

			opts := append(DefaultMetricOptions(), WithTime(ts))
			pts = append(pts, NewPointV2(name, kvs, opts...))
		}
	}

	return pts, nil
}

// GatherPoints gather all metrics in global registry, but convert these metrics
// to Point.
func GatherPoints(reg prometheus.Gatherer) ([]*Point, error) {
	return doGatherPoints(reg)
}

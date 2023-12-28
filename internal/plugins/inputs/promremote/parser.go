// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promremote

import (
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
)

type Parser struct {
	MetricNameFilter      []string `toml:"metric_name_filter"`
	MeasurementNameFilter []string `toml:"measurement_name_filter"`
	MeasurementPrefix     string   `toml:"measurement_prefix"`
	MeasurementName       string   `toml:"measurement_name"`

	metricNameReFilter      []*regexp.Regexp
	measurementNameReFilter []*regexp.Regexp
}

// Parse parses given byte as protocol buffer. it performs necessary
// metric filtering and prefixing, and returns parsed measurements.
func (p *Parser) Parse(timeSeries []prompb.TimeSeries, ipt *Input, additionalTags map[string]string) ([]*point.Point, error) {
	var err error
	var pts []*point.Point
	now := time.Now()
	var t time.Time
	timeOpt := point.WithTime(now)
	opts := point.DefaultMetricOptions()
	opts = append(opts, timeOpt)

	demoSource, ok := additionalTags["__source"]
	if !ok {
		demoSource = "default"
	}
	delete(additionalTags, "__source")

	var duraSum int64
	var lenDura int64
	var noTime int64

	for _, ts := range timeSeries {
		tags := map[string]string{}

		var ok bool
		var metric string
		for _, l := range ts.Labels {
			lName := l.Name
			if lName == model.MetricNameLabel {
				ok = true
				metric = l.Value
				continue
			}
			if ipt.tagFilter(lName) {
				for oldKey, newKey := range ipt.TagsRename {
					if lName == oldKey {
						lName = newKey
					}
				}
				tags[lName] = l.Value
			}
		}

		if !ok {
			return nil, fmt.Errorf("metric name %q not found in tag-set or empty", model.MetricNameLabel)
		}

		measurementName, metricName := p.getNames(metric)

		if !p.shouldFilterMeasurementName(measurementName) {
			continue
		}
		// This is metric, not metricName. Include prefix measurementName_ .
		if !p.shouldFilterMetricName(metric) {
			continue
		}

		for _, s := range ts.Samples {
			if !math.IsNaN(s.Value) {
				var kvs point.KVs
				for k, v := range ipt.mergedTags {
					kvs = kvs.MustAddTag(k, v)
				}
				for k, v := range additionalTags {
					kvs = kvs.MustAddTag(k, v)
				}
				for k, v := range tags {
					kvs = kvs.MustAddTag(k, v)
				}

				kvs = kvs.Add(metricName, s.Value, false, true)

				if s.Timestamp > 0 {
					t = time.Unix(0, s.Timestamp*1000000)
					opts[len(opts)-1] = point.WithTime(t)
					duraSum += int64(now.Sub(t))
					lenDura++
				} else {
					opts[len(opts)-1] = timeOpt
					noTime++
				}

				pts = append(pts, point.NewPointV2(measurementName, kvs, opts...))
			}
		}
	}

	var diffDuration float64
	if lenDura > 0 {
		diffDuration = float64(duraSum) / float64(lenDura)
	}
	httpTimeDiffVec.WithLabelValues(demoSource).Observe(diffDuration / float64(time.Second))
	noTimePointsVec.WithLabelValues(demoSource).Add(float64(noTime))
	collectPointsTotalVec.WithLabelValues(demoSource).Observe(float64(len(pts)))

	return pts, err
}

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
	"github.com/gogo/protobuf/proto"
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
func (p *Parser) Parse(buf []byte, ipt *Input) ([]*point.Point, error) {
	var err error
	var pts []*point.Point
	var req prompb.WriteRequest
	if err := proto.Unmarshal(buf, &req); err != nil {
		return nil, fmt.Errorf("unable to unmarshal request body: %w", err)
	}
	now := time.Now()
	t := time.Now()
	_ = t
	timeOpt := point.WithTime(now)
	opts := point.DefaultMetricOptions()
	opts = append(opts, timeOpt)

	for _, ts := range req.Timeseries {
		tags := map[string]string{}

		for _, l := range ts.Labels {
			tags[l.Name] = l.Value
		}

		metric := tags[model.MetricNameLabel]
		if metric == "" {
			return nil, fmt.Errorf("metric name %q not found in tag-set or empty", model.MetricNameLabel)
		}
		delete(tags, model.MetricNameLabel)

		if !p.shouldFilterThroughMetricName(metric) {
			continue
		}
		if !p.shouldFilterThroughMeasurementName(metric) {
			continue
		}

		measurementName, metricName := p.getNames(metric)

		for _, s := range ts.Samples {
			if !math.IsNaN(s.Value) {
				var kvs point.KVs

				kvs = kvs.Add(metricName, s.Value, false, true)

				for k, v := range ipt.Tags {
					kvs = kvs.MustAddTag(k, v)
				}

				for k, v := range tags {
					kvs = kvs.MustAddTag(k, v)
				}

				if s.Timestamp > 0 {
					t = time.Unix(0, s.Timestamp*1000000)
					opts[len(opts)-1] = point.WithTime(t)
				} else {
					opts[len(opts)-1] = timeOpt
				}

				pts = append(pts, point.NewPointV2(measurementName, kvs, opts...))
			}
		}
	}
	return pts, err
}

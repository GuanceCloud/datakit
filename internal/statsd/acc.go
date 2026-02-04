// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
)

type accumulator struct {
	ref            *Collector
	points         []*point.Point
	feedMetricName string
	l              *logger.Logger
}

func (a *accumulator) addFields(name string, fields map[string]any, tags map[string]string, ts time.Time) {
	for k, v := range a.ref.opts.tags {
		tags[k] = v // may override tags in real-data
	}

	for _, t := range a.ref.opts.dropTags {
		a.l.Debugf("drop tag %s", t)
		delete(tags, t)
	}

	a.doFeedMetricName(tags)

	a.l.Debugf("on metric %s, got %d fields: %+#v", name, len(fields), fields)

	ptsopt := point.DefaultMetricOptions()

	for k, v := range fields {
		metricName := name
		fieldKey := name // we choose metric name as field name in influxdb's line protocol

		if len(a.ref.mmap) > 0 {
			for from, to := range a.ref.mmap {
				if strings.HasPrefix(name, from) {
					metricName = to
					fieldKey = strings.TrimPrefix(name, from)

					a.l.Debugf("renaming: %s | %s", metricName, fieldKey)
					break
				}
			}
		} else {
			arr := strings.SplitN(name, a.ref.opts.metricSeparator, 2)
			if len(arr) < 2 {
				a.l.Warnf("got metric '%s', accept it", name)
				metricName = name
			} else {
				metricName = arr[0]
				fieldKey = arr[1]
			}

			a.l.Debugf("update naming %s | %s", metricName, fieldKey)
		}

		// Check metric
		if len(metricName) == 0 || len(fieldKey) == 0 {
			a.l.Warnf("error metricName|fieldKey: %s|%s", metricName, fieldKey)
			return
		}

		switch k {
		case "value":
		default:
			// fieldKey append with _mean/_stddev/_sum/... suffix
			fieldKey = fieldKey + "_" + k
		}

		a.l.Debugf("addFields: %s|%s: %v", metricName, fieldKey, v)
		var kvs point.KVs
		kvs = kvs.Add(fieldKey, v)
		for tk, tv := range tags {
			kvs = kvs.SetTag(tk, tv)
		}

		pt := point.NewPoint(metricName, kvs, append(ptsopt, point.WithTime(ts))...)

		a.points = append(a.points, pt)
	}
}

func (a *accumulator) doFeedMetricName(tags map[string]string) {
	a.feedMetricName = "statsd.x.x" // default
	if len(a.ref.opts.statsdSourceKey) > 0 || len(a.ref.opts.statsdHostKey) > 0 {
		sourceKey := tags[a.ref.opts.statsdSourceKey]
		hostKey := tags[a.ref.opts.statsdHostKey]
		if len(sourceKey) == 0 {
			sourceKey = "x"
		}
		if len(hostKey) == 0 {
			hostKey = "x"
		}
		a.feedMetricName = "statsd." + sourceKey + "." + hostKey

		if !a.ref.opts.saveAboveKey {
			delete(tags, a.ref.opts.statsdSourceKey)
			delete(tags, a.ref.opts.statsdHostKey)
		}
	}
}

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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type statsdMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *statsdMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *statsdMeasurement) Info() *inputs.MeasurementInfo {
	return nil
}

type measurementInfo struct {
	FeedMetricName string
	PT             *point.Point
}

type accumulator struct {
	ref              *Collector
	measurementInfos []*measurementInfo
	feedMetricName   string
	l                *logger.Logger
}

func (a *accumulator) addFields(name string, fields map[string]interface{}, tags map[string]string, ts time.Time) {
	for k, v := range a.ref.opts.tags {
		tags[k] = v // may override tags in real-data
	}

	for _, t := range a.ref.opts.dropTags {
		a.l.Debugf("drop tag %s", t)
		delete(tags, t)
	}

	a.doFeedMetricName(tags)

	// Requrements: there shoule be only 1 field, the field key should be 'value'
	if len(fields) != 1 {
		a.l.Warnf("drop metric %s, got %d fields: %+#v", name, len(fields), fields)
		return
	}

	fval, ok := fields["value"]
	if !ok {
		a.l.Warnf("drop metric %s, field 'value' missing", name)
		return
	}

	metricName := name
	fieldKey := name // we choose metric name as field name in influxdb's line protocol

	if len(a.ref.mmap) > 0 {
		for from, to := range a.ref.mmap {
			if strings.HasPrefix(name, from) {
				metricName = to
				fieldKey = strings.TrimPrefix(name, from)
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
	}

	// Check metric
	if len(metricName) == 0 || len(fieldKey) == 0 {
		a.l.Warnf("error metricName|fieldKey: %s|%s", metricName, fieldKey)
		return
	}

	a.l.Debugf("addFields: %s|%s", metricName, fieldKey)
	metric := &statsdMeasurement{
		name: metricName,
		fields: map[string]interface{}{
			fieldKey: fval,
		},
		tags: tags,
		ts:   ts,
	}

	a.measurementInfos = append(a.measurementInfos, &measurementInfo{
		FeedMetricName: a.feedMetricName,
		PT:             metric.Point(),
	})
}

func (a *accumulator) doFeedMetricName(tags map[string]string) {
	a.feedMetricName = "statsd/-/-" // default
	if len(a.ref.opts.statsdSourceKey) > 0 || len(a.ref.opts.statsdHostKey) > 0 {
		sourceKey := tags[a.ref.opts.statsdSourceKey]
		hostKey := tags[a.ref.opts.statsdHostKey]
		if len(sourceKey) == 0 {
			sourceKey = "-"
		}
		if len(hostKey) == 0 {
			hostKey = "-"
		}
		a.feedMetricName = "statsd/" + sourceKey + "/" + hostKey

		if !a.ref.opts.saveAboveKey {
			delete(tags, a.ref.opts.statsdSourceKey)
			delete(tags, a.ref.opts.statsdHostKey)
		}
	}
}

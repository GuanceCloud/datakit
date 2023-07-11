// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type statsdMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
	ipt    *Input
}

// Point implement MeasurementV2.
func (m *statsdMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.ipt.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *statsdMeasurement) LineProto() (*dkpt.Point, error) {
	// return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOpt())
	return nil, fmt.Errorf("not implement")
}

func (m *statsdMeasurement) Info() *inputs.MeasurementInfo {
	return nil
}

type accumulator struct {
	ref            *Input
	measurements   []*point.Point
	feedMetricName string
}

func (a *accumulator) addFields(name string, fields map[string]interface{}, tags map[string]string, ts time.Time) {
	for k, v := range a.ref.Tags {
		tags[k] = v // may override tags in real-data
	}

	for _, t := range a.ref.DropTags {
		l.Debugf("drop tag %s", t)
		delete(tags, t)
	}

	a.doFeedMetricName(tags)

	// Requrements: there shoule be only 1 field, the field key should be `value'
	if len(fields) != 1 {
		l.Warnf("drop metric %s, got %d fields: %+#v", name, len(fields), fields)
		return
	}

	fval, ok := fields["value"]
	if !ok {
		l.Warnf("drop metric %s, field `value' missing", name)
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
		arr := strings.SplitN(name, a.ref.MetricSeparator, 2)
		if len(arr) < 2 {
			l.Warnf("got metric `%s', accept it", name)
			metricName = name
		} else {
			metricName = arr[0]
			fieldKey = arr[1]
		}
	}

	l.Debugf("addFields: %s|%s", metricName, fieldKey)

	metric := &statsdMeasurement{
		name: metricName,
		fields: map[string]interface{}{
			fieldKey: fval,
		},
		tags: tags,
		ts:   ts,
		ipt:  a.ref,
	}
	a.measurements = append(a.measurements, metric.Point())
}

func (a *accumulator) doFeedMetricName(tags map[string]string) {
	a.feedMetricName = "statsd/-/-" // default
	if len(a.ref.StatsdSourceKey) > 0 || len(a.ref.StatsdHostKey) > 0 {
		sourceKey := tags[a.ref.StatsdSourceKey]
		hostKey := tags[a.ref.StatsdHostKey]
		if len(sourceKey) == 0 {
			sourceKey = "-"
		}
		if len(hostKey) == 0 {
			hostKey = "-"
		}
		a.feedMetricName = "statsd/" + sourceKey + "/" + hostKey

		if !a.ref.SaveAboveKey {
			delete(tags, a.ref.StatsdSourceKey)
			delete(tags, a.ref.StatsdHostKey)
		}
	}
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"strings"
	"time"

	dkpoint "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type point struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	tm     time.Time
}

func (p *point) LineProto() (*dkpoint.Point, error) {
	return dkpoint.NewPoint(p.name, p.tags, p.fields, dkpoint.MOpt())
}

func (p *point) Info() *inputs.MeasurementInfo {
	return nil
}

type accumulator struct {
	ref    *input
	points []inputs.Measurement
}

func (a *accumulator) addFields(name string, fields map[string]interface{}, tags map[string]string, ts time.Time) {
	for k, v := range a.ref.Tags {
		tags[k] = v // may override tags in real-data
	}

	for _, t := range a.ref.DropTags {
		l.Debugf("drop tag %s", t)
		delete(tags, t)
	}

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
	a.points = append(a.points, &point{
		name: metricName,
		fields: map[string]interface{}{
			fieldKey: fval,
		},
		tags: tags,
		tm:   ts,
	})
}

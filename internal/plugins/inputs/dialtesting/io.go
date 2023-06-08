// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package dialtesting

import (
	"time"

	pt "github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func (d *dialer) pointsFeed(urlStr string) error {
	startTime := time.Now()
	tags, fields := d.task.GetResults()

	if status, ok := tags["status"]; ok {
		taskCheckCostSummary.WithLabelValues(d.regionName, d.class, status).Observe(float64(time.Since(startTime)) / float64(time.Second))
	}

	for k, v := range d.tags {
		if d.measurementInfo != nil && d.measurementInfo.Tags != nil {
			if _, ok := d.measurementInfo.Tags[k]; !ok {
				continue
			}
		}

		if _, ok := tags[k]; !ok {
			tags[k] = v
		} else {
			l.Debugf("ignore dialer tag %s: %s", k, v)
		}
	}
	data := pt.NewPointV2([]byte(d.task.MetricName()),
		append(pt.NewTags(tags), pt.NewKVs(fields)...), pt.DefaultLoggingOptions()...)

	pts := []*pt.Point{}
	pts = append(pts, data)

	err := d.feeder.Feed(inputName, pt.Metric, pts, &io.Option{
		HTTPHost: urlStr,
	})

	l.Debugf(`url:%s, tags: %+#v, fs: %+#v`, urlStr, tags, fields)

	return err
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package dialtesting

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

func (d *dialer) pointsFeed(urlStr string) error {
	// 获取此次任务执行的基本信息
	startTime := time.Now()
	tags, fields := d.task.GetResults()

	if status, ok := tags["status"]; ok {
		taskCheckCostSummary.WithLabelValues(d.regionName, d.class, status).Observe(float64(time.Since(startTime)) / float64(time.Second))
	}

	for k, v := range d.tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		} else {
			l.Warnf("ignore dialer tag %s: %s", k, v)
		}
	}
	data, err := point.NewPoint(d.task.MetricName(), tags, fields, point.LOpt())
	if err != nil {
		l.Warnf("make metric failed: %s", err.Error)
		return err
	}

	pts := []*point.Point{}
	pts = append(pts, data)

	err = Feed(inputName, datakit.Logging, pts, &io.Option{
		HTTPHost: urlStr,
	})

	l.Debugf(`url:%s, tags: %+#v, fs: %+#v`, urlStr, tags, fields)

	return err
}

func Feed(name, category string, pts []*point.Point, opt *io.Option) error {
	return io.Feed(name, category, pts, opt)
}

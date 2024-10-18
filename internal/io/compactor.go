// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

type compactor struct {
	category      point.Category
	compactTicker *time.Ticker
	points        []*point.Point
	lastCompact   time.Time
}

func (x *dkIO) runCompactor(cat point.Category) {
	r := x.fo.Reader(cat)
	if r == nil && cat != point.DynamicDWCategory {
		log.Panicf("invalid category %q, should not been here", cat.String())
	}

	c := &compactor{
		compactTicker: time.NewTicker(x.flushInterval),
		category:      cat,
	}

	defer c.compactTicker.Stop()

	if cat == point.DynamicDWCategory {
		// NOTE: 目前只有拨测的数据会将数据打到 dynamicDatawayPts 中，而拨测数据
		// 是写日志，故将 category 设置为 logging
		c.category = point.Logging
	}

	log.Infof("run compactor on %s", c.category)
	for {
		select {
		case d := <-r:
			x.cacheData(c, d, true)
			PutFeedOption(d) // release feed options here

		case <-c.compactTicker.C:
			if len(c.points) > 0 {
				log.Debugf("on tick(%s) to compact %s(%d points), last compact %s ago...",
					x.flushInterval, c.category, len(c.points), time.Since(c.lastCompact))
				x.compact(c)
			}

		case <-datakit.Exit.Wait():
			if len(c.points) > 0 {
				log.Debugf("on tick(%s) to compact %s(%d points), last compact %s ago...",
					x.flushInterval, c.category, len(c.points), time.Since(c.lastCompact))
				x.compact(c)
			}
			return
		}
	}
}

func (x *dkIO) cacheData(c *compactor, d *feedOption, tryClean bool) {
	if d == nil {
		log.Warn("get empty data, ignored")
		return
	}

	if len(d.pts) == 0 {
		log.Warnf("no point from %q", d.input)
		return
	}

	defer func() {
		queuePtsVec.WithLabelValues(d.cat.String()).Set(float64(len(c.points)))
	}()

	log.Debugf("get iodata(%d points) from %s|%s", len(d.pts), d.cat, d.input)

	x.recordPoints(d)
	c.points = append(c.points, d.pts...)

	if tryClean && x.compactAt > 0 && len(c.points) > x.compactAt {
		x.compact(c)

		// reset compact ticker to prevent small packages
		c.compactTicker.Reset(x.flushInterval)
	}
}

func (x *dkIO) recordPoints(d *feedOption) {
	if x.recorder != nil && x.recorder.Enabled {
		if err := x.recorder.Record(d.pts, d.cat, d.input); err != nil {
			log.Warnf("record %d points on %q from %q failed: %s", len(d.pts), d.cat, d.input, err)
		} else {
			log.Debugf("record %d points on %q from %q ok", len(d.pts), d.cat, d.input)
		}
	} else {
		log.Debugf("recorder disabled: %+#v", x.recorder)
	}
}

func (x *dkIO) compact(c *compactor) {
	c.lastCompact = time.Now()

	defer func() {
		flushVec.WithLabelValues(c.category.String()).Inc()
	}()

	if err := x.doCompact(c.points, c.category); err != nil {
		log.Warnf("post %d points to %s failed: %s, ignored", len(c.points), c.category, err)
	}

	// I think here is the best position to put back these points.
	datakit.PutbackPoints(c.points...)

	c.points = c.points[:0] // clear
}

func (x *dkIO) doCompact(points []*point.Point, cat point.Category, dynamicURL ...string) error {
	if x.dw == nil {
		return fmt.Errorf("dataway not set")
	}

	if len(points) == 0 {
		return nil
	}

	opts := []dataway.WriteOption{
		dataway.WithPoints(points),
		// max cache size(in memory) upload as a batch
		dataway.WithBatchSize(x.compactAt),
		dataway.WithCategory(cat),
	}

	if len(dynamicURL) > 0 {
		opts = append(opts, dataway.WithDynamicURL(dynamicURL[0]))
	}

	return x.dw.Write(opts...)
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/failcache"
)

type consumer struct {
	fc failcache.Cache

	category point.Category

	flushTiker *time.Ticker

	points    []*point.Point
	lastFlush time.Time
}

func (x *dkIO) runConsumer(cat point.Category) {
	r := x.fo.Reader(cat)
	if r == nil && x.enableCache && cat != point.DynamicDWCategory {
		log.Panicf("invalid category %q, should not been here", cat.String())
	}

	fc, ok := x.fcs[cat.String()]
	if !ok {
		log.Infof("IO local cache not set for %q", cat.String())
	}

	c := &consumer{
		flushTiker: time.NewTicker(x.flushInterval),
		fc:         fc,
		category:   cat,
	}

	defer c.flushTiker.Stop()

	if cat == point.DynamicDWCategory {
		// NOTE: 目前只有拨测的数据会将数据打到 dynamicDatawayPts 中，而拨测数据
		// 是写日志，故将 category 设置为 logging
		c.category = point.Logging
	}

	fcTick := time.NewTicker(x.cacheCleanInterval)
	defer fcTick.Stop()

	// close diskcache on exit.
	defer func() {
		if c.fc != nil {
			if err := c.fc.Close(); err != nil {
				log.Warnf("cache.Close: %s, ignored", err)
			}
		}
	}()

	log.Infof("run consumer on %s", c.category)
	for {
		select {
		case d := <-r:
			x.cacheData(c, d, true)

		case <-c.flushTiker.C:
			if len(c.points) > 0 {
				log.Debugf("on tick(%s) to flush %s(%d points), last flush %s ago...",
					x.flushInterval, c.category, len(c.points), time.Since(c.lastFlush))
				x.flush(c)
			}

		case <-fcTick.C:
			x.flushFailCache(c)

		case <-datakit.Exit.Wait():
			log.Infof("io consumer on %s exit on exit", c.category)
			return
		}
	}
}

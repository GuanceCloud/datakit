// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/failcache"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

type consumer struct {
	ch chan *iodata
	fc failcache.Cache

	category string

	flushTiker *time.Ticker

	pts               []*point.Point
	lastFlush         time.Time
	dynamicDatawayPts map[string][]*point.Point // 拨测数据
}

func (x *dkIO) runConsumer(category string) {
	ch, ok := x.chans[category]
	if !ok {
		log.Panicf("invalid category %s, should not been here", category)
	}

	fc, ok := x.fcs[category]
	if !ok {
		if x.enableCache && category != datakit.DynamicDatawayCategory {
			log.Panicf("invalid category %s, should not been here", category)
		}
	}

	c := &consumer{
		ch:                ch,
		flushTiker:        time.NewTicker(x.flushInterval),
		fc:                fc,
		category:          category,
		dynamicDatawayPts: map[string][]*point.Point{},
	}

	defer c.flushTiker.Stop()

	if category == datakit.DynamicDatawayCategory {
		// NOTE: 目前只有拨测的数据会将数据打到 dynamicDatawayPts 中，而拨测数据
		// 是写日志，故将 category 设置为 logging
		c.category = datakit.Logging
	}

	fcTick := time.NewTicker(x.cacheCleanInterval)
	defer fcTick.Stop()

	log.Infof("run consumer on %s", c.category)
	for {
		select {
		case d := <-ch:
			x.cacheData(c, d, true)

		case <-c.flushTiker.C:
			if len(c.pts) > 0 {
				log.Debugf("on tick(%s) to flush %s(%d pts), last flush %s ago...",
					x.flushInterval, c.category, len(c.pts), time.Since(c.lastFlush))
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

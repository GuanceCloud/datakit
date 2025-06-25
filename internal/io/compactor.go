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

	readyPoints   int // queued point count in arrPoints and indexedPoints
	arrPoints     []*point.Point
	indexedPoints map[string][]*point.Point
}

func (x *dkIO) runCompactor(cat point.Category) {
	r := x.fo.Reader(cat)
	if r == nil && cat != point.DynamicDWCategory {
		log.Panicf("invalid category %q, should not been here", cat.String())
	}

	c := &compactor{
		compactTicker: time.NewTicker(x.flushInterval),
		category:      cat,
		indexedPoints: map[string][]*point.Point{},
	}

	defer c.compactTicker.Stop()

	if cat == point.DynamicDWCategory {
		// NOTE: 目前只有拨测的数据会将数据打到 dynamicDatawayPts 中，而拨测数据
		// 是写日志，故将 category 设置为 logging
		c.category = point.Logging
	}

	log.Infof("run compactor on %s, compact at %s/%d points", c.category, x.flushInterval, x.compactAt)
	for {
		select {
		case d := <-r:
			x.cacheData(c, d, true)
			putFeedData(d) // release feed data here

		case <-c.compactTicker.C:
			if c.readyPoints > 0 {
				log.Debugf("on tick(%s) to compact %s(%d points)", x.flushInterval, c.category, c.readyPoints)
				x.compact(c)
			}

		case <-datakit.Exit.Wait():
			if c.readyPoints > 0 {
				log.Debugf("on tick(%s) to compact %s(%d points)", x.flushInterval, c.category, c.readyPoints)
				x.compact(c)
			}

			log.Infof("compactor on %s exit", c.category)
			return
		}
	}
}

func (x *dkIO) cacheData(c *compactor, d *feedData, tryClean bool) {
	if d == nil {
		log.Warn("get empty data, ignored")
		return
	}

	if len(d.pts) == 0 {
		log.Warnf("no point from %q", d.input)
		return
	}

	c.readyPoints += len(d.pts)

	log.Debugf("get iodata(%d points) from %s|%s", len(d.pts), d.cat, d.input)

	x.recordPoints(d)
	if d.storageIndex != "" {
		c.indexedPoints[d.storageIndex] = append(c.indexedPoints[d.storageIndex], d.pts...)
	} else {
		c.arrPoints = append(c.arrPoints, d.pts...)
	}

	queuePtsVec.WithLabelValues(d.cat.String()).Add(float64(len(d.pts)))

	if tryClean && x.compactAt > 0 && c.readyPoints > x.compactAt {
		x.compact(c)
		// reset compact ticker to prevent small packages
		c.compactTicker.Reset(x.flushInterval)
	}
}

func (x *dkIO) recordPoints(d *feedData) {
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
	if c.readyPoints == 0 {
		log.Debugf("no point to compact on %s", c.category.String())
		return // no points to compact
	}

	defer func() {
		flushVec.WithLabelValues(c.category.String()).Inc()
		queuePtsVec.WithLabelValues(c.category.String()).Sub(float64(c.readyPoints))
		c.readyPoints = 0
	}()

	// compact no index points
	if len(c.arrPoints) > 0 {
		if err := x.doCompact(c.arrPoints, c.category, ""); err != nil {
			log.Warnf("compact %d points on %s failed: %s, ignored", len(c.arrPoints), c.category, err)
		}
		// put back these points.
		datakit.PutbackPoints(c.arrPoints...)

		c.arrPoints = c.arrPoints[:0] // clear
	}

	// compact storage-indexed points
	for k, arr := range c.indexedPoints {
		if len(arr) == 0 {
			log.Debugf("no point, skip storage index %q", k)
			continue
		}

		if err := x.doCompact(arr, c.category, k); err != nil {
			log.Warnf("compact %d points on %s failed: %s, ignored", len(c.arrPoints), c.category, err)
		}

		// put back these points.
		datakit.PutbackPoints(arr...)
		c.indexedPoints[k] = c.indexedPoints[k][:0] // clear: should we delete the key in map?
	}
}

func (x *dkIO) doCompact(points []*point.Point, cat point.Category, indexName string) error {
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

	if indexName != "" {
		opts = append(opts, dataway.WithStorageIndex(indexName))
	}

	return x.dw.Write(opts...)
}

// compactAndUpload build body then upload to dataway directly.
func (x *dkIO) compactAndUpload(points []*point.Point, cat point.Category) error {
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
		dataway.WithNoWAL(true), // send body directly(without WAL)
		dataway.WithGzipDuringBuildBody(true),
	}

	return x.dw.Write(opts...)
}

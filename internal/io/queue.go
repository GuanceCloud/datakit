// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/failcache"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

func (x *dkIO) cacheData(c *consumer, d *iodata, tryClean bool) {
	if d == nil {
		log.Warn("get empty data, ignored")
		return
	}

	defer func() {
		queuePtsVec.WithLabelValues(d.category.String()).Set(float64(len(c.pts)))
	}()

	log.Debugf("get iodata(%d points) from %s|%s", len(d.pts), d.category, d.from)

	if x.fd != nil && x.matchOutputFileInput(d.from) {
		log.Debugf("write %d(%s) points to %s", len(d.pts), d.from, x.outputFile)

		if err := x.fileOutput(d); err != nil {
			log.Errorf("fileOutput: %s", err)
		}

		// do not send data to remote.
		return
	}

	c.pts = append(c.pts, d.pts...)

	if tryClean &&
		x.maxCacheCount > 0 &&
		len(c.pts) > x.maxCacheCount {
		x.flush(c)

		// reset consumer flush ticker to prevent send small packages
		c.flushTiker.Reset(x.flushInterval)
	}
}

func (x *dkIO) flush(c *consumer) {
	c.lastFlush = time.Now()

	defer func() {
		flushVec.WithLabelValues(c.category.String()).Inc()
	}()

	if err := x.doFlush(c.pts, c.category, c.fc); err != nil {
		log.Warnf("post %d points to %s failed: %s, ignored", len(c.pts), c.category, err)
	}

	c.pts = c.pts[:0] // clear
}

func (x *dkIO) flushFailCache(c *consumer) {
	if c.fc == nil {
		return
	}

	if err := x.dw.Write(dataway.WithCacheClean(true),
		dataway.WithCategory(c.category),
		dataway.WithFailCache(c.fc),
	); err != nil {
		log.Warnf("flush cache failed: %s, ignored", err)
	}
}

func (x *dkIO) doFlush(pts []*dkpt.Point, cat point.Category, fc failcache.Cache, dynamicURL ...string) error {
	if x.dw == nil {
		return fmt.Errorf("dataway not set")
	}

	if len(pts) == 0 {
		return nil
	}

	opts := []dataway.WriteOption{
		dataway.WithPoints(pts),
		dataway.WithCategory(cat),
		dataway.WithFailCache(fc),
		dataway.WithCacheAll(x.cacheAll),
	}

	if len(dynamicURL) > 0 {
		opts = append(opts, dataway.WithDynamicURL(dynamicURL[0]))
	}

	return x.dw.Write(opts...)
}

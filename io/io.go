// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package io implements datakits data transfer among inputs.
package io

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sender"
)

var (
	minGZSize   = 1024
	maxKodoPack = 10 * 1024 * 1024
)

var (
	log = logger.DefaultSLogger("io")

	g = datakit.G("io")
)

type IOConfig struct {
	FeedChanSize int `toml:"feed_chan_size"`

	MaxCacheCount        int `toml:"max_cache_count"`
	MaxDynamicCacheCount int `toml:"max_dynamic_cache_count"`

	FlushInterval string `toml:"flush_interval"`

	OutputFile       string   `toml:"output_file"`
	OutputFileInputs []string `toml:"output_file_inputs"`

	EnableCache bool `toml:"enable_cache"`
	CacheSizeGB int  `toml:"cache_max_size_gb"`

	BlockingMode bool `toml:"blocking_mode"`

	Filters map[string][]string `toml:"filters"`
}

type IO struct {
	conf *IOConfig

	dw dataway.DataWay

	chans map[string]chan *iodata

	inLastErr chan *lastError

	inputstats map[string]*InputsStat

	lock sync.RWMutex

	fd *os.File

	flushInterval time.Duration

	droppedTotal int64

	outputFileSize int64
	sender         *sender.Sender
}

func (x *IO) ifMatchOutputFileInput(feedName string) bool {
	for _, v := range x.conf.OutputFileInputs {
		if v == feedName {
			return true
		}
	}
	return false
}

func (x *IO) cacheData(c *consumer, d *iodata, tryClean bool) {
	if d == nil {
		log.Warn("get empty data, ignored")
		return
	}

	log.Debugf("get iodata(%d points) from %s|%s", len(d.pts), d.category, d.from)

	x.updateStats(d)

	if x.conf.OutputFile != "" {
		if err := x.fileOutput(d); err != nil {
			log.Errorf("fileOutput: %s", err)
		}
	}

	if d.opt != nil && d.opt.HTTPHost != "" {
		c.dynamicDatawayPts[d.opt.HTTPHost] = append(c.dynamicDatawayPts[d.opt.HTTPHost], d.pts...)
	} else {
		c.pts = append(c.pts, d.pts...)
	}

	if (tryClean && x.conf.MaxCacheCount > 0 && len(c.pts) > x.conf.MaxCacheCount) ||
		(x.conf.MaxDynamicCacheCount > 0 && len(c.dynamicDatawayPts) > x.conf.MaxDynamicCacheCount) {
		x.flush(c)
	}
}

func (x *IO) StartIO(recoverable bool) {
	g.Go(func(_ context.Context) error {
		StartFilter()
		return nil
	})

	for _, c := range []string{
		datakit.Metric,
		datakit.Network,
		datakit.KeyEvent,
		datakit.Object,
		datakit.CustomObject,
		datakit.Logging,
		datakit.Tracing,
		datakit.RUM,
		datakit.Security,
		datakit.Profile,
		dynamicDatawayCategory,
	} {
		log.Infof("starting consumer on %s...", c)
		func(category string) {
			g.Go(func(_ context.Context) error {
				x.runConsumer(category)
				return nil
			})
		}(c)
	}
}

type consumer struct {
	ch chan *iodata

	sendPts *uint64
	failPts *uint64

	category string

	pts               []*point.Point
	dynamicDatawayPts map[string][]*point.Point
}

func (x *IO) runConsumer(category string) {
	tick := time.NewTicker(x.flushInterval)
	defer tick.Stop()

	ch, ok := x.chans[category]
	if !ok {
		log.Panicf("invalid category %s, should not been here", category)
	}

	c := &consumer{
		ch:                ch,
		category:          category,
		dynamicDatawayPts: map[string][]*point.Point{},
	}

	switch category {
	case datakit.Metric:
		c.sendPts = &MSendPts
		c.failPts = &MFailPts
	case datakit.Network:
		c.sendPts = &NSendPts
		c.failPts = &NFailPts
	case datakit.KeyEvent:
		c.sendPts = &ESendPts
		c.failPts = &EFailPts
	case datakit.Object:
		c.sendPts = &OSendPts
		c.failPts = &OFailPts
	case datakit.CustomObject:
		c.sendPts = &COSendPts
		c.failPts = &COFailPts
	case datakit.Logging:
		c.sendPts = &LSendPts
		c.failPts = &LFailPts
	case datakit.Tracing:
		c.sendPts = &TSendPts
		c.failPts = &TFailPts
	case datakit.RUM:
		c.sendPts = &RSendPts
		c.failPts = &RFailPts
	case datakit.Security:
		c.sendPts = &SSendPts
		c.failPts = &SFailPts
	case datakit.Profile:
		c.sendPts = &PSendPts
		c.failPts = &PFailPts
	case dynamicDatawayCategory:
		c.sendPts = &LSendPts
		c.failPts = &LFailPts

		// NOTE: ????????????????????????????????????????????? dynamicDatawayPts ?????????????????????
		// ????????????????????? category ????????? logging
		c.category = datakit.Logging
	}

	log.Infof("run consumer on %s", category)
	for {
		select {
		case d := <-ch:
			x.cacheData(c, d, true)

		case <-tick.C:
			log.Debugf("try flush pts on %s", c.category)
			x.flush(c)

		case e := <-x.inLastErr:
			x.updateLastErr(e)

		case <-datakit.Exit.Wait():
			log.Infof("io consumer on %s exit on exit", c.category)
			return
		}
	}
}

func (x *IO) updateLastErr(e *lastError) {
	x.lock.Lock()
	defer x.lock.Unlock()

	stat, ok := x.inputstats[e.from]
	if !ok {
		stat = &InputsStat{
			First: time.Now(),
			Last:  time.Now(),
		}
		x.inputstats[e.from] = stat
	}

	stat.LastErr = e.err
	stat.LastErrTS = e.ts
}

func (x *IO) flush(c *consumer) {
	log.Debugf("try flush %d pts on %s", len(c.pts), c.category)

	failed := 0

	if n, err := x.doFlush(c.pts, c.category); err != nil {
		log.Errorf("post %d to %s failed: %s", len(c.pts), c.category, err)
		failed += n
	} else {
		failed += n
	}

	for k, pts := range c.dynamicDatawayPts {
		log.Debugf("try flush dynamic dataway %d pts on %s", len(pts), k)
		if n, err := x.doFlush(pts, c.category); err != nil {
			log.Errorf("post %d to %s failed", len(pts), k)
			failed += n
		} else {
			failed += n
		}
	}

	atomic.AddUint64(c.sendPts, uint64(len(c.pts)+len(c.dynamicDatawayPts)))
	atomic.AddUint64(c.failPts, uint64(failed))

	// clear
	c.pts = nil
	c.dynamicDatawayPts = map[string][]*point.Point{}
}

type body struct {
	buf  []byte
	gzon bool
}

func doBuildBody(pts []*point.Point, outfile string) ([]*body, error) {
	var (
		gz = func(data []byte) (*body, error) {
			body := &body{buf: data}

			if len(data) > minGZSize && outfile == "" {
				if gzbuf, err := datakit.GZip(body.buf); err != nil {
					log.Errorf("Gzip: %s", err.Error())
					return nil, err
				} else {
					log.Debugf("GZip: %d/%d=%f", len(gzbuf), len(body.buf), float64(len(gzbuf))/float64(len(body.buf)))
					body.buf = gzbuf
					body.gzon = true
				}
			}

			return body, nil
		}

		bodies []*body
		lines  = bytes.Buffer{}
	)

	for _, pt := range pts {
		ptstr := pt.Point.String()
		if lines.Len()+len(ptstr)+1 >= maxKodoPack {
			if body, err := gz(lines.Bytes()); err != nil {
				return nil, err
			} else {
				bodies = append(bodies, body)
			}
			lines.Reset()
		}

		if lines.Len() != 0 {
			lines.WriteString("\n")
		}
		lines.WriteString(ptstr)
	}
	if body, err := gz(lines.Bytes()); err != nil {
		return nil, err
	} else {
		return append(bodies, body), nil
	}
}

func (x *IO) buildBody(pts []*point.Point) ([]*body, error) {
	return doBuildBody(pts, x.conf.OutputFile)
}

func (x *IO) doFlush(pts []*point.Point, category string) (int, error) {
	if x.sender == nil {
		return 0, fmt.Errorf("io sender is not initialized")
	}

	if len(pts) == 0 {
		return 0, nil
	}

	return x.sender.Write(category, pts)
}

func (x *IO) fileOutput(d *iodata) error {
	// concurrent write
	x.lock.Lock()
	defer x.lock.Unlock()

	bodies, err := x.buildBody(d.pts)
	if err != nil {
		return err
	}

	for _, body := range bodies {
		if len(x.conf.OutputFileInputs) == 0 || x.ifMatchOutputFileInput(d.from) {
			if _, err := x.fd.Write(append(body.buf, '\n')); err != nil {
				return err
			}

			x.outputFileSize += int64(len(body.buf))
			if x.outputFileSize > 4*1024*1024 {
				if err := x.fd.Truncate(0); err != nil {
					return err
				}
				x.outputFileSize = 0
			}
		}
	}

	return nil
}

func (x *IO) ChanUsage() map[string][2]int {
	usage := map[string][2]int{}
	for k, v := range x.chans {
		usage[k] = [2]int{len(v), cap(v)}
	}
	return usage
}

func (x *IO) CacheSize() string {
	// TODO: return disk failed cache info
	return "TODO"
}

func (x *IO) DroppedTotal() int64 {
	// NOTE: not thread-safe
	return x.droppedTotal
}

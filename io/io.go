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
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sender"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
)

var (
	minGZSize   = 1024
	maxKodoPack = 10 * 1024 * 1024
)

var (
	testAssert                 = false
	highFreqCleanInterval      = time.Millisecond * 500
	datawayListIntervalDefault = 50
	heartBeatIntervalDefault   = 40
	log                        = logger.DefaultSLogger("io")

	g = datakit.G("io")
)

type IOConfig struct {
	FeedChanSize              int                 `toml:"feed_chan_size"`
	HighFreqFeedChanSize      int                 `toml:"high_frequency_feed_chan_size"`
	MaxCacheCount             int64               `toml:"max_cache_count"`
	CacheDumpThreshold        int64               `toml:"cache_dump_threshold"`
	MaxDynamicCacheCount      int64               `toml:"max_dynamic_cache_count"`
	DynamicCacheDumpThreshold int64               `toml:"dynamic_cache_dump_threshold"`
	FlushInterval             string              `toml:"flush_interval"`
	OutputFile                string              `toml:"output_file"`
	OutputFileInputs          []string            `toml:"output_file_inputs"`
	EnableCache               bool                `toml:"enable_cache"`
	Filters                   map[string][]string `toml:"filters"`
}

type Option struct {
	CollectCost time.Duration
	HighFreq    bool
	Version     string
	HTTPHost    string
	PostTimeout time.Duration
	Sample      func(points []*Point) []*Point

	PlScript map[string]string // <measurement>: <script name>
	PlOption *plscript.Option
}

type IO struct {
	conf *IOConfig

	dw dataway.DataWay

	in        chan *iodata
	in2       chan *iodata // high-freq chan
	inLastErr chan *lastError

	SentBytes int

	inputstats map[string]*InputsStat
	lock       sync.RWMutex

	cache        map[string][]*Point
	dynamicCache map[string][]*Point

	fd *os.File

	cacheCnt        int64
	dynamicCacheCnt int64
	droppedTotal    int64
	outputFileSize  int64
	sender          *sender.Sender
}

type IoStat struct {
	SentBytes int `json:"sent_bytes"`
}

type iodata struct {
	category,
	name string
	filtered int
	opt      *Option
	pts      []*Point
}

func TestOutput() {
	testAssert = true
}

func SetTest() {
	testAssert = true
}

//nolint:gocyclo
func (x *IO) DoFeed(pts []*Point, category, name string, opt *Option) error {
	if testAssert {
		return nil
	}

	log.Debugf("io feed %s|%s", name, category)

	ch := x.in
	if opt != nil && opt.HighFreq {
		ch = x.in2
	}

	filtered := 0
	var after []*Point

	switch category {
	case datakit.MetricDeprecated:

	case datakit.Logging,
		datakit.Tracing,
		datakit.Metric,
		datakit.Object,
		datakit.Network,
		datakit.KeyEvent,
		datakit.CustomObject,
		datakit.RUM,
		datakit.Security,
		datakit.Profile:

		// run filters
		after = filterPts(category, pts)
		filtered = len(pts) - len(after)
		pts = after

	default:
		return fmt.Errorf("invalid category `%s'", category)
	}

	// run pipeline
	after, err := runPl(category, pts, opt)
	if err != nil {
		l.Error(err)
	} else {
		pts = after
	}

	// Maybe all points been filtered, but we still send the feeding into io.
	// We can still see some inputs/data are sending to io in monitor. Do not
	// optimize the feeding, or we see nothing on monitor about these filtered
	// points.
	select {
	case ch <- &iodata{
		category: category,
		pts:      pts,
		filtered: filtered,
		name:     name,
		opt:      opt,
	}:
	case <-datakit.Exit.Wait():
		log.Warnf("%s/%s feed skipped on global exit", category, name)
	}

	return nil
}

func (x *IO) ioStop() {
	if x.fd != nil {
		if err := x.fd.Close(); err != nil {
			log.Error(err)
		}
	}
	// stop sender
	if err := x.sender.Stop(); err != nil {
		log.Error(err)
	}
}

func (x *IO) updateStats(d *iodata) {
	x.lock.Lock()
	defer x.lock.Unlock()

	now := time.Now()
	stat, ok := x.inputstats[d.name]

	if !ok {
		stat = &InputsStat{
			Total: int64(len(d.pts)),
			First: now,
		}
		x.inputstats[d.name] = stat
	}

	stat.Total += int64(len(d.pts))
	stat.Count++
	stat.Filtered += int64(d.filtered)
	stat.Last = now
	stat.Category = d.category

	if (stat.Last.Unix() - stat.First.Unix()) > 0 {
		stat.Frequency = fmt.Sprintf("%.02f/min",
			float64(stat.Count)/(float64(stat.Last.Unix()-stat.First.Unix())/60))
	}
	stat.AvgSize = (stat.Total) / stat.Count

	if d.opt != nil {
		stat.Version = d.opt.Version
		stat.totalCost += d.opt.CollectCost
		stat.AvgCollectCost = (stat.totalCost) / time.Duration(stat.Count)
		if d.opt.CollectCost > stat.MaxCollectCost {
			stat.MaxCollectCost = d.opt.CollectCost
		}
	}
}

func (x *IO) ifMatchOutputFileInput(feedName string) bool {
	for _, v := range x.conf.OutputFileInputs {
		if v == feedName {
			return true
		}
	}
	return false
}

func (x *IO) cacheData(d *iodata, tryClean bool) {
	if d == nil {
		log.Warn("get empty data, ignored")
		return
	}

	log.Debugf("get iodata(%d points) from %s|%s", len(d.pts), d.category, d.name)

	x.updateStats(d)

	if d.opt != nil && d.opt.HTTPHost != "" {
		x.dynamicCache[d.opt.HTTPHost] = append(x.dynamicCache[d.opt.HTTPHost], d.pts...)
		x.dynamicCacheCnt += int64(len(d.pts))
	} else {
		x.cache[d.category] = append(x.cache[d.category], d.pts...)
		x.cacheCnt += int64(len(d.pts))
	}

	if x.conf.OutputFile != "" {
		bodies, err := x.buildBody(d.pts)
		if err != nil {
			log.Errorf("build iodata bodies failed: %s", err)
		}
		for _, body := range bodies {
			if len(x.conf.OutputFileInputs) == 0 || x.ifMatchOutputFileInput(d.name) {
				if err := x.fileOutput(body.buf); err != nil {
					log.Error("fileOutput: %s, ignored", err.Error())
				}
			}
		}
	}

	if (tryClean && x.conf.MaxCacheCount > 0 && x.cacheCnt > x.conf.MaxCacheCount) ||
		(x.conf.MaxDynamicCacheCount > 0 && x.dynamicCacheCnt > x.conf.MaxDynamicCacheCount) {
		x.flushAll()
	}
}

func (x *IO) cleanHighFreqIOData() {
	if len(x.in2) > 0 {
		log.Debugf("clean %d cache on high-freq-chan", len(x.in2))
	}

	for {
		select {
		case d := <-x.in2: // eat all cached data
			x.cacheData(d, false)
		default:
			return
		}
	}
}

func (x *IO) init() error {
	if x.conf.OutputFile != "" {
		f, err := os.OpenFile(x.conf.OutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644) //nolint:gosec
		if err != nil {
			log.Error(err)
			return err
		}

		x.fd = f
	}

	return nil
}

func (x *IO) StartIO(recoverable bool) {
	du, err := time.ParseDuration(x.conf.FlushInterval)
	if err != nil {
		l.Warnf("time.ParseDuration: %s, ignored", err)
		du = time.Second * 10
	}

	if sender, err := sender.NewSender(
		&sender.Option{
			Cache:              x.conf.EnableCache,
			FlushCacheInterval: du,
			ErrorCallback: func(err error) {
				FeedEventLog(&DKEvent{Message: err.Error(), Status: "error", Category: "dataway"})
			},
		}); err != nil {
		log.Errorf("init sender error: %s", err.Error())
	} else {
		x.sender = sender
	}

	g.Go(func(_ context.Context) error {
		StartFilter()
		return nil
	})

	g.Go(func(ctx context.Context) error {
		if err := x.init(); err != nil {
			log.Errorf("init io err %v", err)
			return nil
		}

		defer x.ioStop()

		tick := time.NewTicker(du)
		defer tick.Stop()

		highFreqRecvTicker := time.NewTicker(highFreqCleanInterval)
		defer highFreqRecvTicker.Stop()

		heartBeatTick := time.NewTicker(time.Second * time.Duration(heartBeatIntervalDefault))
		defer heartBeatTick.Stop()

		datawaylistTick := time.NewTicker(time.Second * time.Duration(datawayListIntervalDefault))
		defer datawaylistTick.Stop()

		for {
			if x.dw != nil {
				select {
				case <-heartBeatTick.C:
					log.Debugf("### enter heartBeat")
					heartBeatInterval, err := x.dw.HeartBeat()
					if err != nil {
						log.Warnf("dw.HeartBeat: %s, ignored", err.Error())
					}
					if heartBeatInterval != heartBeatIntervalDefault {
						heartBeatTick.Reset(time.Second * time.Duration(heartBeatInterval))
						heartBeatIntervalDefault = heartBeatInterval
					}

				case <-datawaylistTick.C:
					log.Debugf("### enter dataway list")
					var dws []string
					var err error
					var datawayListInterval int
					dws, datawayListInterval, err = x.dw.DatawayList()
					if err != nil {
						log.Warnf("DatawayList(): %s, ignored", err)
					}
					dataway.AvailableDataways = dws
					if datawayListInterval != datawayListIntervalDefault {
						datawaylistTick.Reset(time.Second * time.Duration(datawayListInterval))
						datawayListIntervalDefault = datawayListInterval
					}
				default:
				}
			}

			select {
			case d := <-x.in:
				x.cacheData(d, true)

			case e := <-x.inLastErr:
				x.updateLastErr(e)

			case <-highFreqRecvTicker.C:
				x.cleanHighFreqIOData()

			case <-tick.C:
				x.flushAll()

			case <-datakit.Exit.Wait():
				log.Info("io exit on exit")
				return nil
			}
		}
	})

	log.Info("starting...")
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

func (x *IO) flushAll() {
	x.flush()

	if x.cacheCnt > 0 {
		log.Warnf("post failed cache count: %d", x.cacheCnt)
	}

	// dump cache pts
	if x.conf.CacheDumpThreshold > 0 && x.cacheCnt > x.conf.CacheDumpThreshold {
		log.Warnf("failed cache count reach max limit(%d), cleanning cache...", x.conf.MaxCacheCount)
		for k := range x.cache {
			x.cache[k] = nil
		}
		x.droppedTotal += x.cacheCnt
		x.cacheCnt = 0
	}
	// dump dynamic cache pts
	if x.conf.DynamicCacheDumpThreshold > 0 && x.dynamicCacheCnt > x.conf.DynamicCacheDumpThreshold {
		log.Warnf("failed dynamicCache count reach max limit(%d), cleanning cache...", x.conf.MaxDynamicCacheCount)
		for k := range x.dynamicCache {
			x.dynamicCache[k] = nil
		}
		x.droppedTotal += x.dynamicCacheCnt
		x.dynamicCacheCnt = 0
	}
}

func (x *IO) flush() {
	for k, v := range x.cache {
		if err := x.doFlush(v, k); err != nil {
			log.Errorf("post %d to %s failed", len(v), k)
			continue
		}

		if len(v) > 0 {
			x.cacheCnt -= int64(len(v))
			log.Debugf("clean %d cache on %s, remain: %d", len(v), k, x.cacheCnt)
			x.cache[k] = nil
		}
	}

	// flush dynamic cache: __not__ post to default dataway
	for k, v := range x.dynamicCache {
		if err := x.doFlush(v, k); err != nil {
			log.Errorf("post %d to %s failed", len(v), k)
			continue
		}

		if len(v) > 0 {
			x.dynamicCacheCnt -= int64(len(v))
			log.Debugf("clean %d dynamicCache on %s, remain: %d", len(v), k, x.dynamicCacheCnt)
			x.dynamicCache[k] = nil
		}
	}
}

type body struct {
	buf  []byte
	gzon bool
}

func doBuildBody(pts []*Point, outfile string) ([]*body, error) {
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

func (x *IO) buildBody(pts []*Point) ([]*body, error) {
	return doBuildBody(pts, x.conf.OutputFile)
}

func (x *IO) doFlush(pts []*Point, category string) error {
	if x.sender == nil {
		return fmt.Errorf("io sender is not initialized")
	}

	points := []sinkcommon.ISinkPoint{}

	for _, pt := range pts {
		points = append(points, pt)
	}

	return x.sender.Write(category, points)
}

func (x *IO) fileOutput(body []byte) error {
	if _, err := x.fd.Write(append(body, '\n')); err != nil {
		return err
	}

	x.outputFileSize += int64(len(body))
	if x.outputFileSize > 4*1024*1024 {
		if err := x.fd.Truncate(0); err != nil {
			return err
		}
		x.outputFileSize = 0
	}

	return nil
}

func (x *IO) DroppedTotal() int64 {
	// NOTE: not thread-safe
	return x.droppedTotal
}

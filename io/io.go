package io

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

const (
	minGZSize   = 1024
	maxKodoPack = 10 * 1000 * 1000
)

var (
	testAssert            = false
	highFreqCleanInterval = time.Millisecond * 500
	l                     = logger.DefaultSLogger("io")

	DisableLogFilter   bool
	DisableHeartbeat   bool
	DisableDatawayList bool
)

type Option struct {
	CollectCost time.Duration
	HighFreq    bool

	HTTPHost    string
	PostTimeout time.Duration
	Sample      func(points []*Point) []*Point
}

type lastErr struct {
	from, err string
	ts        time.Time
}

type IO struct {
	FeedChanSize              int
	HighFreqFeedChanSize      int
	MaxCacheCount             int64
	CacheDumpThreshold        int64
	MaxDynamicCacheCount      int64
	DynamicCacheDumpThreshold int64
	FlushInterval             time.Duration
	OutputFile                string

	dw *dataway.DataWayCfg

	in        chan *iodata
	in2       chan *iodata // high-freq chan
	inLastErr chan *lastErr

	lastBodyBytes int
	SentBytes     int

	inputstats map[string]*InputsStat
	qstatsCh   chan *qinputStats

	cache        map[string][]*Point
	dynamicCache map[string][]*Point

	cacheCnt        int64
	dynamicCacheCnt int64
	droppedTotal    int64
	fd              *os.File
	outputFileSize  int64
}

type IoStat struct {
	SentBytes int `json:"sent_bytes"`
}

func NewIO() *IO {
	x := &IO{
		FeedChanSize:         1024,
		HighFreqFeedChanSize: 2048,
		MaxCacheCount:        1024,
		MaxDynamicCacheCount: 1024,
		FlushInterval:        10 * time.Second,
		in:                   make(chan *iodata, 128),
		in2:                  make(chan *iodata, 128*8),
		inLastErr:            make(chan *lastErr, 128),

		inputstats: map[string]*InputsStat{},
		qstatsCh:   make(chan *qinputStats), // blocking

		cache:        map[string][]*Point{},
		dynamicCache: map[string][]*Point{},
	}

	l.Debugf("IO: %+#v", x)

	return x
}

type iodata struct {
	category, name string
	opt            *Option
	pts            []*Point
	url            string
	isProxy        bool
}

func TestOutput() {
	testAssert = true
}

func SetTest() {
	testAssert = true
}

func (x *IO) DoFeed(pts []*Point, category, name string, opt *Option) error {
	if testAssert {
		return nil
	}

	ch := x.in
	if opt != nil && opt.HighFreq {
		ch = x.in2
	}

	switch category {
	case datakit.MetricDeprecated:
	case datakit.Metric:
	case datakit.Network:
	case datakit.KeyEvent:
	case datakit.Object:
	case datakit.CustomObject:
	case datakit.Logging:
		if x.dw.ClientsCount() == 1 {
			if !DisableLogFilter {
				pts = defLogfilter.filter(pts)
			}
		} else {
			// TODO: add multiple dataway config support
			l.Infof("multiple dataway config %d for log filter not support yet", x.dw.ClientsCount())
		}
	case datakit.Tracing:
	case datakit.Security:
	case datakit.Rum:
	default:
		return fmt.Errorf("invalid category `%s'", category)
	}

	l.Debugf("io feed %s", name)

	select {
	case ch <- &iodata{
		category: category,
		pts:      pts,
		name:     name,
		opt:      opt,
	}:
	case <-datakit.Exit.Wait():
		l.Warnf("%s/%s feed skipped on global exit", category, name)
	}

	return nil
}

func (x *IO) ioStop() {
	if x.fd != nil {
		if err := x.fd.Close(); err != nil {
			l.Error(err)
		}
	}
}

func (x *IO) updateLastErr(e *lastErr) {
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

func (x *IO) updateStats(d *iodata) {
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
	stat.Last = now
	stat.Category = d.category

	if (stat.Last.Unix() - stat.First.Unix()) > 0 {
		stat.Frequency = fmt.Sprintf("%.02f/min",
			float64(stat.Count)/(float64(stat.Last.Unix()-stat.First.Unix())/60))
	}
	stat.AvgSize = (stat.Total) / stat.Count

	if d.opt != nil {
		stat.totalCost += d.opt.CollectCost
		stat.AvgCollectCost = (stat.totalCost) / time.Duration(stat.Count)
		if d.opt.CollectCost > stat.MaxCollectCost {
			stat.MaxCollectCost = d.opt.CollectCost
		}
	}
}

func (x *IO) cacheData(d *iodata, tryClean bool) {
	if d == nil {
		l.Warn("get empty data, ignored")
		return
	}

	l.Debugf("get iodata(%d points) from %s|%s", len(d.pts), d.category, d.name)

	x.updateStats(d)

	if d.opt != nil && d.opt.HTTPHost != "" {
		x.dynamicCache[d.opt.HTTPHost] = append(x.dynamicCache[d.opt.HTTPHost], d.pts...)
		x.dynamicCacheCnt += int64(len(d.pts))
	} else {
		x.cache[d.category] = append(x.cache[d.category], d.pts...)
		x.cacheCnt += int64(len(d.pts))
	}

	if (tryClean && x.MaxCacheCount > 0 && x.cacheCnt > x.MaxCacheCount) || (x.MaxDynamicCacheCount > 0 && x.dynamicCacheCnt > x.MaxDynamicCacheCount) {
		x.flushAll()
	}
}

func (x *IO) cleanHighFreqIOData() {
	if len(x.in2) > 0 {
		l.Debugf("clean %d cache on high-freq-chan", len(x.in2))
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
	if x.OutputFile != "" {
		f, err := os.OpenFile(x.OutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			l.Error(err)
			return err
		}

		x.fd = f
	}

	return nil
}

func (x *IO) StartIO(recoverable bool) {
	g := datakit.G("io")
	g.Go(func(ctx context.Context) error {
		if err := x.init(); err != nil {
			l.Errorf("init io err %v", err)
			return nil
		}

		defer x.ioStop()

		tick := time.NewTicker(x.FlushInterval)
		defer tick.Stop()

		highFreqRecvTicker := time.NewTicker(highFreqCleanInterval)
		defer highFreqRecvTicker.Stop()

		heartBeatTick := time.NewTicker(time.Second * 30)
		defer heartBeatTick.Stop()

		datawaylistTick := time.NewTicker(time.Minute)
		defer datawaylistTick.Stop()

		for {
			select {
			case d := <-x.in:
				x.cacheData(d, true)

			case e := <-x.inLastErr:
				x.updateLastErr(e)

			case q := <-x.qstatsCh:

				res := dumpStats(x.inputstats)
				select {
				case <-q.ch:
					l.Warnf("qid(%s) client canceled, ignored", q.qid)
				case q.ch <- res: // XXX: reference
					l.Debugf("qid(%s) response ok", q.qid)
				}

			case <-highFreqRecvTicker.C:
				x.cleanHighFreqIOData()

			case <-heartBeatTick.C:
				if !DisableHeartbeat {
					x.dw.HeartBeat()
				}

			case <-datawaylistTick.C:
				if !DisableDatawayList {
					dws, err := x.dw.DatawayList()
					if err != nil {
						l.Warnf("DatawayList(): %s, ignored", err)
					}
					dataway.AvailableDataways = dws
				}

			case <-tick.C:
				x.flushAll()

			case <-datakit.Exit.Wait():
				l.Info("io exit on exit")
				return nil
			}
		}
	})

	// start log filter
	if !DisableLogFilter {
		defLogfilter.start()
	}

	l.Info("starting...")
}

func (x *IO) flushAll() {
	x.flush()

	if x.cacheCnt > 0 {
		l.Warnf("post failed cache count: %d", x.cacheCnt)
	}

	// dump cache pts
	if x.CacheDumpThreshold > 0 && x.cacheCnt > x.CacheDumpThreshold {
		l.Warnf("failed cache count reach max limit(%d), cleanning cache...", x.MaxCacheCount)
		for k := range x.cache {
			x.cache[k] = nil
		}
		atomic.AddInt64(&x.droppedTotal, x.cacheCnt)
		x.cacheCnt = 0
	}
	// dump dynamic cache pts
	if x.DynamicCacheDumpThreshold > 0 && x.dynamicCacheCnt > x.DynamicCacheDumpThreshold {
		l.Warnf("failed dynamicCache count reach max limit(%d), cleanning cache...", x.MaxDynamicCacheCount)
		for k := range x.dynamicCache {
			x.dynamicCache[k] = nil
		}
		atomic.AddInt64(&x.droppedTotal, x.dynamicCacheCnt)
		x.dynamicCacheCnt = 0
	}
}

func (x *IO) flush() {
	for k, v := range x.cache {
		if err := x.doFlush(v, k); err != nil {
			l.Errorf("post %d to %s failed", len(v), k)
			continue
		}

		if len(v) > 0 {
			x.cacheCnt -= int64(len(v))
			l.Debugf("clean %d cache on %s, remain: %d", len(v), k, x.cacheCnt)
			x.cache[k] = nil
		}
	}

	// flush dynamic cache: __not__ post to default dataway
	for k, v := range x.dynamicCache {
		if err := x.doFlush(v, k); err != nil {
			l.Errorf("post %d to %s failed", len(v), k)
			// clear data
			x.dynamicCache[k] = nil
			continue
		}

		if len(v) > 0 {
			x.dynamicCacheCnt -= int64(len(v))
			l.Debugf("clean %d dynamicCache on %s, remain: %d", len(v), k, x.dynamicCacheCnt)
			x.dynamicCache[k] = nil
		}
	}
}

type body struct {
	buf  []byte
	gzon bool
}

var lines = bytes.Buffer{}

func (x *IO) buildBody(pts []*Point) ([]*body, error) {
	var (
		gz = func(lines []byte) (*body, error) {
			var (
				body = &body{buf: lines}
				err  error
			)
			l.Debugf("### io body size before GZ: %dM %dK", len(body.buf)/1000/1000, len(body.buf)/1000)
			if len(lines) > minGZSize && x.OutputFile == "" {
				if body.buf, err = datakit.GZip(body.buf); err != nil {
					l.Errorf("gz: %s", err.Error())

					return nil, err
				}
				body.gzon = true
			}

			return body, nil
		}
		// lines  bytes.Buffer
		bodies []*body
	)
	lines.Reset()
	for _, pt := range pts {
		ptstr := pt.String()
		if lines.Len()+len(ptstr)+1 >= maxKodoPack {
			if body, err := gz(lines.Bytes()); err != nil {
				return nil, err
			} else {
				bodies = append(bodies, body)
			}
			lines.Reset()
		}
		lines.WriteString(ptstr)
		lines.WriteString("\n")
	}
	if body, err := gz(lines.Bytes()); err != nil {
		return nil, err
	} else {
		return append(bodies, body), nil
	}
}

func (x *IO) doFlush(pts []*Point, category string) error {
	if testAssert {
		return nil
	}

	if pts == nil {
		return nil
	}

	bodies, err := x.buildBody(pts)
	if err != nil {
		return err
	}
	for _, body := range bodies {
		if x.OutputFile != "" {
			x.fileOutput(body.buf)
			continue
		}

		if err := x.dw.Send(category, body.buf, body.gzon); err != nil {
			return err
		}
		x.SentBytes += x.lastBodyBytes
		x.lastBodyBytes = 0
	}

	return nil
}

func (x *IO) fileOutput(body []byte) error {
	if _, err := x.fd.Write(append(body, '\n')); err != nil {
		l.Error(err)
		return err
	}

	x.outputFileSize += int64(len(body))
	if x.outputFileSize > 4*1024*1024 {
		if err := x.fd.Truncate(0); err != nil {
			l.Error(err)
			return err
		}
		x.outputFileSize = 0
	}

	return nil
}

func (x *IO) DroppedTotal() int64 {
	return atomic.LoadInt64(&x.droppedTotal)
}

// Package sender manages io data storage and data cache when failed.
package sender

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cache"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
	pb "google.golang.org/protobuf/proto"
)

type Point struct {
	*influxdb.Point
}

var _ sinkcommon.ISinkPoint = new(Point)

func (p *Point) ToPoint() *influxdb.Point {
	return p.Point
}

func WrapPoint(pts []*influxdb.Point) (x []sinkcommon.ISinkPoint) {
	for _, pt := range pts {
		x = append(x, &Point{pt})
	}
	return
}

var (
	cacheBucket = "io_upload_metric"
	l           = logger.DefaultSLogger("sender")
)

type WriteFunc func(string, []sinkcommon.ISinkPoint) error

type Option struct {
	Cache              bool
	CacheDir           string
	ErrorCallback      func(error)
	FlushCacheInterval time.Duration
	Write              WriteFunc
}

type senderStat struct {
	FailCount    int64
	SuccessCount int64
}

type Sender struct {
	Stat   senderStat
	opt    *Option
	write  WriteFunc
	group  *goroutine.Group
	stopCh chan interface{}
}

// Write receive input data and then call worker to save the data.
func (s *Sender) Write(category string, pts []sinkcommon.ISinkPoint) error {
	if len(pts) == 0 {
		return nil
	}

	if err := s.worker(category, pts); err != nil {
		return err
	}

	return nil
}

// worker create a groutine to run write job.
func (s *Sender) worker(category string, pts []sinkcommon.ISinkPoint) error {
	if s.group == nil {
		return fmt.Errorf("sender is not initialized correctly, missing worker group")
	}

	s.group.Go(func(ctx context.Context) error {
		if err := s.write(category, pts); err != nil {
			atomic.AddInt64(&s.Stat.FailCount, 1)
			l.Error("sink write error: ", err)

			if s.opt.ErrorCallback != nil {
				s.opt.ErrorCallback(err)
			}

			if s.opt.Cache {
				err := s.cache(category, pts)
				if err == nil {
					l.Debugf("sink write cached: %s(%d)", category, len(pts))
				}
			}
		} else {
			l.Debugf("sink write ok: %s(%d)", category, len(pts))
			atomic.AddInt64(&s.Stat.SuccessCount, 1)
		}

		return nil
	})

	return nil
}

// Wait waits all worker to stop.
func (s *Sender) Wait() error {
	return s.group.Wait()
}

// cache save points to cache.
func (s *Sender) cache(category string, pts []sinkcommon.ISinkPoint) error {
	if len(pts) == 0 {
		return nil
	}

	ptList := []string{}
	for _, pt := range pts {
		ptList = append(ptList, pt.String())
	}

	ptStr := strings.Join(ptList, "\n")

	id := cliutils.XID("cache_")

	data := PBData{
		Category: category,
		Lines:    []byte(ptStr),
	}

	dataBuffer, err := pb.Marshal(&data)
	if err != nil {
		l.Warnf("marshal data error: %s", err.Error())
		return err
	}

	if err := cache.Put(cacheBucket, []byte(id), dataBuffer); err != nil {
		l.Warnf("cache data error: %s", err.Error())
		return err
	}

	return nil
}

// init setup sender instance.
func (s *Sender) init(opt *Option) error {
	s.stopCh = make(chan interface{})

	if opt != nil {
		s.opt = opt
	} else {
		s.opt = &Option{}
	}

	if s.group == nil {
		s.group = datakit.G("sender")
	}

	if s.opt.Write != nil {
		s.write = s.opt.Write
	} else {
		s.write = sink.Write
	}

	if s.write == nil {
		return fmt.Errorf("sender init error: write method is required")
	}

	if s.opt.Cache {
		cacheDir := datakit.CacheDir
		if len(s.opt.CacheDir) != 0 {
			cacheDir = s.opt.CacheDir
		}
		s.initCache(cacheDir)
		s.group.Go(func(ctx context.Context) error {
			s.startFlushCache()
			return nil
		})
	}

	return nil
}

// initCache init cache instance.
func (s *Sender) initCache(cacheDir string) {
	if err := cache.Initialize(cacheDir, nil); err != nil {
		l.Warnf("initialized cache: %s, ignored", err.Error())
	} else { //nolint
		if err := cache.CreateBucketIfNotExists(cacheBucket); err != nil {
			l.Warnf("create bucket: %s", err.Error())
		}
	}
}

// startFlushCache start flush cache loop.
func (s *Sender) startFlushCache() {
	interval := s.opt.FlushCacheInterval
	if interval == 0 {
		interval = 10 * time.Second
	}

	tick := time.NewTicker(interval)

	for {
		select {
		case <-tick.C:
			s.flushCache()
		case <-datakit.Exit.Wait():
			l.Warnf("flush cache exit on global exit")
			return
		case <-s.stopCh:
			l.Warn("flush cache stop")
			return
		}
	}
}

// flushCache flush cache when necessary.
func (s *Sender) flushCache() {
	l.Debugf("flush cache start")

	cacheInfo, err := cache.GetInfo()
	if err != nil {
		l.Warnf("get cache info error: %s", err.Error())
		return
	}

	if cacheInfo.CacheCount <= cacheInfo.FlushedCount {
		l.Debug("cache count is less than flushed count, no need to flush")
		return
	}

	const clean = false

	toCleanKeys := [][]byte{}

	l.Debugf("cache info before: %s", cache.Info())
	fn := func(k, v []byte) error {
		time.Sleep(100 * time.Millisecond)
		d := PBData{}
		if err := pb.Unmarshal(v, &d); err != nil {
			return err
		}
		pts, err := lp.ParsePoints(d.Lines, nil)
		if err != nil {
			l.Warnf("parse cache points error : %s", err.Error())
		}
		points := WrapPoint(pts)
		err = s.write(d.Category, points)
		if err != nil {
			l.Warnf("cache sink write error: %s", err.Error())
		} else {
			toCleanKeys = append(toCleanKeys, k)
		}

		return err
	}

	if err := cache.ForEach(cacheBucket, fn, clean); err != nil {
		l.Warnf("upload cache: %s, ignore", err.Error())
	}

	if len(toCleanKeys) > 0 {
		if err := cache.Del(cacheBucket, toCleanKeys); err != nil {
			l.Warnf("cache upload ok , but clean cache failed: ", err.Error())
		}
	}

	l.Debugf("cache info after: %s", cache.Info())
}

// Stop stop cache interval and stop cache.
func (s *Sender) Stop() error {
	if s.stopCh != nil {
		close(s.stopCh)
	}

	return cache.Stop()
}

// NewSender init sender with sinker instance and custom opt.
func NewSender(opt *Option) (*Sender, error) {
	l = logger.SLogger("sender")
	s := &Sender{
		group: datakit.G("sender"),
	}
	err := s.init(opt)
	return s, err
}

package io

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cache"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	pb "google.golang.org/protobuf/proto"
)

var cacheBucket = "io_upload_metric"

type SenderOption struct {
	Cache              bool
	CacheDir           string
	FlushCacheInterval time.Duration
}

type senderStat struct {
	failCount    int64
	successCount int64
}

type Sender struct {
	WriteFunc func(string, []*Point) error
	Stat      senderStat
	opt       *SenderOption
	group     *goroutine.Group
	stopCh    chan interface{}
}

func (s *Sender) Write(category string, pts []*Point) error {
	if len(pts) == 0 {
		return nil
	}

	if s.WriteFunc == nil {
		return fmt.Errorf("missing sinker instance")
	}

	if err := s.worker(category, pts); err != nil {
		return err
	}

	return nil
}

func (s *Sender) worker(category string, pts []*Point) error {
	if s.group == nil {
		return fmt.Errorf("sender is not initialized correctly, missing worker group")
	}

	s.group.Go(func(ctx context.Context) error {
		if err := s.WriteFunc(category, pts); err != nil {
			atomic.AddInt64(&s.Stat.failCount, 1)
			l.Error("sink write error", err)

			if s.opt.Cache {
				err := s.cache(category, pts)
				if err == nil {
					l.Debugf("sink write cached: %s(%d)", category, len(pts))
				}
			}
		} else {
			l.Debugf("sink write ok: %s(%d)", category, len(pts))
			atomic.AddInt64(&s.Stat.successCount, 1)
		}

		return nil
	})

	return nil
}

func (s *Sender) Wait() error {
	return s.group.Wait()
}

// cache save points to cache.
func (s *Sender) cache(category string, pts []*Point) error {
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

func (s *Sender) init(opt *SenderOption) {
	s.stopCh = make(chan interface{})

	if opt != nil {
		s.opt = opt
	} else {
		s.opt = &SenderOption{}
	}

	if s.group == nil {
		s.group = datakit.G("sender")
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
}

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

func (s *Sender) flushCache() {
	l.Debugf("flush cache")
	const clean = false

	fn := func(k, v []byte) error {
		d := PBData{}
		if err := pb.Unmarshal(v, &d); err != nil {
			return err
		}
		pts, err := lp.ParsePoints(d.Lines, nil)
		if err != nil {
			l.Warnf("parse cache points error : %s", err.Error())
		}

		points := WrapPoint(pts)

		err = s.WriteFunc(d.Category, points)
		if err != nil {
			l.Warnf("cache sink write error: %s", err.Error())
		} else {
			if err := cache.Del(cacheBucket, k); err != nil {
				l.Warnf("cache send ok, but delete cache error: %s", string(k))
			}
		}

		return err
	}

	if err := cache.ForEach(cacheBucket, fn, clean); err != nil {
		l.Warnf("upload cache: %s, ignore", err)
	}

	l.Debug(cache.Info())
}

// Stop stop cache interval and stop cache.
func (s *Sender) Stop() error {
	if s.stopCh != nil {
		close(s.stopCh)
	}

	return cache.Stop()
}

// NewSender init sender with sinker instance and custom opt.
func NewSender(writeFunc func(string, []*Point) error, opt *SenderOption) *Sender {
	l = logger.SLogger("sender")
	s := &Sender{
		WriteFunc: writeFunc,
		group:     datakit.G("sender"),
	}
	s.init(opt)
	return s
}

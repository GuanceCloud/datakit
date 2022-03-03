// Package sender mainly save io data.
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
	pb "google.golang.org/protobuf/proto"
)

var l = logger.DefaultSLogger("sender")

var cacheBucket = "io_upload_metric"

type Option struct {
	Cache              bool
	CacheDir           string
	FlushCacheInterval time.Duration
}

type stat struct {
	failCount    int64
	successCount int64
}

type Sender struct {
	Stat           stat
	opt            *Option
	sinkerInstance *Sinker
	group          *goroutine.Group
}

func (s *Sender) Write(category string, pts []*influxdb.Point) error {
	if s.sinkerInstance == nil {
		return fmt.Errorf("missing sinker instance")
	}

	if err := s.worker(category, pts); err != nil {
		return err
	}

	return nil
}

func (s *Sender) worker(category string, pts []*influxdb.Point) error {
	if s.group == nil {
		return fmt.Errorf("sender is not initialized correctly, missing worker group")
	}

	s.group.Go(func(ctx context.Context) error {
		if err := s.sinkerInstance.Write(category, pts); err != nil {
			atomic.AddInt64(&s.Stat.failCount, 1)
			l.Error("sender...", err)

			if s.opt.Cache {
				s.cache(category, pts)
			}

		} else {
			l.Debug("sender....", category)
			atomic.AddInt64(&s.Stat.successCount, 1)
		}

		return nil
	})

	return nil
}

func (s *Sender) Wait() error {
	return s.group.Wait()
}

func (s *Sender) cache(category string, pts []*influxdb.Point) {
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
	}

	if err := cache.Put(cacheBucket, []byte(id), dataBuffer); err != nil {
		l.Warnf("cache data error: %s", err.Error())
	}
}

func (s *Sender) init(opt *Option) {
	if opt != nil {
		s.opt = opt
	} else {
		s.opt = &Option{}
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
		l.Warn("initialized cache: %s, ignored", err)
	} else { //nolint
		if err := cache.CreateBucketIfNotExists(cacheBucket); err != nil {
			l.Warn("create bucket: %s", err)
		}
	}
}

func (s *Sender) startFlushCache() {
	interval := s.opt.FlushCacheInterval
	if interval == 0 {
		interval = time.Duration(10 * time.Second)
	}

	tick := time.NewTicker(interval)

	for {
		select {
		case <-tick.C:
			s.flushCache()
		case <-datakit.Exit.Wait():
			l.Warnf("flush cache exit")
		}
	}

}

func (s *Sender) flushCache() {
	l.Debugf("flush cache")
	const clean = true

	fn := func(k, v []byte) error {
		d := PBData{}
		if err := pb.Unmarshal(v, &d); err != nil {
			return err
		}
		pts, err := lp.ParsePoints(d.Lines, nil)
		if err != nil {
			l.Warnf("parse cace points error : %s", err.Error())
		}

		err = s.sinkerInstance.Write(d.Category, pts)
		if err != nil {
			l.Warnf("cache sink write error: %s", err.Error())
		}
		return err
	}

	if err := cache.ForEach(cacheBucket, fn, clean); err != nil {
		l.Warnf("upload cache: %s, ignore", err)
	}
}

func NewSender(sinker *Sinker, opt *Option) *Sender {
	l = logger.SLogger("sender")
	s := &Sender{
		sinkerInstance: sinker,
		group:          datakit.G("sender"),
	}
	s.init(opt)
	return s
}

// func getDataType(category string) (dataType string) {
// 	switch category {
// 	case datakit.MetricDeprecated:
// 		dataType = "metric"
// 	case datakit.Metric:
// 		dataType = "metric"
// 	case datakit.Network:
// 		dataType = "network"
// 	case datakit.KeyEvent:
// 		dataType = "keyevent"
// 	case datakit.Object:
// 		dataType = "object"
// 	case datakit.CustomObject:
// 		dataType = "custom_object"
// 	case datakit.Logging:
// 		dataType = "logging"
// 	case datakit.Tracing:
// 		dataType = "tracing"
// 	case datakit.Security:
// 		dataType = "security"
// 	case datakit.RUM:
// 		dataType = "rum"
// 	default:
// 		if _, err := url.ParseRequestURI(category); err == nil {
// 			dataType = "dataway"
// 		}
// 	}
// 	return
// }

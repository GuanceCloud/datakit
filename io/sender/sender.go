// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sender manages io data storage and data cache when failed.
package sender

import (
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
	pb "google.golang.org/protobuf/proto"
)

type Writer interface {
	Write(string, []*point.Point) (*sinkcommon.Failed, error)
}

const GB = uint64(1024 * 1024 * 1024)

var l = logger.DefaultSLogger("sender")

type WriteFunc func(string, []*point.Point) (*sinkcommon.Failed, error)

type Option struct {
	Cache              bool
	CacheSizeGB        int
	CacheDir           string
	ErrorCallback      func(error)
	FlushCacheInterval time.Duration
	Write              WriteFunc
}

type Sender struct {
	opt   *Option
	write WriteFunc
	fc    *failCache
}

// Write receive input data and then call worker to save the data.
func (s *Sender) Write(category string, pts []*point.Point) (int, error) {
	if len(pts) == 0 {
		l.Debugf("empty job on %s", category)
		return 0, nil
	}

	l.Debugf("sending %s(%d pts)...", category, len(pts))

	if failed, err := s.write(category, pts); err != nil {
		l.Error("sender write error: ", err)

		if s.opt.ErrorCallback != nil {
			s.opt.ErrorCallback(err)
		}

		return len(pts), err
	} else {
		var failedPts []*point.Point
		if failed != nil {
			failedPts = selectPts(pts, failed)
		}

		if len(failedPts) > 0 && s.opt.Cache && false { // disable cache
			switch category {
			case datakit.Metric, datakit.MetricDeprecated:
				l.Warnf("drop %d pts on %s, not cached", len(pts), category)

			default:
				l.Infof("caching %s(%d pts)...", category, len(pts))
				if err := s.cache(category, failedPts); err != nil {
					l.Errorf("caching %s(%d pts) failed", category, len(pts))
				}
			}
		}

		l.Debugf("sink write %s(%d) done", category, len(pts))

		return len(failedPts), nil
	}
}

func selectPts(pts []*point.Point, f *sinkcommon.Failed) []*point.Point {
	size := len(pts)

	var failed []*point.Point
	if f.Ranges != nil {
		for _, rng := range f.Ranges {
			if rng[0] >= size || rng[1] > size {
				l.Warnf("invalid range [%d:%d], exceed size: %d", rng[0], rng[1], size)
				continue
			}

			failed = append(failed, pts[rng[0]:rng[1]]...)
		}
	}

	if f.Indexes != nil {
		for _, idx := range f.Indexes {
			if idx >= size {
				l.Warnf("invalid index %d, exceed size: %d", idx, size)
				continue
			}
			failed = append(failed, pts[idx])
		}
	}

	l.Debugf("select %d pts from %d pts according to failed job %+#v", len(failed), len(pts), f)
	return failed
}

// init setup sender instance.
func (s *Sender) init(opt *Option) error {
	if opt != nil {
		s.opt = opt
	} else {
		s.opt = &Option{}
	}

	if s.opt.Write != nil {
		s.write = s.opt.Write
	} else {
		s.write = sink.Write
	}

	if s.write == nil {
		return fmt.Errorf("sender init error: write method is required")
	}

	if opt.Cache {
		cacheDir := datakit.CacheDir
		if len(s.opt.CacheDir) != 0 {
			cacheDir = s.opt.CacheDir
		}

		fc, err := initFailCache(cacheDir, int64(s.opt.CacheSizeGB*1024*1024*1024))
		if err != nil {
			return err
		}
		s.fc = fc
	}

	return nil
}

//nolint: unused
func clean(data []byte) error {
	// TODO
	return nil
}

//nolint: unused
func (s *Sender) cleanCache() error {
	return s.fc.get(clean)
}

func (s *Sender) cache(category string, pts []*point.Point) error {
	if len(pts) == 0 {
		return nil
	}

	arr := []string{}
	for _, pt := range pts {
		arr = append(arr, pt.String())
	}

	l.Debugf("get %s cache: %d pts", category, len(pts))

	buf, err := pb.Marshal(&PBData{
		Category: category,
		Lines:    []byte(strings.Join(arr, "\n")),
	})
	if err != nil {
		l.Warnf("dump %s cache(%d) failed: %s(%d pts)", category, len(pts), err)
		return err
	}

	if err := s.fc.put(buf); err != nil {
		l.Warnf("dump %s cache(%d) failed: %s(%d pts)", category, len(pts), err)
		return err
	}

	l.Debugf("put %s cache ok, %d pts", category, len(pts))
	return nil
}

// NewSender init sender with sinker instance and custom opt.
func NewSender(opt *Option) (*Sender, error) {
	l = logger.SLogger("sender")

	s := &Sender{}
	err := s.init(opt)

	return s, err
}

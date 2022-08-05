// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sender manages io data storage and data cache when failed.
package sender

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink"
)

type Writer interface {
	Write(string, []*point.Point) (*point.Failed, error)
}

const GB = uint64(1024 * 1024 * 1024)

var l = logger.DefaultSLogger("sender")

type WriteFunc func(string, []*point.Point) (*point.Failed, error)

type Option struct {
	Cache              bool
	ErrorCallback      func(error)
	FlushCacheInterval time.Duration
	Write              WriteFunc
}

type Sender struct {
	opt   *Option
	write WriteFunc
}

// Write receive input data and then call worker to save the data.
func (s *Sender) Write(category string, pts []*point.Point) ([]*point.Point, error) {
	if len(pts) == 0 {
		l.Debugf("empty job on %s", category)
		return nil, nil
	}

	l.Debugf("sending %s(%d pts)...", category, len(pts))

	failed, err := s.write(category, pts)
	if err != nil {
		// When error happens, must be data problem nor network, so just write log.
		if s.opt.ErrorCallback != nil {
			s.opt.ErrorCallback(err)
		}
		return pts, err
	}

	var failedPts []*point.Point
	if failed != nil {
		failedPts = selectPts(pts, failed)
	}
	return failedPts, nil
}

func selectPts(pts []*point.Point, f *point.Failed) []*point.Point {
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

	return nil
}

// NewSender init sender with sinker instance and custom opt.
func NewSender(opt *Option) (*Sender, error) {
	l = logger.SLogger("sender")

	s := &Sender{}
	err := s.init(opt)

	return s, err
}

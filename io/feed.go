// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
)

var ErrIOBusy = errors.New("io busy")

type Option struct {
	CollectCost time.Duration

	// HighFreq deprecated
	// 之前的高频通道是避免部分不老实的采集器每次只 feed 少数几个点
	// 容易导致 chan feed 效率低，所以选择惩罚性的定期才去消费该
	// 高频通道。
	// 目前基本不太会有这种采集器了，而且 io 本身也不会再有阻塞操作，
	// 故移除了高频通道。
	HighFreq bool

	Version  string
	HTTPHost string

	PostTimeout time.Duration
	Sample      func(points []*point.Point) []*point.Point

	Blocking bool

	PlScript map[string]string // <measurement>: <script name>
	PlOption *plscript.Option
}

type iodata struct {
	category,
	from string
	filtered int
	opt      *Option
	pts      []*point.Point
}

func Feed(name, category string, pts []*point.Point, opt *Option) error {
	if len(pts) == 0 {
		return nil
	}

	return defaultIO.DoFeed(pts, category, name, opt)
}

type lastError struct {
	from, err string
	ts        time.Time
}

// FeedLastError feed some error message(*unblocking*) to inputs stats
// we can see the error in monitor.
// NOTE: the error may be skipped if there is too many error.
func FeedLastError(inputName string, err string) {
	select {
	case defaultIO.inLastErr <- &lastError{
		from: inputName,
		err:  err,
		ts:   time.Now(),
	}:

	// NOTE: the defaultIO.inLastErr is unblock channel, so make it
	// unblock feed here, to prevent inputs blocked when IO blocked(and
	// the bug we have to fix)
	default:
		log.Warnf("FeedLastError(%s, %s) skipped, ignored", inputName, err)
	}
}

func SelfError(err string) {
	FeedLastError(datakit.DatakitInputName, err)
}

//nolint:gocyclo
func (x *IO) DoFeed(pts []*point.Point, category, from string, opt *Option) error {
	log.Debugf("io feed %s|%s", from, category)

	var ch chan *iodata

	filtered := 0
	var after []*point.Point

	switch category {
	case datakit.Logging,
		datakit.Tracing,
		datakit.Object,
		datakit.Network,
		datakit.KeyEvent,
		datakit.CustomObject,
		datakit.RUM,
		datakit.Security,
		datakit.Profiling:
		if opt == nil {
			opt = &Option{Blocking: true}
		} else {
			opt.Blocking = true
		}
	case datakit.Metric, datakit.MetricDeprecated:
	default:
		return fmt.Errorf("invalid category `%s'", category)
	}

	// run pipeline
	var plopt *plscript.Option
	var scriptConfMap map[string]string
	if opt != nil {
		plopt = opt.PlOption
		scriptConfMap = opt.PlScript
	}
	after, err := pipeline.RunPl(category, pts, plopt, scriptConfMap)
	if err != nil {
		log.Error(err)
	} else {
		pts = after
	}

	// run filters
	after = filter.FilterPts(category, pts)
	filtered = len(pts) - len(after)
	pts = after

	if opt != nil && opt.HTTPHost != "" {
		ch = x.chans[datakit.DynamicDatawayCategory]
	} else {
		ch = x.chans[category]
	}

	job := &iodata{
		category: category,
		pts:      pts,
		filtered: filtered,
		from:     from,
		opt:      opt,
	}

	if opt != nil && opt.Blocking {
		return blockingFeed(job, ch)
	}

	return unblockingFeed(job, ch)
}

func unblockingFeed(job *iodata, ch chan *iodata) error {
	// Maybe all points been filtered, but we still send the feeding into io.
	// We can still see some inputs/data are sending to io in monitor. Do not
	// optimize the feeding, or we see nothing on monitor about these filtered
	// points.
	select {
	case ch <- job:
		return nil
	case <-datakit.Exit.Wait():
		log.Warnf("%s/%s feed skipped on global exit", job.category, job.from)
		return fmt.Errorf("feed on global exit")

	default:
		atomic.AddUint64(&FeedDropPts, uint64(len(job.pts)))
		log.Warnf("io busy, %d (%s/%s) points maybe dropped", len(job.pts), job.from, job.category)
		return ErrIOBusy
	}
}

func blockingFeed(job *iodata, ch chan *iodata) error {
	select {
	case ch <- job:
		return nil

	case <-datakit.Exit.Wait():
		log.Warnf("%s/%s feed skipped on global exit", job.category, job.from)
		return fmt.Errorf("feed on global exit")
	}
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"
	"sync/atomic"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
)

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
		l.Warnf("FeedLastError(%s, %s) skipped, ignored", inputName, err)
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
		datakit.Metric,
		datakit.MetricDeprecated,
		datakit.Object,
		datakit.Network,
		datakit.KeyEvent,
		datakit.CustomObject,
		datakit.RUM,
		datakit.Security,
		datakit.Profiling:

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

	// run filters
	after = filterPts(category, pts)
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

	if x.conf.BlockingMode {
		return blockingFeed(job, ch)
	} else {
		return unblockingFeed(job, ch)
	}
}

// 重试机制：这里做三次重试，主要考虑：
//  - 尽量不丢数据，io goroutine 在处理 job 的时候，不太会超过 100ms
//    还不能完成，而且重试三次
//  - 最大三次重试，在一定程度上，尽量不阻塞住采集端以及数据接收端。这里
//    的数据接收端可能是用户系统将数据打到 datakit（日志/Tracing 等），不能
//    阻塞这些用户系统的 HTTP 调用
func unblockingFeed(job *iodata, ch chan *iodata) error {
	retry := 0
	for {
		if retry >= 3 {
			log.Warnf("feed retry %d, dropped %d point on %s", retry, len(job.pts), job.category)
			atomic.AddUint64(&FeedDropPts, uint64(len(job.pts)))
			return fmt.Errorf("io busy")
		}

		// Maybe all points been filtered, but we still send the feeding into io.
		// We can still see some inputs/data are sending to io in monitor. Do not
		// optimize the feeding, or we see nothing on monitor about these filtered
		// points.
		select {
		case ch <- job:
			if retry > 0 {
				log.Warnf("feed retry %d ok", retry)
			}

			return nil

		case <-datakit.Exit.Wait():
			log.Warnf("%s/%s feed skipped on global exit", job.category, job.from)
			return fmt.Errorf("feed on global exit")

		default:
			time.Sleep(time.Millisecond * 100)
			retry++
		}
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

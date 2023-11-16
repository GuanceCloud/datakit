// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package io implements datakits data transfer among inputs.
package io

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/failcache"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/recorder"
)

var (
	log = logger.DefaultSLogger("io")
	g   = datakit.G("io")

	// default dkIO singleton.
	defIO = getIO()
)

type dkIO struct {
	//////////////////////////
	// optional fields
	//////////////////////////
	dw      dataway.IDataway
	filters map[string]filter.FilterConditions

	cacheSizeGB        int
	cacheCleanInterval time.Duration
	enableCache,
	cacheAll bool

	withFilter,
	withConsumer bool

	recorder *recorder.Recorder

	flushInterval time.Duration
	flushWorkers  int

	maxCacheCount int

	//////////////////////////
	// inner fields
	//////////////////////////
	fo FeederOutputer

	fcs map[string]failcache.Cache

	lock sync.RWMutex

	droppedTotal int64
}

func Start(opts ...IOOption) {
	log = logger.SLogger("io")

	for _, opt := range opts {
		if opt != nil {
			opt(defIO)
		}
	}

	log.Debugf("default io config: %v", defIO)
	defIO.start()
}

func getIO() *dkIO {
	x := &dkIO{
		cacheSizeGB:        1 * 1024 * 1024,
		cacheCleanInterval: 30 * time.Second,
		enableCache:        false,

		withFilter:   true,
		withConsumer: true,

		flushInterval: time.Second * 10,
		maxCacheCount: 1024,

		fcs: map[string]failcache.Cache{},

		lock: sync.RWMutex{},
	}

	return x
}

func (x *dkIO) start() {
	if x.withFilter {
		g.Go(func(_ context.Context) error {
			if defIO.filters != nil {
				log.Infof("use local filters")
				filter.StartFilter(filter.NewLocalFilter(defIO.filters))
			} else {
				log.Infof("use remote filters")
				filter.StartFilter(defIO.dw)
			}

			return nil
		})
	}

	if x.withConsumer {
		fn := func(cat point.Category, n int) {
			log.Infof("start %d workers on %q", n, cat)
			for i := 0; i < n; i++ {
				g.Go(func(_ context.Context) error {
					x.runConsumer(cat)
					return nil
				})
			}
		}

		nworker := runtime.NumCPU()*2 + 1
		if x.flushWorkers > 0 {
			nworker = x.flushWorkers
		}

		for _, c := range point.AllCategories() {
			log.Infof("starting consumer on %q...", c.String())

			//nolint:exhaustive
			switch c {
			case point.Metric,
				point.Network,
				point.Logging,
				point.Tracing,
				point.RUM:
				fn(c, nworker)

				flushWorkersVec.WithLabelValues(c.String()).Set(float64(nworker))
			default:
				fn(c, 1)
				flushWorkersVec.WithLabelValues(c.String()).Set(1)
			}
		}
	}
}

func (x *dkIO) DroppedTotal() int64 {
	// NOTE: not thread-safe
	return x.droppedTotal
}

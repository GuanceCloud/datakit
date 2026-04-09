// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package io implements datakits data transfer among inputs.
package io

import (
	"context"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/aggr"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/compact"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/endpoint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/recorder"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/remotejob"
)

var (
	log = logger.DefaultSLogger("io")

	// default dkIO singleton.
	defIO = getIO()

	correctPointTimeAtDuration = float64(time.Hour * 2)
)

type dkIO struct {
	dw      dataway.IDataway
	filters map[string]filter.FilterConditions

	withTimeCorrect,
	withFilter,
	withCompactor bool

	recorder *recorder.Recorder

	flushInterval time.Duration
	availableCPUs,
	flushWorkers int

	compactAt int

	foDataway FeederOutputer
	Aggr      *aggr.Aggregator

	remoteManager *remotejob.Manager
	lock          sync.RWMutex
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
		withFilter:      true,
		withCompactor:   true,
		withTimeCorrect: true,

		flushInterval: time.Second * 10,
		compactAt:     1024,

		lock: sync.RWMutex{},
	}

	return x
}

func (x *dkIO) start() {
	compact.Setup()
	endpoint.Setup()

	if x.withFilter {
		g := goroutine.G("io/filter")
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

	if x.withCompactor {
		compactorWorker := func(cat point.Category, n int) {
			log.Infof("start %dth workers on %q", n, cat)
			g := goroutine.G("io/compactor/" + cat.Alias())
			for i := 0; i < n; i++ {
				g.Go(func(_ context.Context) error {
					x.runCompactor(cat)
					return nil
				})
			}
		}

		nworker := x.availableCPUs * 2
		if x.flushWorkers > 0 {
			nworker = x.flushWorkers
		}

		for _, c := range point.AllCategories() {
			log.Infof("starting consumer on %q...", c.String())

			//nolint:exhaustive
			switch c {
			case
				point.Metric,
				point.Network,
				point.Logging,
				point.Tracing,
				point.RUM:
				compactorWorker(c, nworker)

				flushWorkersVec.WithLabelValues(c.String()).Set(float64(nworker))
			default:
				compactorWorker(c, 1)
				flushWorkersVec.WithLabelValues(c.String()).Set(1)
			}
		}
	}

	if x.remoteManager != nil {
		g := goroutine.G("io/remote_job")
		g.Go(func(_ context.Context) error {
			// x.remoteManager.AddJob(remotejob.NewJVMJob(x.remoteManager.Envs, ""))
			x.remoteManager.Start()
			return nil
		})
	}
	// 定时下拉聚合配置
	g := goroutine.G("io/aggr")
	g.Go(func(_ context.Context) error {
		if x.Aggr == nil {
			log.Warnf("aggr is nil,return")
		} else {
			x.Aggr.StartAggr()
		}

		return nil
	})
}

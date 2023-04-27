// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package io implements datakits data transfer among inputs.
package io

import (
	"context"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/failcache"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/filter"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
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

	outputFile       string
	outputFileInputs []string

	flushInterval time.Duration
	flushWorkers  int

	feedChanSize  int
	maxCacheCount int

	//////////////////////////
	// inner fields
	//////////////////////////
	chans map[string]chan *iodata
	fcs   map[string]failcache.Cache

	lock sync.RWMutex

	fd *os.File

	droppedTotal   int64
	outputFileSize int64
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

	plscript.SetUploadFn(plAggFeed)
}

func getIO() *dkIO {
	x := &dkIO{
		cacheSizeGB:        1 * 1024 * 1024,
		cacheCleanInterval: 30 * time.Second,
		enableCache:        false,

		flushInterval: time.Second * 10,
		feedChanSize:  128,
		maxCacheCount: 1024,

		chans: map[string]chan *iodata{},
		fcs:   map[string]failcache.Cache{},

		lock: sync.RWMutex{},
	}

	x.chanSetup()
	return x
}

func (x *dkIO) chanSetup() {
	for _, c := range []string{
		datakit.Metric,
		datakit.Network,
		datakit.KeyEvent,
		datakit.Object,
		datakit.CustomObject,
		datakit.Logging,
		datakit.Tracing,
		datakit.RUM,
		datakit.Security,
		datakit.Profiling,
		datakit.DynamicDatawayCategory,
	} {
		x.chans[c] = make(chan *iodata, x.feedChanSize)
	}
}

func (x *dkIO) start() {
	x.chanSetup() // reset chan size

	ioChanCap.WithLabelValues("all-the-same").Set(float64(defIO.feedChanSize))

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

	fn := func(category string, n int) {
		log.Infof("start %d workers on %q", n, category)
		for i := 0; i < n; i++ {
			g.Go(func(_ context.Context) error {
				x.runConsumer(category)
				return nil
			})
		}
	}

	nworker := runtime.NumCPU()*2 + 1
	if x.flushWorkers > 0 {
		nworker = x.flushWorkers
	}

	for _, c := range []string{
		datakit.Metric,
		datakit.Network,
		datakit.KeyEvent,
		datakit.Object,
		datakit.CustomObject,
		datakit.Logging,
		datakit.Tracing,
		datakit.RUM,
		datakit.Security,
		datakit.Profiling,
		datakit.DynamicDatawayCategory,
	} {
		log.Infof("starting consumer on %s...", c)
		switch c {
		case datakit.Metric, datakit.Network, datakit.Logging, datakit.Tracing, datakit.RUM:
			fn(c, nworker)

			flushWorkersVec.WithLabelValues(point.CatURL(c).String()).Set(float64(nworker))
		default:
			fn(c, 1)
			flushWorkersVec.WithLabelValues(point.CatURL(c).String()).Set(1.0)
		}
	}
}

func (x *dkIO) DroppedTotal() int64 {
	// NOTE: not thread-safe
	return x.droppedTotal
}

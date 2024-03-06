// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package diskcache wrap cache/storage functions
package diskcache

import (
	"context"
	"fmt"
	sync "sync"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"github.com/GuanceCloud/cliutils/pipeline/manager"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"google.golang.org/protobuf/proto"
)

var (
	// globalCache is the shared cache instance.
	globalCache Cache

	// globalPBChan.
	globalPBChan chan *PBData

	initOnce sync.Once

	l = logger.DefaultSLogger("logging-cache")
)

const (
	// storage 需要给每一种数据定义 key，不能重复.
	storageKey uint8 = 8

	pbChanLength = 128
)

func Start(opts ...Option) error {
	var err error

	initOnce.Do(func() {
		l = logger.SLogger("logging-cache")

		c := defaultConfig()
		for _, opt := range opts {
			opt(c)
		}

		config := &storage.StorageConfig{Path: c.path, Capacity: c.capacity}
		globalCache, err = storage.NewStorage(config, l)
		if err != nil {
			return
		}

		globalPBChan = make(chan *PBData, pbChanLength)
		globalCache.RegisterConsumer(storageKey, func(buf []byte) error {
			pbdata := &PBData{}
			if err := proto.Unmarshal(buf, pbdata); err != nil {
				return fmt.Errorf("unable parsed PBData, err: %w", err)
			}
			globalPBChan <- pbdata
			return nil
		})

		err = globalCache.RunConsumeWorker()
		if err != nil {
			return
		}

		l.Infof("init diskcache success, path: %s, capacity: %dMiB", c.path, c.capacity)

		for i := 0; i < c.consumerNum; i++ {
			g := goroutine.NewGroup(goroutine.Option{Name: "logtail-dickcache"})
			g.Go(func(ctx context.Context) error {
				for {
					select {
					case <-datakit.Exit.Wait():
						l.Info("consumer goroutine exit")
						return nil

					case pbdata := <-globalPBChan:
						pts, cfg, err := c.handle(pbdata)
						if err != nil {
							l.Warnf("handle err:%s", err)
							continue
						}

						if err := c.feeder.FeedV2(point.Logging, pts,
							dkio.WithInputName("logging/"+cfg.Source),
							dkio.WithPipelineOption(&manager.Option{
								DisableAddStatusField: cfg.DisableAddStatusField,
								IgnoreStatus:          cfg.IgnoreStatus,
								ScriptMap:             map[string]string{cfg.Source: cfg.Pipeline},
							}),
							dkio.WithBlocking(cfg.Blocking),
						); err != nil {
							l.Warnf("feed %d pts failed: %w", len(pts), err)
						}
					}
				}
			})
		}
	})

	return err
}

func Put(buf []byte) error {
	if globalCache == nil {
		return fmt.Errorf("invalid logging cache")
	}
	return globalCache.Put(storageKey, buf)
}

func DecodeToPoint(p *PBData) ([]*point.Point, *PBConfig, error) {
	cfg := p.Config
	l.Debugf("consume pbdata, config: %#v", cfg)

	dec := point.GetDecoder(point.WithDecEncoding(point.Protobuf))
	defer point.PutDecoder(dec)

	pts, err := dec.Decode(p.Points)
	if err != nil {
		return nil, nil, fmt.Errorf("decode pbpoint err: %w", err)
	}

	return pts, cfg, nil
}

func Close() error {
	return globalCache.Close()
}

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
	pbpoint "github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/script"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"google.golang.org/protobuf/proto"
)

var (
	// globalCache is the shared cache instance.
	globalCache Cache

	// globalPBChan.
	globalPBChan chan *PBData

	initOnce sync.Once
	//nolint
	initErr error

	l = logger.DefaultSLogger("logging-cache")
)

const (
	// storage 需要给每一种数据定义 key，不能重复.
	storageKey uint8 = 8

	pbChanLength = 128
)

func Start(opts ...Option) error {
	initOnce.Do(func() {
		l = logger.SLogger("logging-cache")

		c := defaultConfig()
		for _, opt := range opts {
			opt(c)
		}

		config := &storage.StorageConfig{Path: c.path, Capacity: c.capacity}
		globalCache, initErr = storage.NewStorage(config, l)
		if initErr != nil {
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

		initErr = globalCache.RunConsumeWorker()
		if initErr != nil {
			return
		}

		for i := 0; i < c.consumerNum; i++ {
			g := goroutine.NewGroup(goroutine.Option{Name: "logtail-dickcache"})
			g.Go(func(ctx context.Context) error {
				for {
					select {
					case <-datakit.Exit.Wait():
						l.Debug("consumer goroutine exit")
						return nil

					case pbdata := <-globalPBChan:
						if err := c.handle(pbdata); err != nil {
							l.Warnf("handle err:%s", err)
						}
					}
				}
			})
		}
	})

	return initErr
}

func Put(buf []byte) error {
	if globalCache == nil {
		return fmt.Errorf("invalid logging cache")
	}
	return globalCache.Put(storageKey, buf)
}

func HandleFeedIO(p *PBData) error {
	cfg := p.Config
	l.Debugf("consume pbdata, config: %#v", cfg)

	pts, err := pbpoint.GetDecoder(pbpoint.WithDecEncoding(pbpoint.Protobuf)).Decode(p.Points)
	if err != nil {
		return fmt.Errorf("decode pbpoint err: %w", err)
	}

	res := []*point.Point{}

	for _, pt := range pts {
		lpt, err := point.NewPoint(
			string(pt.Name()),
			pt.Tags().InfluxTags(),
			pt.Fields().InfluxFields(),
			&point.PointOption{Time: pt.Time(), Category: datakit.Logging},
		)
		if err != nil {
			l.Warnf("new point err:%s, skip", err)
			continue
		}

		l.Debugf("consume pbdata, point %s", lpt.String())
		res = append(res, lpt)
	}

	if err := iod.Feed(
		"logging/"+cfg.Source,
		datakit.Logging,
		res,
		&iod.Option{
			PlScript: map[string]string{cfg.Source: cfg.Pipeline},
			PlOption: &script.Option{
				DisableAddStatusField: cfg.DisableAddStatusField,
				IgnoreStatus:          cfg.IgnoreStatus,
			},
			Blocking: cfg.Blocking,
		},
	); err != nil {
		return fmt.Errorf("feed %d pts failed: %w", len(res), err)
	}

	return nil
}

func Close() error {
	return globalCache.Close()
}

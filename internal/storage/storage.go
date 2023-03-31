// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package storage define local-cache for tracing, logging, rum high frequency data buffering
package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils"
	dc "github.com/GuanceCloud/cliutils/diskcache"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
)

const (
	HTTP_KEY             uint8 = iota + 1 // nolint: stylecheck
	OTEL_GRPC_KEY                         // nolint: stylecheck
	SKY_WALKING_GRPC_KEY                  // nolint: stylecheck
	ZIPKIN_HTTP_V1_KEY                    // nolint: stylecheck
	ZIPKIN_HTTP_V2_KEY                    // nolint: stylecheck
)

type StorageConfig struct {
	Path     string `json:"storage"`
	Capacity int    `json:"capacity"`
}

type ConsumerFunc func(buf []byte) error

type Storage struct {
	path      string
	cache     *dc.DiskCache
	log       *logger.Logger
	exit      *cliutils.Sem
	consumers map[uint8]ConsumerFunc
	enabled   bool
}

func NewStorage(config *StorageConfig, log *logger.Logger) (*Storage, error) {
	if config == nil || config.Path == "" || config.Capacity < 0 {
		return nil, errors.New("storage config error")
	}

	cache, err := dc.Open(
		dc.WithPath(datakit.JoinToCacheDir(config.Path)),
		dc.WithCapacity(int64(config.Capacity)<<20))
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		path:      config.Path,
		cache:     cache,
		log:       log,
		exit:      cliutils.NewSem(),
		consumers: make(map[uint8]ConsumerFunc),
		enabled:   false,
	}
	if storage.log == nil {
		storage.log = logger.DefaultSLogger("local-cache")
	}

	return storage, nil
}

func (s *Storage) RunConsumeWorker() error {
	if s.enabled {
		return errors.New("local-cache is already enabled")
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "internal_trace"})
	g.Go(func(ctx context.Context) error {
		for {
			select {
			case <-s.exit.Wait():
				s.log.Infof("on exit, storage stop on path '%s' exit", s.path)

				return nil
			default:
			}

			key, buf, err := s.Get()
			if err != nil {
				if errors.Is(err, dc.ErrEOF) {
					s.log.Debug("local-cache empty")
					time.Sleep(time.Second)
					continue
				}
				s.log.Error(err.Error())
				continue
			}
			consumer, ok := s.consumers[key]
			if !ok {
				s.log.Errorf("consumer of key: %d not found", key)
				continue
			}
			if err = consumer(buf); err != nil {
				s.log.Error(err.Error())
			}
		}
	})
	s.enabled = true

	return nil
}

func (s *Storage) Put(key uint8, buf []byte) error {
	if !s.enabled {
		return errors.New("local-cache consume worker not running")
	}

	newbuf := make([]byte, len(buf)+1)
	newbuf[0] = key
	copy(newbuf[1:], buf)

	return s.cache.Put(newbuf)
}

func (s *Storage) Get() (key uint8, buf []byte, err error) {
	err = s.cache.Get(func(b []byte) error {
		if len(b) < 1 {
			return errors.New("bytes not long enough")
		}
		buf = make([]byte, len(b))
		copy(buf, b)

		return nil
	})
	if err != nil { // nolint: gocritic
		return 0, nil, err
	} else if len(buf) < 1 {
		return 0, nil, fmt.Errorf("get wrong bytes with prefix: %d len: %d", key, len(buf))
	} else {
		return buf[0], buf[1:], nil
	}
}

func (s *Storage) RegisterConsumer(key uint8, consumer ConsumerFunc) {
	if consumer != nil {
		s.consumers[key] = consumer
	}
}

func (s *Storage) Close() error {
	if s.exit != nil {
		s.exit.Close()
	}
	if s.cache != nil {
		return s.cache.Close()
	}
	s.enabled = false

	return nil
}

func (s *Storage) Enabled() bool {
	return s.enabled
}

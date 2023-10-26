// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package offload

import (
	"context"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
)

const (
	maxCustomer   = 16
	chanSize      = maxCustomer
	ptsBuf        = 128
	flushInterval = time.Second * 15
)

var l = logger.DefaultSLogger("pl-offload")

func InitLog() {
	l = logger.SLogger("pl-offload")
}

type OffloadConfig struct {
	Receiver  string   `toml:"receiver"`
	Addresses []string `toml:"addresses"`
}

type Receiver interface {
	// thread safety
	Send(s uint64, cat point.Category, data []*point.Point) error
}

type dataChan struct {
	unknownCategory,
	dynamicDWCategory,
	metric,
	metricDeprecated,
	network,
	keyEvent,
	object,
	customObject,
	logging,
	tracing,
	rum,
	security,
	profiling chan []*point.Point
}

func newDataChan() *dataChan {
	newChan := func() chan []*point.Point {
		return make(chan []*point.Point, chanSize)
	}

	return &dataChan{
		unknownCategory:   newChan(),
		dynamicDWCategory: newChan(),
		metric:            newChan(),
		metricDeprecated:  newChan(),
		network:           newChan(),
		keyEvent:          newChan(),
		object:            newChan(),
		customObject:      newChan(),
		logging:           newChan(),
		tracing:           newChan(),
		rum:               newChan(),
		security:          newChan(),
		profiling:         newChan(),
	}
}

type OffloadWorker struct {
	ch       *dataChan
	stopChan chan struct{}

	sender Receiver
}

func NewOffloader(cfg *OffloadConfig) (*OffloadWorker, error) {
	if cfg == nil || len(cfg.Receiver) == 0 {
		return nil, fmt.Errorf("no config")
	}

	wrk := &OffloadWorker{
		ch:       newDataChan(),
		stopChan: make(chan struct{}),
	}

	switch cfg.Receiver {
	case DKRcv:
		if s, err := NewDKRecver(cfg.Addresses); err != nil {
			return nil, err
		} else {
			wrk.sender = s
		}
	default:
		return nil, fmt.Errorf("unsupported receiver")
	}

	return wrk, nil
}

func (offload *OffloadWorker) Customer(ctx context.Context, cat point.Category) error {
	flushTicker := time.NewTicker(flushInterval)
	var ch chan []*point.Point

	switch cat { //nolint:exhaustive
	case point.Logging:
		ch = offload.ch.logging
	default:
		return fmt.Errorf("unsupported category")
	}

	ptsCache := make([]*point.Point, 0, ptsBuf)

	var lbID uint64 = 0 // taking modulus to achieve load balancing

	for {
		select {
		case pts := <-ch:
			ptsCache, lbID = offload.sendOrCache(lbID, cat, ptsCache, pts)

		case <-flushTicker.C:
			if len(ptsCache) > 0 {
				if err := offload.sender.Send(lbID, cat, ptsCache); err != nil {
					l.Errorf("offload send failed: %w", err)
				}
				ptsCache = make([]*point.Point, 0, ptsBuf)
				lbID++
			}
		case <-offload.stopChan:
			if err := offload.sender.Send(lbID, cat, ptsCache); err != nil {
				l.Errorf("offload send failed: %w", err)
			}
			return nil

		case <-ctx.Done():
			if err := offload.sender.Send(lbID, cat, ptsCache); err != nil {
				l.Errorf("offload send failed: %w", err)
			}
			return nil
		}
	}
}

func (offload *OffloadWorker) sendOrCache(s uint64, cat point.Category, cache []*point.Point, ptsInput []*point.Point) ([]*point.Point, uint64) {
	diff := ptsBuf - len(cache)
	switch {
	case diff > len(ptsInput):
		// append
		cache = append(cache, ptsInput...)

	case diff == len(ptsInput):
		// append and send
		cache = append(cache, ptsInput...)
		if err := offload.sender.Send(s, cat, cache); err != nil {
			l.Errorf("offload send failed: %w", err)
		}
		// new slice
		cache = make([]*point.Point, 0, ptsBuf)
		s++

	case diff < len(ptsInput):
		cache = append(cache, ptsInput[:diff]...)
		if err := offload.sender.Send(s, cat, cache); err != nil {
			l.Errorf("offload send failed: %w", err)
		}
		cache = make([]*point.Point, 0, ptsBuf)

		ptsInput = ptsInput[diff:]
		for i := 0; i < len(ptsInput)/ptsBuf; i++ {
			cache = append(cache, ptsInput[i*ptsBuf:(i+1)*ptsBuf]...)
			if err := offload.sender.Send(s, cat, cache); err != nil {
				l.Errorf("offload send failed: %w", err)
			}
			cache = make([]*point.Point, 0, ptsBuf)
		}

		if i := len(ptsInput) % ptsBuf; i > 0 {
			cache = append(cache, ptsInput[len(ptsInput)-i:]...)
		}
		s++
	}
	return cache, s
}

func (offload *OffloadWorker) Send(cat point.Category, pts []*point.Point) error {
	if cat != point.Logging {
		return fmt.Errorf("unsupported category")
	}

	if offload.ch == nil || offload.ch.logging == nil {
		return fmt.Errorf("logging data chan not ready")
	}

	switch cat { //nolint:exhaustive
	case point.Logging:
		offload.ch.logging <- pts
	default:
	}

	return nil
}

func (offload *OffloadWorker) Stop() {
	if offload.stopChan != nil {
		close(offload.stopChan)
	}
}

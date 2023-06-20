// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package offload

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
)

const (
	maxCustomer   = 16
	chanSize      = maxCustomer
	ptsBuf        = 128
	flushInterval = time.Second * 15
)

var (
	_offloadWkr *OffloadWorker

	l = logger.DefaultSLogger("pl-offload")
)

func InitOffloaWorker(cfg *OffloadConfig) error {
	l = logger.SLogger("pl-offload")

	if wkr, err := newOffloader(cfg); err != nil {
		return err
	} else {
		_offloadWkr = wkr
		for i := 0; i < int(math.Ceil(float64(runtime.NumCPU())*1.5)); i++ {
			if i >= maxCustomer {
				break
			}
			// logging only
			_offloadWkr.g.Go(func(ctx context.Context) error {
				return _offloadWkr.Customer(ctx, point.Logging)
			})
		}
		return nil
	}
}

func Enabled() bool {
	return _offloadWkr != nil
}

func OffloadSend(cat point.Category, pt []*dkpt.Point) error {
	if _offloadWkr == nil {
		return fmt.Errorf("offload worker not inited")
	}
	return _offloadWkr.Send(cat, pt)
}

type OffloadConfig struct {
	Receiver  string   `toml:"receiver"`
	Addresses []string `toml:"addresses"`
}

type Receiver interface {
	// thread safety
	Send(s uint64, cat point.Category, data []*dkpt.Point) error
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
	profiling chan []*dkpt.Point
}

func newDataChan() *dataChan {
	newChan := func() chan []*dkpt.Point {
		return make(chan []*dkpt.Point, chanSize)
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

	g *goroutine.Group
}

func newOffloader(cfg *OffloadConfig) (*OffloadWorker, error) {
	if cfg == nil || len(cfg.Receiver) == 0 {
		return nil, fmt.Errorf("no config")
	}

	wrk := &OffloadWorker{
		ch:       newDataChan(),
		stopChan: make(chan struct{}),
		g: goroutine.NewGroup(goroutine.Option{
			Name: "pipeline-offload",
		}),
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
	var ch chan []*dkpt.Point

	switch cat { //nolint:exhaustive
	case point.Logging:
		ch = offload.ch.logging
	default:
		return fmt.Errorf("unsupported category")
	}

	ptsCache := make([]*dkpt.Point, 0, ptsBuf)

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
				ptsCache = make([]*dkpt.Point, 0, ptsBuf)
				lbID++
			}
		case <-offload.stopChan:
			if err := offload.sender.Send(lbID, cat, ptsCache); err != nil {
				l.Errorf("offload send failed: %w", err)
			}
			return nil

		case <-datakit.Exit.Wait():
			if err := offload.sender.Send(lbID, cat, ptsCache); err != nil {
				l.Errorf("offload send failed: %w", err)
			}
			return nil
		}
	}
}

func (offload *OffloadWorker) sendOrCache(s uint64, cat point.Category, cache []*dkpt.Point, ptsInput []*dkpt.Point) ([]*dkpt.Point, uint64) {
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
		cache = make([]*dkpt.Point, 0, ptsBuf)
		s++

	case diff < len(ptsInput):
		cache = append(cache, ptsInput[:diff]...)
		if err := offload.sender.Send(s, cat, cache); err != nil {
			l.Errorf("offload send failed: %w", err)
		}
		cache = make([]*dkpt.Point, 0, ptsBuf)

		ptsInput = ptsInput[diff:]
		for i := 0; i < len(ptsInput)/ptsBuf; i++ {
			cache = append(cache, ptsInput[i*ptsBuf:(i+1)*ptsBuf]...)
			if err := offload.sender.Send(s, cat, cache); err != nil {
				l.Errorf("offload send failed: %w", err)
			}
			cache = make([]*dkpt.Point, 0, ptsBuf)
		}

		if i := len(ptsInput) % ptsBuf; i > 0 {
			cache = append(cache, ptsInput[len(ptsInput)-i:]...)
		}
		s++
	}
	return cache, s
}

func (offload *OffloadWorker) Send(cat point.Category, pts []*dkpt.Point) error {
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

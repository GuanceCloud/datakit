// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package trace

import (
	"context"
	"errors"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
)

type SpanSender struct {
	name            string
	duration        time.Duration
	afterGatherFunc AfterGatherFunc
	log             *logger.Logger
	p               chan *DkSpan
	sig             chan struct{}
}

func NewSpanSender(name string, capacity int, duration time.Duration, afterGatherFunc AfterGatherFunc, log *logger.Logger) (*SpanSender, error) {
	if name == "" || capacity < 1 || duration < 1 || afterGatherFunc == nil {
		return nil, errors.New("span sender config error")
	}

	ss := &SpanSender{
		name:            name,
		duration:        duration,
		afterGatherFunc: afterGatherFunc,
		log:             log,
		p:               make(chan *DkSpan, capacity),
		sig:             make(chan struct{}),
	}
	if ss.log == nil {
		ss.log = logger.DefaultSLogger("span_sender")
	}

	return ss, nil
}

func (ss *SpanSender) Start() {
	g := goroutine.NewGroup(goroutine.Option{Name: "span_sender"})
	g.Go(func(ctx context.Context) error {
		ss.flushWorker()

		return nil
	})
}

func (ss *SpanSender) Append(dkspans ...*DkSpan) {
	for i := range dkspans {
		if len(ss.p) == cap(ss.p) {
			ss.sig <- struct{}{}
		}
		ss.p <- dkspans[i]
	}
}

func (ss *SpanSender) Close() {
	close(ss.sig)
}

func (ss *SpanSender) flushWorker() {
	for {
		timer := time.NewTimer(ss.duration)
		select {
		case _, ok := <-ss.sig:
			timer.Stop()
			if !ok {
				return
			}
		case <-timer.C:
		}

		l := len(ss.p)
		if l != 0 {
			trace := make(DatakitTrace, l)
			for i := 0; i < l; i++ {
				trace[i] = <-ss.p
			}
			ss.afterGatherFunc.Run(ss.name, DatakitTraces{trace})
		}
	}
}

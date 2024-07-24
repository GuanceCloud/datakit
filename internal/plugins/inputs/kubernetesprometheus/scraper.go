// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"sync"
	"time"

	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

type scraper struct {
	proms map[string]map[string]context.CancelFunc
	mu    sync.Mutex
}

func newScraper() *scraper {
	return &scraper{
		proms: make(map[string]map[string]context.CancelFunc),
	}
}

func (s *scraper) runProm(ctx context.Context, key, urlstr string, interval time.Duration, opts []iprom.PromOption) {
	// lock
	s.mu.Lock()
	if _, exist := s.proms[key]; !exist {
		s.proms[key] = make(map[string]context.CancelFunc)
	}
	ctx, cancel := context.WithCancel(ctx)
	s.proms[key][urlstr] = cancel
	length := len(s.proms[key])
	// unlock
	s.mu.Unlock()

	klog.Infof("create prom url %s for %s, interval %s, current len(%d)", urlstr, key, interval, length)

	if err := runPromCollect(ctx, interval, urlstr, opts); err != nil {
		klog.Warnf("failed of prom %s", err)
	}
}

func (s *scraper) terminateProms(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cancels, exist := s.proms[key]
	if !exist {
		return
	}

	klog.Infof("terminate prom len(%d) from key %s", len(cancels), key)

	for _, cancel := range cancels {
		cancel()
	}
	delete(s.proms, key)
}

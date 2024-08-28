// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const workInterval = time.Second * 5

type scrapeTarget interface {
	url() string
	scrape() error
}

type scrapeWorker struct {
	role    Role
	keys    map[string]string
	targets map[string][]scrapeTarget

	targetChan chan scrapeTarget
	mu         sync.Mutex
}

func newScrapeWorker(role Role) *scrapeWorker {
	return &scrapeWorker{
		role:    role,
		keys:    make(map[string]string),
		targets: make(map[string][]scrapeTarget),
	}
}

func (s *scrapeWorker) startWorker(ctx context.Context, workerNum int) {
	if s.targetChan == nil {
		s.targetChan = make(chan scrapeTarget, workerNum)
	}

	for i := 0; i < workerNum; i++ {
		workerGo.Go(func(_ context.Context) error {
			for target := range s.targetChan {
				if err := target.scrape(); err != nil {
					klog.Warn(err)
				}
			}
			return nil
		})
	}

	workerGo.Go(func(_ context.Context) error {
		defer func() {
			close(s.targetChan)
			klog.Infof("role-%s worker exit", s.role)
		}()

		tick := time.NewTicker(workInterval)
		defer tick.Stop()
		klog.Infof("role-%s worker start", s.role)

		for {
			s.mu.Lock()
			for _, targets := range s.targets {
				for _, target := range targets {
					select {
					case <-datakit.Exit.Wait():
						return nil

					case <-ctx.Done():
						return nil

					default:
						// next
					}
					s.targetChan <- target
				}
			}
			s.mu.Unlock()

			select {
			case <-datakit.Exit.Wait():
				return nil

			case <-ctx.Done():
				return nil

			case <-tick.C:
				// next
			}
		}
	})
}

func (s *scrapeWorker) registerKey(key, feature string) {
	s.mu.Lock()
	s.keys[key] = feature
	s.mu.Unlock()
}

func (s *scrapeWorker) matchesKey(key, feature string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, exists := s.keys[key]
	return exists && f == feature
}

func (s *scrapeWorker) registerTarget(key string, target scrapeTarget) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.targets[key] = append(s.targets[key], target)
	klog.Infof("added prom url %s for %s, current len(%d)", target.url(), key, len(s.targets[key]))
	scrapeTargetNumber.WithLabelValues(string(s.role), key).Add(1)
}

func (s *scrapeWorker) terminate(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.keys, key)

	if _, ok := s.targets[key]; ok {
		length := len(s.targets[key])
		delete(s.targets, key)
		klog.Infof("terminated prom len(%d) for key %s", length, key)
		scrapeTargetNumber.WithLabelValues(string(s.role), key).Sub(float64(length))
	}
}

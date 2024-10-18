// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"strconv"
	"sync"
	"time"
)

const scraperChanNum = 32

type scraper interface {
	targetURL() string
	shouldScrape() bool
	scrape() error
	isTerminated() bool
	markAsTerminated()
}

type scrapeManager struct {
	role     Role
	keys     map[string]string
	scrapers map[string][]scraper

	scraperChan chan scraper
	mu          sync.Mutex
}

func newScrapeManager(role Role) *scrapeManager {
	return &scrapeManager{
		role:        role,
		keys:        make(map[string]string),
		scrapers:    make(map[string][]scraper),
		scraperChan: make(chan scraper, scraperChanNum),
	}
}

func (s *scrapeManager) registerScrape(key, feature string, sp scraper) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.keys[key]; !ok {
		s.keys[key] = feature
	}

	select {
	case s.scraperChan <- sp:
		s.scrapers[key] = append(s.scrapers[key], sp)
	default:
		klog.Warnf("manager channel is closed, register failed for key %s", key)
		return
	}

	klog.Infof("added scraper url %s for %s, current len(%d)", sp.targetURL(), key, len(s.scrapers[key]))
	scrapeTargetNumber.WithLabelValues(string(s.role), key).Add(1)
}

func (s *scrapeManager) matchesKey(key, feature string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, exists := s.keys[key]
	return exists && f == feature
}

func (s *scrapeManager) terminateScrape(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.keys, key)

	if _, ok := s.scrapers[key]; ok {
		for _, sp := range s.scrapers[key] {
			sp.markAsTerminated()
		}

		length := len(s.scrapers[key])
		klog.Infof("terminated scraper len(%d) for key %s", length, key)
		scrapeTargetNumber.WithLabelValues(string(s.role), key).Sub(float64(length))

		delete(s.scrapers, key)
	}
}

func (s *scrapeManager) run(ctx context.Context, workerNum int) {
	managerGo.Go(func(_ context.Context) error {
		<-ctx.Done()
		s.stop()
		klog.Infof("role-%s manager exit", s.role)
		return nil
	})

	for i := 0; i < workerNum; i++ {
		name := "worker-" + strconv.Itoa(i)
		workerGo.Go(func(_ context.Context) error {
			s.doWork(ctx, name)
			return nil
		})
	}
}

func (s *scrapeManager) doWork(ctx context.Context, name string) {
	var tasks []scraper
	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case sp, ok := <-s.scraperChan:
			if !ok {
				return
			}
			if len(tasks) >= 100 {
				klog.Warnf("%s scrape is over limit", s.role)
			} else {
				tasks = append(tasks, sp)
			}

		case <-tick.C:
			// next
		}

		removeIndex := []int{}
		for idx, task := range tasks {
			if task.isTerminated() {
				removeIndex = append(removeIndex, idx)
				continue
			}
			if task.shouldScrape() {
				if err := task.scrape(); err != nil {
					klog.Warnf("failed to scrape url %s, err %s", task.targetURL(), err)
					removeIndex = append(removeIndex, idx)
				}
			}
		}

		if len(removeIndex) != 0 {
			var newtasks []scraper
			for _, rmIdx := range removeIndex {
				for idx := range tasks {
					if idx == rmIdx {
						continue
					}
					newtasks = append(newtasks, tasks[idx])
				}
			}
			tasks = newtasks
		}

		activeWorkerTasks.WithLabelValues(string(s.role), name).Set(float64(len(tasks)))
	}
}

func (s *scrapeManager) stop() {
	select {
	case <-s.scraperChan:
		// nil
	default:
		close(s.scraperChan)
	}
}

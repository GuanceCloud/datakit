// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const scraperChanNum = 32

type scraper interface {
	targetURL() string
	shouldScrape() bool
	scrape(int64) error
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
		klog.Warnf("manager channel is stuffed, register failed for key %s", key)
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

func (s *scrapeManager) cleanDeadScrape() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.scrapers {
		var newScrapers []scraper
		var count int
		for _, sp := range s.scrapers[key] {
			if !sp.isTerminated() {
				newScrapers = append(newScrapers, sp)
				continue
			}
			count += 1
		}
		if count != 0 {
			s.scrapers[key] = newScrapers
			scrapeTargetNumber.WithLabelValues(string(s.role), key).Sub(float64(count))
			klog.Infof("clean %d scraper for %s", count, key)
		}
	}
}

func (s *scrapeManager) existScrape(key string, urlstr string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sp := range s.scrapers[key] {
		if sp.targetURL() == urlstr {
			return true
		}
	}
	return false
}

func (s *scrapeManager) tryCleanScrapes(key string, urlstrList []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sp := range s.scrapers[key] {
		found := false
		for _, urlstr := range urlstrList {
			if sp.targetURL() == urlstr {
				found = true
			}
		}
		if !found {
			sp.markAsTerminated()
		}
	}
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
		klog.Infof("%s terminated scraper len(%d) for key %s", s.role, length, key)
		scrapeTargetNumber.WithLabelValues(string(s.role), key).Sub(float64(length))

		delete(s.scrapers, key)
	}
}

func (s *scrapeManager) run(ctx context.Context, workerNum int) {
	managerGo.Go(func(_ context.Context) error {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				s.stop()
				klog.Infof("%s manager exit", s.role)
				return nil
			case <-ticker.C:
				s.cleanDeadScrape()
			}
		}
	})

	for i := 0; i < workerNum; i++ {
		name := "worker-" + strconv.Itoa(i)
		randSleep := rand.Intn(100 /*100 ms*/) //nolint:gosec
		time.Sleep(time.Duration(randSleep) * time.Millisecond)
		workerGo.Go(func(_ context.Context) error {
			s.doWork(ctx, name)
			return nil
		})
	}
}

func (s *scrapeManager) doWork(ctx context.Context, name string) {
	tasks := make(map[string]scraper)
	timestamp := time.Now().UnixNano() / 1e6

	ticker := time.NewTicker(globalScrapeInterval)
	defer ticker.Stop()

	for {
		timestamp += globalScrapeInterval.Milliseconds()
		select {
		case <-ctx.Done():
			return

		case sp, ok := <-s.scraperChan:
			if !ok {
				klog.Warnf("%s(%s) scrape channel is closed, exit", s.role, name)
				return
			}
			if len(tasks) >= 100 {
				klog.Warnf("%s(%s) scrape is over limit", s.role, name)
			} else {
				if _, ok := tasks[sp.targetURL()]; !ok {
					tasks[sp.targetURL()] = sp
				}
			}

		case tt := <-ticker.C:
			t := tt.UnixNano() / 1e6
			if d := math.Abs(float64(t - timestamp)); d > 0 && d/float64(globalScrapeInterval.Milliseconds()) > 0.1 {
				timestamp = t
			}

			var removeTasks []string
			for _, task := range tasks {
				if task.isTerminated() {
					removeTasks = append(removeTasks, task.targetURL())
					klog.Debugf("%s(%s) urlstr %s terminated", s.role, name, task.targetURL())
					continue
				}
				if task.shouldScrape() {
					if err := task.scrape(timestamp * 1e6 /* To Nanoseconds */); err != nil {
						klog.Warnf("failed to scrape url %s, err %s", task.targetURL(), err)
						task.isTerminated()
						removeTasks = append(removeTasks, task.targetURL())
					}
				}
			}

			if len(removeTasks) != 0 {
				klog.Debugf("%s(%s) remove len(%d) tasks", s.role, name, len(removeTasks))
				for _, taskURL := range removeTasks {
					delete(tasks, taskURL)
				}
			}

			activeWorkerTasks.WithLabelValues(string(s.role), name).Set(float64(len(tasks)))
		}
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

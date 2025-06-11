// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type scrapeManagerInterface interface {
	runWorker(ctx context.Context, workerChan int, interval time.Duration)
	registerScrape(role Role, key string, traits string, sp scraper)
	isTraitsExists(role Role, key string, traits string) bool
	isScrapeExists(role Role, key string, targetURL string) bool
	removeScrape(role Role, key string)
	tryCleanScrapes(role Role, key string, keepTargetURLs []string)
}

type scraper interface {
	targetURL() string
	shouldScrape() bool
	scrape(timestamp int64) error
	shouldRetry(maxScrapeRetry int) (bool, int)
	resetRetryCount()
	isTerminated() bool
	markAsTerminated()
}

type scrapeManager struct {
	nodeStore           *scrapeStore
	serviceStore        *scrapeStore
	endpointsStore      *scrapeStore
	podStore            *scrapeStore
	podmonitorStore     *scrapeStore
	servicemonitorStore *scrapeStore

	workerChan    chan scraper
	runWorkerOnce sync.Once
	mu            sync.Mutex
}

const defaultScraperChanNum = 32

func newScrapeManager() scrapeManagerInterface {
	return &scrapeManager{
		nodeStore:           newScrapeStore(),
		serviceStore:        newScrapeStore(),
		endpointsStore:      newScrapeStore(),
		podStore:            newScrapeStore(),
		podmonitorStore:     newScrapeStore(),
		servicemonitorStore: newScrapeStore(),
		workerChan:          make(chan scraper, defaultScraperChanNum),
	}
}

func (s *scrapeManager) registerScrape(role Role, key, traits string, sp scraper) {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case s.workerChan <- sp:
		s.store(role).addTraits(key, traits)
		s.store(role).addScraper(key, sp)
	default:
		klog.Warnf("manager channel is stuffed, register failed for key %s", key)
		return
	}

	klog.Infof("%s added scraper url %s for %s, current len(%d)", role, sp.targetURL(), key, s.store(role).scraperCount(key))
	scraperNumberVec.WithLabelValues(string(role), key).Add(1)
}

func (s *scrapeManager) isTraitsExists(role Role, key string, traits string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.store(role).hasTraits(key, traits)
}

func (s *scrapeManager) isScrapeExists(role Role, key string, targetURL string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.store(role).hasScraper(key, targetURL)
}

func (s *scrapeManager) removeScrape(role Role, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := s.store(role).deleteKeyAndTerminateScrapers(key)
	if count != 0 {
		klog.Infof("%s terminated scraper len(%d) for key %s", role, count, key)
		scraperNumberVec.WithLabelValues(string(role), key).Sub(float64(count))
	}
}

func (s *scrapeManager) tryCleanScrapes(role Role, key string, keepTargetURLs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sp := range s.store(role).scrapers[key] {
		found := false
		for _, urlstr := range keepTargetURLs {
			if sp.targetURL() == urlstr {
				found = true
			}
		}
		if !found {
			sp.markAsTerminated()
		}
	}
}

func (s *scrapeManager) runWorker(ctx context.Context, workerNum int, scrapeInterval time.Duration) {
	s.runWorkerOnce.Do(func() {
		managerGo.Go(func(_ context.Context) error {
			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					s.stop()
					klog.Info("manager exit")
					return nil

				case <-ticker.C:
					s.cleanDeadScraper()
				}
			}
		})

		for i := 0; i < workerNum; i++ {
			name := "worker-" + strconv.Itoa(i)

			randSleep := rand.Intn(100 /*100 ms*/) //nolint:gosec
			time.Sleep(time.Duration(randSleep) * time.Millisecond)

			workerGo.Go(func(_ context.Context) error {
				s.doWork(ctx, name, scrapeInterval)
				return nil
			})
		}
	})
}

func (s *scrapeManager) doWork(ctx context.Context, name string, scrapeInterval time.Duration) {
	tasks := make(map[string]scraper)
	start := ntp.Now()

	ticker := time.NewTicker(scrapeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case sp, ok := <-s.workerChan:
			if !ok {
				klog.Warnf("worker channel %s is closed, exit", name)
				return
			}

			if len(tasks) > maxTaskNumber {
				klog.Warnf("tasks is over limit %d", maxTaskNumber)
			} else {
				if _, ok := tasks[sp.targetURL()]; !ok {
					tasks[sp.targetURL()] = sp
				}
			}

		case tt := <-ticker.C:
			start = inputs.AlignTime(tt, start, scrapeInterval)

			var removeTasks []string
			for _, task := range tasks {
				if task.isTerminated() {
					removeTasks = append(removeTasks, task.targetURL())
					continue
				}
				if task.shouldScrape() {
					err := task.scrape(start.UnixNano())
					if err == nil {
						task.resetRetryCount()
						continue
					}

					retry, count := task.shouldRetry(maxScrapeRetry)
					klog.Warnf("failed to scrape url %s, err %s, retry count %d of %d", task.targetURL(), err, count, maxScrapeRetry)

					if !retry {
						klog.Warnf("task %s will be removed", task.targetURL())
						task.markAsTerminated()
						removeTasks = append(removeTasks, task.targetURL())
					}
				}
			}

			if len(removeTasks) != 0 {
				for _, taskURL := range removeTasks {
					delete(tasks, taskURL)
				}
			}

			taskNumerVec.WithLabelValues(name).Set(float64(len(tasks)))
		}
	}
}

func (s *scrapeManager) cleanDeadScraper() {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	count += s.nodeStore.cleanDeadScraper(string(RoleNode))
	count += s.serviceStore.cleanDeadScraper(string(RoleService))
	count += s.endpointsStore.cleanDeadScraper(string(RoleEndpoints))
	count += s.podStore.cleanDeadScraper(string(RolePod))
	count += s.podmonitorStore.cleanDeadScraper(string(RolePodMonitor))
	count += s.servicemonitorStore.cleanDeadScraper(string(RoleServiceMonitor))

	if count != 0 {
		klog.Infof("clean %d scraper", count)
	}
}

func (s *scrapeManager) stop() {
	select {
	case <-s.workerChan:
		// nil
	default:
		close(s.workerChan)
	}
}

func (s *scrapeManager) store(role Role) *scrapeStore {
	switch role {
	case RoleNode:
		return s.nodeStore
	case RoleService:
		return s.serviceStore
	case RoleEndpoints:
		return s.endpointsStore
	case RolePod:
		return s.podStore
	case RolePodMonitor:
		return s.podmonitorStore
	case RoleServiceMonitor:
		return s.servicemonitorStore
	default:
		// unreachable
		return nil
	}
}

type scrapeStore struct {
	traits   map[string]string
	scrapers map[string][]scraper
}

func newScrapeStore() *scrapeStore {
	return &scrapeStore{
		traits:   make(map[string]string),
		scrapers: make(map[string][]scraper),
	}
}

func (store *scrapeStore) addTraits(key, traits string) {
	store.traits[key] = traits
}

func (store *scrapeStore) hasTraits(key, traits string) bool {
	tr, exists := store.traits[key]
	return exists && tr == traits
}

func (store *scrapeStore) addScraper(key string, sp scraper) {
	store.scrapers[key] = append(store.scrapers[key], sp)
}

func (store *scrapeStore) hasScraper(key, targetURL string) bool {
	for _, sp := range store.scrapers[key] {
		if sp.targetURL() == targetURL {
			return true
		}
	}
	return false
}

func (store *scrapeStore) cleanDeadScraper(roleName string) int {
	var count int

	for key := range store.scrapers {
		var newScrapers []scraper
		var num int

		for _, sp := range store.scrapers[key] {
			if !sp.isTerminated() {
				newScrapers = append(newScrapers, sp)
				continue
			}
			num += 1
		}

		if num != 0 {
			store.scrapers[key] = newScrapers
			scraperNumberVec.WithLabelValues(roleName, key).Sub(float64(num))
			count += num
		}
	}

	return count
}

func (store *scrapeStore) scraperCount(key string) int {
	return len(store.scrapers[key])
}

func (store *scrapeStore) deleteKeyAndTerminateScrapers(key string) (cleanCount int) {
	delete(store.traits, key)

	for _, sp := range store.scrapers[key] {
		sp.markAsTerminated()
	}

	cleanCount = len(store.scrapers[key])
	delete(store.scrapers, key)

	return
}

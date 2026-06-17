// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testScraper struct {
	url        string
	scrapeFunc func(context.Context) error
	terminated atomic.Bool
}

func (s *testScraper) targetURL() string  { return s.url }
func (s *testScraper) shouldScrape() bool { return true }
func (s *testScraper) isTerminated() bool { return s.terminated.Load() }
func (s *testScraper) markAsTerminated()  { s.terminated.Store(true) }
func (s *testScraper) canScrape(time.Time) bool {
	return true
}
func (s *testScraper) recordFailure(time.Duration) (int, time.Time) {
	return 1, time.Now()
}
func (s *testScraper) resetRetryCount() {}
func (s *testScraper) scrape(ctx context.Context, _ int64) error {
	if s.scrapeFunc != nil {
		return s.scrapeFunc(ctx)
	}
	return nil
}

func TestScrapeManagerRegistrationDoesNotDropTasks(t *testing.T) {
	manager := newScrapeManager().(*scrapeManager)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager.runWorker(ctx, 1, time.Hour)

	const count = 100
	for i := 0; i < count; i++ {
		manager.registerScrape(
			RolePod,
			fmt.Sprintf("default/pod-%d", i),
			"traits",
			&testScraper{url: fmt.Sprintf("http://10.0.0.1:%d/metrics", i)},
		)
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()
	total := 0
	for _, scrapers := range manager.podStore.scrapers {
		total += len(scrapers)
	}
	require.Equal(t, count, total)
}

func TestScrapeBackoff(t *testing.T) {
	require.Equal(t, 10*time.Second, scrapeBackoff(10*time.Second, 1))
	require.Equal(t, 20*time.Second, scrapeBackoff(10*time.Second, 2))
	require.Equal(t, 40*time.Second, scrapeBackoff(10*time.Second, 3))
	require.Equal(t, maxScrapeBackoff, scrapeBackoff(10*time.Second, 20))

	require.True(t, shouldLogScrapeFailure(1))
	require.True(t, shouldLogScrapeFailure(2))
	require.False(t, shouldLogScrapeFailure(3))
	require.True(t, shouldLogScrapeFailure(4))
}

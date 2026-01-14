// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	ingestioncanary "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ingestion_canary"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

const (
	toolName            = "ingestion_canary"
	pollInterval        = 500 * time.Millisecond
	roundInterval       = 10 * time.Second
	errorRetries        = 10 // default error retry count
	defaultStorageIndex = "default"
)

type ingestionCanaryTool struct {
	canary       *ingestioncanary.Canary
	dw           *dataway.Dataway
	round        int64
	mutex        sync.Mutex
	storageIndex string
}

func runIngestionCanaryTool() error {
	if config.Cfg.Dataway == nil {
		return errors.New("dataway not configured")
	}
	if len(config.Cfg.Dataway.URLs) == 0 {
		return errors.New("dataway URLs not configured")
	}

	tool := &ingestionCanaryTool{}
	if flagToolIngestionCanaryIndex != nil {
		tool.storageIndex = *flagToolIngestionCanaryIndex
	}

	// Initialize canary
	tool.canary = ingestioncanary.New(
		ingestioncanary.WithName(toolName),
		ingestioncanary.WithTestType("cmd"),
	)

	// Initialize dataway (similar to import.go)
	tool.dw = dataway.NewDefaultDataway()
	tool.dw.URLs = config.Cfg.Dataway.URLs
	if err := tool.dw.Init(); err != nil {
		return fmt.Errorf("init dataway failed: %w", err)
	}
	if err := config.Cfg.Dataway.Init(); err != nil {
		return fmt.Errorf("init config dataway failed: %w", err)
	}

	// Run test rounds
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	for {
		tool.round++
		if err := tool.runTestRound(ctx); err != nil {
			return err
		}
		cp.Printf("> sleep %s...\n", roundInterval.String())

		select {
		case <-ctx.Done():
			cp.Printf("> interrupted, exiting...\n")
			return nil
		case <-time.After(roundInterval):
		}
	}
}

type categoryInfo struct {
	name     string
	category ingestioncanary.Category
	point    *point.Point
	cat      point.Category
}

func (t *ingestionCanaryTool) runTestRound(ctx context.Context) error {
	cp.Printf("> round(%d): generating probe data...\n", t.round)

	// Generate data points
	ts := time.Now().Truncate(time.Millisecond)
	metricPt := t.canary.Metric(ts, t.round, nil)
	loggingPt := t.canary.Logging(ts, t.round, nil)
	tracingPt := t.canary.Tracing(ts, t.round, nil)

	// Print generated data points
	cp.Printf("M: %s\n", metricPt.LineProto())
	cp.Printf("L: %s\n", loggingPt.LineProto())
	cp.Printf("T: %s\n", tracingPt.LineProto())

	cp.Printf("> round(%d): feeding data and querying...\n", t.round)

	// Feed and query each category concurrently

	categories := []categoryInfo{
		{"M", ingestioncanary.MetricCategory, metricPt, point.Metric},
		{"L", ingestioncanary.LoggingCategory, loggingPt, point.Logging},
		{"T", ingestioncanary.TracingCategory, tracingPt, point.Tracing},
	}

	var wg sync.WaitGroup
	latencies := make(map[string]time.Duration, len(categories))
	pointContents := make(map[string]string, len(categories))
	for _, cat := range categories {
		wg.Add(1)
		go func(c categoryInfo) {
			defer wg.Done()

			// Feed data
			if err := t.feedCategory(c.cat, c.point); err != nil {
				cp.Warnf("%s: feed failed: %s\n", c.name, err)
				return
			}

			// Query data
			latency, err := t.queryData(ctx, c.category, ts)
			if err != nil {
				cp.Warnf("%s: query failed: %s\n", c.name, err)
				return
			} else {
				t.mutex.Lock()
				latencies[c.name] = latency
				pointContents[c.name] = fmt.Sprintf("round:%d,feedTime:%d", t.round, ts.UnixMilli())
				t.mutex.Unlock()
			}
		}(cat)
	}
	wg.Wait()

	cp.Println("Point contents:")
	for _, cat := range categories {
		if content, ok := pointContents[cat.name]; ok {
			cp.Printf("%s: %s\n", cat.name, content)
		}
	}
	cp.Println("---")
	cp.Println("Latencies:")
	for _, cat := range categories {
		if latency, ok := latencies[cat.name]; ok {
			cp.Printf("%s: %.1fs\n", cat.name, latency.Seconds())
		}
	}
	cp.Printf("> round(%d): ended.\n", t.round)

	return nil
}

func (t *ingestionCanaryTool) feedCategory(cat point.Category, pt *point.Point) error {
	opts := []dataway.WriteOption{
		dataway.WithPoints([]*point.Point{pt}),
		dataway.WithCategory(cat),
		dataway.WithNoWAL(true),
		dataway.WithGzipDuringBuildBody(true),
		dataway.WithHTTPHeader("X-Sub-Category", toolName),
	}
	if t.storageIndex != "" && cat == point.Logging {
		opts = append(opts, dataway.WithStorageIndex(t.storageIndex))
	}
	return t.dw.Write(opts...)
}

// nolint: lll
func (t *ingestionCanaryTool) queryData(ctx context.Context, category ingestioncanary.Category, feedTime time.Time) (latency time.Duration, err error) {
	storageIndex := ""
	if category == ingestioncanary.LoggingCategory {
		storageIndex = t.storageIndex
		if storageIndex == "" {
			storageIndex = defaultStorageIndex
		}
	}

	errorCount := 0
	for {
		found, err := t.canary.CheckLast(category, &ingestioncanary.Feed{
			TimeMs: feedTime.UnixMilli(),
			Round:  t.round,
		}, storageIndex)
		if err != nil {
			errorCount++
			cp.Warnf("%s: query failed (retry %d/%d): %s\n", category, errorCount, errorRetries, err)
			if errorCount >= errorRetries {
				return 0, fmt.Errorf("query failed after %d retries: %w", errorRetries, err)
			}
		}

		if found {
			latency = time.Since(feedTime)
			return latency, nil
		}

		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("exited")
		case <-time.After(pollInterval):
		}
	}
}

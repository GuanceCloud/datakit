// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import (
	"context"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

type SD interface {
	SetLogger(log *logger.Logger)
	StartScraperProducer(ctx context.Context, cfg *ScrapeConfig, opts []promscrape.Option, out chan<- scraper)
}

type Auth struct {
	BearerTokenFile string `toml:"bearer_token_file"`
	*dknet.TLSClientConfig
}

type ScrapeConfig struct {
	Scheme              string            `toml:"scheme"`
	MetricsPath         string            `toml:"metrics_path"`
	Params              string            `toml:"params"`
	Interval            time.Duration     `toml:"interval"`
	KeepExistMetricName bool              `toml:"keep_exist_metric_name"`
	HTTPHeaders         map[string]string `toml:"http_headers"`
	Auth                *Auth             `toml:"auth"`
}

func startScraperConsumer(ctx context.Context, logger *logger.Logger, workerName string, scrapeInterval time.Duration, in <-chan scraper) {
	var scrapers []scraper
	start := ntp.Now()

	ticker := time.NewTicker(scrapeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Infof("%s: worker stopped", workerName)
			return

		case sp, ok := <-in:
			if !ok {
				logger.Warnf("%s: scraper channel closed, exiting", workerName)
				return
			}
			if len(scrapers) > maxScrapersPerWorker {
				logger.Warnf("%s: scraper count exceeds limit (%d), dropping new scraper", workerName, maxScrapersPerWorker)
			} else {
				scrapers = append(scrapers, sp)
				logger.Debugf("%s: added new scraper, total: %d", workerName, len(scrapers))
			}

		case tt := <-ticker.C:
			var activeScrapers []scraper
			for _, sp := range scrapers {
				if !sp.isTerminated() {
					activeScrapers = append(activeScrapers, sp)
				} else {
					logger.Debugf("%s: removing terminated scraper: %s", workerName, sp.targetURL())
				}
			}

			start = inputs.AlignTime(tt, start, scrapeInterval)
			for _, sp := range activeScrapers {
				if err := sp.scrape(start.UnixNano()); err != nil {
					logger.Warnf("%s: scrape failed for %s: %s", workerName, sp.targetURL(), err)
				}
			}

			if len(scrapers) != len(activeScrapers) {
				logger.Infof("%s: cleaned up terminated scrapers (%d -> %d)", workerName, len(scrapers), len(activeScrapers))
			}
			scrapers = activeScrapers
		}
	}
}

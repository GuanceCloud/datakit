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
	Interval            time.Duration     `toml:"interval"`
	KeepExistMetricName bool              `toml:"keep_exist_metric_name"`
	Auth                *Auth             `toml:"auth"`
	HTTPHeaders         map[string]string `toml:"http_headers"`
}

func startScraperConsumer(ctx context.Context, log *logger.Logger, name string, scrapeInterval time.Duration, in <-chan scraper) {
	var scrapers []scraper
	start := ntp.Now()

	ticker := time.NewTicker(scrapeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case sp, ok := <-in:
			if !ok {
				log.Warnf("%s: channel is closed, exit", name)
				return
			}
			if len(scrapers) > maxScrapersPerWorker {
				log.Warnf("%s: scrapers is over limit %d", name, maxScrapersPerWorker)
			} else {
				scrapers = append(scrapers, sp)
			}

		case tt := <-ticker.C:
			var activeScrapers []scraper
			for _, sp := range scrapers {
				if !sp.isTerminated() {
					activeScrapers = append(activeScrapers, sp)
				}
			}

			start = inputs.AlignTime(tt, start, scrapeInterval)
			for _, sp := range activeScrapers {
				if err := sp.scrape(start.UnixNano()); err != nil {
					log.Warnf("%s: failed of scrape, err: %s", name, err)
				}
			}

			log.Infof("%s: removed terminated scrapers, count(%d-%d)", name, len(scrapers), len(activeScrapers))
			scrapers = activeScrapers
		}
	}
}

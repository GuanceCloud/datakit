// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import (
	"context"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

type ConsulSD struct {
	Auth *Auth `toml:"auth"`
	log  *logger.Logger
}

func (sd *ConsulSD) SetLogger(log *logger.Logger) { sd.log = log }

func (sd *ConsulSD) StartScraperProducer(ctx context.Context, cfg *ScrapeConfig, opts []promscrape.Option, out chan<- scraper) {
	// nil
}

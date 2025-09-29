// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

type HTTPSD struct {
	ServiceURL      string        `toml:"service_url"`
	RefreshInterval time.Duration `toml:"refresh_interval"`
	Auth            *Auth         `toml:"auth"`

	targetGroups TargetGroups
	tasks        []scraper
	log          *logger.Logger
}

func (sd *HTTPSD) SetLogger(log *logger.Logger) { sd.log = log }

func (sd *HTTPSD) StartScraperProducer(ctx context.Context, cfg *ScrapeConfig, opts []promscrape.Option, out chan<- scraper) {
	sd.log.Infof("http_sd: start %s", sd.ServiceURL)

	ticker := time.NewTicker(sd.RefreshInterval)
	defer ticker.Stop()

	for {
		if err := sd.produceScrapers(ctx, cfg, opts, out); err != nil {
			sd.log.Warnf("http_sd: failed of produce scrapers, err: %s", err)
		}

		select {
		case <-ctx.Done():
			sd.terminatedTasks()
			sd.log.Info("http_sd: terminating all tasks and exitting")
			return

		case <-ticker.C:
			// next
		}
	}
}

func (sd *HTTPSD) produceScrapers(ctx context.Context, cfg *ScrapeConfig, opts []promscrape.Option, out chan<- scraper) error {
	newTargetGroups, err := sd.discoveryTargetGroups()
	if err != nil {
		return err
	}

	if !sd.targetGroupsChanged(newTargetGroups) {
		sd.log.Debugf("http_sd: targetGroups unchanged")
		return nil
	}

	scrapers, err := convertTargetGroupsToScraper(cfg, opts, newTargetGroups)
	if err != nil {
		return err
	}

	for _, scraper := range scrapers {
		if ctx.Err() != nil {
			return err
		}

		select {
		case out <- scraper:
			// next
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	sd.terminatedTasks()
	sd.targetGroups = newTargetGroups
	sd.tasks = scrapers
	sd.log.Infof("http_sd: found new targetGroups and replaced")
	return nil
}

func (sd *HTTPSD) discoveryTargetGroups() (TargetGroups, error) {
	req, err := http.NewRequest("GET", sd.ServiceURL, nil)
	if err != nil {
		return nil, err
	}

	if sd.Auth != nil && sd.Auth.BearerTokenFile != "" {
		token, err := os.ReadFile(sd.Auth.BearerTokenFile)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+string(bytes.TrimSpace(token)))
	}

	clientOpts := httpcli.NewOptions()

	if sd.Auth != nil && sd.Auth.TLSClientConfig != nil {
		config, err := sd.Auth.TLSClientConfig.TLSConfig()
		if err != nil {
			return nil, err
		}
		clientOpts.TLSClientConfig = config
	}

	client := httpcli.Cli(clientOpts)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code returned when get %q: %d", sd.ServiceURL, resp.StatusCode)
	}
	defer resp.Body.Close() // nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var groups TargetGroups
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (sd *HTTPSD) terminatedTasks() {
	for _, task := range sd.tasks {
		task.markAsTerminated()
	}
}

func (sd *HTTPSD) targetGroupsChanged(newTargetGroups TargetGroups) bool {
	return !reflect.DeepEqual(sd.targetGroups, newTargetGroups)
}

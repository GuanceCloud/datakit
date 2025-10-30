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
	ServiceURL      string            `toml:"service_url"`
	RefreshInterval time.Duration     `toml:"refresh_interval"`
	HTTPHeaders     map[string]string `toml:"http_headers"`
	Auth            *Auth             `toml:"auth"`

	targetGroups TargetGroups
	tasks        []scraper
	logger       *logger.Logger
}

func (sd *HTTPSD) SetLogger(logger *logger.Logger) { sd.logger = logger }

func (sd *HTTPSD) StartScraperProducer(ctx context.Context, cfg *ScrapeConfig, opts []promscrape.Option, out chan<- scraper) {
	sd.logger.Infof("http_sd: starting service discovery for %s", sd.ServiceURL)

	ticker := time.NewTicker(sd.RefreshInterval)
	defer ticker.Stop()

	for {
		if err := sd.produceScrapers(ctx, cfg, opts, out); err != nil {
			sd.logger.Warnf("http_sd: failed to produce scrapers: %s", err)
		}

		select {
		case <-ctx.Done():
			sd.terminateTasks()
			sd.logger.Info("http_sd: terminating all tasks and exiting")
			return

		case <-ticker.C:
		}
	}
}

func (sd *HTTPSD) produceScrapers(ctx context.Context, cfg *ScrapeConfig, opts []promscrape.Option, out chan<- scraper) error {
	newTargetGroups, err := sd.discoverTargetGroups()
	if err != nil {
		return err
	}

	if !sd.targetGroupsChanged(newTargetGroups) {
		sd.logger.Debugf("http_sd: target groups unchanged")
		return nil
	}

	scrapers, err := convertTargetGroupsToScraper(cfg, opts, newTargetGroups)
	if err != nil {
		return err
	}

	for _, scraper := range scrapers {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		select {
		case out <- scraper:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	sd.terminateTasks()
	sd.targetGroups = newTargetGroups
	sd.tasks = scrapers
	sd.logger.Infof("http_sd: updated target groups, found %d new scrapers", len(scrapers))
	return nil
}

func (sd *HTTPSD) discoverTargetGroups() (TargetGroups, error) {
	req, err := http.NewRequest("GET", sd.ServiceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range sd.HTTPHeaders {
		req.Header.Set(k, v)
	}

	if err := sd.addAuthToRequest(req); err != nil {
		return nil, err
	}

	clientOpts := httpcli.NewOptions()
	if err := sd.configureTLS(clientOpts); err != nil {
		return nil, err
	}

	client := httpcli.Cli(clientOpts)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, sd.ServiceURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var groups TargetGroups
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	sd.logger.Debugf("http_sd: discovered %d target groups", len(groups))
	return groups, nil
}

func (sd *HTTPSD) addAuthToRequest(req *http.Request) error {
	if sd.Auth == nil || sd.Auth.BearerTokenFile == "" {
		return nil
	}

	token, err := os.ReadFile(sd.Auth.BearerTokenFile)
	if err != nil {
		return fmt.Errorf("failed to read bearer token file: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+string(bytes.TrimSpace(token)))
	return nil
}

func (sd *HTTPSD) configureTLS(clientOpts *httpcli.Options) error {
	if sd.Auth == nil || sd.Auth.TLSClientConfig == nil {
		return nil
	}

	config, err := sd.Auth.TLSClientConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("failed to create TLS config: %w", err)
	}

	clientOpts.TLSClientConfig = config
	return nil
}

func (sd *HTTPSD) terminateTasks() {
	for _, task := range sd.tasks {
		task.markAsTerminated()
	}
}

func (sd *HTTPSD) targetGroupsChanged(newTargetGroups TargetGroups) bool {
	return !reflect.DeepEqual(sd.targetGroups, newTargetGroups)
}

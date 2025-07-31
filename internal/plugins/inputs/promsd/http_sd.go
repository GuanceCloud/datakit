// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

type HTTPSD struct {
	ServiceURL      string        `toml:"service_url"`
	RefreshInterval time.Duration `toml:"refresh_interval"`
	Auth            *Auth         `toml:"auth"`

	targetGroups HTTPSDTargetGroups
	tasks        []scraper
	log          *logger.Logger
}

func (sd *HTTPSD) SetLogger(log *logger.Logger) { sd.log = log }

func (sd *HTTPSD) StartScraperProducer(ctx context.Context, cfg *ScrapeConfig, opts []promscrape.Option, out chan<- scraper) {
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
	newTargetGroups, err := sd.fetchNewTargetGroups()
	if err != nil {
		return err
	}

	if !sd.targetGroupsChanged(newTargetGroups) {
		sd.log.Debugf("http_sd: targetGroups unchanged")
		return nil
	}

	scrapers, err := sd.convertTargetGroupsToScraper(cfg, opts, newTargetGroups)
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
	sd.tasks = scrapers
	sd.log.Infof("http_sd: found new targetGroups and replaced")
	return nil
}

func (sd *HTTPSD) fetchNewTargetGroups() (HTTPSDTargetGroups, error) {
	clientOpts := httpcli.NewOptions()

	if sd.Auth.TLSClientConfig != nil {
		config, err := sd.Auth.TLSClientConfig.TLSConfig()
		if err != nil {
			return nil, err
		}
		clientOpts.TLSClientConfig = config
	}

	client := httpcli.Cli(clientOpts)
	resp, err := client.Get(sd.ServiceURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code returned when get %q: %d", sd.ServiceURL, resp.StatusCode)
	}
	defer resp.Body.Close() //nolint

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var groups HTTPSDTargetGroups
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (sd *HTTPSD) convertTargetGroupsToScraper(cfg *ScrapeConfig, opts []promscrape.Option, newTargetGroups HTTPSDTargetGroups) ([]scraper, error) {
	var scrapers []scraper

	for _, group := range newTargetGroups {
		var urls []string

		scheme := extractScrapeSchemeFromHTTPSDLabels(group.Labels)
		if scheme == "" {
			scheme = cfg.Scheme
		}
		path := extractScrapeMetricsPathFromHTTPSDLabels(group.Labels)
		params := extractScrapeParamsFromHTTPSDLabels(group.Labels)
		paramsQuery := url.Values(params).Encode()

		for _, target := range group.Targets {
			u := &url.URL{
				Scheme:   scheme,
				Host:     target,
				Path:     path,
				RawQuery: paramsQuery,
			}
			urls = append(urls, u.String())
		}

		for _, u := range urls {
			scraper, err := newPromScraper(u, append(opts, promscrape.WithExtraTags(group.Labels)))
			if err != nil {
				return nil, err
			}
			scrapers = append(scrapers, scraper)
		}
	}

	return scrapers, nil
}

func (sd *HTTPSD) terminatedTasks() {
	for _, task := range sd.tasks {
		task.markAsTerminated()
	}
}

func (sd *HTTPSD) targetGroupsChanged(newTargetGroups HTTPSDTargetGroups) bool {
	return !reflect.DeepEqual(sd.targetGroups, newTargetGroups)
}

type HTTPSDTargetGroups []HTTPSDTargetGroup

type HTTPSDTargetGroup struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func extractScrapeSchemeFromHTTPSDLabels(labels map[string]string) string {
	scheme := labels["__scheme__"]
	s := strings.ToLower(scheme)

	if s == "" || (s != "http" && s != "https") {
		return ""
	}
	return s
}

func extractScrapeMetricsPathFromHTTPSDLabels(labels map[string]string) string {
	path := labels["__metrics_path__"]
	if path == "" {
		return "/metrics"
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

func extractScrapeParamsFromHTTPSDLabels(labels map[string]string) map[string][]string {
	params := make(map[string][]string)

	for key, value := range labels {
		if !strings.HasPrefix(key, "__param_") {
			continue
		}
		paramName := strings.TrimPrefix(key, "__param_")
		if paramName == "" {
			continue
		}
		params[paramName] = append(params[paramName], value)
	}
	return params
}

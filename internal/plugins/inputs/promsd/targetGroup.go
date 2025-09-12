// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import (
	"fmt"
	"net/url"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

type TargetGroups []TargetGroup

type TargetGroup struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func convertTargetGroupsToScraper(cfg *ScrapeConfig, opts []promscrape.Option, newTargetGroups TargetGroups) ([]scraper, error) {
	var scrapers []scraper

	for _, group := range newTargetGroups {
		var urls []string

		scheme := extractScrapeSchemeFromHTTPSDLabels(group.Labels)
		if scheme == "" {
			scheme = cfg.Scheme
		}

		path := extractScrapeMetricsPathFromHTTPSDLabels(group.Labels)
		if path == "" {
			path = cfg.MetricsPath
		}

		params := extractScrapeParamsFromHTTPSDLabels(group.Labels)
		paramValues := url.Values(params)
		if values, err := url.ParseQuery(cfg.Params); err != nil {
			return nil, fmt.Errorf("unexpected scrape params: %s", cfg.Params)
		} else {
			for k, valueSlice := range values {
				for _, value := range valueSlice {
					paramValues.Add(k, value)
				}
			}
		}

		for _, target := range group.Targets {
			u := &url.URL{
				Scheme:   scheme,
				Host:     target,
				Path:     path,
				RawQuery: paramValues.Encode(),
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
	if path != "" && !strings.HasPrefix(path, "/") {
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

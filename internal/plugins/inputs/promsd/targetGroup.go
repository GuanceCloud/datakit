// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import (
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
		groupScrapers, err := buildScrapersFromGroup(cfg, opts, group)
		if err != nil {
			return nil, err
		}
		scrapers = append(scrapers, groupScrapers...)
	}

	return scrapers, nil
}

func buildScrapersFromGroup(cfg *ScrapeConfig, opts []promscrape.Option, group TargetGroup) ([]scraper, error) {
	var scrapers []scraper

	scheme := extractSchemeFromLabels(group.Labels, cfg.Scheme)
	path := extractMetricsPathFromLabels(group.Labels, cfg.MetricsPath)
	params := mergeParamsFromLabelsAndConfig(group.Labels, cfg.Params)

	for _, target := range group.Targets {
		url := buildScrapeURL(scheme, target, path, params)
		scraper, err := newPromScraper(url, append(opts, promscrape.WithExtraTags(group.Labels)))
		if err != nil {
			return nil, err
		}
		scrapers = append(scrapers, scraper)
	}

	return scrapers, nil
}

func buildScrapeURL(scheme, target, path string, params url.Values) string {
	u := &url.URL{
		Scheme:   scheme,
		Host:     target,
		Path:     path,
		RawQuery: params.Encode(),
	}
	return u.String()
}

func extractSchemeFromLabels(labels map[string]string, defaultScheme string) string {
	scheme := labels["__scheme__"]
	s := strings.ToLower(scheme)

	if s == "" || (s != "http" && s != "https") {
		return defaultScheme
	}
	return s
}

func extractMetricsPathFromLabels(labels map[string]string, defaultPath string) string {
	path := labels["__metrics_path__"]
	if path != "" && !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	if path == "" {
		return defaultPath
	}
	return path
}

func mergeParamsFromLabelsAndConfig(labels map[string]string, configParams string) url.Values {
	params := extractParamsFromLabels(labels)
	paramValues := url.Values(params)

	if values, err := url.ParseQuery(configParams); err == nil {
		for k, valueSlice := range values {
			for _, value := range valueSlice {
				paramValues.Add(k, value)
			}
		}
	}

	return paramValues
}

func extractParamsFromLabels(labels map[string]string) map[string][]string {
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

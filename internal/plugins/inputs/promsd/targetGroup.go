// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/prometheus/common/model"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

type TargetGroups []TargetGroup

type TargetGroup struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

type preparedTarget struct {
	URL  string
	Tags map[string]string
}

func convertTargetGroupsToScraper(
	cfg *ScrapeConfig,
	opts []promscrape.Option,
	newTargetGroups TargetGroups,
	log *logger.Logger,
) ([]scraper, error) {
	var scrapers []scraper

	for _, group := range newTargetGroups {
		groupScrapers, err := buildScrapersFromGroup(cfg, opts, group, log)
		if err != nil {
			return nil, err
		}
		scrapers = append(scrapers, groupScrapers...)
	}

	return scrapers, nil
}

func buildScrapersFromGroup(
	cfg *ScrapeConfig,
	opts []promscrape.Option,
	group TargetGroup,
	log *logger.Logger,
) ([]scraper, error) {
	var scrapers []scraper

	if len(cfg.RelabelConfigs) > 0 {
		return buildRelabeledScrapersFromGroup(cfg, opts, group, log)
	}

	scheme := extractSchemeFromLabels(group.Labels, cfg.Scheme)
	path := extractMetricsPathFromLabels(group.Labels, cfg.MetricsPath)
	params := mergeParamsFromLabelsAndConfig(group.Labels, cfg.Params)

	for _, target := range group.Targets {
		url := buildScrapeURL(scheme, target, path, params)
		targetOpts := append([]promscrape.Option{}, opts...)
		targetOpts = append(targetOpts, promscrape.WithExtraTags(group.Labels))
		scraper, err := newPromScraper(url, targetOpts)
		if err != nil {
			return nil, err
		}
		scrapers = append(scrapers, scraper)
	}

	return scrapers, nil
}

func buildRelabeledScrapersFromGroup(
	cfg *ScrapeConfig,
	opts []promscrape.Option,
	group TargetGroup,
	log *logger.Logger,
) ([]scraper, error) {
	configParams, err := url.ParseQuery(cfg.Params)
	if err != nil {
		return nil, fmt.Errorf("parse scrape params %q: %w", cfg.Params, err)
	}

	var scrapers []scraper

	for _, address := range group.Targets {
		target, keep, err := prepareTarget(cfg, address, group.Labels, configParams)
		if err != nil {
			if log != nil {
				log.Warnf("skip target %q: failed to prepare after relabeling: %s", address, err)
			}
			continue
		}
		if !keep {
			continue
		}

		scraper, err := newPromScraperWithTags(target.URL, opts, target.Tags)
		if err != nil {
			return nil, fmt.Errorf("create scraper for target %q: %w", address, err)
		}
		scrapers = append(scrapers, scraper)
	}

	return scrapers, nil
}

func prepareTarget(
	cfg *ScrapeConfig,
	address string,
	targetLabels map[string]string,
	configParams url.Values,
) (preparedTarget, bool, error) {
	labels := newRelabelLabels(targetLabels)
	labels.Set("__address__", address)
	if labels.Get("__scheme__") == "" {
		labels.Set("__scheme__", cfg.Scheme)
	}
	if labels.Get("__metrics_path__") == "" {
		labels.Set("__metrics_path__", cfg.MetricsPath)
	}

	for key, values := range configParams {
		labelName := "__param_" + key
		if len(values) > 0 && labels.Get(labelName) == "" {
			labels.Set(labelName, values[0])
		}
	}

	if !applyRelabelConfigs(labels, cfg.compiledRelabelConfigs) {
		return preparedTarget{}, false, nil
	}

	targetAddress := labels.Get("__address__")
	if targetAddress == "" {
		return preparedTarget{}, false, fmt.Errorf("target address is empty after relabeling")
	}
	if err := validateRelabeledLabels(labels); err != nil {
		return preparedTarget{}, false, err
	}
	scheme := labels.Get("__scheme__")
	path := labels.Get("__metrics_path__")
	urlstr := buildScrapeURL(scheme, targetAddress, path, buildRelabeledTargetParams(labels, configParams))

	tags := make(map[string]string)
	for key, value := range labels {
		if !strings.HasPrefix(key, "__") {
			tags[key] = value
		}
	}
	return preparedTarget{URL: urlstr, Tags: tags}, true, nil
}

func validateRelabeledLabels(labels map[string]string) error {
	for name, value := range labels {
		if !model.LabelValue(value).IsValid() {
			return fmt.Errorf("invalid label value for %q after relabeling", name)
		}
	}
	return nil
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

func buildRelabeledTargetParams(labels map[string]string, configParams url.Values) url.Values {
	params := make(url.Values, len(configParams))
	for key, values := range configParams {
		params[key] = append([]string(nil), values...)
	}

	for key, value := range labels {
		if !strings.HasPrefix(key, "__param_") {
			continue
		}
		paramName := strings.TrimPrefix(key, "__param_")
		if paramName == "" {
			continue
		}
		if len(params[paramName]) > 0 {
			params[paramName][0] = value
		} else {
			params.Set(paramName, value)
		}
	}
	return params
}

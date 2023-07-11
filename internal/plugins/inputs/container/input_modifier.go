// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"regexp"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/multiline"
)

func (i *Input) setLoggingAutoMultilineToLogConfigs(configs logConfigs) {
	if !i.LoggingAutoMultilineDetection {
		return
	}
	for _, cfg := range configs {
		if cfg.Multiline != "" {
			cfg.MultilinePatterns = []string{cfg.Multiline}
		} else {
			if len(i.LoggingAutoMultilineExtraPatterns) != 0 {
				cfg.MultilinePatterns = i.LoggingAutoMultilineExtraPatterns
			} else {
				cfg.MultilinePatterns = multiline.GlobalPatterns
			}
		}
	}
}

func (i *Input) setLoggingExtraSourceMapToLogConfigs(configs logConfigs) {
	// gitlab #903
	for re, newSource := range i.LoggingExtraSourceMap {
		for _, cfg := range configs {
			match, err := regexp.MatchString(re, cfg.Source)
			if err != nil {
				l.Warnf("invalid global_extra_source_map '%s', err %s, skip", re, err)
			}
			if match {
				l.Debugf("replaced source '%s' with '%s'", cfg.Source, newSource)
				cfg.Source = newSource
				break
			}
		}
	}
}

func (i *Input) setLoggingSourceMultilineMapToLogConfigs(configs logConfigs) {
	if len(i.LoggingSourceMultilineMap) == 0 {
		return
	}
	for _, cfg := range configs {
		if cfg.Multiline != "" {
			continue
		}

		source := cfg.Source
		mult := i.LoggingSourceMultilineMap[source]
		if mult != "" {
			l.Debugf("replaced multiline '%s' with '%s' to source %s", cfg.Multiline, mult, source)
			cfg.Multiline = mult
		}
	}
}

func (i *Input) setExtractK8sLabelAsTagsToLogConfigs(configs logConfigs, labels map[string]string) {
	if !i.ExtractK8sLabelAsTags {
		return
	}

	for _, cfg := range configs {
		if cfg.Tags == nil {
			cfg.Tags = make(map[string]string)
		}
		for k, v := range labels {
			if _, ok := cfg.Tags[k]; !ok {
				// replace dot
				k := strings.ReplaceAll(k, ".", "_")
				cfg.Tags[k] = v
			}
		}
	}
}

func (i *Input) setGlobalTagsToLogConfigs(configs logConfigs) {
	i.setTagsToLogConfigs(configs, i.Tags)
}

func (i *Input) setTagsToLogConfigs(configs logConfigs, m map[string]string) {
	for _, cfg := range configs {
		if cfg.Tags == nil {
			cfg.Tags = make(map[string]string)
		}
		for k, v := range m {
			if _, ok := cfg.Tags[k]; !ok {
				cfg.Tags[k] = v
			}
		}
	}
}

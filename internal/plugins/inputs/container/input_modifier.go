// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"regexp"

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

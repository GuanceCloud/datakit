// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package mapper contains graphite mapper default config.
package mapper

import "time"

type MapperConfigDefaults struct {
	ObserverType        ObserverType     `toml:"observer_type"`
	MatchType           MatchType        `toml:"match_type"`
	GlobDisableOrdering bool             `toml:"glob_disable_ordering"`
	TTL                 time.Duration    `toml:"ttl"`
	SummaryOptions      SummaryOptions   `toml:"summary_options"`
	HistogramOptions    HistogramOptions `toml:"histogram_options"`
}

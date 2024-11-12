// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package mapper contains graphite mapper
package mapper

import (
	"regexp"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/graphite/mapper/fsm"
)

type Labels map[string]string

type MetricMapping struct {
	Match           string `toml:"match"`
	Name            string `toml:"name"`
	MeasurementName string `toml:"measurement_name"`
	nameFormatter   *fsm.TemplateFormatter
	regex           *regexp.Regexp
	Labels          Labels `toml:"labels"`
	HonorLabels     bool   `toml:"honor_labels"`
	labelKeys       []string
	labelFormatters []*fsm.TemplateFormatter
	ObserverType    ObserverType      `toml:"observer_type"`
	LegacyBuckets   []float64         `toml:"buckets"`
	LegacyQuantiles []MetricObjective `toml:"quantiles"`

	MatchType        MatchType         `toml:"match_type"`
	Action           ActionType        `toml:"action_type"`
	MatchMetricType  MetricType        `toml:"match_metric_type"`
	TTL              time.Duration     `toml:"ttl"`
	SummaryOptions   *SummaryOptions   `toml:"summary_options"`
	HistogramOptions *HistogramOptions `toml:"histogram_options"`
	Sacle            MaybeFloat64      `toml:"scale"`
}

type MatchType string

const (
	MatchTypeGlob    MatchType = "glob"
	MatchTypeRegex   MatchType = "regex"
	MatchTypeDefault MatchType = ""
)

type MaybeFloat64 struct {
	Set bool
	Val float64
}

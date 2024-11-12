// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package mapper contains graphite mapper
package mapper

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/graphite/mapper/fsm"
)

type MetricMapper struct {
	Name      string `toml:"name"`
	Defaults  MapperConfigDefaults
	Mappings  []MetricMapping `toml:"mappings"`
	FSM       *fsm.FSM
	CacheSize int       `toml:"cache_size"`
	CacheType CacheType `toml:"cache_type"`
	doFSM     bool
	doRegex   bool
	mu        sync.RWMutex
	cache     MetricMapperCache

	logger *logger.Logger
}

type SummaryOptions struct {
	Quantiles  []MetricObjective `toml:"quantiles"`
	MaxAge     time.Duration     `toml:"max_age"`
	AgeBuckets uint32            `toml:"age_buckets"`
	BufCap     uint32            `toml:"buf_cap"`
}

type HistogramOptions struct {
	Buckets                     []float64 `toml:"buckets"`
	NativeHistogramBucketFactor float64   `toml:"native_histogram_bucket_factor"`
	NativeHistogramMaxBuckets   uint32    `toml:"native_histogram_max_buckets"`
}

type MetricObjective struct {
	Quantile float64 `toml:"quantile"`
	Error    float64 `toml:"error"`
}

var defaultQuantiles = []MetricObjective{
	{Quantile: 0.5, Error: 0.05},
	{Quantile: 0.9, Error: 0.01},
	{Quantile: 0.99, Error: 0.001},
}

func (m *MetricMapper) GetMapping(metric string, metricType MetricType) (*MetricMapping, Labels, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// only use a cache if one is present
	if m.cache != nil {
		result, cached := m.cache.Get(formatKey(metric, metricType))
		if cached {
			r := result.(MetricMapperCacheResult)
			return r.Mapping, r.Labels, r.Matched
		}
	}

	// global matching
	if m.doFSM {
		finalState, captures := m.FSM.GetMapping(metric, string(metricType))
		if finalState != nil && finalState.Result != nil {
			v := finalState.Result.(*MetricMapping)
			result := copyMetricMapping(v)
			result.Name = result.nameFormatter.Format(captures)

			labels := Labels{}
			for index, formatter := range result.labelFormatters {
				labels[result.labelKeys[index]] = formatter.Format(captures)
			}

			r := MetricMapperCacheResult{
				Mapping: result,
				Matched: true,
				Labels:  labels,
			}
			// add match to cache
			if m.cache != nil {
				m.cache.Add(formatKey(metric, metricType), r)
			}

			return result, labels, true
		} else if !m.doRegex {
			// if there's no regex match type, return immediately
			// Add miss to cache
			if m.cache != nil {
				m.cache.Add(formatKey(metric, metricType), MetricMapperCacheResult{})
			}
			return nil, nil, false
		}
	}

	// regex matching
	for _, mapping := range m.Mappings {
		mapp := mapping
		// if a rule don't have regex matching type, the regex field is unset
		if mapp.regex == nil {
			continue
		}
		matches := mapp.regex.FindStringSubmatchIndex(metric)
		if len(matches) == 0 {
			continue
		}

		mapp.Name = string(mapp.regex.ExpandString(
			[]byte{},
			mapping.Name,
			metric,
			matches,
		))

		if mt := mapp.MatchMetricType; mt != "" && mt != metricType {
			continue
		}

		labels := Labels{}
		for label, valueExpr := range mapp.Labels {
			value := mapp.regex.ExpandString([]byte{}, valueExpr, metric, matches)
			labels[label] = string(value)
		}

		r := MetricMapperCacheResult{
			Mapping: &mapp,
			Matched: true,
			Labels:  labels,
		}
		// Add Match to cache
		if m.cache != nil {
			m.cache.Add(formatKey(metric, metricType), r)
		}

		return &mapp, labels, true
	}

	// Add Miss to cache
	if m.cache != nil {
		m.cache.Add(formatKey(metric, metricType), MetricMapperCacheResult{})
	}
	return nil, nil, false
}

func (m *MetricMapper) MapperSetup() error {
	remainingMappingsCount := len(m.Mappings)

	if len(m.Defaults.HistogramOptions.Buckets) == 0 {
		m.Defaults.HistogramOptions.Buckets = prometheus.DefBuckets
	}
	if m.Defaults.HistogramOptions.NativeHistogramBucketFactor == 0 {
		m.Defaults.HistogramOptions.NativeHistogramBucketFactor = 1.1
	}
	if m.Defaults.HistogramOptions.NativeHistogramMaxBuckets <= 0 {
		m.Defaults.HistogramOptions.NativeHistogramMaxBuckets = 256
	}

	if len(m.Defaults.SummaryOptions.Quantiles) == 0 {
		m.Defaults.SummaryOptions.Quantiles = defaultQuantiles
	}

	if m.Defaults.MatchType == MatchTypeDefault {
		m.Defaults.MatchType = MatchTypeGlob
	}

	if m.CacheType == "" {
		m.CacheType = CacheLRU
	}

	if m.CacheSize == 0 {
		m.CacheSize = DefaultCacheSize
	}

	m.FSM = fsm.NewFSM([]string{string(MetricTypeCounter), string(MetricTypeGauge), string(MetricTypeObserver)},
		remainingMappingsCount, false)

	for i := range m.Mappings {
		currentMapping := &m.Mappings[i]

		// check that label is correct
		for k := range currentMapping.Labels {
			if !labelNameRE.MatchString(k) {
				return fmt.Errorf("invalid label key: %s", k)
			}
		}

		if currentMapping.MeasurementName == "" {
			currentMapping.MeasurementName = defaultMeasurement
			m.logger.Infof("line %d: metric mapping didn't set a measurement name, default to graphite", i)
		}

		if currentMapping.Name == "" {
			return fmt.Errorf("line %d: metric mapping didn't set a metric name", i)
		}

		if !metricNameRE.MatchString(currentMapping.Name) {
			return fmt.Errorf("metric name '%s' doesn't match regex '%s'", currentMapping.Name, metricNameRE)
		}

		if currentMapping.MatchType == "" {
			currentMapping.MatchType = MatchTypeGlob
		}

		if currentMapping.Action == "" {
			currentMapping.Action = ActionTypeMap
		}

		if currentMapping.MatchType == MatchTypeGlob {
			m.doFSM = true
			if !metricLineRE.MatchString(currentMapping.Match) {
				return fmt.Errorf("invalid match: %s", currentMapping.Match)
			}

			captureCount := m.FSM.AddState(currentMapping.Match, string(currentMapping.MatchMetricType),
				remainingMappingsCount, currentMapping)

			currentMapping.nameFormatter = fsm.NewTemplateFormatter(currentMapping.Name, captureCount)

			labelKeys := make([]string, len(currentMapping.Labels))
			labelFormatters := make([]*fsm.TemplateFormatter, len(currentMapping.Labels))
			labelIndex := 0
			for label, valueExpr := range currentMapping.Labels {
				labelKeys[labelIndex] = label
				labelFormatters[labelIndex] = fsm.NewTemplateFormatter(valueExpr, captureCount)
				labelIndex++
			}
			currentMapping.labelFormatters = labelFormatters
			currentMapping.labelKeys = labelKeys
		} else {
			if regex, err := regexp.Compile(currentMapping.Match); err != nil {
				return fmt.Errorf("invalid regex %s in mapping: %w", currentMapping.Match, err)
			} else {
				currentMapping.regex = regex
			}
			m.doRegex = true
		}

		if currentMapping.SummaryOptions != nil &&
			currentMapping.LegacyQuantiles != nil &&
			currentMapping.SummaryOptions.Quantiles != nil {
			return fmt.Errorf("cannot use quantiles in both the top level and summary options at the same time in %s", currentMapping.Match)
		}

		if currentMapping.HistogramOptions != nil &&
			currentMapping.LegacyBuckets != nil &&
			currentMapping.HistogramOptions.Buckets != nil {
			return fmt.Errorf("cannot use buckets in both the top level and histogram options at the same time in %s", currentMapping.Match)
		}

		if currentMapping.ObserverType == ObserverTypeHistogram {
			if currentMapping.SummaryOptions != nil {
				return fmt.Errorf("cannot use histogram observer and summary options at the same time")
			}
			if currentMapping.HistogramOptions == nil {
				currentMapping.HistogramOptions = &HistogramOptions{}
			}
			if currentMapping.LegacyBuckets != nil && len(currentMapping.LegacyBuckets) != 0 {
				currentMapping.HistogramOptions.Buckets = currentMapping.LegacyBuckets
			}
			if currentMapping.HistogramOptions.Buckets == nil || len(currentMapping.HistogramOptions.Buckets) == 0 {
				currentMapping.HistogramOptions.Buckets = m.Defaults.HistogramOptions.Buckets
			}
		}

		if currentMapping.ObserverType == ObserverTypeSummary {
			if currentMapping.HistogramOptions != nil {
				return fmt.Errorf("cannot use summary observer and histogram options at the same time")
			}
			if currentMapping.SummaryOptions == nil {
				currentMapping.SummaryOptions = &SummaryOptions{}
			}
			if currentMapping.LegacyQuantiles != nil && len(currentMapping.LegacyQuantiles) != 0 {
				currentMapping.SummaryOptions.Quantiles = currentMapping.LegacyQuantiles
			}
			if currentMapping.SummaryOptions.Quantiles == nil || len(currentMapping.SummaryOptions.Quantiles) == 0 {
				currentMapping.SummaryOptions.Quantiles = m.Defaults.SummaryOptions.Quantiles
			}
			if currentMapping.SummaryOptions.MaxAge == 0 {
				currentMapping.SummaryOptions.MaxAge = m.Defaults.SummaryOptions.MaxAge
			}
			if currentMapping.SummaryOptions.AgeBuckets == 0 {
				currentMapping.SummaryOptions.AgeBuckets = m.Defaults.SummaryOptions.AgeBuckets
			}
			if currentMapping.SummaryOptions.BufCap == 0 {
				currentMapping.SummaryOptions.BufCap = m.Defaults.SummaryOptions.BufCap
			}
		}
	}

	if m.cache == nil {
		if cache, err := getCache(m.CacheSize, m.CacheType); err != nil {
			return fmt.Errorf("error initializing mapper cache, err: %w", err)
		} else {
			m.cache = cache
		}
	}

	if m.logger == nil {
		m.logger = logger.DefaultSLogger("graphite_mapper")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.doFSM {
		var mappings []string
		for _, mapping := range m.Mappings {
			if mapping.MatchType == MatchTypeGlob {
				mappings = append(mappings, mapping.Match)
			}
		}
		m.FSM.BacktrackingNeeded = fsm.TestIfNeedBacktracking(mappings, false, m.logger)
	}

	return nil
}

func (m *MetricMapper) UseCache(cache MetricMapperCache) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = cache
}

// make a shallow copy so that we do not overwrite name
// as multiple names can be matched by same mapping.
func copyMetricMapping(in *MetricMapping) *MetricMapping {
	out := *in
	return &out
}

type MetricType string

const (
	defaultMeasurement = "grahite"
)

const (
	MetricTypeCounter  MetricType = "counter"
	MetricTypeGauge    MetricType = "gauge"
	MetricTypeObserver MetricType = "observer"
	MetricTypeTimer    MetricType = "timer" // DEPRECATED
)

var (
	metricRE = `[a-zA-Z_]([a-zA-Z0-9_\-])*`
	// The subsequent segments of a match can start with a number.
	metricSubRE       = `[a-zA-Z0-9_]([a-zA-Z0-9_\-])*`
	templateReplaceRE = `(\$\{?\d+\}?)`

	labelNameRE  = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]+$`)
	metricLineRE = regexp.MustCompile(`^(\*|` + metricRE + `)(\.\*|\.` + metricSubRE + `)*$`)
	metricNameRE = regexp.MustCompile(`^([a-zA-Z_]|` + templateReplaceRE + `)([a-zA-Z0-9_]|` + templateReplaceRE + `)*$`)
)

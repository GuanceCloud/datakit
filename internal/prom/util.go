// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dto "github.com/prometheus/client_model/go"
)

const statusInfo = "INFO"

type nameAndFamily struct {
	metricName   string
	metricFamily *dto.MetricFamily
}

func (p *Prom) getMetricTypeName(familyType dto.MetricType) string {
	var metricTypeName string
	switch familyType {
	case dto.MetricType_COUNTER:
		metricTypeName = "counter"
	case dto.MetricType_GAUGE:
		metricTypeName = "gauge"
	case dto.MetricType_HISTOGRAM:
		metricTypeName = "histogram"
	case dto.MetricType_SUMMARY:
		metricTypeName = "summary"
	case dto.MetricType_UNTYPED:
		metricTypeName = "untyped"
	case dto.MetricType_GAUGE_HISTOGRAM:
		// TODO
		// passed lint
	case dto.MetricType_INFO:
		metricTypeName = "info"
	}
	return metricTypeName
}

func (p *Prom) validMetricType(familyType dto.MetricType) bool {
	if len(p.opt.metricTypes) == 0 {
		return true
	}
	typeName := p.getMetricTypeName(familyType)
	for _, mt := range p.opt.metricTypes {
		if strings.ToLower(mt) == typeName {
			return true
		}
	}
	return false
}

func (p *Prom) validMetricName(name string) bool {
	// blacklist first
	if len(p.opt.metricNameReFilterIgnore) > 0 {
		for _, p := range p.opt.metricNameReFilterIgnore {
			match := p.MatchString(name)
			if match {
				return false
			}
		}
	}

	// whitelist second
	if len(p.opt.metricNameReFilter) == 0 {
		return true
	}
	for _, p := range p.opt.metricNameReFilter {
		match := p.MatchString(name)
		if match {
			return true
		}
	}
	return false
}

// getNames prioritizes naming rules as follows:
// 1. Check if any measurement rule is matched.
// 2. Check if measurement name is configured.
// 3. Check if measurement/field name can be split by the first '_' met.
// 4. If no term above matches, set both measurement name and field name to name.
func (p *Prom) getNames(name string) (measurementName string, fieldName string) {
	measurementName, fieldName = p.doGetNames(name)
	if measurementName == "" {
		measurementName = "prom"
	}
	return p.opt.measurementPrefix + measurementName, fieldName
}

func (p *Prom) doGetNames(name string) (measurementName string, fieldName string) {
	// Check if it matches custom rules.
	if mName, fName, matchAny := p.getNamesByRules(name); matchAny {
		return mName, fName
	}

	// Check if measurement name is set.
	if len(p.opt.measurementName) > 0 {
		return p.opt.measurementName, name
	}

	if mName, fName, matchAny := p.getNamesByDefault(name); matchAny {
		return mName, fName
	}

	return name, name
}

func (p *Prom) getNamesByRules(name string) (measurementName string, fieldName string, matchAny bool) {
	for _, rule := range p.opt.measurements {
		if len(rule.Prefix) > 0 && strings.HasPrefix(name, rule.Prefix) {
			if rule.Name != "" {
				measurementName = rule.Name
			} else {
				// If rule name is not set, use rule prefix as measurement name but remove all trailing _.
				measurementName = strings.TrimRight(rule.Prefix, "_")
			}
			return measurementName, name[len(rule.Prefix):], true
		}
	}
	return
}

func (p *Prom) getNamesByDefault(name string) (measurementName string, fieldName string, matchAny bool) {
	// By default, measurement name and metric name are split according to the first '_' met.
	pattern := "(^[^_]+)_(.*)$"
	reg := regexp.MustCompile(pattern)
	if reg != nil {
		result := reg.FindAllStringSubmatch(name, -1)
		if len(result) == 1 {
			return result[0][1], result[0][2], true
		}
	}
	return
}

func (p *Prom) filterIgnoreTagKV(tags map[string]string) map[string]string {
	if p.opt.ignoreTagKV == nil {
		return tags
	}

	newTags := map[string]string{}

	for k, v := range tags {
		newTags[k] = v
		if res, ok := p.opt.ignoreTagKV[k]; ok {
			for _, re := range res {
				if re.MatchString(v) {
					delete(newTags, k)
					break
				}
			}
		}
	}

	return newTags
}

func (p *Prom) getTags(labels []*dto.LabelPair, measurementName string, u string) map[string]string {
	tags := map[string]string{}

	if !p.opt.disableInfoTag {
		for k, v := range p.InfoTags {
			tags[k] = v
		}
	}

	// Add custom tags.
	for k, v := range p.opt.tags {
		tags[k] = v
	}

	// Add prometheus labels as tags.
	for _, l := range labels {
		tags[l.GetName()] = l.GetValue()
	}

	p.removeIgnoredTags(tags)
	p.renameTags(tags)

	// Configure service tag if metrics are fed as logging.
	if p.opt.asLogging != nil && p.opt.asLogging.Enable {
		if p.opt.asLogging.Service != "" {
			tags["service"] = p.opt.asLogging.Service
		} else {
			tags["service"] = measurementName
		}
	}

	return tags
}

func (p *Prom) getTagsWithLE(labels []*dto.LabelPair, measurementName string, b *dto.Bucket) map[string]string {
	tags := map[string]string{}

	// Add custom tags.
	for k, v := range p.opt.tags {
		tags[k] = v
	}

	// Add prometheus labels as tags.
	for _, lab := range labels {
		tags[lab.GetName()] = lab.GetValue()
	}

	tags["le"] = fmt.Sprint(b.GetUpperBound())

	p.removeIgnoredTags(tags)
	p.renameTags(tags)

	// Configure service tag if metrics are fed as logging.
	if p.opt.asLogging != nil && p.opt.asLogging.Enable {
		if p.opt.asLogging.Service != "" {
			tags["service"] = p.opt.asLogging.Service
		} else {
			tags["service"] = measurementName
		}
	}

	return tags
}

func (p *Prom) removeIgnoredTags(tags map[string]string) {
	for t := range tags {
		for _, ignoredTag := range p.opt.tagsIgnore {
			if t == ignoredTag {
				delete(tags, t)
			}
		}
	}
}

func (p *Prom) renameTags(tags map[string]string) {
	if tags == nil || p.opt.tagsRename == nil {
		return
	}

	for oldKey, newKey := range p.opt.tagsRename.Mapping {
		if v, ok := tags[oldKey]; ok { // rename the tag
			if _, exists := tags[newKey]; exists && !p.opt.tagsRename.OverwriteExistTags {
				continue
			}

			delete(tags, oldKey)
			tags[newKey] = v
		}
	}
}

func (p *Prom) swapTypeInfoToFront(nf []nameAndFamily) {
	var i int
	for j := 0; j < len(nf); j++ {
		temp := nf[j].metricFamily.GetType()
		_ = temp
		if nf[j].metricFamily.GetType() == dto.MetricType_INFO {
			nf[i], nf[j] = nf[j], nf[i]
			i++
		}
	}
}

func (p *Prom) filterMetricFamilies(metricFamilies map[string]*dto.MetricFamily) []nameAndFamily {
	var filteredMetricFamilies []nameAndFamily
	for name, value := range metricFamilies {
		if p.validMetricName(name) && p.validMetricType(value.GetType()) {
			filteredMetricFamilies = append(filteredMetricFamilies, nameAndFamily{metricName: name, metricFamily: value})
		}
	}
	return filteredMetricFamilies
}

// text2Metrics converts raw prometheus metric text to line protocol point.
func (p *Prom) text2Metrics(in io.Reader, u string) (pts []*point.Point, lastErr error) {
	p.ptCount = 0
	for k := range p.InfoTags {
		delete(p.InfoTags, k)
	}

	if p.opt.batchCallback != nil {
		return p.text2MetricsBatch(in, u)
	} else {
		return p.text2MetricsNoBatch(in, u)
	}
}

func (p *Prom) text2MetricsBatch(in io.Reader, u string) (pts []*point.Point, lastErr error) {
	err := p.parser.StreamingParse(in)
	return nil, err
}

func (p *Prom) text2MetricsNoBatch(in io.Reader, u string) (pts []*point.Point, lastErr error) {
	metricFamilies, err := p.parser.TextToMetricFamilies(in)
	if err != nil {
		return nil, err
	}

	return p.MetricFamilies2points(metricFamilies, u)
}

func (p *Prom) MetricFamilies2points(metricFamilies map[string]*dto.MetricFamily, u string) (pts []*point.Point, lastErr error) {
	filteredMetricFamilies := p.filterMetricFamilies(metricFamilies)
	p.swapTypeInfoToFront(filteredMetricFamilies)
	var err error

	opts := point.DefaultMetricOptions()
	for _, nf := range filteredMetricFamilies {
		name, value := nf.metricName, nf.metricFamily
		measurementName, fieldName := p.getNames(name)

		switch value.GetType() {
		case dto.MetricType_GAUGE, dto.MetricType_UNTYPED, dto.MetricType_COUNTER:
			for _, m := range value.GetMetric() {
				v := getValue(m, value.GetType())
				if math.IsInf(v, 0) || math.IsNaN(v) {
					continue
				}

				fields := map[string]interface{}{
					fieldName: v,
				}
				if p.opt.asLogging != nil && p.opt.asLogging.Enable {
					opts = point.DefaultLoggingOptions()
					fields["status"] = statusInfo
				}

				tags := p.filterIgnoreTagKV(p.getTags(m.GetLabel(), measurementName, u))
				pt := point.NewPointV2(measurementName,
					append(point.NewTags(tags), point.NewKVs(fields)...),
					opts...)
				if err != nil {
					lastErr = err
				} else {
					pts = append(pts, pt)
				}
			}

		case dto.MetricType_SUMMARY:
			for _, m := range value.GetMetric() {
				fields := map[string]interface{}{
					fieldName + "_count": float64(m.GetSummary().GetSampleCount()),
					fieldName + "_sum":   m.GetSummary().GetSampleSum(),
				}
				if p.opt.asLogging != nil && p.opt.asLogging.Enable {
					fields["status"] = statusInfo
				}

				tags := p.filterIgnoreTagKV(p.getTags(m.GetLabel(), measurementName, u))
				pt := point.NewPointV2(measurementName,
					append(point.NewTags(tags), point.NewKVs(fields)...),
					opts...)
				if err != nil {
					lastErr = err
				} else {
					pts = append(pts, pt)
				}

				for _, q := range m.GetSummary().Quantile {
					v := q.GetValue()
					if math.IsNaN(v) {
						continue
					}

					fields := map[string]interface{}{
						fieldName: v,
					}

					if p.opt.asLogging != nil && p.opt.asLogging.Enable {
						fields["status"] = statusInfo
					}

					tags := p.filterIgnoreTagKV(p.getTags(m.GetLabel(), measurementName, u))
					tags["quantile"] = fmt.Sprint(q.GetQuantile())
					pt := point.NewPointV2(measurementName,
						append(point.NewTags(tags), point.NewKVs(fields)...),
						opts...)
					if err != nil {
						lastErr = err
					} else {
						pts = append(pts, pt)
					}
				}
			}

		case dto.MetricType_HISTOGRAM:
			for _, m := range value.GetMetric() {
				fields := map[string]interface{}{
					fieldName + "_count": float64(m.GetHistogram().GetSampleCount()),
					fieldName + "_sum":   m.GetHistogram().GetSampleSum(),
				}
				if p.opt.asLogging != nil && p.opt.asLogging.Enable {
					fields["status"] = statusInfo
				}

				tags := p.filterIgnoreTagKV(p.getTags(m.GetLabel(), measurementName, u))
				pt := point.NewPointV2(measurementName,
					append(point.NewTags(tags), point.NewKVs(fields)...),
					opts...)
				if err != nil {
					lastErr = err
				} else {
					pts = append(pts, pt)
				}

				for _, b := range m.GetHistogram().GetBucket() {
					fields := map[string]interface{}{
						fieldName + "_bucket": b.GetCumulativeCount(),
					}
					if p.opt.asLogging != nil && p.opt.asLogging.Enable {
						fields["status"] = statusInfo
					}

					tags := p.filterIgnoreTagKV(p.getTagsWithLE(m.GetLabel(), measurementName, b))
					pt := point.NewPointV2(measurementName,
						append(point.NewTags(tags), point.NewKVs(fields)...),
						opts...)
					if err != nil {
						lastErr = err
					} else {
						pts = append(pts, pt)
					}
				}
			}
		case dto.MetricType_GAUGE_HISTOGRAM:
			// TODO
			// passed lint
		case dto.MetricType_INFO:
			// Info metrics are used to expose textual information which SHOULD NOT change during process lifetime.
			// Info may be used to encode ENUMs whose values do not change over time, such as the type of a network interface.
			// https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#info
			for _, m := range value.GetMetric() {
				for _, l := range m.GetLabel() {
					p.InfoTags[l.GetName()] = l.GetValue()
				}
			}
		}
	}

	if lastErr != nil {
		return pts, fmt.Errorf("text2Metrics encountered make point error: %w", lastErr)
	}
	p.ptCount += len(pts)
	return pts, nil
}

func getValue(m *dto.Metric, metricType dto.MetricType) float64 {
	switch metricType { //nolint:exhaustive
	case dto.MetricType_GAUGE:
		return m.GetGauge().GetValue()
	case dto.MetricType_UNTYPED:
		return m.GetUntyped().GetValue()
	case dto.MetricType_COUNTER:
		return m.GetCounter().GetValue()
	default:
		// Shouldn't get here.
		return 0
	}
}

func getTimestampS(m *dto.Metric, startTime time.Time) time.Time {
	if m.GetTimestampMs() != 0 {
		return time.Unix(m.GetTimestampMs()/1000, 0)
	}
	return startTime
}

func (p *Prom) getMode() string {
	if p.opt.streamSize > 0 {
		return "stream"
	} else {
		return "no_stream"
	}
}

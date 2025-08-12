// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dto "github.com/prometheus/client_model/go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
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
			if p.MatchString(name) {
				return false
			}
		}
	}

	// whitelist second
	if len(p.opt.metricNameReFilter) == 0 {
		return true
	}

	for _, p := range p.opt.metricNameReFilter {
		if p.MatchString(name) {
			return true
		}
	}

	return false
}

// getNames prioritizes naming rules as follows:
// 1. Check if any measurement rule is matched.
// 2. Check if measurement name is configured.
// 3. Check if measurement/field name can be split by the first '_' met.
// 3.1 If the useExistMetricName is true, keep the raw value for field names.
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
	return p.getNamesByDefault(name)
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

func (p *Prom) getNamesByDefault(name string) (measurementName string, fieldName string) {
	startPosition := strings.IndexFunc(name, func(r rune) bool {
		return r != '_'
	})
	if startPosition == -1 || startPosition == len(name)-1 {
		return "unknown", "unknown"
	}

	name = name[startPosition:]
	// By default, measurement name and metric name are split according to the first '_' met.
	index := strings.Index(name, "_")

	switch index {
	case -1:
		return name, name
	case 0:
		return name[index:], name[index:]
	case len(name) - 1:
		return name[:index], name[:index]
	}

	// If the keepExistMetricName is true, keep the raw value for field names.
	if p.opt.keepExistMetricName {
		return name[:index], name
	}
	return name[:index], name[index+1:]
}

func (p *Prom) filterIgnoreTagKV(tags point.KVs) point.KVs {
	if p.opt.ignoreTagKV == nil {
		return tags
	}

	for key, arr := range p.opt.ignoreTagKV {
		if f := tags.Get(key); f != nil {
			for _, re := range arr {
				if re.MatchString(f.GetS()) { // drop the tag
					tags = tags.Del(key)
					break
				}
			}
		}
	}

	return tags
}

func (p *Prom) getTags(labels []*dto.LabelPair, measurementName string, u string) point.KVs {
	var kvs point.KVs

	if !p.opt.disableInfoTag {
		for k, v := range p.InfoTags {
			kvs = kvs.AddTag(k, v)
		}
	}

	// Add custom tags.
	for k, v := range p.opt.tags {
		kvs = kvs.AddTag(k, v)
	}

	// Add prometheus labels as tags.
	for _, l := range labels {
		kvs = kvs.AddTag(l.GetName(), l.GetValue())
	}

	kvs = p.removeIgnoredTags(kvs)
	kvs = p.renameTags(kvs)

	// Configure service tag if metrics are fed as logging.
	if p.opt.asLogging != nil && p.opt.asLogging.Enable {
		if p.opt.asLogging.Service != "" {
			kvs = kvs.AddTag("service", p.opt.asLogging.Service)
		} else {
			kvs = kvs.AddTag("service", measurementName)
		}
	}

	return kvs
}

func (p *Prom) getTagsWithLE(labels []*dto.LabelPair, measurementName string, b *dto.Bucket) point.KVs {
	var kvs point.KVs

	// Add custom tags.
	for k, v := range p.opt.tags {
		kvs = kvs.AddTag(k, v)
	}

	// Add prometheus labels as tags.
	for _, lab := range labels {
		kvs = kvs.AddTag(lab.GetName(), lab.GetValue())
	}

	kvs = kvs.AddTag("le", fmt.Sprint(b.GetUpperBound()))
	kvs = p.removeIgnoredTags(kvs)
	kvs = p.renameTags(kvs)

	// Configure service tag if metrics are fed as logging.
	if p.opt.asLogging != nil && p.opt.asLogging.Enable {
		if p.opt.asLogging.Service != "" {
			kvs = kvs.AddTag("service", p.opt.asLogging.Service)
		} else {
			kvs = kvs.AddTag("service", measurementName)
		}
	}

	return kvs
}

func (p *Prom) removeIgnoredTags(tags point.KVs) point.KVs {
	for _, key := range p.opt.tagsIgnore {
		if f := tags.Get(key); f != nil && f.IsTag {
			tags = tags.Del(key)
		}
	}

	return tags
}

func (p *Prom) renameTags(tags point.KVs) point.KVs {
	if len(tags) == 0 || p.opt.tagsRename == nil {
		return tags
	}

	for oldKey, newKey := range p.opt.tagsRename.Mapping {
		if x := tags.Get(oldKey); x != nil { // rename the tag
			if tags.Get(newKey) != nil && !p.opt.tagsRename.OverwriteExistTags {
				continue
			}

			tags = tags.Del(oldKey)
			x.Key = newKey // update key name
			tags = tags.SetKV(x)
		}
	}

	return tags
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

	ptts := p.opt.ptts
	if ptts == 0 {
		ptts = ntp.Now().UnixNano()
	}

	opts := append(point.DefaultMetricOptions(), point.WithTimestamp(ptts))
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

				kvs := p.getTags(m.GetLabel(), measurementName, u)

				kvs = p.filterIgnoreTagKV(kvs)
				kvs = kvs.Add(fieldName, v)

				if p.opt.asLogging != nil && p.opt.asLogging.Enable {
					opts = point.DefaultLoggingOptions() // we do not need timestamp alignment on logging
					kvs = kvs.Add("status", statusInfo)
				}

				pts = append(pts, point.NewPoint(measurementName, kvs, opts...))
			}

		case dto.MetricType_SUMMARY:
			for _, m := range value.GetMetric() {
				kvs := p.filterIgnoreTagKV(p.getTags(m.GetLabel(), measurementName, u))
				kvs = kvs.Add(fieldName+"_count", float64(m.GetSummary().GetSampleCount()))
				kvs = kvs.Add(fieldName+"_sum", m.GetSummary().GetSampleSum())
				if p.opt.asLogging != nil && p.opt.asLogging.Enable {
					kvs = kvs.Add("status", statusInfo)
				}

				pts = append(pts, point.NewPoint(measurementName, kvs, opts...))

				for _, q := range m.GetSummary().Quantile {
					v := q.GetValue()
					if math.IsNaN(v) {
						continue
					}

					kvs := p.filterIgnoreTagKV(p.getTags(m.GetLabel(), measurementName, u))
					kvs = kvs.AddTag("quantile", fmt.Sprint(q.GetQuantile()))
					if p.opt.asLogging != nil && p.opt.asLogging.Enable {
						kvs = kvs.Add("status", statusInfo)
					}

					kvs = kvs.Add(fieldName, v)
					pts = append(pts, point.NewPoint(measurementName, kvs, opts...))
				}
			}

		case dto.MetricType_HISTOGRAM:
			for _, m := range value.GetMetric() {
				kvs := p.filterIgnoreTagKV(p.getTags(m.GetLabel(), measurementName, u))
				kvs = kvs.Add(fieldName+"_count", float64(m.GetHistogram().GetSampleCount()))
				kvs = kvs.Add(fieldName+"_sum", m.GetHistogram().GetSampleSum())

				if p.opt.asLogging != nil && p.opt.asLogging.Enable {
					kvs = kvs.Add("status", statusInfo)
				}

				pts = append(pts, point.NewPoint(measurementName, kvs, opts...))

				for _, b := range m.GetHistogram().GetBucket() {
					kvs := p.filterIgnoreTagKV(p.getTagsWithLE(m.GetLabel(), measurementName, b))
					kvs = kvs.Add(fieldName+"_bucket", float64(b.GetCumulativeCount()))
					if p.opt.asLogging != nil && p.opt.asLogging.Enable {
						kvs = kvs.Add("status", statusInfo)
					}

					pts = append(pts, point.NewPoint(measurementName, kvs, opts...))
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

func (p *Prom) getMode() string {
	if p.opt.streamSize > 0 {
		return "stream"
	} else {
		return "no_stream"
	}
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

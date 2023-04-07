// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

const statusInfo = "INFO"

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
	}
	return metricTypeName
}

func (p *Prom) validMetricType(familyType dto.MetricType) bool {
	if len(p.opt.MetricTypes) == 0 {
		return true
	}
	typeName := p.getMetricTypeName(familyType)
	for _, mt := range p.opt.MetricTypes {
		if strings.ToLower(mt) == typeName {
			return true
		}
	}
	return false
}

func (p *Prom) validMetricName(name string) bool {
	if len(p.opt.MetricNameFilter) == 0 {
		return true
	}
	for _, p := range p.opt.MetricNameFilter {
		match, err := regexp.MatchString(p, name)
		if err != nil {
			continue
		}
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
	return p.opt.MeasurementPrefix + measurementName, fieldName
}

func (p *Prom) doGetNames(name string) (measurementName string, fieldName string) {
	// Check if it matches custom rules.
	if mName, fName, matchAny := p.getNamesByRules(name); matchAny {
		return mName, fName
	}

	// Check if measurement name is set.
	if len(p.opt.MeasurementName) > 0 {
		return p.opt.MeasurementName, name
	}

	if mName, fName, matchAny := p.getNamesByDefault(name); matchAny {
		return mName, fName
	}

	return name, name
}

func (p *Prom) getNamesByRules(name string) (measurementName string, fieldName string, matchAny bool) {
	for _, rule := range p.opt.Measurements {
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

func (p *Prom) tagKVMatched(tags map[string]string) bool {
	if p.opt.IgnoreTagKV == nil {
		return false
	}

	for k, v := range tags {
		if res, ok := p.opt.IgnoreTagKV[k]; ok {
			for _, re := range res {
				if re.MatchString(v) {
					return true
				}
			}
		}
	}

	return false
}

func (p *Prom) getTags(labels []*dto.LabelPair, measurementName string, u string) map[string]string {
	tags := map[string]string{}

	if !p.opt.DisableHostTag {
		setHostTagIfNotLoopback(tags, u)
	}

	if !p.opt.DisableInstanceTag {
		setInstanceTag(tags, u)
	}

	// Add custom tags.
	for k, v := range p.opt.Tags {
		tags[k] = v
	}

	// Add prometheus labels as tags.
	for _, lab := range labels {
		tags[lab.GetName()] = lab.GetValue()
	}

	p.removeIgnoredTags(tags)
	p.renameTags(tags)

	// Configure service tag if metrics are fed as logging.
	if p.opt.AsLogging != nil && p.opt.AsLogging.Enable {
		if p.opt.AsLogging.Service != "" {
			tags["service"] = p.opt.AsLogging.Service
		} else {
			tags["service"] = measurementName
		}
	}

	return tags
}

func setInstanceTag(tags map[string]string, u string) {
	uu, err := url.Parse(u)
	if err != nil {
		return
	}
	tags["instance"] = uu.Host
}

func setHostTagIfNotLoopback(tags map[string]string, u string) {
	uu, err := url.Parse(u)
	if err != nil {
		return
	}
	host, _, err := net.SplitHostPort(uu.Host)
	if err != nil {
		return
	}
	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		tags["host"] = host
	}
}

func shouldDisableGlobalHostTag(u string) bool {
	uu, err := url.Parse(u)
	if err != nil {
		return false
	}
	host, _, err := net.SplitHostPort(uu.Host)
	if err != nil {
		return false
	}
	// disable global host tag if ip is not a loopback address
	return host == "localhost" || net.ParseIP(host).IsLoopback()
}

func (p *Prom) getTagsWithLE(labels []*dto.LabelPair, measurementName string, b *dto.Bucket) map[string]string {
	tags := map[string]string{}

	// Add custom tags.
	for k, v := range p.opt.Tags {
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
	if p.opt.AsLogging != nil && p.opt.AsLogging.Enable {
		if p.opt.AsLogging.Service != "" {
			tags["service"] = p.opt.AsLogging.Service
		} else {
			tags["service"] = measurementName
		}
	}

	return tags
}

func (p *Prom) removeIgnoredTags(tags map[string]string) {
	for t := range tags {
		for _, ignoredTag := range p.opt.TagsIgnore {
			if t == ignoredTag {
				delete(tags, t)
			}
		}
	}
}

func (p *Prom) renameTags(tags map[string]string) {
	if tags == nil || p.opt.RenameTags == nil {
		return
	}

	for oldKey, newKey := range p.opt.RenameTags.Mapping {
		if v, ok := tags[oldKey]; ok { // rename the tag
			if _, exists := tags[newKey]; exists && !p.opt.RenameTags.OverwriteExistTags {
				continue
			}

			delete(tags, oldKey)
			tags[newKey] = v
		}
	}
}

func (p *Prom) filterMetricFamilies(metricFamilies map[string]*dto.MetricFamily) map[string]*dto.MetricFamily {
	filteredMetricFamilies := make(map[string]*dto.MetricFamily)
	for name, value := range metricFamilies {
		if p.validMetricName(name) && p.validMetricType(value.GetType()) {
			filteredMetricFamilies[name] = value
		}
	}
	return filteredMetricFamilies
}

// text2Metrics converts raw prometheus metric text to line protocol points.
// It handles the case in which the same kind of comment for a metric name appears repeatedly.
// E.g. multiple '# TYPE node_network_info gauge' appears in one piece of text.
func (p *Prom) text2Metrics(in io.Reader, u string) ([]*point.Point, error) {
	var (
		buf         []byte
		index       int
		points      []*point.Point
		comment     = make(map[string]struct{})
		split       bool
		startOfLine = true
	)
	text, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}
	for index < len(text) {
		if text[index] == '#' && startOfLine {
			// Lines with a # as the first non-whitespace character are comments.
			ti := newTokenIterator(text, index+1)
			tkn := ti.readNextToken()
			if tkn == "HELP" || tkn == "TYPE" {
				// If the token is HELP, at least one more token is expected, which is the metric name. If the token is
				// TYPE, exactly two more tokens are expected. The first is the metric name, and the second is either
				// counter, gauge, histogram, summary, or untyped, defining the type for the metric of that name. In
				// both cases, # is followed by token HELP or TYPE, and followed by at least one more token which is the
				// metric name. If it reaches end of line before reading out metric name, This line is treated as a
				// common comment.
				metricName := ti.readNextToken()
				if metricName != "" {
					if _, has := comment[tkn+metricName]; has {
						split = true
					} else {
						comment[tkn+metricName] = struct{}{}
					}
				}
			}
			if split {
				pts, err := p.doText2Metrics(bytes.NewReader(buf), u)
				if err != nil {
					return nil, err
				}
				points = append(points, pts...)
				comment = map[string]struct{}{}
				buf = make([]byte, 0)
				split = false
				continue
			}
		}
		if text[index] == '\n' {
			startOfLine = true
		} else {
			startOfLine = false
		}
		buf = append(buf, text[index])
		index++
	}

	if len(buf) != 0 {
		pts, err := p.doText2Metrics(bytes.NewReader(buf), u)
		if err != nil {
			return nil, err
		}
		points = append(points, pts...)
	}
	return points, nil
}

type tokenIterator struct {
	text  []byte
	index int
}

func (ti *tokenIterator) readNextToken() string {
	if ti.index >= len(ti.text) {
		return ""
	}
	var l, r int
	cur := ti.index
	for cur < len(ti.text) && ti.text[cur] == ' ' {
		cur++
	}
	l = cur
	for cur < len(ti.text) && ti.text[cur] != ' ' && ti.text[cur] != '\n' {
		cur++
	}
	r = cur
	ti.index = cur
	return string(ti.text[l:r])
}

func newTokenIterator(text []byte, start int) tokenIterator {
	return tokenIterator{
		text:  text,
		index: start,
	}
}

// doText2Metrics converts raw prometheus metric text to line protocol point.
func (p *Prom) doText2Metrics(in io.Reader, u string) (pts []*point.Point, lastErr error) {
	metricFamilies, err := p.parser.TextToMetricFamilies(in)
	if err != nil {
		return nil, err
	}

	filteredMetricFamilies := p.filterMetricFamilies(metricFamilies)

	for name, value := range filteredMetricFamilies {
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
				if p.opt.AsLogging != nil && p.opt.AsLogging.Enable {
					fields["status"] = statusInfo
				}
				tags := p.getTags(m.GetLabel(), measurementName, u)

				if !p.tagKVMatched(tags) {
					pointOpt := *p.opt.pointOpt
					if shouldDisableGlobalHostTag(u) {
						pointOpt.DisableGlobalTags = true
					}
					pt, err := point.NewPoint(measurementName, tags, fields, &pointOpt)
					if err != nil {
						lastErr = err
					} else {
						pts = append(pts, pt)
					}
				}
			}

		case dto.MetricType_SUMMARY:
			for _, m := range value.GetMetric() {
				fields := map[string]interface{}{
					fieldName + "_count": float64(m.GetSummary().GetSampleCount()),
					fieldName + "_sum":   m.GetSummary().GetSampleSum(),
				}
				if p.opt.AsLogging != nil && p.opt.AsLogging.Enable {
					fields["status"] = statusInfo
				}
				tags := p.getTags(m.GetLabel(), measurementName, u)

				if !p.tagKVMatched(tags) {
					pointOpt := *p.opt.pointOpt
					if shouldDisableGlobalHostTag(u) {
						pointOpt.DisableGlobalTags = true
					}
					pt, err := point.NewPoint(measurementName, tags, fields, &pointOpt)
					if err != nil {
						lastErr = err
					} else {
						pts = append(pts, pt)
					}
				}

				for _, q := range m.GetSummary().Quantile {
					fields := map[string]interface{}{
						fieldName: q.GetValue(),
					}
					if p.opt.AsLogging != nil && p.opt.AsLogging.Enable {
						fields["status"] = statusInfo
					}

					tags := p.getTags(m.GetLabel(), measurementName, u)
					tags["quantile"] = fmt.Sprint(q.GetQuantile())

					if !p.tagKVMatched(tags) {
						pointOpt := *p.opt.pointOpt
						if shouldDisableGlobalHostTag(u) {
							pointOpt.DisableGlobalTags = true
						}
						pt, err := point.NewPoint(measurementName, tags, fields, &pointOpt)
						if err != nil {
							lastErr = err
						} else {
							pts = append(pts, pt)
						}
					}
				}
			}

		case dto.MetricType_HISTOGRAM:
			for _, m := range value.GetMetric() {
				fields := map[string]interface{}{
					fieldName + "_count": float64(m.GetHistogram().GetSampleCount()),
					fieldName + "_sum":   m.GetHistogram().GetSampleSum(),
				}
				if p.opt.AsLogging != nil && p.opt.AsLogging.Enable {
					fields["status"] = statusInfo
				}

				tags := p.getTags(m.GetLabel(), measurementName, u)

				if !p.tagKVMatched(tags) {
					pointOpt := *p.opt.pointOpt
					if shouldDisableGlobalHostTag(u) {
						pointOpt.DisableGlobalTags = true
					}
					pt, err := point.NewPoint(measurementName, tags, fields, &pointOpt)
					if err != nil {
						lastErr = err
					} else {
						pts = append(pts, pt)
					}
				}

				for _, b := range m.GetHistogram().GetBucket() {
					fields := map[string]interface{}{
						fieldName + "_bucket": b.GetCumulativeCount(),
					}
					if p.opt.AsLogging != nil && p.opt.AsLogging.Enable {
						fields["status"] = statusInfo
					}

					tags := p.getTagsWithLE(m.GetLabel(), measurementName, b)

					if !p.tagKVMatched(tags) {
						pointOpt := *p.opt.pointOpt
						if shouldDisableGlobalHostTag(u) {
							pointOpt.DisableGlobalTags = true
						}
						pt, err := point.NewPoint(measurementName, tags, fields, &pointOpt)
						if err != nil {
							lastErr = err
						} else {
							pts = append(pts, pt)
						}
					}
				}
			}
		case dto.MetricType_GAUGE_HISTOGRAM:
			// TODO
			// passed lint
		}
	}
	if lastErr != nil {
		return pts, fmt.Errorf("doText2Metrics encountered make point error: %w", lastErr)
	}
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

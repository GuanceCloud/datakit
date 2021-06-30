package prom

import (
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func isValid(familyType dto.MetricType, name string, metricTypes []string, metricNameFilter []string) bool {
	metricType := ""
	typeValid := false
	nameValid := false

	switch familyType {
	case dto.MetricType_COUNTER:
		metricType = "counter"
	case dto.MetricType_GAUGE:
		metricType = "gauge"
	case dto.MetricType_HISTOGRAM:
		metricType = "histogram"
	case dto.MetricType_SUMMARY:
		metricType = "summary"
	case dto.MetricType_UNTYPED:
		metricType = "untyped"
	default:
		return false
	}

	// check metric type
	if len(metricTypes) == 0 {
		typeValid = true
	} else {
		for _, mt := range metricTypes {
			mtLower := strings.ToLower(mt)
			if mtLower == metricType {
				typeValid = true
				break
			}
		}
	}

	// check name
	// set nameValid true when metricNameFilter is empty
	if len(metricNameFilter) == 0 {
		nameValid = true
	} else {
		for _, p := range metricNameFilter {
			match, err := regexp.MatchString(p, name)
			if err != nil {
				continue
			}
			if match {
				nameValid = true
				break
			}
		}
	}

	return typeValid && nameValid
}

func getNames(name string, customMeasurementRules []Rule, measurementName, measurementPrefix string) (string, string) {
	// 1. check custom rules
	if len(customMeasurementRules) > 0 {
		for _, rule := range customMeasurementRules {
			prefix := rule.Prefix
			if len(prefix) > 0 {
				if strings.HasPrefix(name, prefix) {
					ruleName := rule.Name
					if len(ruleName) == 0 {
						ruleName = strings.TrimRight(prefix, "_")
					}
					ruleName = measurementPrefix + ruleName
					fieldName := strings.Replace(name, prefix, "", 1)
					return ruleName, fieldName
				}
			}
		}
	}

	// 2. check measurementName
	if len(measurementName) > 0 {
		return measurementPrefix + measurementName, name
	}

	// default
	pattern := "(^[^_]+)_(.*)$"
	reg := regexp.MustCompile(pattern)
	if reg == nil {
		return measurementPrefix + name, name
	}
	result := reg.FindAllStringSubmatch(name, -1)
	if len(result) == 1 {
		measurementName := measurementPrefix + result[0][1]
		fieldName := result[0][2]
		return measurementName, fieldName
	}

	return measurementPrefix + name, name
}

func getTags(labels []*dto.LabelPair, tags, extraTags map[string]string, ignoreTags []string) map[string]string {
	if tags == nil {
		tags = map[string]string{}
	}

	for k, v := range extraTags {
		tags[k] = v
	}

	for _, lab := range labels {
		tags[lab.GetName()] = lab.GetValue()
	}

	for k := range tags {
		for _, ignoreTag := range ignoreTags {
			if k == ignoreTag {
				delete(tags, k)
			}
		}
	}

	return tags
}

func PromText2Metrics(text interface{}, prom *Input, extraTags map[string]string) ([]inputs.Measurement, error) {
	var reader io.Reader
	r, ok := text.(string)
	if ok {
		reader = strings.NewReader(r)
	} else {
		r, ok := text.(io.Reader)
		if !ok {
			return nil, fmt.Errorf("invalid text")
		} else {
			reader = r
		}
	}
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		return nil, err
	}

	metricTypes := prom.MetricTypes
	metricNameFilter := prom.MetricNameFilter

	customMeasurementRules := prom.Measurements
	measurementName := prom.MeasurementName
	measurementPrefix := prom.MeasurementPrefix

	collectTime := prom.collectTime
	if collectTime.Unix() < 0 {
		collectTime = time.Now()
	}

	points := []inputs.Measurement{}

	// iterate all metrics
	for name, value := range metricFamilies {
		familyType := value.GetType()
		valid := isValid(familyType, name, metricTypes, metricNameFilter)
		if !valid {
			continue
		}

		measurementName, fieldName := getNames(name, customMeasurementRules, measurementName, measurementPrefix)

		// set default name when measurementName is empty
		if len(measurementName) == 0 {
			measurementName = "prom"
		}

		metrics := value.GetMetric()

		switch familyType {
		case dto.MetricType_GAUGE:
			for _, m := range metrics {
				v := m.GetGauge().GetValue()
				if math.IsInf(v, 0) {
					continue
				}

				fields := make(map[string]interface{})
				fields[fieldName] = v

				labels := m.GetLabel()
				tags := getTags(labels, prom.Tags, extraTags, prom.TagsIgnore)

				points = append(points, Measurement{
					name:   measurementName,
					tags:   tags,
					fields: fields,
					ts:     collectTime,
				})

			}
		case dto.MetricType_UNTYPED:
			for _, m := range metrics {
				v := m.GetUntyped().GetValue()
				if math.IsInf(v, 0) {
					continue
				}

				fields := make(map[string]interface{})
				fields[fieldName] = v

				labels := m.GetLabel()
				tags := getTags(labels, prom.Tags, extraTags, prom.TagsIgnore)

				points = append(points, Measurement{
					name:   measurementName,
					tags:   tags,
					fields: fields,
					ts:     collectTime,
				})

			}
		case dto.MetricType_COUNTER:
			for _, m := range metrics {
				fields := make(map[string]interface{})
				v := m.GetCounter().GetValue()
				if math.IsInf(v, 0) {
					continue
				}
				fields[fieldName] = v

				labels := m.GetLabel()
				tags := getTags(labels, prom.Tags, extraTags, prom.TagsIgnore)

				points = append(points, Measurement{
					name:   measurementName,
					tags:   tags,
					fields: fields,
					ts:     collectTime,
				})
			}
		case dto.MetricType_SUMMARY:
			for _, m := range metrics {
				fields := make(map[string]interface{})
				count := m.GetSummary().GetSampleCount()
				sum := m.GetSummary().GetSampleSum()
				quantiles := m.GetSummary().Quantile

				fields[fieldName+"_count"] = float64(count)
				fields[fieldName+"_sum"] = sum

				labels := m.GetLabel()
				tags := getTags(labels, prom.Tags, extraTags, prom.TagsIgnore)

				points = append(points, Measurement{
					name:   measurementName,
					tags:   tags,
					fields: fields,
					ts:     collectTime,
				})

				for _, q := range quantiles {
					quantile := q.GetQuantile() // 0 0.25 0.5 0.75 1
					val := q.GetValue()         // value

					fields := make(map[string]interface{})
					fields[fieldName] = val

					labels := m.GetLabel()
					tags := getTags(labels, prom.Tags, extraTags, prom.TagsIgnore)

					tags["quantile"] = fmt.Sprint(quantile)

					points = append(points, Measurement{
						name:   measurementName,
						tags:   tags,
						fields: fields,
						ts:     collectTime,
					})
				}

			}
		case dto.MetricType_HISTOGRAM:
			for _, m := range metrics {

				fields := make(map[string]interface{})
				count := m.GetHistogram().GetSampleCount()
				sum := m.GetHistogram().GetSampleSum()
				buckets := m.GetHistogram().GetBucket()

				fields[fieldName+"_count"] = float64(count)
				fields[fieldName+"_sum"] = sum

				labels := m.GetLabel()
				tags := getTags(labels, prom.Tags, extraTags, prom.TagsIgnore)

				points = append(points, Measurement{
					name:   measurementName,
					tags:   tags,
					fields: fields,
					ts:     collectTime,
				})

				for _, b := range buckets {
					count := b.GetCumulativeCount()
					bond := b.GetUpperBound()

					fields := make(map[string]interface{})
					fields[fieldName+"_bucket"] = count

					labels := m.GetLabel()
					tags := getTags(labels, prom.Tags, extraTags, prom.TagsIgnore)
					tags["le"] = fmt.Sprint(bond)

					points = append(points, Measurement{
						name:   measurementName,
						tags:   tags,
						fields: fields,
						ts:     collectTime,
					})
				}

			}
		}
	}
	return points, nil
}

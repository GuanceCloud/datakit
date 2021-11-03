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
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

//nolint:cyclop
func isValid(familyType dto.MetricType, name string, metricTypes, metricNameFilter []string) bool {
	var metricType string
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

//nolint:gocritic
func getNames(name string, customMeasurementRules []Rule,
	measurementName, measurementPrefix string) (string, string) {
	// 1. check custom rules
	if len(customMeasurementRules) > 0 {
		for _, rule := range customMeasurementRules {
			prefix := rule.Prefix
			if len(prefix) > 0 {
				if strings.HasPrefix(name, prefix) {
					ruleName := rule.Name
					if ruleName == "" {
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

func getTags(labels []*dto.LabelPair, promTags, extraTags map[string]string, ignoreTags []string) map[string]string {
	tags := map[string]string{}

	for k, v := range promTags {
		tags[k] = v
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

// TODO: refact me
//nolint:funlen,gocyclo,cyclop
func Text2Metrics(in io.Reader,
	prom *Option,
	extraTags map[string]string) ([]*iod.Point, error) {
	var lastErr error
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(in)
	if err != nil {
		return nil, err
	}

	metricTypes := prom.MetricTypes
	metricNameFilter := prom.MetricNameFilter

	customMeasurementRules := prom.Measurements
	measurementName := prom.MeasurementName
	measurementPrefix := prom.MeasurementPrefix

	var pts []*iod.Point

	// iterate all metrics
	for name, value := range metricFamilies {
		familyType := value.GetType()
		var fieldName string

		valid := isValid(familyType, name, metricTypes, metricNameFilter)
		if !valid {
			continue
		}

		msName, fieldName := getNames(name,
			customMeasurementRules, measurementName, measurementPrefix)

		// set default name when measurementName is empty
		if msName == "" {
			msName = "prom" //nolint:goconst
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

				pt, err := iod.MakePoint(msName, tags, fields, time.Now())
				if err != nil {
					lastErr = err
				} else {
					pts = append(pts, pt)
				}
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

				pt, err := iod.MakePoint(msName, tags, fields, time.Now())
				if err != nil {
					lastErr = err
				} else {
					pts = append(pts, pt)
				}
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

				pt, err := iod.MakePoint(msName, tags, fields, time.Now())
				if err != nil {
					lastErr = err
				} else {
					pts = append(pts, pt)
				}
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

				pt, err := iod.MakePoint(msName, tags, fields, time.Now())
				if err != nil {
					lastErr = err
				} else {
					pts = append(pts, pt)
				}

				for _, q := range quantiles {
					quantile := q.GetQuantile() // 0 0.25 0.5 0.75 1
					val := q.GetValue()         // value

					fields := make(map[string]interface{})
					fields[fieldName] = val

					labels := m.GetLabel()
					tags := getTags(labels, prom.Tags, extraTags, prom.TagsIgnore)

					tags["quantile"] = fmt.Sprint(quantile)

					pt, err := iod.MakePoint(msName, tags, fields, time.Now())
					if err != nil {
						lastErr = err
					} else {
						pts = append(pts, pt)
					}
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

				pt, err := iod.MakePoint(msName, tags, fields, time.Now())
				if err != nil {
					lastErr = err
				} else {
					pts = append(pts, pt)
				}
				for _, b := range buckets {
					count := b.GetCumulativeCount()
					bond := b.GetUpperBound()

					fields := make(map[string]interface{})
					fields[fieldName+"_bucket"] = count

					labels := m.GetLabel()
					tags := getTags(labels, prom.Tags, extraTags, prom.TagsIgnore)
					tags["le"] = fmt.Sprint(bond)

					pt, err := iod.MakePoint(msName, tags, fields, time.Now())
					if err != nil {
						lastErr = err
					} else {
						pts = append(pts, pt)
					}
				}
			}
		}
	}

	return pts, lastErr
}

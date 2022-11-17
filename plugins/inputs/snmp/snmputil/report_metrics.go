// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import "fmt"

// MetricSample is a collected metric sample with its metadata, ready to be submitted through the metric sender.
type MetricSample struct {
	value      ResultValue
	tags       []string
	symbol     SymbolConfig
	forcedType string
	options    MetricsConfigOption
}

// ReportMetrics reports metrics using Sender.
func ReportMetrics(metrics []MetricsConfig, values *ResultValueStore, tags []string, outData *MetricDatas) {
	scalarSamples := make(map[string]MetricSample)
	columnSamples := make(map[string]map[string]MetricSample)

	for _, metric := range metrics {
		if metric.IsScalar() {
			sample, err := reportScalarMetrics(metric, values, tags, outData)
			if err != nil {
				continue
			}
			if _, ok := EvaluatedSampleDependencies[sample.symbol.Name]; !ok {
				continue
			}
			scalarSamples[sample.symbol.Name] = sample
		} else if metric.IsColumn() {
			samples := reportColumnMetrics(metric, values, tags, outData)

			for name, sampleRows := range samples {
				if _, ok := EvaluatedSampleDependencies[name]; !ok {
					continue
				}
				columnSamples[name] = sampleRows
			}
		}
	}

	err := tryReportMemoryUsage(scalarSamples, columnSamples, outData)
	if err != nil {
		l.Debugf("error reporting memory usage : %v", err)
	}
}

// GetCheckInstanceMetricTags returns check instance metric tags.
func GetCheckInstanceMetricTags(metricTags []MetricTagConfig, values *ResultValueStore) []string {
	var globalTags []string

	for _, metricTag := range metricTags {
		// TODO: Support extract value see II-635
		value, err := values.GetScalarValue(metricTag.OID)
		if err != nil {
			continue
		}
		strValue, err := value.ToString()
		if err != nil {
			l.Debugf("error converting value (%#v) to string : %v", value, err)
			continue
		}
		globalTags = append(globalTags, metricTag.GetTags(strValue)...)
	}
	return globalTags
}

func reportScalarMetrics(metric MetricsConfig, values *ResultValueStore, tags []string, outData *MetricDatas) (MetricSample, error) {
	value, err := getScalarValueFromSymbol(values, metric.Symbol)
	if err != nil {
		l.Debugf("report scalar: error getting scalar value: %v", err)
		return MetricSample{}, err
	}

	scalarTags := CopyStrings(tags)
	scalarTags = append(scalarTags, metric.GetSymbolTags()...)
	sample := MetricSample{
		value:      value,
		tags:       scalarTags,
		symbol:     metric.Symbol,
		forcedType: metric.ForcedType,
		options:    metric.Options,
	}
	sendMetric(sample, outData)
	return sample, nil
}

//nolint:lll
func reportColumnMetrics(metricConfig MetricsConfig, values *ResultValueStore, tags []string, outData *MetricDatas) map[string]map[string]MetricSample {
	rowTagsCache := make(map[string][]string)
	samples := map[string]map[string]MetricSample{}
	for _, symbol := range metricConfig.Symbols {
		metricValues, err := getColumnValueFromSymbol(values, symbol)
		if err != nil {
			continue
		}
		for fullIndex, value := range metricValues {
			// cache row tags by fullIndex to avoid rebuilding it for every column rows
			if _, ok := rowTagsCache[fullIndex]; !ok {
				tmpTags := CopyStrings(tags)
				tmpTags = append(tmpTags, metricConfig.StaticTags...)
				tmpTags = append(tmpTags, getTagsFromMetricTagConfigList(metricConfig.MetricTags, fullIndex, values)...)
				rowTagsCache[fullIndex] = tmpTags
			}
			rowTags := rowTagsCache[fullIndex]
			sample := MetricSample{
				value:      value,
				tags:       rowTags,
				symbol:     symbol,
				forcedType: metricConfig.ForcedType,
				options:    metricConfig.Options,
			}
			sendMetric(sample, outData)
			if _, ok := samples[sample.symbol.Name]; !ok {
				samples[sample.symbol.Name] = make(map[string]MetricSample)
			}
			samples[sample.symbol.Name][fullIndex] = sample
			trySendBandwidthUsageMetric(symbol, fullIndex, values, rowTags, outData)
		}
	}
	return samples
}

func sendMetric(metricSample MetricSample, outData *MetricDatas) {
	metricFullName := metricSample.symbol.Name
	forcedType := metricSample.forcedType
	if forcedType == "" {
		if metricSample.value.SubmissionType != "" {
			forcedType = metricSample.value.SubmissionType
		} else {
			forcedType = "gauge"
		}
	} else if forcedType == "flag_stream" {
		strValue, err := metricSample.value.ToString()
		if err != nil {
			l.Debugf("error converting value (%#v) to string : %v", metricSample.value, err)
			return
		}
		options := metricSample.options
		floatValue, err := getFlagStreamValue(options.Placement, strValue)
		if err != nil {
			l.Debugf("metric `%s`: failed to get flag stream value: %s", metricFullName, err)
			return
		}
		metricFullName = metricFullName + "." + options.MetricSuffix
		metricSample.value = ResultValue{Value: floatValue}
		forcedType = "gauge"
	}

	floatValue, err := metricSample.value.ToFloat64()
	if err != nil {
		l.Debugf("metric `%s`: failed to convert to float64: %s", metricFullName, err)
		return
	}

	scaleFactor := metricSample.symbol.ScaleFactor
	if scaleFactor != 0 {
		floatValue *= scaleFactor
	}

	switch forcedType {
	case "gauge", "counter", "monotonic_count":
		outData.Add(metricFullName, floatValue, metricSample.tags)
	case "percent":
		outData.Add(metricFullName, floatValue*100, metricSample.tags)
	case "monotonic_count_and_rate":
		outData.Add(metricFullName, floatValue, metricSample.tags)
		outData.Add(metricFullName+".rate", floatValue, metricSample.tags)
	default:
		l.Debugf("metric `%s`: unsupported forcedType: %s", metricFullName, forcedType)
		return
	}
}

func getFlagStreamValue(placement uint, strValue string) (float64, error) {
	index := placement - 1
	if int(index) >= len(strValue) {
		return 0, fmt.Errorf("flag stream index `%d` not found in `%s`", index, strValue)
	}
	floatValue := 0.0
	if strValue[index] == '1' {
		floatValue = 1.0
	}
	return floatValue, nil
}

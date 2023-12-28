// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promremote

import (
	"strings"
)

// Checks whether a metric name should be filtered by checking whether it matches any regex
// specified in p.MetricNameFilter. It passes if metric name filter is empty.
func (p *Parser) shouldFilterMetricName(metric string) bool {
	if len(p.metricNameReFilter) == 0 {
		return true
	}
	for _, filter := range p.metricNameReFilter {
		match := filter.MatchString(metric)
		if match {
			return true
		}
	}
	// Did not match any regex.
	return false
}

// Checks if by default rule, the measurement name we've gotten should be filtered through.
func (p *Parser) shouldFilterMeasurementName(metric string) bool {
	if len(p.measurementNameReFilter) == 0 {
		return true
	}
	for _, filter := range p.measurementNameReFilter {
		match := filter.MatchString(metric)
		if match {
			return true
		}
	}
	return false
}

func (p *Parser) getNames(metric string) (measurementName, metricName string) {
	if p.MeasurementName != "" {
		return p.MeasurementPrefix + p.MeasurementName, metric
	}
	// Split measurement name and metric name by the first '_' met.
	measurementName, metricName = p.getNamesByDefaultRule(metric)
	return p.MeasurementPrefix + measurementName, metricName
}

func (*Parser) getNamesByDefaultRule(metric string) (measurementName, metricName string) {
	measurementName, metricName = metric, metric
	index := strings.Index(metric, "_")
	if index != -1 {
		measurementName = metric[:index]
		metricName = metric[index+1:]
	}
	return
}

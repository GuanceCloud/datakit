// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promremote

import (
	"regexp"
	"strings"
)

// isValid checks whether a metric name is valid by
// checking whether it matches any pattern in p.MetricNameFilter.
func (p *Parser) isValid(name string) bool {
	metricNameFilter := p.MetricNameFilter
	nameValid := false
	if len(metricNameFilter) == 0 {
		nameValid = true
	} else {
		for _, filter := range metricNameFilter {
			match, err := regexp.MatchString(filter, name)
			if err != nil {
				continue
			}
			if match {
				nameValid = true
				break
			}
		}
	}
	return nameValid
}

func (p *Parser) getNames(metric string) (measurementName, metricName string) {
	if p.MeasurementName != "" {
		return p.MeasurementPrefix + p.MeasurementName, metric
	}
	// Split measurement name and metric name by the first '_' met.
	index := strings.Index(metric, "_")
	measurementName, metricName = metric, metric
	if index != -1 {
		measurementName = metric[:index]
		metricName = metric[index+1:]
	}
	return p.MeasurementPrefix + measurementName, metricName
}

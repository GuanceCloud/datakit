package promremote

import "regexp"

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

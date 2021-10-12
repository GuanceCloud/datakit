package prom_remote_write

import "regexp"

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

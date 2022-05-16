// Package convertutil contains convert util
package convertutil

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

// example "metric" -> "/v1/write/metrics"
func GetMapCategoryShortToFull(categoryShort string) (string, error) {
	var out string

	switch categoryShort {
	case datakit.CategoryMetric:
		out = datakit.Metric
	case datakit.CategoryNetwork:
		out = datakit.Network
	case datakit.CategoryKeyEvent:
		out = datakit.KeyEvent
	case datakit.CategoryObject:
		out = datakit.Object
	case datakit.CategoryCustomObject:
		out = datakit.CustomObject
	case datakit.CategoryLogging:
		out = datakit.Logging
	case datakit.CategoryTracing:
		out = datakit.Tracing
	case datakit.CategoryRUM:
		out = datakit.RUM
	case datakit.CategorySecurity:
		out = datakit.Security
	default:
		return "", fmt.Errorf("unrecognized category")
	}

	return out, nil
}

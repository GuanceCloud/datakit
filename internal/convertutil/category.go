// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package convertutil contains convert util
package convertutil

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

// GetMapCategoryShortToFull ...
// example "metric" -> "/v1/write/metrics".
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
	case datakit.CategoryProfiling:
		out = datakit.Profiling
	default:
		return "", fmt.Errorf("unrecognized category")
	}

	return out, nil
}

// GetMapCategoryFullToShort ...
// example "/v1/write/metrics" -> "metric".
func GetMapCategoryFullToShort(categoryFull string) (string, error) {
	var out string

	switch categoryFull {
	case datakit.Metric:
		out = datakit.CategoryMetric
	case datakit.Network:
		out = datakit.CategoryNetwork
	case datakit.KeyEvent:
		out = datakit.CategoryKeyEvent
	case datakit.Object:
		out = datakit.CategoryObject
	case datakit.CustomObject:
		out = datakit.CategoryCustomObject
	case datakit.Logging:
		out = datakit.CategoryLogging
	case datakit.Tracing:
		out = datakit.CategoryTracing
	case datakit.RUM:
		out = datakit.CategoryRUM
	case datakit.Security:
		out = datakit.CategorySecurity
	case datakit.Profiling:
		out = datakit.CategoryProfiling
	default:
		return "", fmt.Errorf("unrecognized category")
	}

	return out, nil
}

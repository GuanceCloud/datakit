// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package utils wraps utilities used by pipeline
package utils

import (
	"os"

	"github.com/GuanceCloud/cliutils/point"
)

func PtCatOption(cat point.Category) []point.Option {
	var opt []point.Option
	switch cat {
	case point.Logging, point.DialTesting:
		opt = point.DefaultLoggingOptions()
	case point.Tracing,
		point.Network,
		point.KeyEvent,
		point.RUM,
		point.Security,
		point.Profiling:
		opt = point.CommonLoggingOptions()
	case point.Object,
		point.CustomObject:
		opt = point.DefaultObjectOptions()
	case point.Metric:
		opt = point.DefaultMetricOptions()

	case point.DynamicDWCategory,
		point.MetricDeprecated,
		point.UnknownCategory: // pass

	default:
	}
	return opt
}

func FileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

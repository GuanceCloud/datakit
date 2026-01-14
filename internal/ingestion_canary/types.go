// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ingestioncanary

type Category string

const (
	MetricCategory  Category = "M"
	LoggingCategory Category = "L"
	TracingCategory Category = "T"
)

type Feed struct {
	TimeMs int64
	Round  int64
}

type QueryResult struct {
	TimeMs int64
	Round  int64
}

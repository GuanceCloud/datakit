// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package mapper contains graphite mapper observer
package mapper

type ObserverType string

const (
	ObserverTypeHistogram ObserverType = "histogram"
	ObserverTypeSummary   ObserverType = "summary"
	ObserverTypeDefault   ObserverType = ""
)

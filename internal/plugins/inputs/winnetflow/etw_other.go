// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows

package winnetflow

import "errors"

func newETWCollector(_ *flowAggregator, _ etwConfig) (flowSource, error) {
	return nil, errors.New("ETW collector is only supported on Windows")
}

func newHTTPCollector(_ *httpAggregator, _ etwConfig, _, _ int) (flowSource, error) {
	return nil, errors.New("HTTP.sys ETW collector is only supported on Windows")
}

// httpStats mirrors the Windows-only stats type so the input can compile on
// non-Windows targets; the HTTP collector is never started there.
type httpStats struct {
	decoded          uint64
	dropped          uint64
	parseErrors      uint64
	completed        uint64
	missedConn       uint64
	missedReq        uint64
	evictedReq       uint64
	droppedReq       uint64
	requestsSkipped  uint64
	invalidFiltered  uint64
	loopbackFiltered uint64
	session          etwSessionStats
}

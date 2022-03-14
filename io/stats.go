// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"time"
)

type InputsStat struct {
	// Name      string    `json:"name"`
	Category       string        `json:"category"`
	Frequency      string        `json:"frequency,omitempty"`
	AvgSize        int64         `json:"avg_size"`
	Total          int64         `json:"total"`
	Count          int64         `json:"count"`
	First          time.Time     `json:"first"`
	Last           time.Time     `json:"last"`
	LastErr        string        `json:"last_error,omitempty"`
	LastErrTS      time.Time     `json:"last_error_ts,omitempty"`
	Version        string        `json:"version,omitempty"`
	MaxCollectCost time.Duration `json:"max_collect_cost"`
	AvgCollectCost time.Duration `json:"avg_collect_cost"`

	totalCost time.Duration `json:"-"`
}

func dumpStats(is map[string]*InputsStat) (res map[string]*InputsStat) {
	res = map[string]*InputsStat{}
	for x, y := range is {
		res[x] = y
	}
	return
}

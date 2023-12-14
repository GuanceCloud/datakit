// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package time wraps time related functions
package time

import (
	"time"
)

func NewAlignedTicker(interval time.Duration) *time.Ticker {
	ts := time.Now().UnixNano()
	sl := int64(interval) - ts%int64(interval)
	time.Sleep(time.Nanosecond * time.Duration(sl))
	return time.NewTicker(interval)
}

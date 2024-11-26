// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

import (
	"math"
	"time"
)

func AlignTimeMillSec(triggerTime time.Time, lastAlignTime int64, intervalMillSec int64) int64 {
	tT := triggerTime.UnixMilli()
	lastAlignTime += intervalMillSec
	if d := math.Abs(float64(tT - lastAlignTime)); d > 0 && d/float64(intervalMillSec) > 0.1 {
		return tT
	}
	return lastAlignTime
}

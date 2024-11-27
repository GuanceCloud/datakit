// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

import (
	"math"
	"time"
)

func AlignTimeMillSec(triggerTime time.Time, lastts, intervalMillSec int64) (nextts int64) {
	tt := triggerTime.UnixMilli()
	nextts = lastts + intervalMillSec
	if d := math.Abs(float64(tt - nextts)); d > 0 && d/float64(intervalMillSec) > 0.1 {
		nextts = tt
	}
	return nextts
}

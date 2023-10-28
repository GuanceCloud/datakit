// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"golang.org/x/exp/slices"
)

// SortByTime sort(ASC) pts according to time.
func SortByTime(pts []*Point) {
	slices.SortFunc(pts, func(l, r *Point) int {
		diff := l.Time().Sub(r.Time())
		if diff == 0 {
			return 0
		} else if diff > 0 {
			return 1
		} else {
			return -1
		}
	})
}

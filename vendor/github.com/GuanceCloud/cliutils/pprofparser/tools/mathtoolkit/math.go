// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package mathtoolkit is a collection of math utils.
package mathtoolkit

import "math"

func Trunc(x float64) int64 {
	return int64(math.Trunc(x))
}

func Floor(x float64) int64 {
	return int64(math.Floor(x))
}

func Ceil(x float64) int64 {
	return int64(math.Ceil(x))
}

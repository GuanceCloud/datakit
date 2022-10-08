// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package testutils

import (
	"testing"
)

func TestRandInt64(t *testing.T) {
	for i := 1; i <= 30; i++ {
		for j := 0; j < 10; j++ {
			t.Log(RandInt64(i))
		}
	}
}

func TestRandWithinInts(t *testing.T) {
	data := []int{2, 3, 45, 9, 67, 8, 9}
	for i := 0; i < 10; i++ {
		t.Log(RandWithinInts(data))
	}
}

func TestRandStrID(t *testing.T) {
	for i := 1; i <= 30; i++ {
		for j := 0; j < 10; j++ {
			t.Log(RandStrID(i))
		}
	}
}

func TestRandTime(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Log(RandTime().String())
	}
}

func TestRandPoint(t *testing.T) {
	pnt := RandPoint("test_utils", 30, 90)
	t.Log(pnt.String())
	pnts := RandPoints(100, 10, 30)
	for i := range pnts {
		t.Log(pnts[i].String())
	}
}

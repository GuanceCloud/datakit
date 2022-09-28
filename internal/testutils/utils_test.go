// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package testutils

import (
	"fmt"
	"log"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
)

func TestRandInt64(t *testing.T) {
	for i := 1; i <= 30; i++ {
		for j := 0; j < 10; j++ {
			fmt.Println(RandInt64(i))
		}
	}
}

func TestRandWithinInts(t *testing.T) {
	data := []int{2, 3, 45, 9, 67, 8, 9}
	for i := 0; i < 10; i++ {
		log.Println(RandWithinInts(data))
	}
}

func TestRandStrID(t *testing.T) {
	for i := 1; i <= 30; i++ {
		for j := 0; j < 10; j++ {
			fmt.Println(RandStrID(i))
		}
	}
}

func TestRandTime(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(RandTime().String())
	}
}

func TestRandPoint(t *testing.T) {
	encoder := lineproto.NewLineEncoder()
	pnt := RandPoint("test_utils", 30, 90)
	if err := encoder.AppendPoint(pnt); err != nil {
		t.Fatal(err)
	}
	line, err := encoder.UnsafeString()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(line)
	pnts := RandPoints(100, 10, 30)
	for i := range pnts {
		encoder.Reset()
		if err := encoder.AppendPoint(pnts[i]); err != nil {
			t.Fatal(err)
		}
		line, err := encoder.UnsafeStringWithoutLn()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(line)
	}
}

package testutils

import (
	"fmt"
	"testing"
)

func TestRandInt64(t *testing.T) {
	for i := 1; i < 10; i++ {
		for j := 0; j < 10; j++ {
			fmt.Println(RandInt64(i))
		}
	}
}

func TestRandTime(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(RandTime().String())
	}
}

func TestRandPoint(t *testing.T) {
	pnt := RandPoint("test_utils", 30, 90)
	fmt.Println(pnt.String())
	pnts := RandPoints(100, 10, 30)
	for i := range pnts {
		fmt.Println(pnts[i].String())
	}
}

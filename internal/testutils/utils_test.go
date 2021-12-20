package testutils

import (
	"fmt"
	"testing"
)

func TestRandPoint(t *testing.T) {
	pnt := RandPoint("test_utils", 30, 90)
	fmt.Println(pnt.String())
	pnts := RandPoints(100, 10, 30)
	for i := range pnts {
		fmt.Println(pnts[i].String())
	}
}

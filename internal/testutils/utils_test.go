package testutils

import (
	"fmt"
	"testing"
)

func TestRamPoint(t *testing.T) {
	pnt := RamPoint("test_utils", 30, 90)
	fmt.Println(pnt.String())
	pnts := RamPoints(100, 10, 30)
	for i := range pnts {
		fmt.Println(pnts[i].String())
	}
}

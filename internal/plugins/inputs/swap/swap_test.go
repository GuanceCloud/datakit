// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package swap

import (
	"testing"

	"github.com/shirou/gopsutil/mem"
)

func TestCollec(t *testing.T) {
	i := &Input{swapStat: PSSwapStat4Test}

	for x := 0; x < 6; x++ {
		if err := i.Collect(); err != nil {
			t.Errorf("collect swap stat error:%s", err)
		}
	}

	if len(i.collectCache) != 1 {
		t.Errorf("Need to clear collectCache")
	}

	expected := map[string]interface{}{
		"total":        int64(111),
		"used":         int64(222),
		"free":         uint(333),
		"used_percent": float64(50),
		"sin":          int64(100),
		"sout":         int64(100),
	}

	actual := i.collectCacheLast1Ptr.InfluxFields()

	t.Log("actual =", actual)
	t.Log("expected =", expected)

	AssertFields(t, actual, expected)
}

func PSSwapStat4Test() (*mem.SwapMemoryStat, error) {
	return &mem.SwapMemoryStat{
		Total:       111,
		Used:        222,
		Free:        333,
		UsedPercent: 50,
		Sin:         100,
		Sout:        100,
	}, nil
}

func AssertFields(t *testing.T, actual, expected map[string]interface{}) {
	t.Helper()

	for k := range expected {
		ed := expected[k]
		if al, ok := actual[k]; ok {
			switch v := ed.(type) {
			case uint64:
				if v != al.(uint64) {
					t.Errorf("error: "+k+" expected: %f \t actual %f", v, al.(uint64))
				}
			case float64:
				if v != al.(float64) {
					t.Errorf("error: "+k+" expected: %f \t actual %f", v, al.(float64))
				}
			}
		}
	}
}

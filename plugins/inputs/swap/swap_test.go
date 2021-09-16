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
		"total":        uint64(111),
		"used":         uint64(222),
		"free":         uint(333),
		"used_percent": float64(50),
		"sin":          uint64(100),
		"sout":         uint64(100),
	}

	actual := i.collectCacheLast1Ptr.(*swapMeasurement).fields
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
	for k := range expected {
		ed := expected[k]
		if al, ok := actual[k]; ok {
			switch ed.(type) {
			case uint64:
				if ed.(uint64) != al.(uint64) {
					t.Errorf("error: "+k+" expected: %f \t actual %f", ed.(uint64), al.(uint64))
				}
			case float64:
				if ed.(float64) != al.(float64) {
					t.Errorf("error: "+k+" expected: %f \t actual %f", ed.(float64), al.(float64))
				}
			}
		}
	}
}

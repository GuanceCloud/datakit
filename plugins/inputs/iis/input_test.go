// +build windows,amd64

package iis

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/win_utils/pdh"
	"golang.org/x/sys/windows"
)

// 非英语环境下，对于低版本 Windows 操作系统
// 测试可能失败，
// 原因是性能计数器名被本地化
func TestPerfObjMetricMap(t *testing.T) {
	for _, counterMap := range PerfObjMetricMap {
		for k, v := range counterMap {
			_, counterList, ret := pdh.PdhEnumObjectItems(k)
			if ret != uint32(windows.ERROR_SUCCESS) {
				t.Errorf("Failed to enumerate the instance and counter of object %s", k)
			}
			var counterList2Map = map[string]string{}
			for c := range counterList {
				counterList2Map[counterList[c]] = ""
			}
			for c := range v {
				if _, ok := counterList2Map[c]; !ok {
					t.Errorf("The counter %s in PerfObjMetricMap is not in the list of collected counters", c)
				}
			}
		}
	}
}

func TestCollect(t *testing.T) {
	i := Input{Interval: datakit.Duration{Duration: time.Second * 15}}
	if err := i.Collect(); err != nil {
		t.Error(err)
	}
}

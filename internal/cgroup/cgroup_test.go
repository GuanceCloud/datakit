// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cgroup

import (
	T "testing"
	"time"
)

func TestProcessInfo(t *T.T) {
	t.Skip()
	t.Run("ctx-switch", func(t *T.T) {
		for i := 0; i < 10; i++ {
			ctxswitch := MyCtxSwitch()
			if ctxswitch == nil {
				return
			}

			t.Logf("switch: %+#v", ctxswitch)
			t.Logf("switch: %+#v", ctxswitch)
			t.Logf("switch: %+#v", ctxswitch)
			t.Logf("switch: %+#v", ctxswitch)

			time.Sleep(time.Second)
		}
	})
}

func TestCPUUsage(t *T.T) {
	t.Skip()
	t.Run("cpu-100", func(t *T.T) {
		for i := 0; i < 4; i++ {
			go func() {
				n := 0
				for {
					if n%10000000000 == 0 {
						time.Sleep(time.Second * 3)
						t.Logf("sleep once...")
					}
					n++
				}
			}()
		}

		for {
			cpu, _ := MyCPUPercent(time.Second)
			mem, _ := MyMemPercent()
			t.Logf("cpu perc: %f, mem perc: %f", cpu, mem)
			time.Sleep(time.Second)
		}
	})
}

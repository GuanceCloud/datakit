// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package resourcelimit

import (
	"runtime"
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestSetup(t *T.T) {
	t.Run("cpumax-to-cpucores", func(t *T.T) {
		c := ResourceLimitOptions{
			CPUMax: 10.0,
		}

		cores := float64(runtime.NumCPU())
		c.Setup()
		assert.Equal(t, cores*c.CPUMax/100.0, c.CPUCores)
	})

	t.Run("cpucores-to-cpumax", func(t *T.T) {
		c := ResourceLimitOptions{
			CPUCores: 1.0,
		}

		cores := float64(runtime.NumCPU())
		c.Setup()
		assert.Equal(t, c.CPUCores/cores*100.0, c.CPUMax)
	})

	t.Run("100%-cpumax", func(t *T.T) {
		c := ResourceLimitOptions{
			CPUMax: 101.0,
		}

		cores := float64(runtime.NumCPU())
		c.Setup()
		assert.Equal(t, 100.0, c.CPUMax)
		assert.Equal(t, cores, c.CPUCores)
	})

	t.Run("cpucores-plus-1", func(t *T.T) {
		c := ResourceLimitOptions{
			CPUCores: float64(runtime.NumCPU() + 1),
		}

		cores := float64(runtime.NumCPU())
		c.Setup()
		assert.Equal(t, 100.0, c.CPUMax)
		assert.Equal(t, cores, c.CPUCores)
	})
}

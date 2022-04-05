// Package cgroup wraps Linux cgroup functions.
package cgroup

import (
	"context"
	"time"

	"github.com/containerd/cgroups"
	"github.com/shirou/gopsutil/cpu"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var l = logger.DefaultSLogger("cgroup")

type Cgroup struct {
	Enable bool `toml:"enable"`

	Path      string  `toml:"path"`
	CPUMax    float64 `toml:"cpu_max"`
	CPUMin    float64 `toml:"cpu_min"`
	cpuHigh   float64
	cpuLow    float64
	quotaHigh int64
	quotaLow  int64
	waitNum   int
	level     string

	MemMax     int64 `toml:"mem_max_mb"`
	memMaxSwap int64

	DisableOOM bool `toml:"disable_oom"`

	control cgroups.Cgroup
}

func Run(c *Cgroup) {
	l = logger.SLogger("cgroup")

	if c == nil || !c.Enable {
		return
	}

	if !(0 < c.CPUMax && c.CPUMax < 100) ||
		!(0 < c.CPUMin && c.CPUMin < 100) {
		l.Errorf("CPUMax and CPUMin should be in range of (0.0, 100.0)")
		return
	}

	if c.CPUMax < c.CPUMin {
		l.Errorf("CPUMin should less than CPUMax of the cgroup")
		return
	}

	g := datakit.G("cgroup")

	g.Go(func(ctx context.Context) error {
		start(c)
		return nil
	})
}

func GetCPUPercent(interval time.Duration) (float64, error) {
	percent, err := cpu.Percent(interval, false)
	if err != nil {
		return 0, err
	}

	if len(percent) == 0 {
		return 0, nil
	}
	return percent[0] / 100, nil //nolint:gomnd
}

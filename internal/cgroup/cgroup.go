// Package cgroup wraps Linux cgroup functions.
package cgroup

import (
	"time"

	"github.com/shirou/gopsutil/cpu"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var l = logger.DefaultSLogger("cgroup")

type Cgroup struct {
	Enable bool    `toml:"enable"`
	CPUMax float64 `toml:"cpu_max"`
	CPUMin float64 `toml:"cpu_min"`
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

	start(c)
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

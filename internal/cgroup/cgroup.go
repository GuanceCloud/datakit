// Package cgroup wraps Linux cgroup functions.
package cgroup

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	cg *Cgroup
	l  = logger.DefaultSLogger("cgroup")
)

const (
	MB = 1024 * 1024
)

type CgroupOptions struct {
	Enable     bool    `toml:"enable"`
	Path       string  `toml:"path"`
	CPUMax     float64 `toml:"cpu_max"`
	CPUMin     float64 `toml:"cpu_min"`
	MemMax     int64   `toml:"mem_max_mb"`
	DisableOOM bool    `toml:"disable_oom,omitempty"`
}

func Run(c *CgroupOptions) {
	l = logger.SLogger("cgroup")

	if c == nil || !c.Enable {
		return
	}

	cg = &Cgroup{opt: c}

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
		cg.start()
		return nil
	})
}

func (c *Cgroup) String() string {
	if !c.opt.Enable {
		return "-"
	}

	return fmt.Sprintf("path: %s, mem: %dMB, cpu: [%.2f:%.2f]",
		c.opt.Path, c.opt.MemMax/MB, c.opt.CPUMin, c.opt.CPUMax)
}

func Info() string {
	if cg == nil {
		return "not ready"
	}

	switch runtime.GOOS {
	case "linux":
		if cg.err != nil {
			return cg.err.Error()
		} else {
			return cg.String()
		}

	default:
		return "-"
	}
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

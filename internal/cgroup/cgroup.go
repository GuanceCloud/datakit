// Package cgroup wraps Linux cgroup functions.
package cgroup

import (
	"time"

	"github.com/shirou/gopsutil/cpu"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

var l = logger.DefaultSLogger("cgroup")

func Run() {
	l = logger.SLogger("cgroup")

	if config.Cfg.Cgroup == nil || !config.Cfg.Cgroup.Enable {
		return
	}

	if !(0 < config.Cfg.Cgroup.CPUMax && config.Cfg.Cgroup.CPUMax < 100) ||
		!(0 < config.Cfg.Cgroup.CPUMin && config.Cfg.Cgroup.CPUMin < 100) {
		l.Errorf("CPUMax and CPUMin should be in range of (0.0, 100.0)")
		return
	}

	if config.Cfg.Cgroup.CPUMax < config.Cfg.Cgroup.CPUMin {
		l.Errorf("CPUMin should less than CPUMax of the cgroup")
		return
	}

	start()
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

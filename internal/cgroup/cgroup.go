package cgroup

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

var (
	l = logger.DefaultSLogger("cgroup")
)

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

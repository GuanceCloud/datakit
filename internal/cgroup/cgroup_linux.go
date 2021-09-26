package cgroup

import (
	"os"
	"runtime"
	"time"

	"github.com/containerd/cgroups"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

// 1 second
var period = uint64(1000000) //nolint:gomnd

const (
	H = "high"
	L = "low"
)

//nolint:cyclop
func start() {
	// config.Cfg.Cgroup 是百分比
	high := config.Cfg.Cgroup.CPUMax * float64(runtime.NumCPU()) / 100 //nolint:gomnd
	low := config.Cfg.Cgroup.CPUMin * float64(runtime.NumCPU()) / 100  //nolint:gomnd

	quotaHigh := int64(float64(period) * high)
	quotaLow := int64(float64(period) * low)

	pid := os.Getpid()

	l.Infof("with %d CPU, set CPU limimt %.2f%%",
		runtime.NumCPU(), float64(quotaLow)/float64(period)*100.0) //nolint:gomnd

	control, err := cgroups.New(cgroups.V1, cgroups.StaticPath("/datakit"),
		&specs.LinuxResources{
			CPU: &specs.LinuxCPU{
				Period: &period,
				Quota:  &quotaLow,
			},
		})
	if err != nil {
		l.Errorf("failed of new cgroup: %s", err)
		return
	}
	defer func() {
		if err := control.Delete(); err != nil {
			l.Warnf("control.Delete(): %s, ignored", err.Error())
		}
	}()

	if err := control.Add(cgroups.Process{Pid: pid}); err != nil {
		l.Errorf("faild of add cgroup: %s", err)
		return
	}

	l.Infof("add PID %d to cgroup", pid)

	level := L
	waitNum := 0
	for {
		percpu, err := GetCPUPercent(time.Second * 3) //nolint:gomnd
		if err != nil {
			l.Warnf("GetCPUPercent: %s, ignored", err)
			continue
		}

		var q int64

		// 当前 cpu 使用率 + 设定的最大使用率 超过 95% 时，将使用 low 模式
		// 否则如果连续 3 次判断小于 95%，则使用 high 模式
		if 95 < percpu+high {
			if level == L {
				continue
			}
			q = quotaLow
			level = L
		} else {
			if level == H {
				continue
			}
			if waitNum < 3 { //nolint:gomnd
				waitNum++
				continue
			}
			q = quotaHigh
			level = H
			waitNum = 0
		}

		err = control.Update(&specs.LinuxResources{
			CPU: &specs.LinuxCPU{
				Period: &period,
				Quota:  &q,
			},
		})
		if err != nil {
			l.Warnf("failed of update cgroup: %s", err)
			continue
		}

		l.Debugf("switch to quota %.2f%%",
			float64(q)/float64(period)*100.0) //nolint:gomnd
	}
}

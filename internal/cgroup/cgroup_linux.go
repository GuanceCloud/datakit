// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cgroup

import (
	"os"
	"path"
	"runtime"
	"time"

	"github.com/containerd/cgroups"
	"github.com/containerd/cgroups/v3/cgroup2"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	H                  = "high"
	L                  = "low"
	cgroupName         = "datakit"
	defaultCgroup2Path = "/sys/fs/cgroup"
)

type Cgroup struct {
	opt *CgroupOptions

	cpuHigh   float64
	cpuLow    float64
	quotaHigh int64
	quotaLow  int64
	waitNum   int
	level     string

	err error

	control cgroups.Cgroup
	manager *cgroup2.Manager
}

// 1 second.
var period = uint64(1000000) //nolint:gomnd
var cgroupV2 bool

func (c *Cgroup) cpuSetup() {
	c.cpuHigh = c.opt.CPUMax * float64(runtime.NumCPU()) / 100 //nolint:gomnd
	c.cpuLow = c.opt.CPUMin * float64(runtime.NumCPU()) / 100  //nolint:gomnd

	c.quotaHigh = int64(float64(period) * c.cpuHigh)
	c.quotaLow = int64(float64(period) * c.cpuLow)
}

func (c *Cgroup) memSetup() {
	c.opt.MemMax *= MB
}

func cgroupEnabled(mountPoint, name string) bool {
	_, err := os.Stat(path.Join(mountPoint, name))
	return err == nil
}

func (c *Cgroup) setup() error {
	c.cpuSetup()
	c.memSetup()
	// ------------ cgroup v2

	r := &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Period: &period,
			Quota:  &c.quotaLow,
		},
	}

	if c.opt.MemMax > 0 {
		r.Memory = &specs.LinuxMemory{
			Limit:            &c.opt.MemMax,
			DisableOOMKiller: &c.opt.DisableOOM,
		}
		if cgroupEnabled(c.opt.Path, "memory.memsw.limit_in_bytes") {
			r.Memory.Swap = &c.opt.MemMax
		}
	} else {
		l.Infof("memory limit not set")
	}
	pid := os.Getpid()

	if cgroups.Mode() == cgroups.Unified {
		// use cgroupV2
		l.Infof("use cgroup2")
		cgroupV2 = true
		resource := cgroup2.ToResources(r)
		return c.setup2(resource, pid)
	}

	var control cgroups.Cgroup
	var err error
	if control = c.load(); control == nil {
		control, err = cgroups.New(cgroups.V1, cgroups.StaticPath(c.opt.Path), r)
		if err != nil {
			l.Errorf("cgroups.New(%+#v): %s", r, err)
			return err
		}
	}

	c.control = control

	if err := c.control.Add(cgroups.Process{Pid: pid, Subsystem: cgroupName}); err != nil {
		l.Errorf("faild of add cgroup: %s", err)
		return err
	}

	l.Infof("add PID %d to cgroup", pid)

	return nil
}

func (c *Cgroup) setup2(resource *cgroup2.Resources, pid int) error {
	manager, err := cgroup2.Load(c.opt.Path)
	if err != nil {
		l.Infof("can not load form /sys/fs/cgroup/datakit, use new manager.")
	} else {
		// 先删除已有配置，再重新配置。
		if stat, err := manager.Stat(); err == nil {
			l.Infof("old manager stat is :%s", stat.String())
		}
		err = manager.Delete()
		l.Infof("try to delete old cgroup2 manager err=%v", err)
	}

	manager, err = cgroup2.NewManager(defaultCgroup2Path, c.opt.Path, resource)
	if err != nil {
		l.Warnf("new cgroup2 err=%v", err)
		return err
	}

	c.manager = manager
	return manager.AddProc(uint64(pid))
}

func (c *Cgroup) load() cgroups.Cgroup {
	control, err := cgroups.Load(cgroups.V1, cgroups.StaticPath(c.opt.Path))
	if err != nil {
		l.Infof("can not load cgroup Systemd limit config. Use New()")
	} else {
		l.Infof("datakit cgroups load")
		metrics, err := control.Stat()
		if err == nil && metrics.Memory != nil {
			l.Infof("cgroup: MEM =%s  cgroup Memory.Usage=%s  cgroup Memory.Swap=%s",
				metrics.Memory.String(),
				metrics.Memory.Usage.String(),
				metrics.Memory.Swap.String())
			/*
			 如果和配置的不一样，使用Delete删除掉配置， 不可用 Update 修改.
			 直接修改会报错：write /sys/fs/cgroup/memory/datakit/memory.limit_in_bytes: invalid argument
			 原因是：修改值 不可低于现有配置的(limit 不可低于 Swap).
			*/
			// 如果不相同的话，与其修改，倒不如直接删除。
			if c.opt.MemMax != int64(metrics.Memory.Usage.Usage) {
				if err = control.Delete(); err == nil {
					l.Infof("del cgroup config,use new()")
				} else {
					l.Errorf("del cgroup err=%v", err)
				}
				control = nil
			}
		} else {
			l.Infof("control.Stat err =%v", err)
		}
	}

	return control
}

func (c *Cgroup) stop() {
	var err error
	if cgroupV2 {
		//	err = c.manager.Delete() // auto delete
	} else if c.control != nil {
		err = c.control.Delete()
	}

	if err != nil {
		l.Warnf("control.Delete(): %s, ignored", err.Error())
	} else {
		l.Info("cgroup delete OK")
	}
}

func (c *Cgroup) refreshCPULimit() {
	percpu, err := MyCPUPercent(0) //nolint:gomnd
	if err != nil {
		l.Warnf("GetCPUPercent: %s, ignored", err)
		return
	}

	var q int64

	// 当前 cpu 使用率 + 设定的最大使用率 超过 95% 时，将使用 low 模式
	// 否则如果连续 3 次判断小于 95%，则使用 high 模式
	if 95 < percpu+c.cpuHigh {
		if c.level == L {
			return
		}
		q = c.quotaLow
		c.level = L
	} else {
		if c.level == H {
			return
		}

		if c.waitNum < 3 { //nolint:gomnd
			c.waitNum++
			return
		}
		q = c.quotaHigh
		c.level = H
		c.waitNum = 0
	}

	l.Infof("with %d CPU, set CPU limimt [%.2f%%, %.2f%%], Memory limit: %dMB",
		runtime.NumCPU(),
		float64(c.quotaLow)/float64(period)*100.0,
		float64(c.quotaHigh)/float64(period)*100.0,
		c.opt.MemMax/MB) //nolint:gomnd

	r := &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Period: &period,
			Quota:  &q,
		},
	}

	if cgroupV2 {
		err = c.manager.Update(cgroup2.ToResources(r))
	} else {
		err = c.control.Update(r)
	}
	if err != nil {
		l.Warnf("failed of update cgroup(%+#v): %s", r, err)
		return
	}

	l.Debugf("switch to quota %.5f%%",
		float64(q)/float64(period)*100.0) //nolint:gomnd
}

func (c *Cgroup) start() {
	if err := c.setup(); err != nil {
		c.err = err
		return
	}

	defer c.stop()

	tick := time.NewTicker(time.Second * 3)
	for {
		c.refreshCPULimit()
		select {
		case <-tick.C:
			c.show()
		case <-datakit.Exit.Wait():
			l.Info("cgroup exited")
			return
		}
	}
}

func (c *Cgroup) show() {
	if cgroupV2 {
		stat, err := c.manager.Stat()
		if err == nil {
			l.Debugf("cgroup state: %s", stat.String())
		}
	} else {
		l.Debugf("cgroup state: %s", c.control.State())
	}
}

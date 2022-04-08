package cgroup

import (
	"os"
	"runtime"
	"time"

	"github.com/containerd/cgroups"
	"github.com/dustin/go-humanize"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	H               = "high"
	L               = "low"
	defaultMemLimit = 1024 // 1GB
	cgroupName      = "datakit"
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
}

// 1 second.
var period = uint64(1000000) //nolint:gomnd

func (c *Cgroup) cpuSetup() {
	c.cpuHigh = c.opt.CPUMax * float64(runtime.NumCPU()) / 100 //nolint:gomnd
	c.cpuLow = c.opt.CPUMin * float64(runtime.NumCPU()) / 100  //nolint:gomnd

	c.quotaHigh = int64(float64(period) * c.cpuHigh)
	c.quotaLow = int64(float64(period) * c.cpuLow)
}

func (c *Cgroup) memSetup() {
	if c.opt.MemMax <= 0 {
		c.opt.MemMax = defaultMemLimit
		l.Infof("reset Memory limit to %d(%s)",
			defaultMemLimit, humanize.IBytes(uint64(defaultMemLimit)))
	}

	c.opt.MemMax = c.opt.MemMax * MB
}

func (c *Cgroup) setup() error {
	c.cpuSetup()
	c.memSetup()

	r := &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Period: &period,
			Quota:  &c.quotaLow,
		},
		Memory: &specs.LinuxMemory{
			Swap:             &c.opt.MemMax,
			DisableOOMKiller: &c.opt.DisableOOM,
		},
	}

	control, err := cgroups.New(cgroups.V1, cgroups.StaticPath(c.opt.Path), r)
	if err != nil {
		l.Errorf("cgroups.New(%+#v): %s", r, err)
		return err
	}

	c.control = control

	pid := os.Getpid()
	if err := c.control.Add(cgroups.Process{Pid: pid, Subsystem: cgroupName}); err != nil {
		l.Errorf("faild of add cgroup: %s", err)
		return err
	}

	l.Infof("add PID %d to cgroup", pid)

	return nil
}

func (c *Cgroup) stop() {
	if err := c.control.Delete(); err != nil {
		l.Warnf("control.Delete(): %s, ignored", err.Error())
	} else {
		l.Info("cgroup delete OK")
	}
}

func (c *Cgroup) refreshCPULimit() {
	percpu, err := GetCPUPercent(0) //nolint:gomnd
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

	err = c.control.Update(r)
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
	l.Debugf("cgroup state: %s", c.control.State())
}

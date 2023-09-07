// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

// Package cgroup set cgroup limit in linux
package cgroup

import (
	"fmt"
	"os"
	"runtime"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/containerd/cgroups/v3/cgroup2"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

var (
	cg     *Cgroup
	l      = logger.DefaultSLogger("cgroup")
	period = uint64(1000000) //nolint:gomnd
)

const (
	cgroupName         = "datakit"
	defaultCgroup2Path = "/sys/fs/cgroup"
	MB                 = 1024 * 1024
)

type CgroupOptions struct {
	Path   string
	CPUMax float64
	MemMax int64

	DisableOOM bool
	Enable     bool
}

type Cgroup struct {
	opt       *CgroupOptions
	cpuHigh   float64
	quotaHigh int64
	err       error
}

func (c *Cgroup) cpuSetup() {
	c.cpuHigh = c.opt.CPUMax * float64(runtime.NumCPU()) / 100 //nolint:gomnd
	c.quotaHigh = int64(float64(period) * c.cpuHigh)
}

func (c *Cgroup) memSetup() {
	c.opt.MemMax *= MB
}

func (c *Cgroup) makeLinuxResource() *specs.LinuxResources {
	c.cpuSetup()
	c.memSetup()

	resource := &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Period: &period,
			Quota:  &c.quotaHigh,
		},
	}

	if c.opt.MemMax > 0 {
		swappiness := uint64(0)
		swap := c.opt.MemMax
		resource.Memory = &specs.LinuxMemory{
			Limit:            &c.opt.MemMax,
			Swap:             &swap,
			Swappiness:       &swappiness,
			DisableOOMKiller: &c.opt.DisableOOM,
		}
	}

	return resource
}

func (c *Cgroup) setupV1(resource *specs.LinuxResources, pid int) error {
	c.delControl()

	control, err := cgroup1.New(cgroup1.StaticPath(c.opt.Path), resource)
	if err != nil {
		return fmt.Errorf("cgroups.New(%+#v): %w", resource, err)
	}

	return control.Add(cgroup1.Process{Pid: pid, Subsystem: cgroupName})
}

func (c *Cgroup) setupV2(resource *specs.LinuxResources, pid int) error {
	resource2 := cgroup2.ToResources(resource)
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

	manager, err = cgroup2.NewManager(defaultCgroup2Path, c.opt.Path, resource2)
	if err != nil {
		l.Warnf("new cgroup2 err=%v", err)
		return err
	}

	return manager.AddProc(uint64(pid))
}

// delControl delete cgroups config
// 为避免 limit<swap,这里先删除掉内存的配置，再重新使用 New().
func (c *Cgroup) delControl() {
	control, err := cgroup1.Load(cgroup1.StaticPath(c.opt.Path))
	if err != nil {
		l.Infof("can not load cgroup Systemd limit config. Use New()")
	} else {
		if err = control.Delete(); err != nil {
			l.Infof("del cgroup err=%v", err)
			_ = os.RemoveAll(defaultCgroup2Path + "/memory/datakit")
		}
	}
}

func (c *Cgroup) start() error {
	resource := c.makeLinuxResource()
	pid := os.Getpid()
	if cgroups.Mode() == cgroups.Unified {
		l.Infof("use cgroup V2")
		c.err = c.setupV2(resource, pid)
	} else {
		l.Infof("use cgroup V1")
		c.err = c.setupV1(resource, pid)
	}

	if c.err != nil {
		return fmt.Errorf("cgroup setup err=%w", c.err)
	} else {
		l.Infof("add Datakit pid:%d to cgroup", pid)
	}

	return nil
}

func Run(opt *CgroupOptions) error {
	l = logger.SLogger("cgroup")
	cg = &Cgroup{opt: opt}
	return cg.start()
}

func (c *Cgroup) String() string {
	return fmt.Sprintf("path: %s, mem: %dMB, cpu: %.2f",
		c.opt.Path, c.opt.MemMax/MB, c.opt.CPUMax)
}

func Info() string {
	if cg == nil {
		return "not ready"
	}

	if cg.err != nil {
		return cg.err.Error()
	} else {
		return cg.String()
	}
}

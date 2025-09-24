// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package resourcelimit limit datakit cpu or memory usage
package resourcelimit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/shirou/gopsutil/v3/process"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var l = logger.DefaultSLogger("resourcelimit")

var (
	self                 *process.Process
	resourceLimitOpt     *ResourceLimitOptions
	errProcessInitFailed = errors.New("process init failed")
	userName             string
)

const (
	MB = 1024 * 1024
)

type ResourceLimitOptions struct {
	Path string `toml:"path"`

	// CPU usage percent(max is 100)
	// Deprecated: use CPUCores
	CPUMaxDeprecated float64 `toml:"cpu_max,omitzero"`

	cpuMax   float64
	CPUCores float64 `toml:"cpu_cores"` // CPU cores, 1.0 means 1 core

	MemMax int64 `toml:"mem_max_mb"`

	DisableOOM bool `toml:"disable_oom,omitempty"`
	Enable     bool `toml:"enable"`
}

func CPUMaxToCores(x float64) float64 {
	if x > 100.0 {
		x = 100.0
	}
	return x * float64(runtime.NumCPU()) / 100.0
}

func CPUCoresToCPUMax(x float64) float64 {
	return x / float64(runtime.NumCPU()) * 100.0
}

func (c *ResourceLimitOptions) CPUMax() float64 {
	return c.cpuMax
}

// Setup set internal values of resource limits.
func (c *ResourceLimitOptions) Setup() {
	if c.CPUCores <= 0 { // CPU-cores not set, default to 2C.
		c.CPUCores = 2
	}

	if c.CPUMaxDeprecated > 0 {
		c.CPUCores = CPUMaxToCores(c.CPUMaxDeprecated)
		c.CPUMaxDeprecated = 0.0
	}

	if c.CPUCores > float64(runtime.NumCPU()) { // should not > physical CPU num.
		c.CPUCores = float64(runtime.NumCPU())
		c.cpuMax = 100.0
	} else {
		c.cpuMax = 100.0 * c.CPUCores / float64(runtime.NumCPU())
	}
}

//nolint:gochecknoinits,lll
func init() {
	var err error
	self, err = process.NewProcess(int32(os.Getpid()))
	if err != nil {
		l.Warnf("new process failed: %s, this probably happened in the pod environment when the hostIPC and hostPID fields of the pod spec are set to false", err.Error())
	}
}

func Run(c *ResourceLimitOptions, username string) {
	l = logger.SLogger("resourcelimit")

	c.Setup()

	resourceLimitOpt = c
	userName = username
	if c == nil || !c.Enable {
		return
	}

	l.Infof("set CPU max to %f%%(%d cores)", c.CPUCores, runtime.NumCPU())

	g := datakit.G("internal_resourcelimit")

	g.Go(func(ctx context.Context) error {
		if err := run(c); err != nil {
			l.Errorf("run resource limit error: %s", err.Error())
		} else {
			l.Infof("resource limit: %s", Info())
		}

		return nil
	})
}

func MyMemPercent() (float32, error) {
	if self == nil {
		return 0, errProcessInitFailed
	}

	return self.MemoryPercent()
}

func MyMemStat() (*process.MemoryInfoStat, error) {
	if self == nil {
		return nil, errProcessInitFailed
	}

	return self.MemoryInfo()
}

func MyCPUPercent(du time.Duration) (float64, error) {
	if self == nil {
		return 0, errProcessInitFailed
	}

	return self.Percent(du)
}

func MyCtxSwitch() *process.NumCtxSwitchesStat {
	if self == nil {
		return nil
	}

	if x, err := self.NumCtxSwitches(); err == nil {
		return x
	} else {
		return nil
	}
}

func MyIOCountersStat() *process.IOCountersStat {
	if self == nil {
		return nil
	}

	if x, err := self.IOCounters(); err == nil {
		return x
	} else {
		return nil
	}
}

func Info() string {
	if resourceLimitOpt == nil || !resourceLimitOpt.Enable {
		return "-"
	}

	if userName != "root" {
		return fmt.Sprintf("path: %s, mem: %dMB, cpus: %.2f",
			resourceLimitOpt.Path, resourceLimitOpt.MemMax, resourceLimitOpt.CPUCores)
	}
	return info()
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package resourcelimit limit datakit cpu or memory usage
package resourcelimit

import (
	"context"
	"os"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/shirou/gopsutil/v3/process"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var l = logger.DefaultSLogger("resourcelimit")

var (
	self             *process.Process
	resourceLimitOpt *ResourceLimitOptions
)

const (
	MB = 1024 * 1024
)

type ResourceLimitOptions struct {
	Path   string  `toml:"path"`
	CPUMax float64 `toml:"cpu_max"`
	MemMax int64   `toml:"mem_max_mb"`

	DisableOOM bool `toml:"disable_oom,omitempty"`
	Enable     bool `toml:"enable"`
}

//nolint:gochecknoinits
func init() {
	var err error
	self, err = process.NewProcess(int32(os.Getpid()))
	if err != nil {
		panic(err.Error())
	}
}

func Run(c *ResourceLimitOptions) {
	l = logger.SLogger("resourcelimit")

	resourceLimitOpt = c

	if c == nil || !c.Enable {
		return
	}

	if !(0 < c.CPUMax && c.CPUMax < 100) {
		l.Errorf("CPUMax and CPUMin should be in range of (0.0, 100.0)")
		return
	}

	g := datakit.G("internal_resourcelimit")

	g.Go(func(ctx context.Context) error {
		if err := run(c); err != nil {
			l.Warnf("run resource limit error: %s", err.Error)
		}
		return nil
	})
}

func MyMemPercent() (float32, error) {
	return self.MemoryPercent()
}

func MyCPUPercent(du time.Duration) (float64, error) {
	return self.Percent(du)
}

func MyCtxSwitch() *process.NumCtxSwitchesStat {
	if x, err := self.NumCtxSwitches(); err == nil {
		return x
	} else {
		return nil
	}
}

func MyIOCountersStat() *process.IOCountersStat {
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

	return info()
}

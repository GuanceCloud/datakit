// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package resourcelimit limit datakit cpu or memory usage
package resourcelimit

import (
	"context"
	"errors"
	"os"
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

//nolint:gochecknoinits,lll
func init() {
	var err error
	self, err = process.NewProcess(int32(os.Getpid()))
	if err != nil {
		l.Warnf("new process failed: %s, this probably happened in the pod environment when the hostIPC and hostPID fields of the pod spec are set to false", err.Error())
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

	return info()
}

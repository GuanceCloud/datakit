// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package core

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/resourcelimit"
)

// getCurrentCPULimits determines the current CPU limits.
func (c *Core) getCurrentCPULimits() float64 {
	if datakit.Docker {
		if limit, err := c.getCurrentCPUMaxFromCgroupv2(); err != nil {
			log.Warnf("failed to get cpu limit from cgroupv2, %s", err)
		} else {
			log.Infof("used cpu limit %v from cgroupv2", limit)
			return limit
		}

		if limit, err := c.getCurrentCPUMaxFromCgroupv1(); err != nil {
			log.Warnf("failed to get cpu limit from cgroupv1, %s", err)
		} else {
			log.Infof("used cpu limit %v from cgroupv1", limit)
			return limit
		}

		log.Warn("used default limit 1.0")
		return 1.0
	} else {
		if c.cfg.ResourceLimitOptions.Enable {
			if c.cfg.ResourceLimitOptions.CPUCores == 0.0 {
				return resourcelimit.CPUMaxToCores(c.cfg.ResourceLimitOptions.CPUMax())
			}

			return c.cfg.ResourceLimitOptions.CPUCores
		} else {
			return float64(runtime.NumCPU()) // if no limit, set it to full-CPU cores
		}
	}
}

// Reference: https://docs.kernel.org/admin-guide/cgroup-v2.html#cpu-interface-files.
func (c *Core) getCurrentCPUMaxFromCgroupv2() (float64, error) {
	const cpuMax = "/sys/fs/cgroup/cpu.max"

	data, err := os.ReadFile(cpuMax)
	if err != nil {
		return 0, err
	}

	content := strings.TrimSuffix(string(data), "\n")
	array := strings.Split(content, " ")
	if len(array) != 2 {
		return 0, fmt.Errorf("invalid cgroupv2 file")
	}

	return c.parseCurrentCPUMax(array[0], array[1])
}

// Reference: https://docs.kernel.org/scheduler/sched-bwc.html#management.
func (c *Core) getCurrentCPUMaxFromCgroupv1() (float64, error) {
	const (
		cpuQuota  = "/sys/fs/cgroup/cpu/cpu.cfs_quota_us"
		cpuPeriod = "/sys/fs/cgroup/cpu/cpu.cfs_period_us"
	)

	quota, err := os.ReadFile(cpuQuota)
	if err != nil {
		return 0, err
	}
	period, err := os.ReadFile(cpuPeriod)
	if err != nil {
		return 0, err
	}

	quotaStr := strings.TrimSuffix(string(quota), "\n")
	periodStr := strings.TrimSuffix(string(period), "\n")

	return c.parseCurrentCPUMax(quotaStr, periodStr)
}

// getNumCPU returns the number of CPU cores.
var getNumCPU = func() int { return runtime.NumCPU() } //nolint:gocritic

func (c *Core) parseCurrentCPUMax(quotaStr, periodStr string) (float64, error) {
	var err error
	var quota, period int

	if quotaStr == "max" || quotaStr == "-1" {
		quota = getNumCPU() * 100000 /*time quota in microseconds*/
	} else {
		quota, err = strconv.Atoi(quotaStr)
		if err != nil {
			return 0, fmt.Errorf("not parse quota, %w", err)
		}
	}
	if quota <= 0 {
		return 0, fmt.Errorf("unexpected quota %s", quotaStr)
	}

	period, err = strconv.Atoi(periodStr)
	if err != nil {
		return 0, fmt.Errorf("not parse period, %w", err)
	}
	if period <= 0 {
		return 0, fmt.Errorf("unexpected period %s", quotaStr)
	}

	maxCPU := math.Ceil(float64(quota) / float64(period))
	// Not more than NumCPU
	if numCPU := getNumCPU(); numCPU < int(maxCPU) {
		return float64(numCPU), nil
	}
	return maxCPU, nil
}

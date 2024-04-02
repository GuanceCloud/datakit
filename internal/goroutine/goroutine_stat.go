// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package goroutine

import (
	"fmt"
	"time"
)

// StatInfo represents each group statistic info.
type StatInfo struct {
	Total       int64         `json:"finished_goroutines"`
	CostTime    time.Duration `json:"total_cost_time"`
	MinCostTime time.Duration `json:"min_cost_time"`
	MaxCostTime time.Duration `json:"max_cost_time"`
	ErrCount    int64         `json:"err_count"`
	totalJobs   int64         // total jobs containing non-finished jobs
}

// RunningStatInfo represents each running group information.
type RunningStatInfo struct {
	Total               int64         `json:"finished_goroutines"`
	RunningTotal        int64         `json:"running_goroutines"`
	CostTime            string        `json:"total_cost_time"`
	MinCostTime         string        `json:"min_cost_time"`
	MaxCostTime         string        `json:"max_cost_time"`
	ErrCount            int64         `json:"err_count"`
	CostTimeDuration    time.Duration `json:"cost_time_ns"`
	MinCostTimeDuration time.Duration `json:"min_cost_time_ns"`
	MaxCostTimeDuration time.Duration `json:"max_cost_time_ns"`
}

// Summary represents the total statistic information.
type Summary struct {
	Total        int64  `json:"finished_goroutines"`
	RunningTotal int64  `json:"running_goroutines"`
	CostTime     string `json:"total_cost_time"`
	AvgCostTime  string `json:"avg_cost_time"`

	Items map[string]RunningStatInfo
}

// GetStat return the group summary.
func GetStat() *Summary {
	summary := &Summary{
		Items: make(map[string]RunningStatInfo),
	}

	mu.Lock()
	defer mu.Unlock()

	costTime := time.Duration(0)

	for k, v := range stat {
		runningTotal := v.totalJobs - v.Total
		summary.Items[k] = RunningStatInfo{
			Total:               v.Total,
			RunningTotal:        runningTotal,
			CostTime:            fmt.Sprint(v.CostTime),
			MinCostTime:         fmt.Sprint(v.MinCostTime),
			MaxCostTime:         fmt.Sprint(v.MaxCostTime),
			ErrCount:            v.ErrCount,
			CostTimeDuration:    v.CostTime,
			MinCostTimeDuration: v.MinCostTime,
			MaxCostTimeDuration: v.MaxCostTime,
		}
		summary.Total += v.Total
		summary.RunningTotal += runningTotal
		costTime += v.CostTime
	}

	if summary.Total != 0 {
		summary.AvgCostTime = fmt.Sprint(costTime / time.Duration(summary.Total))
	} else {
		summary.AvgCostTime = "0s"
	}

	summary.CostTime = fmt.Sprint(costTime)

	return summary
}

// GetInputName return the group name for each inputs.
func GetInputName(name string) string {
	return "inputs_" + name
}

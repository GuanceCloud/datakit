package goroutine

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// StatInfo represents each group statistic info
type StatInfo struct {
	Total       int64         `json:"finished_goroutines"`
	CostTime    time.Duration `json:"total_cost_time"`
	MinCostTime time.Duration `json:"min_cost_time"`
	MaxCostTime time.Duration `json:"max_cost_time"`
	ErrCount    int64         `json:"err_count"`
	totalJobs   int64         // total jobs containing non-finished jobs
}

// stat cache the group statistic info
var (
	stat = make(map[string]*StatInfo)
	mu   sync.Mutex
)

// Option provides the setup of a group
type Option struct {
	Name         string
	PanicCb      func([]byte) bool
	PanicTimes   int8
	PanicTimeout time.Duration
}

func beforeCb(name string) func() {
	return func() {
		mu.Lock()
		groupStat, ok := stat[name]
		if !ok {
			stat[name] = &StatInfo{}
			groupStat = stat[name]
		}
		atomic.AddInt64(&groupStat.totalJobs, 1)
		mu.Unlock()
	}
}

func postCb(name string) func(error, time.Duration) {
	return func(err error, costTime time.Duration) {
		mu.Lock()
		groupStat, ok := stat[name]
		if !ok {
			stat[name] = &StatInfo{}
			groupStat = stat[name]
		}
		if costTime < groupStat.MinCostTime || groupStat.MinCostTime == 0 {
			groupStat.MinCostTime = costTime
		}

		if costTime > groupStat.MaxCostTime {
			groupStat.MaxCostTime = costTime
		}

		atomic.AddInt64(&groupStat.Total, 1)
		if err != nil {
			atomic.AddInt64(&groupStat.ErrCount, 1)
		}
		groupStat.CostTime += costTime
		mu.Unlock()
	}
}

// NewGroup create a custom group
func NewGroup(option Option) *Group {
	name := "default"
	if len(option.Name) > 0 {
		name = option.Name
	}
	g := &Group{
		name:         name,
		panicCb:      option.PanicCb,
		panicTimes:   option.PanicTimes,
		panicTimeout: option.PanicTimeout,
	}

	g.postCb = postCb(name)
	g.beforeCb = beforeCb(name)

	return g
}

// RunningStatInfo represents each running group information
type RunningStatInfo struct {
	Total        int64  `json:"finished_goroutines"`
	RunningTotal int64  `json:"running_goroutines"`
	CostTime     string `json:"total_cost_time"`
	MinCostTime  string `json:"min_cost_time"`
	MaxCostTime  string `json:"max_cost_time"`
	ErrCount     int64  `json:"err_count"`
}

// Summary represents the total statistic information
type Summary struct {
	Total        int64  `json:"finished_goroutines"`
	RunningTotal int64  `json:"running_goroutines"`
	CostTime     string `json:"total_cost_time"`
	AvgCostTime  string `json:"avg_cost_time"`

	Items map[string]RunningStatInfo
}

// GetStat return the group summary
func GetStat() *Summary {
	summary := &Summary{
		Items: make(map[string]RunningStatInfo),
	}

	costTime := time.Duration(0)

	for k, v := range stat {
		runningTotal := v.totalJobs - v.Total
		summary.Items[k] = RunningStatInfo{
			Total:        v.Total,
			RunningTotal: runningTotal,
			CostTime:     fmt.Sprint(v.CostTime),
			MinCostTime:  fmt.Sprint(v.MinCostTime),
			MaxCostTime:  fmt.Sprint(v.MaxCostTime),
			ErrCount:     v.ErrCount,
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

// GetInputName return the group name for each inputs
func GetInputName(name string) string {
	return "inputs_" + name
}

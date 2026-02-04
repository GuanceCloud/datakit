// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"container/list"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// 资源监控对象.
type processM struct {
	configProcess   *Process
	Name            string
	Cmdline         string
	Pid             int32
	p               *process.Process
	MaxSize         int       // 环形缓冲区最大容量
	lastProfileTime time.Time // 上次采集profile的时间，不能小于1分钟。

	CPUHistory        *list.List // CPU使用率历史记录（环形缓冲区）
	MemPercentHistory *list.List // 内存使用率历史记录（环形缓冲区）
	MemHistory        *list.List // 内存使用历史记录（环形缓冲区）
	lastCPUTimes      time.Time
	lastCPUTotal      float64
}

// 创建新的监控对象.
func newProcessM(name, cmdline string, pid int32, cp *Process) *processM {
	p, err := process.NewProcess(pid)
	if err != nil {
		log.Errorf("new process error %v  pid:%d", err, pid)
		return nil
	}
	pm := &processM{
		configProcess:     cp,
		Name:              name,
		Cmdline:           cmdline,
		Pid:               pid,
		p:                 p,
		CPUHistory:        list.New(),
		MemHistory:        list.New(),
		MemPercentHistory: list.New(),
		MaxSize:           10, // 维护10个元素的环形缓冲区
	}

	return pm
}

func (pm *processM) updateProcessStats() error {
	// todo 修改cpu使用率 计算方式
	if perc := pm.cpuPercent(); perc != 0 {
		lastCollect := time.Since(pm.lastCPUTimes).Seconds()
		cpuPercent := perc / lastCollect
		log.Debugf("pid %d cpu percent is %f", pm.Pid, cpuPercent)
		pm.AddCPUUsage(cpuPercent)
	}
	pm.lastCPUTimes = time.Now()

	memInfo, err := pm.p.MemoryInfo()
	if err != nil {
		return fmt.Errorf("get memoryInfo err: %w", err)
	}
	memUsage := float64(memInfo.RSS) / 1024 / 1024
	pm.AddMemUsage(memUsage)

	memPercent, err := pm.p.MemoryPercent()
	if err != nil {
		return fmt.Errorf("get memoryPercent err: %w", err)
	}
	log.Debugf("pid %d mem percent is %f , mem used is %f (MB)", pm.Pid, memPercent, memUsage)
	pm.AddMemPercent(float64(memPercent))

	return nil
}

func (pm *processM) cpuPercent() float64 {
	times, err := pm.p.Times()
	if err != nil {
		log.Errorf("get cpu times err: %v", err)
		return 0
	}
	// 采样时已经消耗的时间
	current := times.User + times.System
	if pm.lastCPUTotal == 0 {
		pm.lastCPUTotal = current
		return 0
	}
	proc := 100 * (current - pm.lastCPUTotal)
	pm.lastCPUTotal = current
	return proc
}

// 添加CPU数据到环形缓冲区.
func (pm *processM) AddCPUUsage(usage float64) {
	pm.CPUHistory.PushBack(usage)
	if pm.CPUHistory.Len() > pm.MaxSize {
		pm.CPUHistory.Remove(pm.CPUHistory.Front())
	}
}

// 添加内存数据到环形缓冲区.
func (pm *processM) AddMemUsage(usage float64) {
	pm.MemHistory.PushBack(usage)
	if pm.MemHistory.Len() > pm.MaxSize {
		pm.MemHistory.Remove(pm.MemHistory.Front())
	}
}

func (pm *processM) AddMemPercent(perc float64) {
	pm.MemPercentHistory.PushBack(perc)
	if pm.MemPercentHistory.Len() > pm.MaxSize {
		pm.MemPercentHistory.Remove(pm.MemPercentHistory.Front())
	}
}

func (pm *processM) isTrigger() (bool, []string) {
	if time.Since(pm.lastProfileTime) < time.Minute {
		return false, nil
	}

	trigger := false
	tags := make([]string, 0)

	cpuAvg := getListAvg(pm.CPUHistory, 5)
	if pm.configProcess.CPUUsagePercent > 0 && cpuAvg >= float64(pm.configProcess.CPUUsagePercent) {
		trigger = true
		tags = append(tags, fmt.Sprintf("cpu_avg:%0.2f", cpuAvg))
	}
	memAvg := getListAvg(pm.MemHistory, 5)
	if pm.configProcess.MEMUsageMB > 0 && memAvg >= float64(pm.configProcess.MEMUsageMB) {
		trigger = true
		tags = append(tags, fmt.Sprintf("mem_used:%0.2f", memAvg))
	}
	memPercAvg := getListAvg(pm.MemPercentHistory, 5)
	if pm.configProcess.MEMUsagePercent > 0 && memPercAvg >= float64(pm.configProcess.MEMUsagePercent) {
		trigger = true
		tags = append(tags, fmt.Sprintf("mem_perc_avg:%0.2f", memPercAvg))
	}
	// log.Debugf("cpu len=%d , mem parcent len %d  mem usage len %d", pm.CPUHistory.Len(), pm.MemPercentHistory.Len(), pm.MemHistory.Len())
	if trigger {
		log.Infof("start trigger,because of %+v", tags)
	}
	return trigger, tags
}

func getListAvg(history *list.List, recentCount int) float64 {
	sum := float64(0)
	count := 0
	e := history.Back()
	for i := 0; i < recentCount && e != nil; i++ {
		if usage, ok := e.Value.(float64); ok {
			sum += usage
			count++
		}
		e = e.Prev()
	}

	// 检查是否所有最近的值都超过平均值
	if count == 0 {
		return 0.0
	}
	avg := sum / float64(count)
	// log.Debugf("avg is %f ,sum=%f, count =%d", avg, sum, count)
	return avg
}

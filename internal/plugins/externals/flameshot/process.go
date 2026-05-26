// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"k8s.io/apimachinery/pkg/api/resource"
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
	lastJcmdTime    time.Time // 上次采集 jcmd 轻量快照的时间，避免高水位时频繁执行。
	mu              sync.Mutex

	CPUHistory        *list.List // CPU使用率历史记录（环形缓冲区）
	MemPercentHistory *list.List // 内存使用率历史记录（环形缓冲区）
	MemHistory        *list.List // 内存使用历史记录（环形缓冲区）
	lastCPUTimes      time.Time
	lastCPUTotal      float64

	podCPULimit string
	podMEMLimit string
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
	// 1. 获取两次采集之间的时间间隔（秒）
	now := time.Now()
	interval := now.Sub(pm.lastCPUTimes).Seconds()
	pm.lastCPUTimes = now
	cpuTimeDelta := pm.getCPUTimeDelta()

	if interval > 0 {
		// 计算物理层面的真实核心占用
		// 例如：1秒内消耗了0.5秒CPU时间，说明占用了0.5个核心
		realCoresUsed := cpuTimeDelta / interval

		// 3. 结合 K8s Limit 进行归一化
		limitCores := parseCPULimit(pm.podCPULimit)
		var containerCPUPercent float64
		if limitCores > 0 {
			// 核心公式：实际占用 / 额度
			containerCPUPercent = (realCoresUsed / limitCores) * 100
		} else {
			// 如果没设 limit，可以按物理机单核 100% 展示
			containerCPUPercent = realCoresUsed * 100
		}

		log.Debugf("Real Cores: %.3f, Limit: %.3f, Container CPU: %.2f%%",
			realCoresUsed, limitCores, containerCPUPercent)

		pm.AddCPUUsage(containerCPUPercent)
	}

	// 2. 获取内存使用情况
	memInfo, err := pm.p.MemoryInfo()
	if err != nil {
		return fmt.Errorf("get memoryInfo err: %w", err)
	}

	memUsage := float64(memInfo.RSS) / 1024 / 1024
	pm.AddMemUsage(memUsage)

	// 3. 优先使用 Pod limit 视角的内存百分比，只有没有 limit 时才回退到宿主机视角
	containerMemPercent, memLimitBytes, memPercentSource := pm.getMemoryUsagePercent(memInfo.RSS)
	if memPercentSource == "host" && pm.configProcess != nil &&
		(pm.configProcess.MEMUsagePercent > 0 || pm.configProcess.MEMUsagePercentEmergency > 0) {
		log.Debugf("pid %d: pod mem limit not configured, fallback to host memory percent", pm.Pid)
	}

	log.Debugf("pid %d: RSS=%d Bytes, Limit=%d Bytes, Usage=%.2f%%, Source=%s",
		pm.Pid, memInfo.RSS, memLimitBytes, containerMemPercent, memPercentSource)
	pm.AddMemPercent(containerMemPercent)

	return nil
}

func (pm *processM) getMemoryUsagePercent(rss uint64) (float64, int64, string) {
	memLimitBytes := parseMemoryLimit(pm.podMEMLimit)
	if memLimitBytes > 0 {
		return (float64(rss) / float64(memLimitBytes)) * 100, memLimitBytes, "pod_limit"
	}

	if pm.p == nil {
		return 0, 0, "unknown"
	}

	p, err := pm.p.MemoryPercent()
	if err != nil {
		return 0, 0, "unknown"
	}

	return float64(p), 0, "host"
}

var readProcessRSS = func(pm *processM) (uint64, error) {
	if pm == nil || pm.p == nil {
		return 0, fmt.Errorf("process not available")
	}

	memInfo, err := pm.p.MemoryInfo()
	if err != nil {
		return 0, err
	}

	return memInfo.RSS, nil
}

func (pm *processM) currentRSSBytes() (uint64, bool) {
	if rss, err := readProcessRSS(pm); err == nil {
		return rss, true
	}

	memMB, ok := getLatestValue(pm.MemHistory)
	if !ok {
		return 0, false
	}

	return uint64(memMB * 1024 * 1024), true
}

// 将 "500m" 转为 0.5, 将 "2" 转为 2.0.
func parseCPULimit(limitStr string) float64 {
	if limitStr == "" {
		return 0
	}
	limitStr = strings.TrimSpace(limitStr)
	q, err := resource.ParseQuantity(limitStr)
	if err != nil {
		return 0
	}
	// AsApproximateFloat64 是最方便的转换方式
	return q.AsApproximateFloat64()
}

// 修改后的 cpuPercent 返回这段时间内消耗的 CPU 秒数.
func (pm *processM) getCPUTimeDelta() float64 {
	times, err := pm.p.Times()
	if err != nil {
		return 0
	}
	current := times.User + times.System
	if pm.lastCPUTotal == 0 {
		pm.lastCPUTotal = current
		return 0
	}
	delta := current - pm.lastCPUTotal
	pm.lastCPUTotal = current
	return delta // 单位是秒
}

// parseMemoryLimit 将 K8s 内存字符串（如 100Mi, 1G）转换为字节数 (int64).
func parseMemoryLimit(limitStr string) int64 {
	if limitStr == "" {
		return 0
	}
	q, err := resource.ParseQuantity(limitStr)
	if err != nil {
		return 0
	}
	// Value() 返回字节数，例如 1Mi 返回 1048576
	return q.Value()
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
	trigger, tags, _ := pm.triggerDecision()
	return trigger, tags
}

func (pm *processM) triggerDecision() (bool, []string, bool) {
	if pm.inCooldown(time.Minute) {
		return false, nil, false
	}

	trigger := false
	emergency := false
	tags := make([]string, 0)

	memCurrent, hasMemCurrent := getLatestValue(pm.MemHistory)
	if hasMemCurrent && pm.configProcess.MEMUsageMBEmergency > 0 && memCurrent >= float64(pm.configProcess.MEMUsageMBEmergency) {
		trigger = true
		emergency = true
		tags = append(tags, fmt.Sprintf("mem_used_emergency:%0.2f", memCurrent))
	}

	memPercCurrent, hasMemPercCurrent := getLatestValue(pm.MemPercentHistory)
	if hasMemPercCurrent && pm.configProcess.MEMUsagePercentEmergency > 0 && memPercCurrent >= float64(pm.configProcess.MEMUsagePercentEmergency) {
		trigger = true
		emergency = true
		tags = append(tags, fmt.Sprintf("mem_perc_emergency:%0.2f", memPercCurrent))
	}

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
	return trigger, tags, emergency
}

func (pm *processM) inCooldown(window time.Duration) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	return time.Since(pm.lastProfileTime) < window
}

func (pm *processM) markProfileTriggered(now time.Time) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.lastProfileTime = now
}

func (pm *processM) inJcmdCooldown(window time.Duration) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	return time.Since(pm.lastJcmdTime) < window
}

func (pm *processM) markJcmdTriggered(now time.Time) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.lastJcmdTime = now
}

func getLatestValue(history *list.List) (float64, bool) {
	if history == nil {
		return 0, false
	}

	last := history.Back()
	if last == nil {
		return 0, false
	}

	usage, ok := last.Value.(float64)
	if !ok {
		return 0, false
	}

	return usage, true
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

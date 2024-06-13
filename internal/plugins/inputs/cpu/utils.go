// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cpu

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

type CPUStatInfo interface {
	CPUTimes(perCPU, totalCPU bool) ([]cpu.TimesStat, error)
}

type CPUInfo struct{}

type CPUInfoTest struct {
	timeStat [][]cpu.TimesStat
	index    int
}

// CPUTimes collect cpu time.
func (c *CPUInfo) CPUTimes(perCPU, totalCPU bool) ([]cpu.TimesStat, error) {
	var cpuTimes []cpu.TimesStat
	PerTotal := [2]bool{perCPU, totalCPU}
	for i := 0; i < 2; i++ {
		if PerTotal[i] {
			if cpuTime, err := cpu.Times(i == 0); err == nil {
				cpuTimes = append(cpuTimes, cpuTime...)
			} else {
				return nil, err
			}
		}
	}
	return cpuTimes, nil
}

func (c *CPUInfoTest) CPUTimes(perCPU, totalCPU bool) ([]cpu.TimesStat, error) {
	if c.index < len(c.timeStat) {
		r := c.timeStat[c.index]
		c.index += 1
		return r, nil
	}
	return nil, fmt.Errorf("")
}

var coreTempRegex = regexp.MustCompile(`coretemp_core[\d]+_input`)

func CoreTemp() (map[string]float64, error) {
	if runtime.GOOS == datakit.OSLinux {
		tempMap := map[string]float64{}
		sensorTempStat, err := host.SensorsTemperatures()
		if err != nil {
			return nil, err
		}

		temp, count := 0.0, 0
		for _, v := range sensorTempStat {
			if coreTempRegex.MatchString(v.SensorKey) {
				temp += v.Temperature
				count++
				cpuID := strings.Replace(strings.Split(v.SensorKey, "_")[1], "core", "cpu", 1)
				tempMap[cpuID] = v.Temperature
			}
		}
		if count > 0 {
			tempMap["cpu-total"] = temp / float64(count)
			return tempMap, nil
		} else {
			return nil, fmt.Errorf("no coretemp data")
		}
	}

	return nil, fmt.Errorf("os is not supported")
}

type UsageStat struct {
	CPU       string
	User      float64
	System    float64
	Idle      float64
	Nice      float64
	Iowait    float64
	Irq       float64
	Softirq   float64
	Steal     float64
	Guest     float64
	GuestNice float64
	Total     float64
}

func usageCPU(cur, pre, total float64) float64 {
	v := 100 * (cur - pre) / total
	// 此处规避异常数值，Windows 上存在原始数据异常的情况
	if v < 0 || v > 100 {
		return 0
	}
	return v
}

// CalculateUsage calculate cpu usage.
func CalculateUsage(nowT cpu.TimesStat, lastT cpu.TimesStat, totalDelta float64) (*UsageStat, error) {
	if nowT.CPU != lastT.CPU {
		err := fmt.Errorf("warning. Not the same CPU. CPU:%s %s", nowT.CPU, lastT.CPU)
		return nil, err
	}
	c := new(UsageStat)

	c.CPU = nowT.CPU
	c.User = usageCPU(nowT.User-nowT.Guest, lastT.User-lastT.Guest, totalDelta)
	c.System = usageCPU(nowT.System, lastT.System, totalDelta)
	c.Idle = usageCPU(nowT.Idle, lastT.Idle, totalDelta)
	c.Nice = usageCPU(nowT.Nice-nowT.GuestNice, lastT.Nice-lastT.GuestNice, totalDelta)
	c.Iowait = usageCPU(nowT.Iowait, lastT.Iowait, totalDelta)
	c.Irq = usageCPU(nowT.Irq, lastT.Irq, totalDelta)
	c.Softirq = usageCPU(nowT.Softirq, lastT.Softirq, totalDelta)
	c.Steal = usageCPU(nowT.Steal, lastT.Steal, totalDelta)
	c.Guest = usageCPU(nowT.Guest, lastT.Guest, totalDelta)
	c.GuestNice = usageCPU(nowT.GuestNice, lastT.GuestNice, totalDelta)
	c.Total = c.User + c.System + c.Nice + c.Iowait + c.Irq + c.Softirq + c.Steal + c.Guest + c.GuestNice
	if c.Total < 0 || c.Total > 100 {
		c.Total = 0
	}
	return c, nil
}

// CPUActiveTotalTime return cpu active, total time.
func CPUActiveTotalTime(t cpu.TimesStat) (float64, float64) {
	total := t.Total()
	active := total - t.Idle
	return active, total
}

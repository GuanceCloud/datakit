package cpu

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
)

type CPUStatInfo interface {
	CPUTimes(perCPU, totalCPU bool) ([]cpu.TimesStat, error)
}

type CPUInfo struct{}

type CPUInfoTest struct {
	timeStat [][]cpu.TimesStat
	index    int
}

// collect cpu time
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

func CoreTemp() (map[string]float64, error) {
	if runtime.GOOS == "linux" {
		tempMap := map[string]float64{}
		sensorTempStat, err := host.SensorsTemperatures()
		if err != nil {
			return nil, err
		}
		regexpC, err := regexp.Compile(`coretemp_core[\d]+_input`)
		if err != nil {
			return nil, err
		}
		temp, count := 0.0, 0
		for _, v := range sensorTempStat {
			if regexpC.MatchString(v.SensorKey) {
				temp += v.Temperature
				count++
				cpuId := strings.Replace(strings.Split(v.SensorKey, "_")[1], "core", "cpu", 1)
				tempMap[cpuId] = v.Temperature
			}
		}
		if count > 0 {
			tempMap["cpu-total"] = temp / float64(count)
			return tempMap, nil
		} else {
			return nil, fmt.Errorf("no coretemp data or regexp error")
		}
	}
	return nil, fmt.Errorf("os is not supported")
}

// cpu usage stat
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

// calculate cpu usage
func CalculateUsage(nowT cpu.TimesStat, lastT cpu.TimesStat, totalDelta float64) (*UsageStat, error) {
	if nowT.CPU != lastT.CPU {
		err := fmt.Errorf("warning. Not the same CPU. CPU:%s %s", nowT.CPU, lastT.CPU)
		return nil, err
	}
	c := new(UsageStat)

	c.CPU = nowT.CPU
	c.User = 100 * (nowT.User - lastT.User - (nowT.Guest - lastT.Guest)) / totalDelta
	c.System = 100 * (nowT.System - lastT.System) / totalDelta
	c.Idle = 100 * (nowT.Idle - lastT.Idle) / totalDelta
	c.Nice = 100 * (nowT.Nice - lastT.Nice - (nowT.GuestNice - lastT.GuestNice)) / totalDelta
	c.Iowait = 100 * (nowT.Iowait - lastT.Iowait) / totalDelta
	c.Irq = 100 * (nowT.Irq - lastT.Irq) / totalDelta
	c.Softirq = 100 * (nowT.Softirq - lastT.Softirq) / totalDelta
	c.Steal = 100 * (nowT.Steal - lastT.Steal) / totalDelta
	c.Guest = 100 * (nowT.Guest - lastT.Guest) / totalDelta
	c.GuestNice = 100 * (nowT.GuestNice - lastT.GuestNice) / totalDelta
	c.Total = 100 * (nowT.Total() - nowT.Idle - lastT.Total() + lastT.Idle) / totalDelta
	return c, nil
}

// return cpu active, total time
func CpuActiveTotalTime(t cpu.TimesStat) (float64, float64) {
	total := t.Total()
	active := total - t.Idle
	return active, total
}

package cpu

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/shirou/gopsutil/cpu"
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

// Hump to underline.
// such as:
// SaRRdDD --> sa_r_rd_dd |
// UserAgent --> user_agent |
// userAgent --> user_agent |
func HumpToUnderline(s string) string {
	lenStrAscii := len(s)
	var upperIndex [][2]int
	for i := 0; i < lenStrAscii; i++ {
		w := s[i : i+1]
		if w >= "A" && w <= "Z" {
			pos := i
			count := 1
			for i++; i < lenStrAscii; i++ {
				w1 := s[i : i+1]
				if !(w1 >= "A" && w1 <= "Z") {
					break
				}
				count++
			}
			upperIndex = append(upperIndex, [2]int{pos, count})
		}
	}
	r := s
	lenU := len(upperIndex)
	for i := 0; i < lenU; i++ {
		if i == 0 {
			r = ""
			r += s[0:upperIndex[i][0]]
		}
		if upperIndex[i][1] > 1 {
			r += "_" + s[upperIndex[i][0]:upperIndex[i][0]+upperIndex[i][1]-1]
		}
		if i == lenU-1 {
			if upperIndex[i][0]+upperIndex[i][1] == len(s) {
				r += s[upperIndex[i][0]+upperIndex[i][1]-1:]
			} else {
				r += "_" + s[upperIndex[i][0]+upperIndex[i][1]-1:]
			}
			break
		}
		r += "_" + s[upperIndex[i][0]+upperIndex[i][1]-1:upperIndex[i+1][0]]
	}
	if r[0:1] == "_" {
		r = r[1:]
	}
	return strings.ToLower(r)
}

// struct -> map
func CPUStatStructToMap(m map[string]interface{}, s interface{}, prefix string) bool {
	ok := true
	defer func() {
		recover()
		ok = false
	}()
	t := reflect.TypeOf(s)
	v := reflect.ValueOf(s)
	if reflect.TypeOf(s).Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	for x := 0; x < t.NumField(); x++ {
		if t.Field(x).Name != "CPU" {
			m[prefix+HumpToUnderline(t.Field(x).Name)] = v.Field(x).Interface()
		}
	}
	return ok
}

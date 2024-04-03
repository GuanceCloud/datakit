//go:build linux
// +build linux

package sysmonitor

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/v3/process"
)

type ResourceLimiter struct {
	CPUTime   float64
	MemBytes  float64
	Bandwidth float64

	proc *process.Process

	ifaceTraffic map[string][2]uint64
	lastNetTn    time.Time
}

func SelfProcess() (*process.Process, error) {
	return process.NewProcess(int32(os.Getpid()))
}

func NewResLimiter(cpuTime float64, memBytes, bandwidth string,
) (*ResourceLimiter, error) {
	var bdw, mem float64
	if bandwidth != "" {
		bdw = GetBandwidth(bandwidth)
	}

	mem = GetBytes(memBytes)

	proc, err := SelfProcess()
	if err != nil {
		return nil, fmt.Errorf("get self process: %w", err)
	} else {
		log.Infof("set bandwidth limit %s, %.3fMiB/s",
			bandwidth, bdw/(1024*1024))
		log.Infof("set memory limit %s, %.3fMiB",
			memBytes, mem/(1024*1024))
		log.Infof("set cpu limit %.3fs", cpuTime)
	}

	return &ResourceLimiter{
		CPUTime:   cpuTime,
		MemBytes:  mem,
		Bandwidth: bdw,
		proc:      proc,
	}, nil
}

func (res *ResourceLimiter) MonitorResource() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second * 3)
		defer ticker.Stop()
		defer close(ch)
		for {
			if res.overResLimit() {
				return
			}
			<-ticker.C
		}
	}()

	return ch
}

func (res *ResourceLimiter) overResLimit() bool {
	if res.MemBytes > 0 {
		m, err := res.proc.MemoryInfo()
		if err != nil {
			log.Errorf("get memory info error: %w", err)
			return true
		}

		if res.MemBytes < float64(m.RSS) {
			return true
		}
	}

	if res.CPUTime > 0 {
		c, err := res.proc.CPUPercent()
		if err != nil {
			log.Errorf("get cpu percent error: %w", err)
			return true
		}
		c /= 100
		if res.CPUTime < c {
			return true
		}
	}
	if res.Bandwidth > 0 {
		tn := time.Now()
		defer func() {
			res.lastNetTn = tn
		}()
		ifaces, err := net.IOCounters(true)
		if err != nil {
			log.Errorf("get net io count error: %w", err)
		}
		newTraffic := map[string][2]uint64{}
		for _, neti := range ifaces {
			newTraffic[neti.Name] = [2]uint64{neti.BytesRecv, neti.BytesSent}
		}
		last := res.ifaceTraffic
		res.ifaceTraffic = newTraffic

		for k, metricCur := range newTraffic {
			if metric, ok := last[k]; ok {
				diffRcv := metricCur[0] - metric[0]
				diffSnd := metricCur[1] - metric[1]
				if diffDur := float64(tn.Sub(res.lastNetTn) / time.Second); diffDur > 0 {
					rateRcv := float64(diffRcv) / diffDur
					rateSnd := float64(diffSnd) / diffDur
					if res.Bandwidth < rateRcv {
						return true
					}
					if res.Bandwidth < rateSnd {
						return true
					}
				}
			}
		}
	}

	return false
}

func GetBytes(v string) float64 {
	v = strings.ToUpper(v)
	v = strings.ReplaceAll(v, "I", "i")

	switch {
	case strings.HasSuffix(v, "K"):
		val, _ := strconv.ParseFloat(v[:len(v)-1], 64)
		return val * 1000
	case strings.HasSuffix(v, "KB"):
		val, _ := strconv.ParseFloat(v[:len(v)-2], 64)
		return val * 1000
	case strings.HasSuffix(v, "KiB"):
		val, _ := strconv.ParseFloat(v[:len(v)-3], 64)
		return val * 1024
	case strings.HasSuffix(v, "M"):
		val, _ := strconv.ParseFloat(v[:len(v)-1], 64)
		return val * 1000 * 1000
	case strings.HasSuffix(v, "MB"):
		val, _ := strconv.ParseFloat(v[:len(v)-2], 64)
		return val * 1000 * 1000
	case strings.HasSuffix(v, "MiB"):
		val, _ := strconv.ParseFloat(v[:len(v)-3], 64)
		return val * 1024 * 1024
	case strings.HasSuffix(v, "G"):
		val, _ := strconv.ParseFloat(v[:len(v)-1], 64)
		return val * 1000 * 1000 * 1000
	case strings.HasSuffix(v, "GB"):
		val, _ := strconv.ParseFloat(v[:len(v)-2], 64)
		return val * 1000 * 1000 * 1000
	case strings.HasSuffix(v, "GiB"):
		val, _ := strconv.ParseFloat(v[:len(v)-3], 64)
		return val * 1024 * 1024 * 1024
	default:
		return 0
	}
}

func GetBandwidth(v string) float64 {
	switch {
	case strings.HasSuffix(v, "/s"):
		v = v[:strings.LastIndex(v, "/s")]
	case strings.HasSuffix(v, "/S"):
		v = v[:strings.LastIndex(v, "/S")]
	}
	return GetBytes(v)
}

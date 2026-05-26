//go:build linux
// +build linux

package procwatch

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/v3/process"
)

type ResourceQuota struct {
	CPUTime   float64
	MemBytes  float64
	Bandwidth float64

	proc *process.Process

	ifaceTraffic map[string][2]uint64
	lastNetAt    time.Time
}

func NewResourceQuota(cpuTime float64, memLimit string, bandwidthLimit string) (*ResourceQuota, error) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, fmt.Errorf("get self process: %w", err)
	}

	memBytes := parseBytes(memLimit)
	bandwidth := parseBandwidth(bandwidthLimit)
	if bandwidth > 0 && !bandwidthLimitSupported() {
		log.Warnf("disable bandwidth limit %s: only isolated network namespaces are supported", bandwidthLimit)
		bandwidth = 0
	}

	if bandwidth > 0 {
		log.Infof("set bandwidth limit %s, %.3fMiB/s", bandwidthLimit, bandwidth/(1024*1024))
	}
	log.Infof("set memory limit %s, %.3fMiB", memLimit, memBytes/(1024*1024))
	log.Infof("set cpu limit %.3f cores", cpuTime)

	return &ResourceQuota{
		CPUTime:   cpuTime,
		MemBytes:  memBytes,
		Bandwidth: bandwidth,
		proc:      proc,
	}, nil
}

func (q *ResourceQuota) Monitor() <-chan string {
	done := make(chan string, 1)
	go func() {
		defer close(done)
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			if exceeded, reason := q.overLimit(); exceeded {
				done <- reason
				return
			}
			<-ticker.C
		}
	}()
	return done
}

func (q *ResourceQuota) overLimit() (bool, string) {
	if q.MemBytes > 0 {
		info, err := q.proc.MemoryInfo()
		if err != nil {
			log.Warnf("skip memory limit check: %v", err)
			return false, ""
		}
		if float64(info.RSS) > q.MemBytes {
			return true, fmt.Sprintf("memory rss %.0f > limit %.0f bytes", float64(info.RSS), q.MemBytes)
		}
	}

	if q.CPUTime > 0 {
		cpu, err := q.proc.CPUPercent()
		if err != nil {
			log.Warnf("skip cpu limit check: %v", err)
			return false, ""
		}
		if cpu/100 > q.CPUTime {
			return true, fmt.Sprintf("cpu %.3f cores > limit %.3f cores", cpu/100, q.CPUTime)
		}
	}

	if q.Bandwidth > 0 {
		now := time.Now()
		defer func() { q.lastNetAt = now }()

		counters, err := net.IOCountersByFile(true, "/proc/self/net/dev")
		if err != nil {
			log.Warnf("skip bandwidth limit check: %v", err)
			return false, ""
		}

		current := make(map[string][2]uint64, len(counters))
		for _, counter := range counters {
			current[counter.Name] = [2]uint64{counter.BytesRecv, counter.BytesSent}
		}

		last := q.ifaceTraffic
		q.ifaceTraffic = current

		for name, metric := range current {
			prev, ok := last[name]
			if !ok {
				continue
			}
			seconds := now.Sub(q.lastNetAt).Seconds()
			if seconds <= 0 {
				continue
			}
			recvRate := float64(metric[0]-prev[0]) / seconds
			if recvRate > q.Bandwidth {
				return true, fmt.Sprintf("bandwidth recv %.0f > limit %.0f B/s on %s", recvRate, q.Bandwidth, name)
			}
			sendRate := float64(metric[1]-prev[1]) / seconds
			if sendRate > q.Bandwidth {
				return true, fmt.Sprintf("bandwidth send %.0f > limit %.0f B/s on %s", sendRate, q.Bandwidth, name)
			}
		}
	}

	return false, ""
}

func bandwidthLimitSupported() bool {
	selfNS, err := os.Readlink("/proc/self/ns/net")
	if err != nil || selfNS == "" {
		return false
	}

	hostNS, err := os.Readlink(HostProc("1", "ns", "net"))
	if err != nil || hostNS == "" {
		return false
	}

	return bandwidthLimitSupportedWithLinks(selfNS, hostNS)
}

func bandwidthLimitSupportedWithLinks(selfNS, hostNS string) bool {
	return selfNS != "" && hostNS != "" && selfNS != hostNS
}

func parseBytes(value string) float64 {
	value = strings.ToUpper(value)
	value = strings.ReplaceAll(value, "I", "i")

	switch {
	case strings.HasSuffix(value, "K"):
		number, _ := strconv.ParseFloat(value[:len(value)-1], 64)
		return number * 1000
	case strings.HasSuffix(value, "KB"):
		number, _ := strconv.ParseFloat(value[:len(value)-2], 64)
		return number * 1000
	case strings.HasSuffix(value, "KiB"):
		number, _ := strconv.ParseFloat(value[:len(value)-3], 64)
		return number * 1024
	case strings.HasSuffix(value, "M"):
		number, _ := strconv.ParseFloat(value[:len(value)-1], 64)
		return number * 1000 * 1000
	case strings.HasSuffix(value, "MB"):
		number, _ := strconv.ParseFloat(value[:len(value)-2], 64)
		return number * 1000 * 1000
	case strings.HasSuffix(value, "MiB"):
		number, _ := strconv.ParseFloat(value[:len(value)-3], 64)
		return number * 1024 * 1024
	case strings.HasSuffix(value, "G"):
		number, _ := strconv.ParseFloat(value[:len(value)-1], 64)
		return number * 1000 * 1000 * 1000
	case strings.HasSuffix(value, "GB"):
		number, _ := strconv.ParseFloat(value[:len(value)-2], 64)
		return number * 1000 * 1000 * 1000
	case strings.HasSuffix(value, "GiB"):
		number, _ := strconv.ParseFloat(value[:len(value)-3], 64)
		return number * 1024 * 1024 * 1024
	default:
		return 0
	}
}

func parseBandwidth(value string) float64 {
	switch {
	case strings.HasSuffix(value, "/s"):
		value = value[:strings.LastIndex(value, "/s")]
	case strings.HasSuffix(value, "/S"):
		value = value[:strings.LastIndex(value, "/S")]
	}
	return parseBytes(value)
}

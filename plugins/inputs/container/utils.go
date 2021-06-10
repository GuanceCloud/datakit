package container

import (
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
)

func RawConnect(host string, port string) error {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return err
	}
	if conn != nil {
		defer conn.Close()
		return nil
	}
	return nil
}

func TrimPodName(podname []string, name string) string {
	for _, n := range podname {
		if !strings.HasPrefix(name, n) {
			continue
		}

		parts := strings.Split(name, "-")
		if len(parts) < 3 {
			continue
		}
		return strings.Join(parts[:len(parts)-1], "-")
	}
	return name
}

func RegexpMatchString(regexps []*regexp.Regexp, s string) bool {
	for _, re := range regexps {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

func getContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return "invalidContainerName"
}

// contianerIsFromKubernetes 判断该容器是否由kubernetes创建
// 所有kubernetes启动的容器的containerNamePrefix都是k8s，依据链接如下
// https://github.com/rootsongjc/kubernetes-handbook/blob/master/practice/monitor.md#%E5%AE%B9%E5%99%A8%E7%9A%84%E5%91%BD%E5%90%8D%E8%A7%84%E5%88%99
func contianerIsFromKubernetes(containerName string) bool {
	const kubernetesContainerNamePrefix = "k8s"
	return strings.HasPrefix(containerName, kubernetesContainerNamePrefix)
}

// https://docs.docker.com/engine/api/v1.41/#operation/ContainerStats

func calculateCPUDelta(v *types.StatsJSON) int64 {
	return int64(v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage)
}

func calculateCPUSystemDelta(v *types.StatsJSON) int64 {
	return int64(v.CPUStats.SystemUsage - v.PreCPUStats.SystemUsage)
}

func calculateCPUNumbers(v *types.StatsJSON) int64 {
	return int64(v.CPUStats.OnlineCPUs)
}

func calculateCPUPercentUnix(previousCPU, previousSystem uint64, v *types.StatsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(previousSystem)
		onlineCPUs  = float64(v.CPUStats.OnlineCPUs)
	)

	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(v.CPUStats.CPUUsage.PercpuUsage))
	}
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}
	return cpuPercent
}

func calculateBlockIO(blkio types.BlkioStats) (int64, int64) {
	var blkRead, blkWrite int64
	for _, bioEntry := range blkio.IoServiceBytesRecursive {
		if len(bioEntry.Op) == 0 {
			continue
		}
		switch bioEntry.Op[0] {
		case 'r', 'R':
			blkRead = blkRead + int64(bioEntry.Value)
		case 'w', 'W':
			blkWrite = blkWrite + int64(bioEntry.Value)
		}
	}
	return blkRead, blkWrite
}

func calculateNetwork(network map[string]types.NetworkStats) (int64, int64) {
	var rx, tx int64

	for _, v := range network {
		rx += int64(v.RxBytes)
		tx += int64(v.TxBytes)
	}
	return rx, tx
}

// calculateMemUsageUnixNoCache calculate memory usage of the container.
// Cache is intentionally excluded to avoid misinterpretation of the output.
//
// On cgroup v1 host, the result is `mem.Usage - mem.Stats["total_inactive_file"]` .
// On cgroup v2 host, the result is `mem.Usage - mem.Stats["inactive_file"] `.
//
// This definition is consistent with cadvisor and containerd/CRI.
// * https://github.com/google/cadvisor/commit/307d1b1cb320fef66fab02db749f07a459245451
// * https://github.com/containerd/cri/commit/6b8846cdf8b8c98c1d965313d66bc8489166059a
//
// On Docker 19.03 and older, the result was `mem.Usage - mem.Stats["cache"]`.
// See https://github.com/moby/moby/issues/40727 for the background.
func calculateMemUsageUnixNoCache(mem types.MemoryStats) int64 {
	// cgroup v1
	if v, isCgroup1 := mem.Stats["total_inactive_file"]; isCgroup1 && v < mem.Usage {
		return int64(mem.Usage - v)
	}
	// cgroup v2
	if v := mem.Stats["inactive_file"]; v < mem.Usage {
		return int64(mem.Usage - v)
	}
	return int64(mem.Usage)
}

func calculateMemPercentUnixNoCache(limit float64, usedNoCache float64) float64 {
	// MemoryStats.Limit will never be 0 unless the container is not running and we haven't
	// got any data from cgroup
	if limit != 0 {
		return usedNoCache / limit * 100.0
	}
	return 0
}

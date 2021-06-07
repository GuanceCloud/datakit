package container

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	metricCategory category = iota + 1
	objectCategory
)

type category int

func (this *Input) gather(category category) ([]*io.Point, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, apiTimeoutDuration)
	defer cancel()

	cList, err := this.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var result struct {
		pts []*io.Point
		mt  sync.Mutex
	}

	for _, container := range cList {
		wg.Add(1)

		go func(c types.Container) {
			defer wg.Done()

			var pt *io.Point
			var err error

			switch category {
			case metricCategory:
				pt, err = this.gatherMetric(c)
				if err != nil {
					l.Errorf("metric gather failed: %s", err)
					return
				}

			case objectCategory:
				pt, err = this.gatherObject(c)
				if err != nil {
					l.Errorf("object gather failed: %s", err)
					return
				}
			default:
				// unreacheable
				return
			}
			result.mt.Lock()
			result.pts = append(result.pts, pt)
			result.mt.Unlock()
		}(container)
	}

	wg.Wait()

	return result.pts, nil
}

func (this *Input) gatherMetric(container types.Container) (*io.Point, error) {
	tags := this.gatherContainerInfo(container)
	for _, key := range this.DropTags {
		if _, ok := tags[key]; ok {
			delete(tags, key)
		}
	}

	fields, err := this.gatherStats(container)
	if err != nil {
		return nil, err
	}
	return io.MakePoint(containerName, tags, fields, time.Now())
}

func (this *Input) gatherObject(container types.Container) (*io.Point, error) {
	tags := this.gatherContainerInfo(container)
	for _, key := range this.DropTags {
		if _, ok := tags[key]; ok {
			delete(tags, key)
		}
	}

	fields, err := this.gatherStats(container)
	if err != nil {
		return nil, err
	}

	// 对象数据需要有 name 和 container_host 标签
	tags["name"] = container.ID
	containerJson, err := this.client.ContainerInspect(context.Background(), container.ID)
	if err != nil {
		l.Warnf("gather container inspect error: %s", err)
		// not have tags["container_host"]
	} else {
		tags["container_host"] = containerJson.Config.Hostname
	}

	// 对象数据包含 message 字段，其值为 tags 和 fields 所有 Key/Value 的 JSON Marshal
	// 将 tags 和 fields 降到一层 K/V（可能会因为字段名相同而冲突，此处不考虑这种情况，因为字段名相同的问题在存储时也会遇到）
	// 需要额外的时空间开销，但 object 采集频率不高，可以接受
	// 另外一种方式是 struct{ Tags map[string]string, Fields map[string]interface{} } { // XX }
	message, err := json.Marshal(func() map[string]interface{} {
		var result = make(map[string]interface{}, len(tags)+len(fields))
		for k, v := range tags {
			result[k] = v
		}
		for k, v := range fields {
			result[k] = v
		}
		return result
	}())

	if err != nil {
		l.Warnf("json marshal failed: %s", err)
	} else {
		fields["message"] = string(message)
	}

	return io.MakePoint(containerName, tags, fields, time.Now())
}

func (this *Input) gatherContainerInfo(container types.Container) map[string]string {
	tags := map[string]string{
		"container_id":   container.ID,
		"container_name": getContainerName(container.Names),
		"docker_image":   container.ImageID,
		"images_name":    container.Image,
		"state":          container.State,
		"container_type": func() string {
			if contianerIsFromKubernetes(getContainerName(container.Names)) {
				return "kubernetes"
			}
			return "docker"
		}(),
	}

	for k, v := range this.Tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	if this.Kubernetes != nil {
		podInfo, err := this.Kubernetes.GatherPodInfo(container.ID)
		if err != nil {
			l.Debugf("gather k8s pod error, %s", err)
		}
		for k, v := range podInfo {
			switch k {
			case "pod_name":
				tags["pod_name"] = TrimPodName(this.PodNameRewrite, v)
			default:
				tags[k] = v
			}
		}
	}

	return tags
}

const streamStats = false

func (this *Input) gatherStats(container types.Container) (map[string]interface{}, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, apiTimeoutDuration)
	defer cancel()

	resp, err := this.client.ContainerStats(ctx, container.ID, streamStats)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.OSType == "windows" {
		return nil, nil
	}

	var v *types.StatsJSON
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	mem := calculateMemUsageUnixNoCache(v.MemoryStats)
	memPercent := calculateMemPercentUnixNoCache(float64(v.MemoryStats.Limit), float64(mem))
	netRx, netTx := calculateNetwork(v.Networks)
	blkRead, blkWrite := calculateBlockIO(v.BlkioStats)

	return map[string]interface{}{
		"cpu_usage":          calculateCPUPercentUnix(v.PreCPUStats.CPUUsage.TotalUsage, v.PreCPUStats.SystemUsage, v), /*float64*/
		"cpu_delta":          calculateCPUDelta(v),
		"cpu_system_delta":   calculateCPUSystemDelta(v),
		"cpu_numbers":        calculateCPUNumbers(v),
		"mem_limit":          int64(v.MemoryStats.Limit),
		"mem_usage":          mem,
		"mem_used_percent":   memPercent, /*float64*/
		"mem_failed_count":   int64(v.MemoryStats.Failcnt),
		"network_bytes_rcvd": netRx,
		"network_bytes_sent": netTx,
		"block_read_byte":    blkRead,
		"block_write_byte":   blkWrite,
	}, nil
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

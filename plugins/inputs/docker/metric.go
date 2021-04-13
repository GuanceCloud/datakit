package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/docker/docker/api/types"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	dockerContainerMeasurement = "docker_containers"
)

func (this *Input) gather(opt *gatherOption) ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, this.timeoutDuration)
	defer cancel()

	cList, err := this.client.ContainerList(ctx, this.opts)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	var buffer bytes.Buffer

	for _, container := range cList {
		tags := this.gatherContainerInfo(container)
		if opt.IsObjectCategory {
			tags["name"] = container.ID
		}

		fields, err := this.gatherStats(container)
		if err != nil {
			l.Error(err)
			continue
		}

		data, err := io.MakeMetric(dockerContainerMeasurement, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
			continue
		}

		buffer.Write(data)
		buffer.WriteString("\n")
	}

	return buffer.Bytes(), nil
}

func (this *Input) gatherContainerInfo(container types.Container) map[string]string {
	tags := map[string]string{
		"container_id":   container.ID,
		"container_name": getContainerName(container.Names),
		"docker_image":   container.ImageID,
		"image_name":     container.Image,
		"stats":          container.State,
	}

	// 内置tags优先
	for k, v := range this.Tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	podInfo, err := this.gatherK8sPodInfo(container.ID)
	if err != nil {
		l.Warnf("gather k8s pod error, %s", err)
	}

	for k, v := range podInfo {
		tags[k] = v
	}

	return tags
}

const streamStats = false

func (this *Input) gatherStats(container types.Container) (map[string]interface{}, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, this.timeoutDuration)
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

	cpuPercent := calculateCPUPercentUnix(v.PreCPUStats.CPUUsage.TotalUsage, v.PreCPUStats.SystemUsage, v)
	mem := calculateMemUsageUnixNoCache(v.MemoryStats)
	memPercent := calculateMemPercentUnixNoCache(float64(v.MemoryStats.Limit), mem)
	netRx, netTx := calculateNetwork(v.Networks)
	blkRead, blkWrite := calculateBlockIO(v.BlkioStats)

	return map[string]interface{}{
		"cpu_usage_percent":  cpuPercent,
		"mem_limit":          int64(v.MemoryStats.Limit),
		"mem_usage":          int64(mem),
		"mem_usage_percent":  memPercent,
		"mem_failed_count":   int64(v.MemoryStats.Failcnt),
		"network_bytes_rcvd": int64(netRx),
		"network_bytes_sent": int64(netTx),
		"block_read_byte":    int64(blkRead),
		"block_write_byte":   int64(blkWrite),
		"from_kubernetes":    contianerIsFromKubernetes(getContainerName(container.Names)),
	}, nil
}

func (this *Input) gatherK8sPodInfo(id string) (map[string]string, error) {
	if this.kubernetes == nil {
		return nil, nil
	}
	return this.kubernetes.GatherPodInfo(id)
}

func getContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return ""
}

// contianerIsFromKubernetes 判断该容器是否由kubernetes创建
// 所有kubernetes启动的容器的containerNamePrefix都是k8s，依据链接如下
// https://github.com/rootsongjc/kubernetes-handbook/blob/master/practice/monitor.md#%E5%AE%B9%E5%99%A8%E7%9A%84%E5%91%BD%E5%90%8D%E8%A7%84%E5%88%99
func contianerIsFromKubernetes(containerName string) bool {
	const kubernetesContainerNamePrefix = "k8s"
	return strings.HasPrefix(containerName, kubernetesContainerNamePrefix)
}

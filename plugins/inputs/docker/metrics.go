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

func (this *Inputs) gather() ([]byte, error) {
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
		data, err := this.gatherContainer(container)
		if err != nil {
			// 忽略某一个container的错误
			// 继续gather下一个
			l.Error(err)
		} else {
			buffer.Write(data)
			buffer.WriteString("\n")
		}
	}

	return buffer.Bytes(), nil
}

func (this *Inputs) gatherContainer(container types.Container) ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, this.timeoutDuration)
	defer cancel()

	tags := map[string]string{
		"container_id":   container.ID,
		"container_name": getContainerName(container.Names),
		"docker_image":   container.ImageID,
		"image_name":     container.Image,
		"stats":          container.State,
	}

	podInfo, err := this.gatherK8sPodInfo(container.ID)
	if err != nil {
		l.Warnf("gather k8s pod error, %s", err)
	}

	for k, v := range podInfo {
		// "pod_name":            "",
		// "pod_phase":           "",
		// TODO:
		// "kube_container_name": "",
		// "kube_daemon_set":     "",
		// "kube_deployment":     "",
		// "kube_namespace":      "",
		// "kube_ownerref_kind ": "",
		// "kube_ownerref_name ": "",
		// "kube_replica_set":    "",
		tags[k] = v
	}

	fields, err := this.gatherStats(ctx, container.ID)
	if err != nil {
		return nil, err
	}

	if podInfo == nil {
		fields["from_kubernetes"] = false
	} else {
		fields["from_kubernetes"] = true
	}

	return io.MakeMetric(inputName, tags, fields, time.Now())
}

const streamStats = false

func (this *Inputs) gatherStats(ctx context.Context, id string) (map[string]interface{}, error) {
	resp, err := this.client.ContainerStats(ctx, id, streamStats)
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
	}, nil
}

func (this *Inputs) composeMessage(ctx context.Context, id string, v *types.ContainerJSON) ([]byte, error) {
	// 容器未启动时，无法进行containerTop，此处会得到error
	// 与 opt.All 冲突，忽略此error即可
	t, _ := this.client.ContainerTop(ctx, id, nil)

	return json.Marshal(struct {
		types.ContainerJSON
		Process containerTop `json:"Process"`
	}{
		*v,
		t,
	})
}

func (this *Inputs) gatherK8sPodInfo(id string) (map[string]string, error) {
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

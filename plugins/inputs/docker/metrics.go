package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	// "github.com/luci/go-render/render"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

// func (d *DockerUtil) do() {
// 	data, err := d.gather()
// 	if err != nil {
// 		return
// 	}
// 	if err := io.NamedFeed(data, io.Object, inputName); err != nil {
// 		l.Error(err)
// 	}
// }

func (d *DockerUtil) gather() ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, defaultAPITimeout)
	defer cancel()

	cList, err := d.client.ContainerList(ctx, types.ContainerListOptions{All: d.IncludeExited})
	if err != nil {
		l.Error(err)
		return nil, err
	}

	var buffer bytes.Buffer

	for _, container := range cList {
		data, err := d.gatherContainer(container)
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

func (d *DockerUtil) gatherContainer(container types.Container) ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, defaultAPITimeout)
	defer cancel()

	tags := map[string]string{
		"container_id":   container.ID,
		"container_name": getContainerName(container.Names),
		"docker_image":   container.ImageID,
		"image_name":     container.Image,
		// "kube_container_name": "",
		// "kube_daemon_set":     "",
		// "kube_deployment":     "",
		// "kube_namespace":      "",
		// "kube_ownerref_kind ": "",
		// "kube_ownerref_name ": "",
		// "kube_replica_set":    "",
		// "pod_name":            "",
		// "pod_phase":           "",
		"stats": container.State,
	}

	fields, _ := d.gatherStats(ctx, container.ID)

	return io.MakeMetric(inputName, tags, fields, time.Now())
}

const streamStats = false

func (d *DockerUtil) gatherStats(ctx context.Context, id string) (map[string]interface{}, error) {
	resp, err := d.client.ContainerStats(ctx, id, streamStats)
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

func (d *DockerUtil) composeMessage(ctx context.Context, id string, v *types.ContainerJSON) ([]byte, error) {
	// 容器未启动时，无法进行containerTop，此处会得到error
	// 与 opt.All 冲突，忽略此error即可
	t, _ := d.client.ContainerTop(ctx, id, nil)

	return json.Marshal(struct {
		types.ContainerJSON
		Process containerTop `json:"Process"`
	}{
		*v,
		t,
	})
}

func getContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return ""
}

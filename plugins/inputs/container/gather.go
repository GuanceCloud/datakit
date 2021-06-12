package container

import (
	"context"
	"encoding/json"
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

			if RegexpMatchString(this.ignoreImageNameRegexps, c.Image) ||
				RegexpMatchString(this.ignoreContainerNameRegexps, getContainerName(c.Names)) {
				l.Debugf("ignore this container, image_name:%s container_name:%s", c.Image, getContainerName(c.Names))
				return
			}

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
		"image_name":     container.Image,
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
			tags[k] = v
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
		"from_kubernetes":    contianerIsFromKubernetes(getContainerName(container.Names)),
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

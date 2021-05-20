package docker

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func (this *Input) gather(option ...*gatherOption) ([]*io.Point, error) {
	var opt *gatherOption
	if len(option) >= 1 {
		opt = option[0]
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, this.apiTimeoutDuration)
	defer cancel()

	cList, err := this.client.ContainerList(ctx, this.opts)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	var pts []*io.Point
	var wg sync.WaitGroup

	for _, container := range cList {
		wg.Add(1)

		go func(c types.Container) {
			defer wg.Done()

			tags := this.gatherContainerInfo(c)

			// 区分指标和对象
			// 对象数据需要有 name 和 container_host 标签
			if opt != nil && opt.IsObjectCategory {
				tags["name"] = c.ID

				containerJson, err := this.client.ContainerInspect(context.Background(), c.ID)
				if err != nil {
					l.Warnf("gather container inspect error: %s", err)
				} else {
					tags["container_host"] = containerJson.Config.Hostname
				}
			}

			fields, err := this.gatherStats(c)
			if err != nil {
				l.Error(err)
				return
			}

			// 对象数据包含 message 字段，其值为 tags 和 fields 所有 Key/Value 的 JSON Marshal
			if opt != nil && opt.IsObjectCategory {
				// 将 tags 和 fields 降维到一层 K/V，避免观看混乱（另外一种方式是 struct{ Tags map[string]string, Fields map[string]interface{} } { // XX }
				// 需要额外的时空间开销，但 object 采集频率不高，可以接受
				message, err := json.Marshal(func() map[string]interface{} {
					var result = make(map[string]interface{}, len(tags)+len(fields))
					for k, v := range tags {
						result[k] = v
					}
					for k, v := range fields {
						result[k] = v
					}
					return result
				})
				if err != nil {
					l.Error(err)
					return
				}
				fields["message"] = string(message)
			}

			pt, err := io.MakePoint(dockerContainersName, tags, fields, time.Now())
			if err != nil {
				l.Error(err)
				return
			}
			pts = append(pts, pt)
		}(container)
	}

	wg.Wait()

	return pts, nil
}

func (this *Input) gatherContainerInfo(container types.Container) map[string]string {
	tags := map[string]string{
		"container_id":   container.ID,
		"container_name": getContainerName(container.Names),
		"docker_image":   container.ImageID,
		"images_name":    container.Image,
		"state":          container.State,
	}

	for k, v := range this.Tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	podInfo, err := this.gatherK8sPodInfo(container.ID)
	if err != nil {
		l.Debugf("gather k8s pod error, %s", err)
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

func (this *Input) gatherK8sPodInfo(id string) (map[string]string, error) {
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

package container

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	v1 "github.com/containerd/cgroups/stats/v1"
	v2 "github.com/containerd/cgroups/v2/stats"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/typeurl"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type containerdInput struct {
	client *containerd.Client
	cfg    *containerdInputConfig
}

type containerdInputConfig struct {
	endpoint  string
	extraTags map[string]string
}

func newContainerdInput(cfg *containerdInputConfig) (*containerdInput, error) {
	cli, err := containerd.New(cfg.endpoint)
	if err != nil {
		return nil, err
	}

	return &containerdInput{client: cli, cfg: cfg}, nil
}

func (c *containerdInput) stop() {
	if err := c.client.Close(); err != nil {
		l.Errorf("closed contianerd, err: %w", err)
	}
}

func (c *containerdInput) gatherObject() ([]inputs.Measurement, error) {
	var res []inputs.Measurement

	nsList, err := c.client.NamespaceService().List(context.TODO())
	if err != nil {
		return nil, err
	}

	l.Debugf("containerd namespaces: %v", nsList)

	for _, ns := range nsList {
		ctx := namespaces.WithNamespace(context.Background(), ns)
		cList, err := c.client.Containers(ctx)
		if err != nil {
			l.Warn("failed to collect containers in containerd, namespace: %s, skip", ns)
			continue
		}

		for _, container := range cList {
			task, err := container.Task(ctx, nil)
			if err != nil {
				l.Warn("failed to create containerd task, err: %w, skip", err)
				continue
			}

			metric, err := task.Metrics(ctx)
			if err != nil {
				l.Warn("failed to get containerd metrics, err: %w, skip", err)
				continue
			}
			metricsData, err := typeurl.UnmarshalAny(metric.Data)
			if err != nil {
				l.Warn("failed to unmarshal containerd metrics, err: %w, skip", err)
				continue
			}

			oldCPU, err := cpuContainerStats(metricsData, time.Now())
			if err != nil {
				l.Warn(err)
				continue
			}

			if oldCPU.usageCoreNanoSeconds == 0 {
				// not running
				continue
			}

			info, err := container.Info(ctx)
			if err != nil {
				l.Warn("failed to get containerd info, err: %w, skip", err)
				continue
			}

			imageName, imageShortName, imageTag := ParseImage(info.Image)
			if imageShortName == "pause" {
				continue
			}

			obj := &containerdObject{time: time.Now()}
			obj.tags = map[string]string{
				"name":             info.ID,
				"namespace":        ns,
				"container_id":     info.ID,
				"image":            info.Image,
				"image_name":       imageName,
				"image_short_name": imageShortName,
				"image_tag":        imageTag,
				"runtime":          info.Runtime.Name,
				"container_type":   "containerd",
			}
			obj.fields = map[string]interface{}{
				// 毫秒除以1000得秒数，不使用Second()因为它返回浮点
				"age": time.Since(info.CreatedAt).Milliseconds() / 1e3,
			}

			if containerName := info.Labels[containerLableForPodContainerName]; containerName != "" {
				obj.tags["container_name"] = containerName
			} else {
				obj.tags["container_name"] = "unknown"
			}

			obj.tags.addValueIfNotEmpty("pod_name", info.Labels[containerLableForPodName])
			obj.tags.addValueIfNotEmpty("pod_namespace", info.Labels[containerLableForPodNamespace])
			obj.tags.append(c.cfg.extraTags)

			mem, err := memoryContainerStats(metricsData)
			if err != nil {
				l.Warn(err)
				continue
			} else {
				obj.fields["mem_usage"] = mem.worksetBytes
				obj.fields["mem_limit"] = mem.limitBytes
				if mem.limitBytes != 0 {
					obj.fields["mem_used_percent"] = float64(mem.worksetBytes) / float64(mem.limitBytes)
				}
			}

			func() {
				//nolint:gosec
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
				metric2, err := task.Metrics(ctx)
				if err != nil {
					l.Warn("failed to get containerd metrics, err: %w, skip", err)
					return
				}
				metricsData2, err := typeurl.UnmarshalAny(metric2.Data)
				if err != nil {
					l.Warn("failed to unmarshal containerd metrics, err: %w, skip", err)
					return
				}
				newCPU, err := cpuContainerStats(metricsData2, time.Now())
				if err != nil {
					l.Warn(err)
					return
				}
				obj.fields["cpu_usage"] = oldCPU.calculatePercent(newCPU)
			}()
			obj.fields.mergeToMessage(obj.tags)

			res = append(res, obj)
		}
	}

	return res, nil
}

type containerdObject struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func (c *containerdObject) LineProto() (*io.Point, error) {
	// 此处使用 docker_containers 不合适
	return io.NewPoint(dockerContainerName, c.tags, c.fields, &io.PointOption{Time: c.time, Category: datakit.Object})
}

func (c *containerdObject) Info() *inputs.MeasurementInfo {
	// return nil
	return &inputs.MeasurementInfo{}
}

type cpuContainerUsage struct {
	usageCoreNanoSeconds int
	timestamp            time.Time
}

func (c *cpuContainerUsage) calculatePercent(currentUsage *cpuContainerUsage) float64 {
	nanoSeconds := currentUsage.timestamp.UnixNano() - c.timestamp.UnixNano()
	if nanoSeconds <= 0 {
		return 0
	}
	return float64(currentUsage.usageCoreNanoSeconds-c.usageCoreNanoSeconds) /
		float64(nanoSeconds) * 100
}

func cpuContainerStats(stats interface{}, currentTimestamp time.Time) (*cpuContainerUsage, error) {
	switch metrics := stats.(type) {
	case *v1.Metrics:
		if metrics.CPU != nil && metrics.CPU.Usage != nil {
			return &cpuContainerUsage{
				usageCoreNanoSeconds: int(metrics.CPU.Usage.Total),
				timestamp:            currentTimestamp,
			}, nil
		}
	case *v2.Metrics:
		if metrics.CPU != nil {
			return &cpuContainerUsage{
				// convert to nano seconds
				usageCoreNanoSeconds: int(metrics.CPU.UsageUsec * 1000),
				timestamp:            currentTimestamp,
			}, nil
		}
	default:
		return nil, fmt.Errorf("unexpected metrics type: %v", metrics)
	}

	return nil, nil
}

type memContainerUsage struct {
	limitBytes   int
	worksetBytes int
}

func memoryContainerStats(stats interface{}) (*memContainerUsage, error) {
	switch metrics := stats.(type) {
	case *v1.Metrics:
		if metrics.Memory != nil && metrics.Memory.Usage != nil {
			workingSetBytes := getWorkingSet(metrics.Memory)
			return &memContainerUsage{
				worksetBytes: int(getAvailableBytes(metrics.Memory, workingSetBytes)),
				limitBytes:   int(metrics.Memory.HierarchicalMemoryLimit),
			}, nil
		}
	case *v2.Metrics:
		if metrics.Memory != nil {
			workingSetBytes := getWorkingSetV2(metrics.Memory)
			return &memContainerUsage{
				worksetBytes: int(getAvailableBytesV2(metrics.Memory, workingSetBytes)),
				limitBytes:   int(metrics.Memory.UsageLimit),
			}, nil
		}
	default:
		return nil, fmt.Errorf("unexpected metrics type: %v", metrics)
	}

	return nil, nil
}

// getWorkingSet calculates workingset memory from cgroup memory stats.
// The caller should make sure memory is not nil.
// workingset = usage - total_inactive_file.
func getWorkingSet(memory *v1.MemoryStat) uint64 {
	if memory.Usage == nil {
		return 0
	}
	var workingSet uint64
	if memory.TotalInactiveFile < memory.Usage.Usage {
		workingSet = memory.Usage.Usage - memory.TotalInactiveFile
	}
	return workingSet
}

// getWorkingSetV2 calculates workingset memory from cgroupv2 memory stats.
// The caller should make sure memory is not nil.
// workingset = usage - inactive_file.
func getWorkingSetV2(memory *v2.MemoryStat) uint64 {
	var workingSet uint64
	if memory.InactiveFile < memory.Usage {
		workingSet = memory.Usage - memory.InactiveFile
	}
	return workingSet
}

//nolint
func isMemoryUnlimited(v uint64) bool {
	// Size after which we consider memory to be "unlimited". This is not
	// MaxInt64 due to rounding by the kernel.
	// TODO: k8s or cadvisor should export this https://github.com/google/cadvisor/blob/2b6fbacac7598e0140b5bc8428e3bdd7d86cf5b9/metrics/prometheus.go#L1969-L1971
	const maxMemorySize = uint64(1 << 62)

	return v > maxMemorySize
}

// https://github.com/kubernetes/kubernetes/blob/b47f8263e18c7b13dba33fba23187e5e0477cdbd/pkg/kubelet/stats/helper.go#L68-L71
func getAvailableBytes(memory *v1.MemoryStat, workingSetBytes uint64) uint64 {
	// memory limit - working set bytes
	if !isMemoryUnlimited(memory.Usage.Limit) {
		return memory.Usage.Limit - workingSetBytes
	}
	return 0
}

func getAvailableBytesV2(memory *v2.MemoryStat, workingSetBytes uint64) uint64 {
	// memory limit (memory.max) for cgroupv2 - working set bytes
	if !isMemoryUnlimited(memory.UsageLimit) {
		return memory.UsageLimit - workingSetBytes
	}
	return 0
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&containerdObject{})
}

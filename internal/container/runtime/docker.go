// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
)

const DockerRuntime = "docker"

var (
	defaultHeaders   = map[string]string{"User-Agent": "engine-api-cli-1.0"}
	dockerListOption = types.ContainerListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("status", "running")),
	}
)

type dockerClient struct {
	endpoint       string
	runtimeName    string
	runtimeVersion string
	client         *docker.Client

	procMountPoint string
}

func VerifyDockerRuntime(endpoint string) error {
	client, err := docker.NewClientWithOpts(docker.WithHost(endpoint))
	if err != nil {
		return err
	}
	_, err = client.Ping(context.TODO())
	return err
}

func NewDockerRuntime(endpoint string, procMountPoint string) (ContainerRuntime, error) {
	client, err := docker.NewClientWithOpts(
		docker.WithHTTPHeaders(defaultHeaders),
		docker.WithAPIVersionNegotiation(),
		docker.WithHost(endpoint))
	if err != nil {
		return nil, err
	}

	_, err = client.Ping(context.TODO())
	if err != nil {
		return nil, err
	}

	info, err := client.Info(context.Background())
	if err != nil {
		return nil, fmt.Errorf("query docker info fail, err: %w", err)
	}

	return &dockerClient{
		endpoint:       endpoint,
		client:         client,
		runtimeName:    DockerRuntime,
		runtimeVersion: info.ServerVersion,
		procMountPoint: procMountPoint,
	}, nil
}

func (d *dockerClient) Version() (*VersionInfo, error) {
	info, err := d.client.ServerVersion(context.TODO())
	if err != nil {
		return nil, err
	}
	return &VersionInfo{
		PlatformName: info.Platform.Name,
		APIVersion:   info.APIVersion,
	}, nil
}

func (d *dockerClient) ListContainers() ([]*Container, error) {
	cList, err := d.client.ContainerList(context.Background(), dockerListOption)
	if err != nil {
		return nil, err
	}

	var containers []*Container
	var lastErr error

	for _, c := range cList {
		container := &Container{
			ID:             c.ID,
			Name:           GetDockerContainerName(c.Names),
			CreatedAt:      time.Unix(c.Created, 0).UnixNano(),
			Labels:         copyMap(c.Labels),
			RuntimeName:    d.runtimeName,
			RuntimeVersion: d.runtimeVersion,
			Image:          c.Image,
			State:          c.State,
			Status:         c.Status,
		}

		status, err := d.ContainerStatus(c.ID)
		if err != nil {
			lastErr = err
		} else {
			container.Pid = status.Pid
			container.LogPath = status.LogPath
			container.Envs = status.Envs
			container.Mounts = status.Mounts
		}

		containers = append(containers, container)
	}

	return containers, lastErr
}

func (d *dockerClient) ContainerStatus(id string) (*ContainerStatus, error) {
	inspect, err := d.client.ContainerInspect(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("inspect docker fail, err: %w", err)
	}

	status := ContainerStatus{
		ID:      id,
		Name:    inspect.Name,
		LogPath: inspect.LogPath,
		Mounts:  make(map[string]string),
	}

	for _, mount := range inspect.Mounts {
		status.Mounts[filepath.Clean(mount.Destination)] = mount.Source
	}

	if inspect.Config != nil {
		status.Envs = parseDockerEnv(inspect.Config.Env)
	}

	if inspect.State != nil {
		status.Pid = inspect.State.Pid
	}

	if inspect.HostConfig != nil {
		status.MemoryLimitInBytes = inspect.HostConfig.Resources.Memory

		if nanoCPUs := inspect.HostConfig.Resources.NanoCPUs; nanoCPUs != 0 {
			status.CPULimitMillicores = nanoCPUs / 1e6 // milli
		} else if inspect.HostConfig.Resources.CPUPeriod != 0 {
			limit := float64(inspect.HostConfig.Resources.CPUQuota) / float64(inspect.HostConfig.Resources.CPUPeriod)
			status.CPULimitMillicores = int64(limit * 1e3) // milli
		}
	}

	return &status, nil
}

func (d *dockerClient) ContainerTop(id string) (*ContainerTop, error) {
	status, err := d.ContainerStatus(id)
	if err != nil {
		return nil, err
	}

	if status.Pid <= 0 {
		return nil, fmt.Errorf("unexpected pid %d for container %s", status.Pid, status.Name)
	}

	pid := status.Pid
	top := ContainerTop{ID: id, Pid: pid}

	stats, err := d.ContainerStats(id)
	if err != nil {
		return nil, err
	}

	// cpu usage
	top.CPUPercent, top.CPUUsageMillicores = calculateCPUUsageUnix(stats)
	// cpu cores
	top.CPUCores = int64(stats.CPUStats.OnlineCPUs)
	// cpu limit millicores
	top.CPULimitMillicores = status.CPULimitMillicores

	// block io
	top.BlockRead, top.BlockWrite = calculateBlockIO(stats.BlkioStats)

	// network
	var rx, tx int64

	if len(stats.Networks) != 0 {
		var rx, tx uint64
		for name, network := range stats.Networks {
			if name == "lo" {
				continue
			}

			rx += network.RxBytes
			tx += network.TxBytes
		}
	} else {
		rx, tx, _ = getNetworkStat(d.procMountPoint, pid)
	}

	top.NetworkRcvd = rx
	top.NetworkSent = tx

	// memory usage
	top.MemoryWorkingSet = calculateMemUsageUnixNoCache(stats.MemoryStats)
	// memory limit
	top.MemoryLimitInBytes = status.MemoryLimitInBytes
	// memory capacity
	if hostMemory, err := getHostMemory(d.procMountPoint); err == nil {
		top.MemoryCapacity = hostMemory
	}

	return &top, nil
}

func (d *dockerClient) ContainerStats(id string) (*types.StatsJSON, error) {
	resp, err := d.client.ContainerStats(context.Background(), id, false /*no stream*/)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.OSType == "windows" {
		return nil, fmt.Errorf("not support windows docker/container")
	}

	var v *types.StatsJSON
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	return v, nil
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

// calculateCPUUsageUnix return usage percent and millicores.
func calculateCPUUsageUnix(v *types.StatsJSON) (float64, int64) {
	var cpuPercent float64
	var cpuUsageMillicores int64

	var (
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(v.PreCPUStats.CPUUsage.TotalUsage)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(v.PreCPUStats.SystemUsage)
		onlineCPUs  = float64(v.CPUStats.OnlineCPUs)
	)

	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(v.CPUStats.CPUUsage.PercpuUsage))
	}
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuUsagePercentage := (cpuDelta / systemDelta) * onlineCPUs
		cpuPercent = cpuUsagePercentage * 100.0
		cpuUsageMillicores = int64(cpuUsagePercentage * 1e3)
	}

	return cpuPercent, cpuUsageMillicores
}

func calculateBlockIO(blkio types.BlkioStats) (int64, int64) {
	var blkRead, blkWrite int64
	for _, bioEntry := range blkio.IoServiceBytesRecursive {
		if len(bioEntry.Op) == 0 {
			continue
		}
		switch bioEntry.Op[0] {
		case 'r', 'R':
			blkRead += int64(bioEntry.Value)
		case 'w', 'W':
			blkWrite += int64(bioEntry.Value)
		}
	}
	return blkRead, blkWrite
}

func GetDockerContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return "unknown"
}

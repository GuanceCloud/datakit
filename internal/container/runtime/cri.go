// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/GuanceCloud/kubernetes/pkg/kubelet/cri/remote"
	internalapi "k8s.io/cri-api/pkg/apis"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

const sampleTime = time.Second * 1

type criClient struct {
	endpoint       string
	runtimeName    string
	runtimeVersion string
	srv            internalapi.RuntimeService

	procMountPoint string
}

func NewCRIRuntime(endpoint string, procMountPoint string) (ContainerRuntime, error) {
	srv, err := remote.NewRemoteRuntimeService(endpoint, time.Second*3)
	if err != nil {
		return nil, fmt.Errorf("invalid container endpoint %s, err: %w", endpoint, err)
	}

	versionResp, err := srv.Version("")
	if err != nil {
		return nil, fmt.Errorf("could not connect endpoint %s, err: %w", endpoint, err)
	}

	return &criClient{
		endpoint:       endpoint,
		runtimeName:    versionResp.RuntimeName,
		runtimeVersion: versionResp.RuntimeVersion,
		srv:            srv,
		procMountPoint: procMountPoint,
	}, nil
}

var criContainerFilter = &runtimeapi.ContainerFilter{
	State: &runtimeapi.ContainerStateValue{
		State: runtimeapi.ContainerState_CONTAINER_RUNNING,
	},
}

var verbose = true

func (ct *criClient) ListContainers() ([]*Container, error) {
	cList, err := ct.srv.ListContainers(criContainerFilter)
	if err != nil {
		return nil, err
	}

	var containers []*Container
	var lastErr error

	for _, c := range cList {
		container := &Container{
			ID:             c.GetId(),
			Name:           c.GetMetadata().GetName(),
			Labels:         copyMap(c.GetLabels()),
			CreatedAt:      c.GetCreatedAt(),
			RuntimeName:    ct.runtimeName,
			RuntimeVersion: ct.runtimeVersion,
			Image:          c.GetImage().GetImage(),
			State:          "Running",
		}

		status, err := ct.ContainerStatus(c.GetId())
		if err != nil {
			lastErr = err
		} else {
			container.Pid = status.Pid
			container.LogPath = status.LogPath
			container.Envs = status.Envs
			container.Mounts = status.Mounts
			if status.Image != "" {
				container.Image = status.Image
			}
		}

		containers = append(containers, container)
	}

	return containers, lastErr
}

func (ct *criClient) ContainerStatus(id string) (*ContainerStatus, error) {
	resp, err := ct.srv.ContainerStatus(id, verbose)
	if err != nil {
		return nil, fmt.Errorf("query cri status fail, err: %w", err)
	}

	info, err := parseCriInfo(resp.GetInfo()["info"])
	if err != nil {
		return nil, fmt.Errorf("parse cri info fail, err: %w", err)
	}

	status := &ContainerStatus{
		ID:      id,
		Name:    resp.GetStatus().GetMetadata().GetName(),
		Pid:     info.getPid(),
		LogPath: resp.GetStatus().GetLogPath(),
		Image:   resp.GetStatus().GetImage().GetImage(),
		Envs:    info.getConfigEnvs(),
		Mounts:  make(map[string]string),
	}

	for _, mount := range resp.GetStatus().GetMounts() {
		status.Mounts[filepath.Clean(mount.GetContainerPath())] = mount.GetHostPath()
	}

	return status, nil
}

// ContainerTop return container stats info.
//
//	Wait for 1 second window time.
func (ct *criClient) ContainerTop(id string) (*ContainerTop, error) {
	status, err := ct.ContainerStatus(id)
	if err != nil {
		return nil, err
	}

	if status.Pid <= 0 {
		return nil, fmt.Errorf("unexpected pid %d for container %s", status.Pid, status.Name)
	}

	pid := status.Pid
	top := ContainerTop{ID: id, Pid: pid}

	stats, err := ct.srv.ContainerStats(id)
	if err != nil {
		return nil, err
	}

	time.Sleep(sampleTime)

	newStats, err := ct.srv.ContainerStats(id)
	if err != nil {
		return nil, err
	}

	// cpu
	if cpu := newStats.GetCpu().GetUsageCoreNanoSeconds().GetValue(); cpu != 0 {
		// Only generate cpuPerc for running container
		duration := newStats.GetCpu().GetTimestamp() - stats.GetCpu().GetTimestamp()
		if duration == 0 {
			return nil, fmt.Errorf("cpu stat is not updated during sample")
		}
		cpuPerc := float64(cpu-stats.GetCpu().GetUsageCoreNanoSeconds().GetValue()) / float64(duration) * 100
		top.CPUUsage = cpuPerc
	}

	// memory
	top.MemoryWorkingSet = int64(stats.GetMemory().GetWorkingSetBytes().GetValue())
	if available := stats.GetMemory().GetAvailableBytes().GetValue(); available != 0 {
		top.MemoryLimit = top.MemoryWorkingSet + int64(available)
	}

	// cpu cores
	if cores, err := getCPUCores(ct.procMountPoint); err == nil {
		top.CPUCores = cores
	}

	// memory capacity
	if hostMemory, err := getHostMemory(ct.procMountPoint); err == nil {
		top.MemoryCapacity = hostMemory
	}

	// network
	if rx, tx, err := getNetworkStat(ct.procMountPoint, pid); err == nil {
		top.NetworkRcvd = rx
		top.NetworkSent = tx
	}

	return &top, nil
}

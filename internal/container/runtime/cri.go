// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/kubernetes/pkg/kubelet/cri/remote"
	internalapi "k8s.io/cri-api/pkg/apis"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

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
			State:          "Running",
		}

		image := digestImage(c.GetImage().GetImage())

		status, err := ct.ContainerStatus(c.GetId())
		if err != nil {
			lastErr = err
		} else {
			container.Pid = status.Pid
			container.LogPath = status.LogPath
			container.Envs = status.Envs
			image = status.Image
		}

		imageName, shortName, tag := ParseImage(image)
		container.Image = Image{
			Image:     image,
			ImageName: imageName,
			ShortName: shortName,
			Tag:       tag,
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

	return &ContainerStatus{
		ID:      id,
		Name:    resp.GetStatus().GetMetadata().GetName(),
		Pid:     info.getPid(),
		LogPath: resp.GetStatus().GetLogPath(),
		Image:   resp.GetStatus().GetImage().GetImage(),
		Envs:    info.getConfigEnvs(),
	}, nil
}

func (ct *criClient) ContainerTop(id string) (*ContainerTop, error) {
	status, err := ct.ContainerStatus(id)
	if err != nil {
		return nil, err
	}

	if status.Pid <= 0 {
		return nil, fmt.Errorf("unexpected pid %d for container %s", status.Pid, status.Name)
	}

	top := ContainerTop{ID: id, Pid: status.Pid}

	stats, err := ct.srv.ContainerStats(id)
	if err != nil {
		return nil, err
	}

	// cpu usage
	// CPU core-nanoseconds per second.
	top.CPUUsage = float64(stats.GetCpu().GetUsageNanoCores().GetValue()/1000/1000/1000) * 100.0
	// memory
	top.MemoryWorkingSet = int64(stats.GetMemory().GetWorkingSetBytes().GetValue())
	if available := stats.GetMemory().GetAvailableBytes().GetValue(); available != 0 {
		top.MemoryLimit = top.MemoryWorkingSet + int64(available)
	}
	// cpu cores
	if err := top.readCPUCores(ct.procMountPoint); err != nil {
		return nil, err
	}
	// memory capacity
	if err := top.readMemoryCapacity(ct.procMountPoint); err != nil {
		return nil, err
	}
	// network
	if err := top.readNetworkStat(ct.procMountPoint); err != nil {
		return nil, err
	}

	return &top, nil
}

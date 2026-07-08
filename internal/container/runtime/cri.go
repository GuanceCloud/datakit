// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/kubernetes/pkg/kubelet/cri/remote"
	componenthealth "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/health"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

const (
	sampleTime           = time.Second
	criConnectionTimeout = 3 * time.Second
	criReconnectInterval = 30 * time.Second
	criFailureTimeout    = 2 * time.Minute
	criSilenceTimeout    = 10 * time.Minute
)

var criLog = logger.DefaultSLogger("container-runtime")

type criRuntimeService interface {
	Version(apiVersion string) (*runtimeapi.VersionResponse, error)
	ListContainers(filter *runtimeapi.ContainerFilter) ([]*runtimeapi.Container, error)
	ContainerStatus(containerID string, verbose bool) (*runtimeapi.ContainerStatusResponse, error)
	ContainerStats(containerID string) (*runtimeapi.ContainerStats, error)
}

type criRuntimeServiceFactory func(string, time.Duration) (criRuntimeService, error)

type criClient struct {
	endpoint       string
	runtimeName    string
	runtimeVersion string

	srvMu sync.RWMutex
	srv   criRuntimeService
	srvID uint64

	reconnectMu          sync.Mutex
	nextReconnectAttempt time.Time
	newRuntimeService    criRuntimeServiceFactory
	healthReporter       componenthealth.Reporter

	procMountPoint string
}

func NewCRIRuntime(endpoint string, procMountPoint string) (ContainerRuntime, error) {
	reporter := componenthealth.Default.Register("container-runtime", endpoint, componenthealth.Options{
		FailureTimeout: criFailureTimeout,
		SilenceTimeout: criSilenceTimeout,
	})
	client, err := newCRIRuntime(endpoint, procMountPoint, func(endpoint string, timeout time.Duration) (criRuntimeService, error) {
		return remote.NewRemoteRuntimeService(endpoint, timeout)
	})
	if err != nil {
		reporter.Close()
		return nil, err
	}
	client.healthReporter = reporter
	reporter.Alive()
	return client, nil
}

func newCRIRuntime(endpoint, procMountPoint string, factory criRuntimeServiceFactory) (*criClient, error) {
	srv, versionResp, err := connectCRIRuntime(endpoint, factory)
	if err != nil {
		return nil, fmt.Errorf("invalid container endpoint %s, err: %w", endpoint, err)
	}

	return &criClient{
		endpoint:          endpoint,
		runtimeName:       versionResp.RuntimeName,
		runtimeVersion:    versionResp.RuntimeVersion,
		srv:               srv,
		newRuntimeService: factory,
		procMountPoint:    procMountPoint,
	}, nil
}

func connectCRIRuntime(endpoint string, factory criRuntimeServiceFactory) (criRuntimeService, *runtimeapi.VersionResponse, error) {
	srv, err := factory(endpoint, criConnectionTimeout)
	if err != nil {
		return nil, nil, err
	}

	versionResp, err := srv.Version("")
	if err != nil {
		closeCRIRuntime(srv)
		return nil, nil, fmt.Errorf("could not connect endpoint %s, err: %w", endpoint, err)
	}
	return srv, versionResp, nil
}

func closeCRIRuntime(srv criRuntimeService) {
	if closer, ok := srv.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			criLog.Warnf("close CRI runtime client failed: %s", err)
		}
	}
}

// Close releases the runtime connection and unregisters its health reporter.
func (ct *criClient) Close() error {
	if ct.healthReporter != nil {
		ct.healthReporter.Close()
	}
	closeCRIRuntime(ct.currentService())
	return nil
}

var criContainerFilter = &runtimeapi.ContainerFilter{
	State: &runtimeapi.ContainerStateValue{
		State: runtimeapi.ContainerState_CONTAINER_RUNNING,
	},
}

func (ct *criClient) Version() (*VersionInfo, error) {
	srv := ct.currentService()
	info, err := srv.Version("")
	if err != nil {
		return nil, err
	}
	return &VersionInfo{
		PlatformName: info.RuntimeName,
		APIVersion:   info.RuntimeApiVersion,
	}, nil
}

var verbose = true

func (ct *criClient) ListContainers() (containers []*Container, err error) {
	recoveryFailed := false
	defer func() {
		if ct.healthReporter == nil {
			return
		}
		if recoveryFailed {
			ct.healthReporter.Failure()
		} else {
			ct.healthReporter.Alive()
		}
	}()

	failedSrv, failedSrvID, runtimeName, runtimeVersion := ct.serviceSnapshot()
	containers, err = ct.listContainers(failedSrv, runtimeName, runtimeVersion)

	if err == nil || !shouldReconnectCRI(err) {
		return containers, err
	}

	if reconnectErr := ct.reconnect(failedSrvID, err); reconnectErr != nil {
		recoveryFailed = true
		return nil, fmt.Errorf("list containers: %w; reconnect CRI runtime: %v", err, reconnectErr)
	}

	srv, _, runtimeName, runtimeVersion := ct.serviceSnapshot()
	containers, err = ct.listContainers(srv, runtimeName, runtimeVersion)
	if err != nil && shouldReconnectCRI(err) {
		recoveryFailed = true
	}
	return containers, err
}

func (ct *criClient) serviceSnapshot() (criRuntimeService, uint64, string, string) {
	ct.srvMu.RLock()
	defer ct.srvMu.RUnlock()
	return ct.srv, ct.srvID, ct.runtimeName, ct.runtimeVersion
}

func (ct *criClient) currentService() criRuntimeService {
	ct.srvMu.RLock()
	defer ct.srvMu.RUnlock()
	return ct.srv
}

func shouldReconnectCRI(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	for current := err; current != nil; current = errors.Unwrap(current) {
		if code := status.Code(current); code == codes.Unavailable || code == codes.DeadlineExceeded {
			return true
		}
	}
	return false
}

func (ct *criClient) reconnect(failedSrvID uint64, cause error) error {
	ct.reconnectMu.Lock()
	defer ct.reconnectMu.Unlock()

	ct.srvMu.RLock()
	serviceChanged := ct.srvID != failedSrvID
	ct.srvMu.RUnlock()
	if serviceChanged {
		return nil
	}

	now := time.Now()
	if now.Before(ct.nextReconnectAttempt) {
		return fmt.Errorf("next reconnect attempt after %s", ct.nextReconnectAttempt.Format(time.RFC3339))
	}
	ct.nextReconnectAttempt = now.Add(criReconnectInterval)

	// Rebuild the whole client instead of relying only on gRPC redial. A runtime
	// restart can recreate its Unix socket while the old ClientConn remains in a
	// prolonged transient-failure state. The interval above prevents reconnect
	// storms while the node is still recovering.
	criLog.Warnf("CRI request failed, rebuilding runtime client for endpoint %s: %s", ct.endpoint, cause)
	newSrv, versionResp, err := connectCRIRuntime(ct.endpoint, ct.newRuntimeService)
	if err != nil {
		return err
	}

	ct.srvMu.Lock()
	oldSrv := ct.srv
	ct.srv = newSrv
	ct.srvID++
	ct.runtimeName = versionResp.RuntimeName
	ct.runtimeVersion = versionResp.RuntimeVersion
	ct.srvMu.Unlock()

	closeCRIRuntime(oldSrv)
	criLog.Infof("CRI runtime client recovered for endpoint %s", ct.endpoint)
	return nil
}

func (ct *criClient) listContainers(srv criRuntimeService, runtimeName, runtimeVersion string) ([]*Container, error) {
	cList, err := srv.ListContainers(criContainerFilter)
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
			RuntimeName:    runtimeName,
			RuntimeVersion: runtimeVersion,
			Image:          c.GetImage().GetImage(),
			State:          "Running",
		}

		status, err := ct.containerStatus(srv, c.GetId())
		if err != nil {
			lastErr = err
		} else {
			container.Pid = status.Pid
			container.LogPath = status.LogPath
			container.Envs = status.Envs
			container.Mounts = status.Mounts
			container.MergedDir = status.MergedDir
			if status.Image != "" {
				container.Image = status.Image
			}
		}

		containers = append(containers, container)
	}

	return containers, lastErr
}

func (ct *criClient) ContainerStatus(id string) (*ContainerStatus, error) {
	srv := ct.currentService()
	return ct.containerStatus(srv, id)
}

func (ct *criClient) containerStatus(srv criRuntimeService, id string) (*ContainerStatus, error) {
	resp, err := srv.ContainerStatus(id, verbose)
	if err != nil {
		return nil, fmt.Errorf("query cri status fail, err: %w", err)
	}

	if resp.GetStatus() == nil {
		return nil, fmt.Errorf("parse cri status fail, err: status is nil")
	}
	if resp.GetInfo() == nil {
		return nil, fmt.Errorf("parse cri info fail, err: info is nil")
	}

	info, err := ParseCriInfo(resp.GetInfo()["info"])
	if err != nil {
		return nil, fmt.Errorf("parse cri info fail, err: %w", err)
	}

	status := &ContainerStatus{
		ID:      id,
		Pid:     info.getPid(),
		LogPath: resp.GetStatus().GetLogPath(),
		Envs:    info.getConfigEnvs(),
		// Currently containerd share similar rootfs mount base on most distros
		MergedDir: fmt.Sprintf("/run/containerd/io.containerd.runtime.v2.task/k8s.io/%s/rootfs", id),
	}

	if metadata := resp.GetStatus().GetMetadata(); metadata != nil {
		status.Name = metadata.GetName()
	}
	if image := resp.GetStatus().GetImage(); image != nil {
		status.Image = image.GetImage()
	}

	if resource := resp.GetStatus().GetResources(); resource != nil {
		if resource.Linux != nil {
			status.MemoryLimitInBytes = resource.Linux.MemoryLimitInBytes

			if resource.Linux.CpuPeriod != 0 {
				limit := float64(resource.Linux.CpuQuota) / float64(resource.Linux.CpuPeriod)
				status.CPULimitMillicores = int64(limit * 1e3) // milli
			}
		}
	}

	for _, mount := range resp.GetStatus().GetMounts() {
		status.Mounts = append(
			status.Mounts,
			Mount{Type: "bind", Destination: mount.GetContainerPath(), Source: mount.GetHostPath()},
		)
	}

	return status, nil
}

// ContainerTop return container stats info.
//
//	Wait for 1 second window time.
func (ct *criClient) ContainerTop(id string) (*ContainerTop, error) {
	srv := ct.currentService()
	status, err := ct.containerStatus(srv, id)
	if err != nil {
		return nil, err
	}

	if status.Pid <= 0 {
		return nil, fmt.Errorf("unexpected pid %d for container %s", status.Pid, status.Name)
	}

	pid := status.Pid
	top := ContainerTop{ID: id, Pid: pid}

	stats, err := srv.ContainerStats(id)
	if err != nil {
		return nil, err
	}

	time.Sleep(sampleTime)

	newStats, err := srv.ContainerStats(id)
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
		cpuUsagePercentage := float64(cpu-stats.GetCpu().GetUsageCoreNanoSeconds().GetValue()) / float64(duration)
		top.CPUPercent = cpuUsagePercentage * 100
		top.CPUUsageMillicores = int64(cpuUsagePercentage * 1e3)
	}
	top.CPULimitMillicores = status.CPULimitMillicores

	// cpu cores
	if cores, err := getCPUCores(ct.procMountPoint); err == nil {
		top.CPUCores = cores
	}

	// memory
	top.MemoryWorkingSet = int64(stats.GetMemory().GetWorkingSetBytes().GetValue())
	top.MemoryLimitInBytes = status.MemoryLimitInBytes
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

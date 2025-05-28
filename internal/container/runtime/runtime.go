// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package runtime wraps docker and CRI functions
package runtime

type ContainerRuntime interface {
	ListContainers() ([]*Container, error)
	ContainerStatus(id string) (*ContainerStatus, error)
	ContainerTop(id string) (*ContainerTop, error)
	Version() (*VersionInfo, error)
}

type VersionInfo struct {
	PlatformName string
	APIVersion   string
}

type Container struct {
	ID      string
	Pid     int // process id on the host
	Name    string
	Image   string
	Labels  map[string]string
	Envs    map[string]string
	LogPath string

	CreatedAt      int64  // unit nanoseconds
	RuntimeName    string // example: "crio"
	RuntimeVersion string // example: "1.20.1"
	State          string // example: "Running"

	// example: "Up 5 hours"
	// https://github.com/moby/moby/blob/73e09ddecf03477c690b2016a613b06156b54969/container/state.go#L76
	Status string

	// Target To Source
	// example: map["/tmp/opt"] = "/var/lib/docker/volumes/<id>/_data"
	Mounts map[string]string
}

type ContainerStatus struct {
	ID      string
	Name    string
	Pid     int
	Image   string
	LogPath string
	Envs    map[string]string

	CPULimitMillicores int64
	MemoryLimitInBytes int64

	// Target To Source
	// example: map["/tmp/opt"] = "/var/lib/docker/volumes/<id>/_data"
	Mounts map[string]string
}

type ContainerTop struct {
	ID  string
	Pid int

	CPUPercent         float64
	CPUUsageMillicores int64
	CPULimitMillicores int64
	CPUCores           int64

	// unit bytes
	MemoryWorkingSet   int64
	MemoryLimitInBytes int64
	MemoryCapacity     int64 // host memory

	// unit bytes
	NetworkRcvd int64
	NetworkSent int64
	BlockRead   int64
	BlockWrite  int64
}

func getCPUCores(procMountPoint string) (int64, error) {
	cpuinfo, err := newCPUInfo(procMountPoint)
	if err != nil {
		return 0, err
	}
	return cpuinfo.cores(), nil
}

const skipLoopback = true

func getNetworkStat(procMountPoint string, pid int) (rx int64, tx int64, err error) {
	networkStat, err := newNetworkStat(procMountPoint, pid)
	if err != nil {
		return 0, 0, err
	}
	rx = networkStat.rxBytes(skipLoopback)
	tx = networkStat.txBytes(skipLoopback)
	return
}

func getHostMemory(procMountPoint string) (int64, error) {
	meminfo, err := newMemInfo(procMountPoint)
	if err != nil {
		return 0, err
	}
	return meminfo.total(), nil
}

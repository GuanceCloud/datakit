// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package runtime wrap docker and CRI functions
package runtime

type ContainerRuntime interface {
	ListContainers() ([]*Container, error)
	ContainerStatus(id string) (*ContainerStatus, error)
	ContainerTop(id string) (*ContainerTop, error)
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
}

type ContainerStatus struct {
	ID      string
	Name    string
	Pid     int
	Image   string
	LogPath string
	Envs    map[string]string
}

type ContainerTop struct {
	ID  string
	Pid int

	CPUUsage float64
	CPUCores int

	// unit bytes
	MemoryWorkingSet int64
	MemoryLimit      int64
	MemoryCapacity   int64 // host memory

	// unit bytes
	NetworkRcvd int64
	NetworkSent int64
	BlockRead   int64
	BlockWrite  int64
}

func (t *ContainerTop) readCPUCores(procMountPoint string) error {
	cpuinfo, err := newCPUInfo(procMountPoint)
	if err != nil {
		return err
	}
	t.CPUCores = cpuinfo.cores()
	return nil
}

const skipLoopback = true

func (t *ContainerTop) readNetworkStat(procMountPoint string) error {
	networkStat, err := newNetworkStat(procMountPoint, t.Pid)
	if err != nil {
		return err
	}
	t.NetworkRcvd = networkStat.rxBytes(skipLoopback)
	t.NetworkSent = networkStat.txBytes(skipLoopback)
	return nil
}

func (t *ContainerTop) readMemoryCapacity(procMountPoint string) error {
	meminfo, err := newMemInfo(procMountPoint)
	if err != nil {
		return err
	}
	t.MemoryCapacity = meminfo.total()
	return nil
}

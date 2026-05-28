//go:build linux
// +build linux

package l4log

import (
	"regexp"

	"github.com/shirou/gopsutil/process"
)

type ContainerInfo struct {
	crt containerRuntime
}

func NewCrtRuntime(endpoint ...string) (*ContainerInfo, error) {
	ep := "unix:///var/run/docker.sock"
	if len(endpoint) > 0 && endpoint[0] != "" {
		ep = endpoint[0]
	}

	crt, err := newContainerRuntime(ep)
	if err != nil {
		return nil, err
	}

	return &ContainerInfo{
		crt: crt,
	}, nil
}

func (ctr *ContainerInfo) ContainerProcessPid(match *regexp.Regexp) ([]int, error) {
	ci, err := ctr.crt.ListContainers()
	if err != nil {
		return nil, err
	}

	r := []int{}
	for _, c := range ci {
		if p, err := process.NewProcess(int32(c.Pid)); err == nil {
			if name, err := p.Name(); err == nil {
				if match.MatchString(name) {
					r = append(r, c.Pid)
				}
			}
		}
	}
	return r, nil
}

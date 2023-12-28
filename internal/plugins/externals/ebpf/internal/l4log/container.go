//go:build linux
// +build linux

package l4log

import (
	"regexp"

	"github.com/shirou/gopsutil/process"
	cruntime "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
)

type ContainerInfo struct {
	crt cruntime.ContainerRuntime
}

func NewCrtRuntime(ed ...string) (*ContainerInfo, error) {
	crt, err := cruntime.NewDockerRuntime("unix:///var/run/docker.sock", "")
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

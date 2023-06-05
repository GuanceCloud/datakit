package sysmonitor

import (
	pr "github.com/shirou/gopsutil/v3/process"
)

func AllProcess() (map[int][2]string, error) {
	pid2name := map[int][2]string{}

	pses, err := pr.Processes()
	if err != nil {
		return nil, err
	}

	for _, p := range pses {
		name, err := p.Name()
		if err != nil {
			continue
		}

		cmdline, err := p.Cmdline()
		if err != nil {
			continue
		}

		pid2name[int(p.Pid)] = [2]string{name, cmdline}
	}

	return pid2name, nil
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package healthcheck

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	pr "github.com/shirou/gopsutil/v3/process"
)

const (
	processMetricName = "host_process_exception"
	defaultMinRunTime = 10 * time.Minute
)

func (ipt *Input) collectProcess() error {
	ts := time.Now()

	for _, process := range ipt.process {
		processes, pids := ipt.getProcesses(process)
		for _, p := range process.processes {
			if _, ok := pids[p.Pid]; !ok {
				if name, err := p.Name(); err != nil {
					l.Warnf("get process(%s) name failed, %s", p.Pid, err.Error())
				} else {
					startDuration, err := getStartDuration(p)
					if err != nil {
						l.Warnf("get process start_duration failed: %s", err.Error())
					}
					var kvs point.KVs
					kvs = kvs.Add("type", "missing", true, true)
					kvs = kvs.Add("process", name, true, true)
					kvs = kvs.Add("start_duration", startDuration.Microseconds(), false, true)
					kvs = kvs.Add("pid", p.Pid, false, true)
					kvs = kvs.Add("exception", true, false, true)

					for k, v := range ipt.mergedTags {
						kvs = kvs.AddTag(k, v)
					}

					opts := point.DefaultMetricOptions()
					opts = append(opts, point.WithTime(ts))

					ipt.collectCache = append(ipt.collectCache, point.NewPointV2(processMetricName, kvs, opts...))
				}
			}
		}
		process.processes = processes
	}
	return nil
}

func (ipt *Input) getProcesses(p *process) (processList []*pr.Process, pids map[int32]bool) {
	pses, err := pr.Processes()
	if err != nil {
		l.Warnf("get process err: %s", err.Error())
		return
	}

	processList, pids = getMatchedProcess(pses, p)

	return
}

func getMatchedProcess(pses []*pr.Process, p *process) (processList []*pr.Process, pids map[int32]bool) {
	pids = map[int32]bool{}

	for _, ps := range pses {
		name, err := ps.Name()
		if err != nil {
			l.Warnf("ps.Name: %s", err)
			continue
		}
		matched := false
		for _, v := range p.Names {
			if v == name {
				matched = true
				break
			}
		}

		if !matched {
			for _, v := range p.namesRegex {
				if v.MatchString(name) {
					matched = true
				}
			}
		}

		if !matched {
			continue
		}

		startDuration, err := getStartDuration(ps)
		if err != nil {
			l.Warnf("ps.CreateTime: %s", err)
			continue
		}

		if startDuration > p.minRunTime {
			processList = append(processList, ps)
			pids[ps.Pid] = true
		}
	}
	return processList, pids
}

func getStartDuration(p *pr.Process) (du time.Duration, err error) {
	t, err := p.CreateTime()
	if err != nil {
		return
	}
	tm := time.Unix(0, t*1000000)
	du = time.Since(tm)
	return
}

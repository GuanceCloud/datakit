// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package healthcheck

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	pr "github.com/shirou/gopsutil/v3/process"
)

const (
	processMetricName = "host_process_exception"
	defaultMinRunTime = 10 * time.Minute
)

type processInfo struct {
	pid           int32
	name          string
	startDuration time.Duration
}

func (ipt *Input) collectProcess() error {
	ts := time.Now()

	pses, err := pr.Processes()
	if err != nil {
		return fmt.Errorf("get process err: %w", err)
	}

	for _, process := range ipt.process {
		runningProcesses := getMatchedProcess(pses, process)
		for oldPid, oldInfo := range process.processes {
			if _, ok := runningProcesses[oldPid]; !ok {
				l.Infof("process %s(%d) is missing", oldInfo.name, oldPid)
				var kvs point.KVs
				kvs = kvs.Add("type", "missing", true, true)
				kvs = kvs.Add("process", oldInfo.name, true, true)
				kvs = kvs.Add("start_duration", oldInfo.startDuration.Microseconds(), false, true)
				kvs = kvs.Add("pid", oldPid, false, true)
				kvs = kvs.Add("exception", true, false, true)

				for k, v := range ipt.mergedTags {
					kvs = kvs.AddTag(k, v)
				}

				opts := point.DefaultMetricOptions()
				opts = append(opts, point.WithTime(ts))

				ipt.collectCache = append(ipt.collectCache, point.NewPointV2(processMetricName, kvs, opts...))
			}
		}
		process.processes = runningProcesses
		l.Debugf("got %d matched processes: %+#v", len(process.processes), process.processes)
	}
	return nil
}

func getMatchedProcess(pses []*pr.Process, p *process) (processMap map[int32]*processInfo) {
	processMap = make(map[int32]*processInfo)

	for _, ps := range pses {
		name, err := ps.Name()
		if err != nil {
			l.Warnf("ps.Name: %s", err)
			continue
		}
		matched := false

		// check name
		for _, v := range p.Names {
			if v == name {
				matched = true
				break
			}
		}

		// check regexp
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
			processMap[ps.Pid] = &processInfo{
				pid:           ps.Pid,
				name:          name,
				startDuration: startDuration,
			}
		}
	}
	return processMap
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

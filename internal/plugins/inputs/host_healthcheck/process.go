// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package healthcheck

import (
	"fmt"
	"regexp"
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
	cmdLine       string
}

func (ipt *Input) collectProcess(ptTS int64) error {
	pses, err := pr.Processes()
	if err != nil {
		return fmt.Errorf("get process err: %w", err)
	}

	for _, process := range ipt.process {
		runningProcesses := getMatchedProcess(pses, process)

		for oldPid, oldInfo := range process.processes {
			var kvs point.KVs
			kvs = kvs.SetTag("process", oldInfo.name)
			kvs = kvs.SetTag("cmd_line", oldInfo.cmdLine)
			kvs = kvs.Set("start_duration", oldInfo.startDuration.Microseconds())
			kvs = kvs.Set("pid", oldPid)
			kvs = kvs.SetTag("type", noneType)
			kvs = kvs.Set("exception", false)

			if _, ok := runningProcesses[oldPid]; !ok {
				l.Infof("process %s(%d) is missing", oldInfo.name, oldPid)
				kvs = kvs.SetTag("type", "missing")
				kvs = kvs.Set("exception", true)
			}
			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}

			opts := point.DefaultMetricOptions()
			opts = append(opts, point.WithTimestamp(ptTS))
			ipt.collectCache = append(ipt.collectCache, point.NewPoint(processMetricName, kvs, opts...))
		}
		process.processes = runningProcesses
		l.Debugf("got %d matched processes: %+#v", len(process.processes), process.processes)
	}
	return nil
}

// isMatched checks whether the target string is matched with the given string or regex.
func isMatchedString(target string, str []string, strRegex []*regexp.Regexp) bool {
	// check str
	for _, v := range str {
		if v == target {
			return true
		}
	}

	// check regexp
	for _, v := range strRegex {
		if v.MatchString(target) {
			return true
		}
	}

	return false
}

func getMatchedProcess(pses []*pr.Process, p *process) (processMap map[int32]*processInfo) {
	processMap = make(map[int32]*processInfo)

	for _, ps := range pses {
		matched := false
		// check name
		name, err := ps.Name()
		if err != nil {
			l.Warnf("ps.Name: %s", err)
		} else {
			matched = isMatchedString(name, p.Names, p.namesRegex)
		}

		// check cmd line
		cmdLine, err := ps.Cmdline()
		if err != nil {
			l.Warnf("ps.Cmdline for process [%s]: %s", name, err.Error())
		} else if !matched {
			matched = isMatchedString(cmdLine, p.CmdLines, p.cmdLinesRegex)
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
				cmdLine:       cmdLine,
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

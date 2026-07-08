// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/v3/process"
)

func (m *monitor) handlerProfile(w http.ResponseWriter, r *http.Request) {
	// 为指定的Pid生成profile文件 /v1/monitor?pid=1234&duration=10&events=all
	// 或者为指定的进程 名生成profile文件 /v1/monitor?command=flameshot&duration=10&events=cpu,alloc
	if m != nil && m.config != nil && !m.config.profilingEnabled() {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("profiling is disabled"))
		return
	}

	queryParams := r.URL.Query()
	d := queryParams.Get("duration")
	if d == "" {
		d = "60s"
	}

	events := queryParams.Get("events")
	if events == "" {
		events = "cpu"
	}
	log.Debugf("http request params is duration=%d, events=%s", d, events)

	if pids := queryParams.Get("pid"); pids != "" {
		log.Debugf("pid is %s", pids)
		pid, err := strconv.Atoi(pids) // nolint
		if err != nil {
			log.Errorf("pid is not a number")
			return
		}
		stats := m.buildManualTriggerStats(m.findProcessByPID(int32(pid)), int32(pid), events, d, []string{"pid:" + pids})
		if stats.CommandName == "" {
			stats.CommandName = getProcessNameByPID(int32(pid))
		}

		m.statsChan <- stats

		w.WriteHeader(200)
		msg := fmt.Sprintf("begin profiling pid=%d, duration=%s, events=%s", pid, d, events)
		_, _ = w.Write([]byte(msg))
		return
	}

	if command := queryParams.Get("command"); command != "" {
		log.Debugf("command is %s", command)
		pms := m.findProcessesForManualCommand(command)
		for _, pm := range pms {
			stats := m.buildManualTriggerStats(pm, pm.Pid, events, d, []string{"command:" + command})
			m.statsChan <- stats
		}
		w.WriteHeader(200)
		msg := fmt.Sprintf("begin profiling command=%s, duration=%s, events=%s", command, d, events)
		_, _ = w.Write([]byte(msg))
		return
	}

	w.WriteHeader(400)
	_, _ = w.Write([]byte("pid or command is required"))
}

func (m *monitor) buildManualTriggerStats(pm *processM, pid int32, events, duration string, reasons []string) *triggerStats {
	tags := make([]string, 0, len(reasons)+1)
	tags = append(tags, reasons...)

	stats := newTriggerStats(events, duration, tags)
	stats.PID = pid
	stats.Triggered = true

	if m != nil && m.config != nil {
		stats.Reason = append(stats.Reason, m.config.Tags...)
	}

	if pm == nil {
		return stats
	}

	stats.CommandName = pm.Name
	m.applyProcessConfigToStats(stats, pm.configProcess)
	if pm.configProcess != nil {
		stats.Service = pm.configProcess.Service
		stats.Reason = append(stats.Reason, pm.configProcess.Tags...)
		stats.Reason = append(stats.Reason, fmt.Sprintf("service:%s", pm.configProcess.Service))
	}

	return stats
}

func (m *monitor) findProcessesForManualCommand(command string) []*processM {
	re, err := regexp.Compile(command) //nolint
	if err != nil {
		log.Errorf("compile [%s] err: %v", command, err)
		return nil
	}

	results := make([]*processM, 0)
	seen := make(map[int32]struct{})
	for _, pm := range m.cs {
		if pm == nil {
			continue
		}
		if re.MatchString(pm.Name) || re.MatchString(pm.Cmdline) {
			results = append(results, pm)
			seen[pm.Pid] = struct{}{}
		}
	}

	for _, pm := range filterProcessesByRegex(&Process{Command: command}) {
		if pm == nil {
			continue
		}
		if monitored := m.findProcessByPID(pm.Pid); monitored != nil {
			if _, ok := seen[pm.Pid]; ok {
				continue
			}
			results = append(results, monitored)
			seen[pm.Pid] = struct{}{}
			continue
		}
		if _, ok := seen[pm.Pid]; ok {
			continue
		}
		results = append(results, pm)
		seen[pm.Pid] = struct{}{}
	}

	return results
}

func getProcessNameByPID(pid int32) string {
	p, err := process.NewProcess(pid)
	if err != nil {
		return ""
	}
	name, err := p.Name()
	if err != nil {
		return ""
	}
	return name
}

func (m *monitor) startHTTPServer() {
	http.HandleFunc("/v1/profile", m.handlerProfile)
	http.Handle("/metrics", promhttp.Handler())
	log.Infof("start http server on %s:%s", m.config.HTTPConfig.LocalHost, m.config.HTTPConfig.LocalPort)
	log.Infof("profile start at /v1/profile")
	log.Infof("prom http start at /metrics")
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", m.config.HTTPConfig.LocalHost, m.config.HTTPConfig.LocalPort), nil)
	if err != nil {
		log.Errorf("http server err: %v", err)
	}
}

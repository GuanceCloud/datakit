// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (m *monitor) handlerProfile(w http.ResponseWriter, r *http.Request) {
	// 为指定的Pid生成profile文件 /v1/monitor?pid=1234&duration=10&events=all
	// 或者为指定的进程 名生成profile文件 /v1/monitor?command=flameshot&duration=10&events=cpu,alloc

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
		stats := newTriggerStats(events, d, []string{"pid:" + pids})

		stats.PID = int32(pid)
		stats.Triggered = true

		m.statsChan <- stats

		w.WriteHeader(200)
		msg := fmt.Sprintf("begin profiling pid=%d, duration=%s, events=%s", pid, d, events)
		_, _ = w.Write([]byte(msg))
		return
	}

	if command := queryParams.Get("command"); command != "" {
		log.Debugf("command is %s", command)
		// 需要先获取pid 再进行 profile操作
		pms := filterProcessesByRegex(&Process{Command: command})
		for _, pm := range pms {
			stats := newTriggerStats(events, d, []string{"command:" + command})
			stats.PID = pm.Pid
			stats.Triggered = true
			stats.Reason = []string{"command:" + command}

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

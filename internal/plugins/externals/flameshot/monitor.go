// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package flameshot  is a monitor for process.
package flameshot

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type monitor struct {
	config    *Config
	cs        []*processM
	csChan    chan *processM
	statsChan chan *triggerStats
}

func NewMonitor(config *Config) *monitor {
	return &monitor{
		config:    config,
		cs:        []*processM{},
		csChan:    make(chan *processM, 10),
		statsChan: make(chan *triggerStats, 5),
	}
}

// 监控单个命令的资源使用情况.
func (m *monitor) MonitorCommand(p *processM) {
	if err := p.updateProcessStats(); err != nil {
		m.cs = removePID(m.cs, p)
		return
	}
	trigger, tags := p.isTrigger()
	tags = append(tags, m.config.Tags...)
	tags = append(tags, p.configProcess.Tags...)
	// 将service作为tag，方便在中心展示。
	tags = append(tags, fmt.Sprintf("%s:%s", "service", p.configProcess.Service))
	if trigger {
		p.lastProfileTime = time.Now()
		stats := newTriggerStats(p.configProcess.Events, p.configProcess.Duration, tags)
		stats.CommandName = p.Name
		stats.PID = p.Pid
		stats.Triggered = true
		stats.Service = p.configProcess.Service
		m.statsChan <- stats
	}
}

func removePID(cs []*processM, command *processM) []*processM {
	for i, c := range cs {
		if c.Pid == command.Pid {
			return append(cs[:i], cs[i+1:]...)
		}
	}
	return cs
}

// start 启动监控
// 检测和trigger是同步进行，防止同一个进程被重复采集profile。
func (m *monitor) Start(osSignal chan os.Signal) {
	filterProcessesTicker := time.NewTicker(time.Minute)
	defer filterProcessesTicker.Stop()
	if m.config.MonitorInterval == "" {
		m.config.MonitorInterval = "1s"
	}
	t, err := time.ParseDuration(m.config.MonitorInterval)
	if err != nil {
		t = time.Second
	}
	log.Infof("start monitor, interval: %s", t.String())
	monitorCommandTicker := time.NewTicker(t)
	defer monitorCommandTicker.Stop()
	if m.config.HTTPConfig != nil {
		go m.startHTTPServer()
	}
	for {
		select {
		case <-filterProcessesTicker.C:
			for _, p := range m.config.Processes {
				pms := filterProcessesByRegex(p)
				for _, pm := range pms {
					m.csChan <- pm
				}
			}
		case pm := <-m.csChan:
			var has bool
			for _, c := range m.cs {
				if c.Pid == pm.Pid {
					has = true
				}
			}
			if !has {
				// 添加到监控列表
				log.Infof("match: PID=%d, name=%s or cmd=%s", pm.Pid, pm.Name, pm.Cmdline)
				pidCount.WithLabelValues(pm.configProcess.Language)
				m.cs = append(m.cs, pm)
			}
		case <-monitorCommandTicker.C:
			for _, c := range m.cs {
				m.MonitorCommand(c)
			}
		case stats := <-m.statsChan:
			// trigger 执行命令不可以超过5分钟，理论上1分钟就可以结束。
			log.Infof("start run profiling, pid: %d, command: %s,", stats.PID, stats.CommandName)
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
			err = runProfiling(ctx, stats)
			cancel()
			if err != nil {
				log.Errorf("run profiling err: %v", err)
			} else {
				// upload
				err = uploadFileToDataKit(stats, m.config.DataKitAddr)
				if err != nil {
					uploadToDK.WithLabelValues(stats.Service, err.Error())
					log.Errorf("upload to DataKit err: %v", err)
				}
				deleteFile(stats)
			}

		case <-osSignal:
			log.Infof("monitor stop")
			return
		}
	}
}

// filterProcessesByRegex 使用正则表达式过滤进程，并通过channel发送匹配的进程.
func filterProcessesByRegex(p *Process) []*processM {
	// 编译正则表达式
	re, err := regexp.Compile(p.Command) //nolint
	if err != nil {
		log.Errorf("compile [%s] err: %v", p.Command, err)
		return nil
	}

	// 获取所有进程ID
	pids, err := process.Pids()
	if err != nil {
		log.Errorf("get all process: %v", err)
		return nil
	}

	log.Debugf("system has %d process ,start match of %s ...", len(pids), p.Command)

	matchedCount := 0
	pms := make([]*processM, 0)
	for _, pid := range pids {
		proc, err := process.NewProcess(pid)
		if err != nil {
			continue // 进程可能已退出，跳过
		}

		name, err := proc.Name()
		if err != nil {
			log.Errorf("process get name err: %v", err)
			continue
		}
		cmd, err := proc.Cmdline()
		if err != nil {
			log.Errorf("process get cmdline err: %v", err)
			continue
		}
		// 使用正则表达式匹配进程名 或者启动命令
		if re.MatchString(name) || re.MatchString(cmd) || name == p.Language {
			matchedCount++
			processInfo := newProcessM(name, cmd, pid, p)
			pms = append(pms, processInfo)
		}
	}

	log.Debugf("filter matched command count: %d", matchedCount)
	return pms
}

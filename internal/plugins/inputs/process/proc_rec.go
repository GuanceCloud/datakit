// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package process

import (
	"sync"
	"time"

	pr "github.com/shirou/gopsutil/v3/process"
)

func newProcRecorder() *procRecorder {
	return &procRecorder{
		recorder: map[int32]procRecStat{},
	}
}

type procRecStat struct {
	Pid int32

	RecorderTime time.Time

	CPUUser   float64
	CPUSystem float64
}

type procRecorder struct {
	recorder map[int32]procRecStat
	sync.RWMutex
}

func (p *procRecorder) flush(process []*pr.Process, recTime time.Time) {
	p.Lock()
	defer p.Unlock()

	p.recorder = map[int32]procRecStat{}

	for _, ps := range process {
		cputime, err := ps.Times()
		if err != nil {
			l.Errorf("pid: %d, err: %v", ps.Pid, err)
			continue
		}
		if cputime == nil {
			l.Errorf("pid: %d, err: nil cputime ", ps.Pid)
			continue
		}
		p.recorder[ps.Pid] = procRecStat{
			Pid:          ps.Pid,
			RecorderTime: recTime,
			CPUUser:      cputime.User,
			CPUSystem:    cputime.System,
		}
	}
}

// calculatePercentTop 计算一个周期内的 cpu 使用率，
// 若无上一次的记录或启动时间不一致的则返回自启动来的使用率.
func (p *procRecorder) calculatePercentTop(ps *pr.Process, pTime time.Time) float64 {
	p.RLock()
	defer p.RUnlock()

	if ps == nil {
		l.Error("nil point: param pr *Process")
		return 0.
	}

	if rec, ok := p.recorder[ps.Pid]; ok {
		return calculatePercentTop(ps, rec, pTime)
	} else {
		return calculatePercent(ps, pTime)
	}
}

func calculatePercent(ps *pr.Process, tn time.Time) float64 {
	if ps == nil {
		l.Error("nil point: param pr *Process")
		return 0.
	}
	cpuTime, err := ps.Times()
	if err != nil {
		l.Error("pid: %d, err: %w", ps.Pid, err)
		return 0.
	}
	created := time.Unix(0, getCreateTime(ps)*int64(time.Millisecond))
	totalTime := tn.Sub(created).Seconds()

	return 100 * ((cpuTime.User + cpuTime.System) / totalTime)
}

func calculatePercentTop(ps *pr.Process, rec procRecStat, tn time.Time) float64 {
	if ps == nil {
		l.Error("nil point: param pr *Process")
		return 0.
	}
	cpuTime, err := ps.Times()
	if err != nil {
		l.Error("pid: %d, err: %w", ps.Pid, err)
		return 0.
	}
	timeDiff := tn.Sub(rec.RecorderTime).Seconds()
	if timeDiff == 0. || timeDiff == -0. ||
		timeDiff == 0 || timeDiff == -0 {
		l.Error("timeDiff == 0s")
		return 0
	}

	// TODO
	cpuDiff := cpuTime.User + cpuTime.System - rec.CPUUser - rec.CPUSystem
	switch {
	case cpuDiff <= 0.0 && cpuDiff >= -0.0000000001:
		return 0.0
	case cpuDiff < -0.0000000001:
	default:
		return 100 * (cpuDiff / timeDiff)
	}
	return calculatePercent(ps, tn)
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"container/list"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/shirou/gopsutil/v3/process"
)

var test_tmall_pid = int32(182882)

func Test_newProcessM(t *testing.T) {
	tmallPID := test_tmall_pid

	pids, err := process.Pids()
	if err != nil {
		t.Error(err)
		return
	}
	has := false
	for _, p := range pids {
		if p == tmallPID {
			has = true
			break
		}
	}
	if !has {
		t.Logf("not find tmall.jar")
		return
	}

	p, err := process.NewProcess(tmallPID)
	if err != nil {
		t.Errorf("newProcess err %v", err)
		return
	}

	name, err := p.Name()
	if err != nil {
		t.Errorf("name err %v", err)
		return
	}
	t.Logf("name is %s", name)

	cmd, err := p.Cmdline()
	if err != nil {
		t.Errorf("cmdline err %v", err)
		return
	}
	t.Logf("tmall cmd is %s", cmd)

	pm := newProcessM(name, cmd, tmallPID, &Process{})
	if pm == nil {
		t.Errorf("newProcessM err")
	}
}

func Test_processM_updateProcessStats(t *testing.T) {
	pm := newProcessM("java", "java -jar tmall.jar", test_tmall_pid, &Process{})
	if pm == nil {
		t.Logf("newProcessM nil, retrun")
		return
	}
	err := pm.updateProcessStats()
	assert.NoError(t, err)
	if e := pm.CPUHistory.Front(); e != nil {
		if usage, ok := e.Value.(float64); ok {
			t.Logf("CPU percent is %0.2f", usage)
		}
	} else {
		t.Errorf("cpu history is nil")
	}

	if e := pm.MemPercentHistory.Front(); e != nil {
		if usage, ok := e.Value.(float64); ok {
			t.Logf("mem Percent usage is %0.2f", usage)
		}
	} else {
		t.Errorf("mem history is nil")
	}

	if e := pm.MemHistory.Front(); e != nil {
		if usage, ok := e.Value.(float64); ok {
			t.Logf("mem usage is %0.2f", usage)
		}
	} else {
		t.Errorf("mem history is nil")
	}
}

func Test_processM_isTrigger(t *testing.T) {
	pm := &processM{
		CPUHistory:        list.New(),
		MemHistory:        list.New(),
		MemPercentHistory: list.New(),
		MaxSize:           10, // 维护10个元素的环形缓冲区
		configProcess: &Process{
			CPUUsagePercent: 100,
			MEMUsagePercent: 90,
			MEMUsageMB:      80,
		},
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	i := float64(0)
	for range ticker.C {
		i++
		pm.AddMemUsage(i * 5)
		pm.AddMemPercent(i * 5)
		pm.AddCPUUsage(i * 5)
		trigger, tags := pm.isTrigger()
		if trigger {
			t.Logf("trigger is %v, tags is %v  count:%f", trigger, tags, i)
			return
		}
	}
}

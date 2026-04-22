// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"container/list"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_newProcessM(t *testing.T) {
	pid := int32(os.Getpid())
	pm := newProcessM("go-test", "go test ./internal/plugins/externals/flameshot", pid, &Process{})
	if pm == nil {
		t.Errorf("newProcessM err")
	}
}

func Test_processM_updateProcessStats(t *testing.T) {
	pm := newProcessM("go-test", "go test ./internal/plugins/externals/flameshot", int32(os.Getpid()), &Process{})
	if pm == nil {
		t.Fatal("newProcessM nil")
	}
	pm.podCPULimit = "1"
	pm.podMEMLimit = "600Mi"
	err := pm.updateProcessStats()
	assert.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	err = pm.updateProcessStats()
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

	for _, usage := range []float64{100, 100, 100, 100, 100} {
		pm.AddMemUsage(usage)
		pm.AddMemPercent(usage)
		pm.AddCPUUsage(usage)
	}

	trigger, tags := pm.isTrigger()
	assert.True(t, trigger)
	assert.Contains(t, tags, "cpu_avg:100.00")
	assert.Contains(t, tags, "mem_used:100.00")
	assert.Contains(t, tags, "mem_perc_avg:100.00")
}

func Test_processM_isTriggerEmergency(t *testing.T) {
	pm := &processM{
		CPUHistory:        list.New(),
		MemHistory:        list.New(),
		MemPercentHistory: list.New(),
		MaxSize:           10,
		configProcess: &Process{
			MEMUsagePercent:          90,
			MEMUsageMB:               1024,
			MEMUsagePercentEmergency: 95,
			MEMUsageMBEmergency:      900,
		},
		lastProfileTime: time.Now().Add(-2 * time.Minute),
	}

	for _, usage := range []float64{100, 200, 300, 400} {
		pm.AddMemUsage(usage)
		pm.AddMemPercent(usage / 10)
	}

	pm.AddMemUsage(950)
	pm.AddMemPercent(96)

	trigger, tags := pm.isTrigger()
	assert.True(t, trigger)
	assert.Contains(t, tags, "mem_used_emergency:950.00")
	assert.Contains(t, tags, "mem_perc_emergency:96.00")
	assert.NotContains(t, tags, "mem_used:950.00")
}

func Test_processM_triggerDecisionEmergency(t *testing.T) {
	pm := &processM{
		CPUHistory:        list.New(),
		MemHistory:        list.New(),
		MemPercentHistory: list.New(),
		MaxSize:           10,
		configProcess: &Process{
			Duration:                 "60s",
			EmergencyDuration:        "12s",
			MEMUsagePercentEmergency: 95,
		},
		lastProfileTime: time.Now().Add(-2 * time.Minute),
	}

	pm.AddMemPercent(96)

	trigger, tags, emergency := pm.triggerDecision()
	assert.True(t, trigger)
	assert.True(t, emergency)
	assert.Contains(t, tags, "mem_perc_emergency:96.00")
}

func TestParseCPULimit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"m core", "500m", 0.5},
		{"1", "1", 1.0},
		{"4", "4", 4.0},
		{"0.25", "0.25", 0.25},
		{"2.0", "2000m", 2.0},
		{"0.01", "10m", 0.01},
		{"null", "", 0.0},
		{"invalid", "invalid", 0.0},
		{"space", " 1000m ", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCPULimit(tt.input)
			if got != tt.expected {
				t.Errorf("parseCPULimit(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseMemoryLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64 // 字节数
	}{
		{"Mi", "1Mi", 1024 * 1024},
		{"M", "1M", 1000 * 1000},
		{"Gi", "1Gi", 1024 * 1024 * 1024},
		{"G", "1G", 1000 * 1000 * 1000},
		{"Ki", "512Ki", 512 * 1024},
		{"int(byte)", "1024", 1024},
		{"invalid", "abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMemoryLimit(tt.input)
			if got != tt.expected {
				t.Errorf("parseMemoryLimit(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func Test_processM_getMemoryUsagePercentPreferPodLimit(t *testing.T) {
	pm := &processM{
		podMEMLimit: "200Mi",
	}

	percent, limit, source := pm.getMemoryUsagePercent(100 * 1024 * 1024)
	assert.Equal(t, int64(200*1024*1024), limit)
	assert.Equal(t, "pod_limit", source)
	assert.InDelta(t, 50.0, percent, 0.0001)
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !(darwin && arm64)

package cpu

import (
	"runtime"
	"testing"

	tu "github.com/GuanceCloud/cliutils/testutil"
	"github.com/shirou/gopsutil/cpu"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestCPUActiveTotalTime(t *testing.T) {
	cputime := cpu.TimesStat{
		CPU:       "cpu-total",
		User:      17105.0, // delta -5.0
		System:    8550.5,  // delta -2.7
		Idle:      83678.4, // delta -56.4
		Nice:      226.9,
		Iowait:    4758.5,
		Irq:       0.0,
		Softirq:   137.8,
		Steal:     0.0,
		Guest:     0.0,
		GuestNice: 0.0,
	}

	active, total := CPUActiveTotalTime(cputime)
	if total != cputime.Total() {
		t.Errorf("error: The CPU total time should be %.2f, but now it's %.2f", cputime.Total(), total)
	}
	if active != cputime.Total()-cputime.Idle {
		t.Errorf("error: The CPU active time should be %.2f, but now it's %.2f", cputime.Total()-cputime.Idle, active)
	}
}

func TestCalculateUsage(t *testing.T) {
	lastT := cpu.TimesStat{
		CPU:       "cpu-total",
		User:      17105.0, // delta -5.0
		System:    8550.5,  // delta -2.7
		Idle:      83678.4, // delta -56.4
		Nice:      226.9,
		Iowait:    4758.5,
		Irq:       0.0,
		Softirq:   137.8,
		Steal:     0.0,
		Guest:     0.0,
		GuestNice: 0.0,
	}
	nowT := cpu.TimesStat{
		CPU:       "cpu-total",
		User:      17110.0,
		System:    8553.2,
		Idle:      83734.8,
		Nice:      226.9,
		Iowait:    4758.5,
		Irq:       0.0,
		Softirq:   137.8,
		Steal:     0.0,
		Guest:     0.0,
		GuestNice: 0.0,
	}
	nowTCpu0 := cpu.TimesStat{
		CPU:       "cpu0",
		User:      17110.0,
		System:    8553.2,
		Idle:      83734.8,
		Nice:      226.9,
		Iowait:    4758.5,
		Irq:       0.0,
		Softirq:   137.8,
		Steal:     0.0,
		Guest:     0.0,
		GuestNice: 0.0,
	}

	_, nowTotal := CPUActiveTotalTime(nowT)
	_, lastTotal := CPUActiveTotalTime(lastT)
	totalDelta := nowTotal - lastTotal

	if _, err := CalculateUsage(nowTCpu0, lastT, totalDelta); err == nil {
		t.Error("error: Use data of different CPUs to calculate CPU utilization should be disabled.")
	}
	uSt, _ := CalculateUsage(nowT, lastT, totalDelta)
	assertEqualFloat64(t, 100*(nowT.User-lastT.User-(nowT.Guest-lastT.Guest))/totalDelta, uSt.User, "usage_user")
	assertEqualFloat64(t, 100*(nowT.System-lastT.System)/totalDelta, uSt.System, "usage_system")
	assertEqualFloat64(t, 100*(nowT.Idle-lastT.Idle)/totalDelta, uSt.Idle, "usage_idle")
	assertEqualFloat64(t, 100*(nowT.Nice-lastT.Nice-(nowT.GuestNice-lastT.GuestNice))/totalDelta, uSt.Nice, "usage_nice")
	assertEqualFloat64(t, 100*(nowT.Iowait-lastT.Iowait)/totalDelta, uSt.Iowait, "usage_iowait")
	assertEqualFloat64(t, 100*(nowT.Irq-lastT.Irq)/totalDelta, uSt.Irq, "usage_irq")
	assertEqualFloat64(t, 100*(nowT.Softirq-lastT.Softirq)/totalDelta, uSt.Softirq, "usage_softirq")
	assertEqualFloat64(t, 100*(nowT.Steal-lastT.Steal)/totalDelta, uSt.Steal, "usage_steal")
	assertEqualFloat64(t, 100*(nowT.Guest-lastT.Guest)/totalDelta, uSt.Guest, "usage_guest")
	assertEqualFloat64(t, 100*(nowT.GuestNice-lastT.GuestNice)/totalDelta, uSt.GuestNice, "usage_guest_nice")
}

func TestCollect(t *testing.T) {
	lastT := cpu.TimesStat{
		CPU:       "cpu-total",
		User:      17105.0, // delta -5.0
		System:    8550.5,  // delta -2.7
		Idle:      83678.4, // delta -56.4
		Nice:      226.9,
		Iowait:    4758.5,
		Irq:       0.0,
		Softirq:   137.8,
		Steal:     0.0,
		Guest:     0.0,
		GuestNice: 0.0,
	}
	nowT := cpu.TimesStat{
		CPU:       "cpu-total",
		User:      17110.0,
		System:    8553.2,
		Idle:      83734.8,
		Nice:      226.9,
		Iowait:    4758.5,
		Irq:       0.0,
		Softirq:   137.8,
		Steal:     0.0,
		Guest:     0.0,
		GuestNice: 0.0,
	}

	timeStats := [][]cpu.TimesStat{
		{lastT}, {nowT},
	}

	i := &Input{
		ps:     &CPUInfoTest{timeStat: timeStats},
		tagger: datakit.DefaultGlobalTagger(),
	}
	i.setup()
	if err := i.collect(); err != nil {
		t.Error(err)
	} else if len(i.collectCache) != 0 {
		tu.Assert(t, 0 != len(i.collectCache), "")
	}

	if err := i.collect(); err != nil {
		t.Error(err)
	} else if len(i.collectCache) != 1 {
		tu.Equals(t, 1, len(i.collectCache))
	}

	// tags
	v := i.collectCache[0].GetTag("cpu")
	tu.Equals(t, "cpu-total", v)

	// fields
	fields := i.collectCache[0].Fields()

	active, nowTotal := CPUActiveTotalTime(nowT)
	lastActive, lastTotal := CPUActiveTotalTime(lastT)
	totalDelta := nowTotal - lastTotal

	assertEqualFloat64(t, 100*(nowT.User-lastT.User-(nowT.Guest-lastT.Guest))/totalDelta, fields.Get("usage_user").GetF(), "usage_user")
	assertEqualFloat64(t, 100*(nowT.System-lastT.System)/totalDelta, fields.Get("usage_system").GetF(), "usage_system")
	assertEqualFloat64(t, 100*(nowT.Idle-lastT.Idle)/totalDelta, fields.Get("usage_idle").GetF(), "usage_idle")
	assertEqualFloat64(t, 100*(nowT.Nice-lastT.Nice-(nowT.GuestNice-lastT.GuestNice))/totalDelta, fields.Get("usage_nice").GetF(), "usage_nice")
	assertEqualFloat64(t, 100*(nowT.Iowait-lastT.Iowait)/totalDelta, fields.Get("usage_iowait").GetF(), "usage_iowait")
	assertEqualFloat64(t, 100*(nowT.Irq-lastT.Irq)/totalDelta, fields.Get("usage_irq").GetF(), "usage_irq")
	assertEqualFloat64(t, 100*(nowT.Softirq-lastT.Softirq)/totalDelta, fields.Get("usage_softirq").GetF(), "usage_softirq")
	assertEqualFloat64(t, 100*(nowT.Steal-lastT.Steal)/totalDelta, fields.Get("usage_steal").GetF(), "usage_steal")
	assertEqualFloat64(t, 100*(nowT.Guest-lastT.Guest)/totalDelta, fields.Get("usage_guest").GetF(), "usage_guest")
	assertEqualFloat64(t, 100*(nowT.GuestNice-lastT.GuestNice)/totalDelta, fields.Get("usage_guest_nice").GetF(), "usage_guest_nice")
	assertEqualFloat64(t, 100*(active-lastActive)/totalDelta, fields.Get("usage_total").GetF(), "usage_total")
}

func assertEqualFloat64(t *testing.T, expected, actual float64, mName string) {
	t.Helper()
	tu.Assert(t, expected == actual, mName+" expected: %f, got %f", expected, actual)
}

func TestSampleMeasurement(t *testing.T) {
	x := &Input{}

	for _, m := range x.SampleMeasurement() {
		_ = m.Info()
	}
}

func TestCoreTemp(t *testing.T) {
	if _, err := CoreTemp(); err != nil {
		switch runtime.GOOS {
		case datakit.OSWindows, datakit.OSDarwin:
			tu.NotOk(t, err, "CoreTemp: mac/windows should not ok")
		default:
			// CI server may be no core temp
			t.Logf("CoreTemp: %s, ignored", err)
		}
	}
}

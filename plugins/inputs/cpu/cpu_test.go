package cpu

import (
	"runtime"
	"testing"

	"github.com/shirou/gopsutil/cpu"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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

	i := &Input{ps: &CPUInfoTest{timeStat: timeStats}}
	if err := i.Collect(); err != nil {
		t.Error(err)
	} else if len(i.collectCache) != 0 {
		tu.Assert(t, 0 != len(i.collectCache), "")
	}

	if err := i.Collect(); err != nil {
		t.Error(err)
	} else if len(i.collectCache) != 1 {
		tu.Equals(t, 1, len(i.collectCache))
	}

	// tags
	tags := i.collectCache[0].(*cpuMeasurement).tags
	tu.Equals(t, "cpu-total", tags["cpu"])

	// fields
	fields := i.collectCache[0].(*cpuMeasurement).fields

	active, nowTotal := CPUActiveTotalTime(nowT)
	lastActive, lastTotal := CPUActiveTotalTime(lastT)
	totalDelta := nowTotal - lastTotal

	assertEqualFloat64(t, 100*(nowT.User-lastT.User-(nowT.Guest-lastT.Guest))/totalDelta, fields["usage_user"].(float64), "usage_user")
	assertEqualFloat64(t, 100*(nowT.System-lastT.System)/totalDelta, fields["usage_system"].(float64), "usage_system")
	assertEqualFloat64(t, 100*(nowT.Idle-lastT.Idle)/totalDelta, fields["usage_idle"].(float64), "usage_idle")
	assertEqualFloat64(t, 100*(nowT.Nice-lastT.Nice-(nowT.GuestNice-lastT.GuestNice))/totalDelta, fields["usage_nice"].(float64), "usage_nice")
	assertEqualFloat64(t, 100*(nowT.Iowait-lastT.Iowait)/totalDelta, fields["usage_iowait"].(float64), "usage_iowait")
	assertEqualFloat64(t, 100*(nowT.Irq-lastT.Irq)/totalDelta, fields["usage_irq"].(float64), "usage_irq")
	assertEqualFloat64(t, 100*(nowT.Softirq-lastT.Softirq)/totalDelta, fields["usage_softirq"].(float64), "usage_softirq")
	assertEqualFloat64(t, 100*(nowT.Steal-lastT.Steal)/totalDelta, fields["usage_steal"].(float64), "usage_steal")
	assertEqualFloat64(t, 100*(nowT.Guest-lastT.Guest)/totalDelta, fields["usage_guest"].(float64), "usage_guest")
	assertEqualFloat64(t, 100*(nowT.GuestNice-lastT.GuestNice)/totalDelta, fields["usage_guest_nice"].(float64), "usage_guest_nice")
	assertEqualFloat64(t, 100*(active-lastActive)/totalDelta, fields["usage_total"].(float64), "usage_total")
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
			tu.Ok(t, err)
		}
	}
}

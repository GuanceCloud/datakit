// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

// Package class provides a WQL interface for WMI on Windows.
package class

import (
	"fmt"

	"github.com/GuanceCloud/cliutils/point"
)

// Win32_Processor cpu 基本信息,不需要转point.
// nolint:stylecheck
type Win32_Processor struct {
	Name                      string `wmi:"Name"`         // module_name
	Manufacturer              string `wmi:"Manufacturer"` // 制造商(如"GenuineIntel"、"AuthenticAMD")
	NumberOfCores             int    `wmi:"NumberOfCores"`
	NumberOfLogicalProcessors int    `wmi:"NumberOfLogicalProcessors"` // 逻辑处理器数量(包含超线程)
	MaxClockSpeed             int    `wmi:"MaxClockSpeed"`             // mhz
	L2CacheSize               int    `wmi:"L2CacheSize"`               // l2 + l3 = "cache_size" 单位kb
	L3CacheSize               int    `wmi:"L3CacheSize"`               // L3缓存大小(KB)
}

func (p *Win32_Processor) ToString() string {
	return fmt.Sprintf(`
CPU Hardware Information:
-------------------------
[Identification]
  Name: %s
  Manufacturer: %s

[Core Configuration]
  Physical Cores: %d
  Logical Processors: %d
  Base Clock: %.2f GHz

[Cache Hierarchy]
  L2 Cache: %d KB
  L3 Cache: %d KB
		`,
		p.Name,
		p.Manufacturer,
		p.NumberOfCores,
		p.NumberOfLogicalProcessors,
		float64(p.MaxClockSpeed)/1000, // 转换为GHz
		p.L2CacheSize,
		p.L3CacheSize,
	)
}

type Win32_PerfFormattedData_PerfOS_Processor struct {
	// 处理器名称，"_Total"表示所有CPU核心的汇总数据
	Name string `wmi:"Name"`

	// 每秒C1低功耗状态转换次数(轻度睡眠状态)
	C1TransitionsPersec uint32 `wmi:"C1TransitionsPersec"`

	// 每秒C2低功耗状态转换次数(中度睡眠状态)
	C2TransitionsPersec uint32 `wmi:"C2TransitionsPersec"`

	// 每秒C3低功耗状态转换次数(深度睡眠状态)
	C3TransitionsPersec uint32 `wmi:"C3TransitionsPersec"`

	// 延迟过程调用(DPC)速率
	DPCRate uint32 `wmi:"DPCRate"`

	// 每秒排队的延迟过程调用(DPC)数量
	DPCsQueuedPersec uint32 `wmi:"DPCsQueuedPersec"`

	// 每秒硬件中断次数
	InterruptsPersec uint32 `wmi:"InterruptsPersec"`

	// 在C1低功耗状态花费的时间百分比
	PercentC1Time uint32 `wmi:"PercentC1Time"`

	// 在C2低功耗状态花费的时间百分比
	PercentC2Time uint32 `wmi:"PercentC2Time"`

	// 在C3低功耗状态花费的时间百分比
	PercentC3Time uint32 `wmi:"PercentC3Time"`

	// 处理延迟过程调用(DPC)的时间百分比
	PercentDPCTime uint32 `wmi:"PercentDPCTime"`

	// CPU空闲时间百分比(包括所有低功耗状态)
	PercentIdleTime uint32 `wmi:"PercentIdleTime"`

	// 处理硬件中断的时间百分比
	PercentInterruptTime uint32 `wmi:"PercentInterruptTime"`

	// 在内核模式(特权模式)运行的时间百分比
	PercentPrivilegedTime uint32 `wmi:"PercentPrivilegedTime"`

	// CPU总使用率百分比(100% - PercentIdleTime)
	PercentProcessorTime uint32 `wmi:"PercentProcessorTime"`

	// 在用户模式运行的时间百分比
	PercentUserTime uint32 `wmi:"PercentUserTime"`
}

func (p *Win32_PerfFormattedData_PerfOS_Processor) ToString() string {
	return fmt.Sprintf(`
CPU 运行时指标:
------------------------
[Basic Info]
  Processor: %s

[CPU Utilization]
  Total Usage: %d%%
  User Mode: %d%%
  Kernel Mode: %d%%
  Interrupt Time: %d%%
  DPC Time: %d%%
  Idle Time: %d%%

[Power States]
  C1 State: %d%% (%d transitions/sec)
  C2 State: %d%% (%d transitions/sec)
  C3 State: %d%% (%d transitions/sec)

[System Activity]
  Interrupts/sec: %d
  DPCs Queued/sec: %d
  DPC Rate: %d

CPU 运行时 信息 %+v

`,
		p.Name,

		p.PercentProcessorTime,
		p.PercentUserTime,
		p.PercentPrivilegedTime,
		p.PercentInterruptTime,
		p.PercentDPCTime,
		p.PercentIdleTime,

		p.PercentC1Time,
		p.C1TransitionsPersec,
		p.PercentC2Time,
		p.C2TransitionsPersec,
		p.PercentC3Time,
		p.C3TransitionsPersec,

		p.InterruptsPersec,
		p.DPCsQueuedPersec,
		p.DPCRate,
		p,
	)
}

func (p *Win32_PerfFormattedData_PerfOS_Processor) ToPoint(host string, cpuInfo *Win32_Processor) *point.Point {
	opts := point.DefaultMetricOptions()
	var kvs point.KVs
	kvs = kvs.AddTag("cpu", cpuInfo.Name).AddTag("host", host).
		AddV2("usage_total", p.PercentProcessorTime, false).
		AddV2("usage_user", p.PercentUserTime, false).
		AddV2("usage_idle", p.PercentIdleTime, false).
		AddV2("usage_c1", p.PercentC1Time, false).
		AddV2("usage_c2", p.PercentC2Time, false).
		AddV2("usage_c3", p.PercentC3Time, false)
	return point.NewPointV2("cpu", kvs, opts...)
}

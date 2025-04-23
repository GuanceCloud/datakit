// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package class

import (
	"fmt"
)

// nolint:stylecheck
type Win32_PerfFormattedData_PerfOS_Memory struct {
	// 可用物理内存量（字节）
	AvailableBytes uint64 `wmi:"AvailableBytes"`

	// 可用物理内存量（KB）
	AvailableKBytes uint64 `wmi:"AvailableKBytes"`

	// 可用物理内存量（MB）
	AvailableMBytes uint64 `wmi:"AvailableMBytes"`

	// 系统缓存使用的物理内存量（字节）
	CacheBytes uint64 `wmi:"CacheBytes"`

	// 系统缓存历史峰值使用量（字节）
	CacheBytesPeak uint64 `wmi:"CacheBytesPeak"`

	// 每秒发生的缓存故障次数
	CacheFaultsPersec uint32 `wmi:"CacheFaultsPersec"`

	// 虚拟内存提交上限（物理内存+分页文件大小，字节）
	CommitLimit uint64 `wmi:"CommitLimit"`

	// 已提交的虚拟内存量（字节）
	CommittedBytes uint64 `wmi:"CommittedBytes"`

	// 每秒发生的需求零页故障次数（需要清零的内存页）
	DemandZeroFaultsPersec uint32 `wmi:"DemandZeroFaultsPersec"`

	// 完全空闲且已清零的物理内存量（字节）
	FreeAndZeroPageListBytes uint64 `wmi:"FreeAndZeroPageListBytes"`

	// 可用的系统页表条目数量
	FreeSystemPageTableEntries uint32 `wmi:"FreeSystemPageTableEntries"`

	// 备用缓存页的平均生命周期（秒）
	LongTermAverageStandbyCacheLifetimes uint32 `wmi:"LongTermAverageStandbyCacheLifetimes"`

	// 已被修改但尚未写入磁盘的内存页大小（字节）
	ModifiedPageListBytes uint64 `wmi:"ModifiedPageListBytes"`

	// 每秒发生的页面故障总数
	PageFaultsPersec uint32 `wmi:"PageFaultsPersec"`

	// 每秒从磁盘读取的页面数
	PageReadsPersec uint32 `wmi:"PageReadsPersec"`

	// 每秒从磁盘读取的页面输入数
	PagesInputPersec uint32 `wmi:"PagesInputPersec"`

	// 每秒写入磁盘的页面输出数
	PagesOutputPersec uint32 `wmi:"PagesOutputPersec"`

	// 每秒处理的页面总数（输入+输出）
	PagesPersec uint32 `wmi:"PagesPersec"`

	// 每秒发生的页面写入次数
	PageWritesPersec uint32 `wmi:"PageWritesPersec"`

	// 已提交内存占总提交限制的百分比（0-100）
	PercentCommittedBytesInUse uint32 `wmi:"PercentCommittedBytesInUse"`

	// 非分页池的内存分配次数
	PoolNonpagedAllocs uint32 `wmi:"PoolNonpagedAllocs"`

	// 非分页池的大小（字节）
	PoolNonpagedBytes uint64 `wmi:"PoolNonpagedBytes"`

	// 分页池的内存分配次数
	PoolPagedAllocs uint32 `wmi:"PoolPagedAllocs"`

	// 分页池的大小（字节）
	PoolPagedBytes uint64 `wmi:"PoolPagedBytes"`

	// 分页池中实际驻留在物理内存中的部分（字节）
	PoolPagedResidentBytes uint64 `wmi:"PoolPagedResidentBytes"`

	// 为关键系统进程保留的备用缓存量（字节）
	StandbyCacheCoreBytes uint64 `wmi:"StandbyCacheCoreBytes"`

	// 普通优先级的备用缓存量（字节）
	StandbyCacheNormalPriorityBytes uint64 `wmi:"StandbyCacheNormalPriorityBytes"`

	// 为高优先级进程保留的备用缓存量（字节）
	StandbyCacheReserveBytes uint64 `wmi:"StandbyCacheReserveBytes"`

	// 系统缓存中实际驻留在物理内存中的部分（字节）
	SystemCacheResidentBytes uint64 `wmi:"SystemCacheResidentBytes"`

	// 系统代码在物理内存中的驻留量（字节）
	SystemCodeResidentBytes uint64 `wmi:"SystemCodeResidentBytes"`

	// 系统代码的总大小（包括未驻留部分，字节）
	SystemCodeTotalBytes uint64 `wmi:"SystemCodeTotalBytes"`

	// 系统驱动程序在物理内存中的驻留量（字节）
	SystemDriverResidentBytes uint64 `wmi:"SystemDriverResidentBytes"`

	// 系统驱动程序的总大小（包括未驻留部分，字节）
	SystemDriverTotalBytes uint64 `wmi:"SystemDriverTotalBytes"`
}

func (m *Win32_PerfFormattedData_PerfOS_Memory) String() string {
	return fmt.Sprintf(`
Memory Performance Metrics:
--------------------------
[Memory Usage]
  Available: %.2f GB (%.2f MB)
  Committed: %.2f GB / %.2f GB (%d%% used)
  Cache: %.2f MB (Peak: %.2f MB)

[Memory Composition]
  Free/Zeroed: %.2f GB
  Modified: %.2f MB
  Standby Cache: %.2f GB (Core: %.2f MB, Normal: %.2f MB, Reserve: %.2f MB)

[System Resources]
  NonPaged Pool: %.2f MB
  Paged Pool: %.2f MB (Resident: %.2f MB)
  System Cache: %.2f MB
  System Code: %.2f MB / %.2f MB
  System Drivers: %.2f MB / %.2f MB

[Page Fault Activity]
  Page Faults/sec: %d
  Pages Input/sec: %d
  Pages Output/sec: %d
  Total Pages/sec: %d
  Demand Zero Faults/sec: %d
`,
		// Memory Usage
		float64(m.AvailableBytes)/1024/1024/1024,
		float64(m.AvailableMBytes),
		float64(m.CommittedBytes)/1024/1024/1024,
		float64(m.CommitLimit)/1024/1024/1024,
		m.PercentCommittedBytesInUse,
		float64(m.CacheBytes)/1024/1024,
		float64(m.CacheBytesPeak)/1024/1024,

		// Memory Composition
		float64(m.FreeAndZeroPageListBytes)/1024/1024/1024,
		float64(m.ModifiedPageListBytes)/1024/1024,
		(float64(m.StandbyCacheCoreBytes)+float64(m.StandbyCacheNormalPriorityBytes)+float64(m.StandbyCacheReserveBytes))/1024/1024/1024,
		float64(m.StandbyCacheCoreBytes)/1024/1024,
		float64(m.StandbyCacheNormalPriorityBytes)/1024/1024,
		float64(m.StandbyCacheReserveBytes)/1024/1024,

		// System Resources
		float64(m.PoolNonpagedBytes)/1024/1024,
		float64(m.PoolPagedBytes)/1024/1024,
		float64(m.PoolPagedResidentBytes)/1024/1024,
		float64(m.SystemCacheResidentBytes)/1024/1024,
		float64(m.SystemCodeResidentBytes)/1024/1024,
		float64(m.SystemCodeTotalBytes)/1024/1024,
		float64(m.SystemDriverResidentBytes)/1024/1024,
		float64(m.SystemDriverTotalBytes)/1024/1024,

		// Page Fault Activity
		m.PageFaultsPersec,
		m.PagesInputPersec,
		m.PagesOutputPersec,
		m.PagesPersec,
		m.DemandZeroFaultsPersec,
	)
}

func (m Win32_PerfFormattedData_PerfOS_Memory) ToShortString() string {
	return fmt.Sprintf(`
	Memory Summary:
  	Available: %.2f GB
  	Used: %d%%
  	Cache: %.2f MB
  	Page Faults: %d/sec

-----------------------------\n
`,
		float64(m.AvailableBytes)/1024/1024/1024,
		m.PercentCommittedBytesInUse,
		float64(m.CacheBytes)/1024/1024,
		m.PageFaultsPersec,
	)
}

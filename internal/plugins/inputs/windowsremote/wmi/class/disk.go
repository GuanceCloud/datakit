// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package class

import (
	"fmt"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
)

// nolint:stylecheck
type Win32_LogicalDisk struct {
	Name       string `wmi:"Name"`
	SystemName string `wmi:"SystemName"`
	FileSystem string `wmi:"FileSystem"`
	DriveType  int    `wmi:"DriveType"` // 只有3是本地磁盘，不需要采集网络磁盘/u盘 等等
	DeviceID   string `wmi:"DeviceID"`
	Size       uint64 `wmi:"Size"`      // 磁盘总大小
	FreeSpace  uint64 `wmi:"FreeSpace"` // 空余大小
}

func (disk Win32_LogicalDisk) String() string {
	return fmt.Sprintf(
		"DeviceID: %s, Name: %s, System: %s, FileSystem: %s, "+
			"Size: %d bytes (%.2f GB), Free: %d bytes (%.2f GB), Used: bytes (%.2f GB)",
		disk.DeviceID,
		disk.Name,
		disk.SystemName,
		disk.FileSystem,
		disk.Size,
		float64(disk.Size)/1024/1024/1024,
		disk.FreeSpace,
		float64(disk.FreeSpace)/1024/1024/1024,
		//disk.Used,
		float64(disk.Size-disk.FreeSpace)/1024/1024/1024,
	)
}

func (disk *Win32_LogicalDisk) ToPoint() *point.Point {
	opts := point.DefaultMetricOptions()
	used := disk.Size - disk.FreeSpace
	var kvs point.KVs
	kvs = kvs.AddTag("device", disk.DeviceID).
		AddTag("host", strings.ToLower(disk.SystemName)).
		AddTag("fstype", disk.FileSystem).
		Add("free", disk.FreeSpace).
		Add("total", disk.Size).
		Add("used", used).
		Add("used_percent", float64(used)/float64(disk.Size))
	return point.NewPoint("disk", kvs, opts...)
}

type Win32_PerfFormattedData_PerfDisk_PhysicalDisk struct {
	// 磁盘名称，"_Total"表示所有物理磁盘的汇总数据
	Name string `wmi:"Name"`

	// 每次读取操作的平均字节数
	AvgDiskBytesPerRead uint64 `wmi:"AvgDiskBytesPerRead"`

	// 每次传输操作的平均字节数(包括读写)
	AvgDiskBytesPerTransfer uint64 `wmi:"AvgDiskBytesPerTransfer"`

	// 每次写入操作的平均字节数
	AvgDiskBytesPerWrite uint64 `wmi:"AvgDiskBytesPerWrite"`

	// 磁盘队列的平均长度(所有请求)
	AvgDiskQueueLength uint64 `wmi:"AvgDiskQueueLength"`
	// 第二次删除
	//	// 读取队列的平均长度
	AvgDiskReadQueueLength uint64 `wmi:"AvgDiskReadQueueLength"`

	// 每次读取操作的平均时间(秒)
	AvgDisksecPerRead uint32 `wmi:"AvgDisksecPerRead"`

	// 每次传输操作的平均时间(秒)
	AvgDisksecPerTransfer uint32 `wmi:"AvgDisksecPerTransfer"`

	// 每次写入操作的平均时间(秒)
	AvgDisksecPerWrite uint32 `wmi:"AvgDisksecPerWrite"`

	// 写入队列的平均长度
	AvgDiskWriteQueueLength uint64 `wmi:"AvgDiskWriteQueueLength"`

	// 当前磁盘队列长度(瞬时值)
	CurrentDiskQueueLength uint32 `wmi:"CurrentDiskQueueLength"`

	// 磁盘每秒传输的字节数(包括读写)
	DiskBytesPersec uint64 `wmi:"DiskBytesPersec"`

	// 磁盘每秒读取的字节数
	DiskReadBytesPersec uint64 `wmi:"DiskReadBytesPersec"`

	// 磁盘每秒读取操作次数
	DiskReadsPersec uint32 `wmi:"DiskReadsPersec"`
	// ---------------------第一次删除------------------------------------
	// 磁盘每秒传输操作次数(包括读写)
	DiskTransfersPersec uint32 `wmi:"DiskTransfersPersec"`

	// 磁盘每秒写入的字节数
	DiskWriteBytesPersec uint64 `wmi:"DiskWriteBytesPersec"`

	// 磁盘每秒写入操作次数
	DiskWritesPersec uint32 `wmi:"DiskWritesPersec"`

	// 磁盘用于读取操作的时间百分比
	PercentDiskReadTime uint64 `wmi:"PercentDiskReadTime"`

	// 磁盘活动时间百分比(包括读写)
	PercentDiskTime uint64 `wmi:"PercentDiskTime"`

	// 磁盘用于写入操作的时间百分比
	PercentDiskWriteTime uint64 `wmi:"PercentDiskWriteTime"`

	// 磁盘空闲时间百分比
	PercentIdleTime uint64 `wmi:"PercentIdleTime"`

	// 每秒拆分的I/O操作次数(通常由磁盘碎片引起)
	SplitIOPerSec uint32 `wmi:"SplitIOPerSec"`
}

func (d *Win32_PerfFormattedData_PerfDisk_PhysicalDisk) ToString() string {
	return fmt.Sprintf(`

本地磁盘 Metrics:

	---------------------------------
	[Basic Info]
	  Disk Name: %s

	[Throughput]
	  Disk Bytes/sec: %.2f MB/s
	  Read Bytes/sec: %.2f MB/s
	  Write Bytes/sec: %.2f MB/s
	  Transfers/sec: %d

	[Latency]
	  Avg. Read Time: %.2f ms
	  Avg. Write Time: %.2f ms
	  Avg. Transfer Time: %.2f ms

	[Queue Length]
	  Current Queue: %d
	  Avg. Queue: %.2f
	  Avg. Read Queue: %.2f
	  Avg. Write Queue: %.2f

	[Utilization]
	  %% Disk Time: %d%%
	  %% Read Time: %d%%
	  %% Write Time: %d%%
	  %% Idle Time: %d%%

	[IO Characteristics]
	  Avg. Bytes/Read: %.2f KB
	  Avg. Bytes/Write: %.2f KB
	  Avg. Bytes/Transfer: %.2f KB
	  Split IO/sec: %d
--------------------------------------
`,

		d.Name,
		float64(d.DiskBytesPersec)/1024/1024,
		float64(d.DiskReadBytesPersec)/1024/1024,
		float64(d.DiskWriteBytesPersec)/1024/1024,
		d.DiskTransfersPersec,
		float64(d.AvgDisksecPerRead)/1000,     // 转换为毫秒
		float64(d.AvgDisksecPerWrite)/1000,    // 转换为毫秒
		float64(d.AvgDisksecPerTransfer)/1000, // 转换为毫秒
		d.CurrentDiskQueueLength,
		float64(d.AvgDiskQueueLength),
		float64(d.AvgDiskReadQueueLength),
		float64(d.AvgDiskWriteQueueLength),
		d.PercentDiskTime,
		d.PercentDiskReadTime,
		d.PercentDiskWriteTime,
		d.PercentIdleTime,
		float64(d.AvgDiskBytesPerRead)/1024,
		float64(d.AvgDiskBytesPerWrite)/1024,
		float64(d.AvgDiskBytesPerTransfer)/1024,
		d.SplitIOPerSec,
	)
}

func (perDisk *Win32_PerfFormattedData_PerfDisk_PhysicalDisk) ToPoint(host string) *point.Point {
	opts := point.DefaultMetricOptions()
	var kvs point.KVs
	kvs = kvs.AddTag("host", host).
		AddTag("name", "total").
		Add("read_bytes", perDisk.AvgDiskBytesPerRead).
		Add("read_bytes/sec", perDisk.DiskReadsPersec).
		Add("write_bytes", perDisk.AvgDiskBytesPerWrite).
		Add("write_bytes/sec", perDisk.DiskWriteBytesPersec).
		Add("read_time", perDisk.AvgDisksecPerRead).
		Add("write_time", perDisk.AvgDisksecPerWrite).
		Add("io_time", perDisk.AvgDisksecPerTransfer)

	return point.NewPoint("diskio", kvs, opts...)
}

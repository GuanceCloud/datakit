// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mem

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

// nolint:lll
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricName,
		Cat:    point.Metric,
		Desc:   "Host virtual memory usage from gopsutil, including common memory totals and platform-specific Darwin or Linux memory counters.",
		DescZh: "通过 gopsutil 采集的主机虚拟内存使用情况，包括通用内存容量以及 Darwin 或 Linux 平台特定的内存计数。",
		Fields: map[string]interface{}{
			"total": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Total physical memory in bytes.",
			},
			"available": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Memory available for starting new applications without swapping, in bytes.",
			},
			"available_percent": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Float,
				Unit:     inputs.Percent,
				Desc:     "Available memory percentage, calculated as `available / total * 100`.",
			},
			"used": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Used physical memory in bytes.",
			},
			"used_percent": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Float,
				Unit:     inputs.Percent,
				Desc:     "Used memory percentage, calculated as `used / total * 100`.",
			},
			"active": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Memory used recently and normally not reclaimed unless necessary. Darwin and Linux only.",
			},
			"free": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Free memory in bytes. Darwin and Linux only.",
			},
			"inactive": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Memory used less recently and more eligible for reclaim. Darwin and Linux only.",
			},
			"wired": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Wired memory that cannot be paged out. Darwin only.",
			},
			"buffered": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Memory used for kernel buffers. Linux only.",
			},
			"cached": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Memory used for the filesystem page cache. Linux only.",
			},
			"commit_limit": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Total memory currently available for allocation under the Linux overcommit policy. Linux only.",
			},
			"committed_as": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Estimated memory committed to allocations on the system. Linux only.",
			},
			"dirty": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Memory waiting to be written back to disk. Linux only.",
			},
			"high_free": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Free high memory in bytes. Linux only.",
			},
			"high_total": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Total high memory in bytes. Linux only.",
			},
			"huge_pages_free": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of huge pages in the pool that are not allocated. Linux only.",
			},
			"huge_pages_size": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Huge page size in bytes. Linux only.",
			},
			"huge_pages_total": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of huge pages in the pool. Linux only.",
			},
			"low_free": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Free low memory in bytes. Linux only.",
			},
			"low_total": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Total low memory in bytes. Linux only.",
			},
			"mapped": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Memory used for mapped files such as libraries. Linux only.",
			},
			"page_tables": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Memory used for the lowest level of page tables. Linux only.",
			},
			"shared": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Shared memory in bytes. Linux only.",
			},
			"slab": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Memory used by in-kernel data structure caches. Linux only.",
			},
			"sreclaimable": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Reclaimable part of slab memory, such as caches. Linux only.",
			},
			"sunreclaim": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Unreclaimable part of slab memory. Linux only.",
			},
			"swap_cached": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Memory that was swapped out, swapped back in, and still also exists in swap. Linux only.",
			},
			"swap_free": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Unused swap space in bytes. Linux only.",
			},
			"swap_total": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Total swap space in bytes. Linux only.",
			},
			"vmalloc_chunk": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Largest contiguous free block in the vmalloc area. Linux only.",
			},
			"vmalloc_total": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Total size of the vmalloc area in bytes. Linux only.",
			},
			"vmalloc_used": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Used vmalloc area in bytes. Linux only.",
			},
			"write_back": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Memory actively being written back to disk. Linux only.",
			},
			"write_back_tmp": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Memory used by FUSE for temporary writeback buffers. Linux only.",
			},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "System hostname."},
		},
	}
}

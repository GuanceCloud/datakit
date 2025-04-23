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
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

// nolint:stylecheck
type Process struct {
	Win32_Process
	Win32_PerfFormattedData_PerfProc_Process
}

func (p Process) ToPoint(tags map[string]string) *point.Point {
	if p.CommandLine == "" {
		p.CommandLine = p.Name
	}
	opts := point.DefaultObjectOptions()
	var kvs point.KVs
	for k, v := range tags {
		kvs = kvs.AddTag(k, v)
	}
	startTime := time.Now().UnixMilli() - int64(p.ElapsedTime*1000)
	kvs = kvs.AddTag("host", strings.ToLower(p.CSName)).
		AddTag("name", fmt.Sprintf("%s_%d", p.CSName, p.IDProcess)).
		AddTag("process_name", p.Name).
		AddV2("pid", p.IDProcess, false).
		AddV2("cmdline", p.CommandLine, false).
		AddV2("proc_syscr", p.IODataBytesPersec, false).
		AddV2("proc_read_bytes", p.IOReadBytesPersec, false).
		AddV2("proc_syscw", p.IOWriteBytesPersec, false).
		AddV2("proc_write_bytes", p.IOWriteOperationsPersec, false).
		AddV2("rss", p.PageFileBytes, false).
		AddV2("threads", p.ThreadCount, false).
		AddV2("cpu_usage", float64(p.PercentProcessorTime)/100, false).
		AddV2("start_time", startTime, false).
		AddV2("started_duration", p.ElapsedTime, false).
		AddV2("open_files", p.HandleCount, false)

	return point.NewPointV2("host_processes", kvs, opts...)
}

type Win32_Process struct {
	Name      string // name
	ProcessId uint32 // pid

	CommandLine string // cmdline
	CSName      string
}

type Win32_PerfFormattedData_PerfProc_Process struct {
	IDProcess         uint32 // pid
	IODataBytesPersec uint64 // proc_syscr
	//IOReadOperationsPersec  uint64 //proc_read_bytes
	IOReadBytesPersec       uint64 //proc_read_bytes
	IOWriteBytesPersec      uint64 // proc_syscw
	IOWriteOperationsPersec uint64 //proc_write_bytes
	VirtualBytes            uint64 // rss
	ThreadCount             uint32 // threads
	PercentProcessorTime    uint64 // cpu_usage
	ElapsedTime             uint64 // start_time 秒
	PageFileBytes           uint64 // 分页内存大小 b
	PageFileBytesPeak       uint64 // 分页内存峰值 b
	HandleCount             uint32 // open_files
	/*
		// -- 保留字段
		IODataOperationsPersec  uint64 //
		Caption           string `json:"caption"`
		CreatingProcessID uint32 `json:"creating_process_id"`
		Description       string `json:"description"`
		//ElapsedTime             uint64 `json:"elapsed_time"`
		FrequencyObject   uint64 `json:"frequency_object"`
		FrequencyPerfTime uint64 `json:"frequency_perf_time"`
		FrequencySys100NS uint64 `json:"frequency_sys_100ns"`
		//HandleCount       uint32 `json:"handle_count"`
		//IODataBytesPersec       uint64 `json:"io_data_bytes_persec"`
		IOOtherBytesPersec      uint64 `json:"io_other_bytes_persec"`
		IOOtherOperationsPersec uint64 `json:"io_other_operations_persec"`
		IOReadBytesPersec       uint64 `json:"io_read_bytes_persec"`
		//IOReadOperationsPersec  uint64 `json:"io_read_operations_persec"`
		//IOWriteBytesPersec    uint64 `json:"io_write_bytes_persec"`
		Name             string `json:"name"`
		PageFaultsPersec uint32 `json:"page_faults_persec"`
		//PageFileBytes         uint64 `json:"page_file_bytes"`
		//PageFileBytesPeak     uint64 `json:"page_file_bytes_peak"`
		PercentPrivilegedTime uint64 `json:"percent_privileged_time"`
		//PercentProcessorTime    uint64 `json:"percent_processor_time"`
		PercentUserTime   uint64 `json:"percent_user_time"`
		PoolNonpagedBytes uint32 `json:"pool_nonpaged_bytes"`
		PoolPagedBytes    uint32 `json:"pool_paged_bytes"`
		PriorityBase      uint32 `json:"priority_base"`
		PrivateBytes      uint64 `json:"private_bytes"`
		//PSComputerName    string `json:"ps_computer_name"`
		//ThreadCount             uint32 `json:"thread_count"`
		TimestampObject   uint64 `json:"timestamp_object"`
		TimestampPerfTime uint64 `json:"timestamp_perf_time"`
		TimestampSys100NS uint64 `json:"timestamp_sys_100ns"`
		//VirtualBytes
		VirtualBytesPeak  uint64 `json:"virtual_bytes_peak"`
		WorkingSet        uint64 `json:"working_set"`
		WorkingSetPeak    uint64 `json:"working_set_peak"`
		WorkingSetPrivate uint64 `json:"working_set_private"`
	*/
}

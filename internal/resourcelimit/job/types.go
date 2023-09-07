// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package job

import "syscall"

type Handle = syscall.Handle

const InvalidHandle = syscall.InvalidHandle

type SecurityAttributes = syscall.SecurityAttributes
type StartupInfo = syscall.StartupInfo
type ProcessInformation = syscall.ProcessInformation

const (
	jobObjectExtendedLimitInformation  = 9
	jobObjectCpuRateControlInformation = 15
)

type ProcessAccessFlags uint32

const (
	STANDARD_RIGHTS_READ      ProcessAccessFlags = 0x00020000
	PROCESS_QUERY_INFORMATION                    = 0x00000400
	SYNCHRONIZE                                  = 0x00100000
	PROCESS_SET_INFORMATION                      = 0x02000000
)

// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_associate_completion_port
type JobObjectAssociateCompletionPort struct {
	CompletionKey  uintptr
	CompletionPort Handle
}

type JobObjectLimitFlags uint32

const (
	JOB_OBJECT_LIMIT_WORKINGSET        JobObjectLimitFlags = 1 << iota // 0x0001
	JOB_OBJECT_LIMIT_PROCESS_TIME                                      // 0x0002
	JOB_OBJECT_LIMIT_JOB_TIME                                          // 0x0004
	JOB_OBJECT_LIMIT_ACTIVE_PROCESS                                    // 0x0008
	JOB_OBJECT_LIMIT_AFFINITY                                          // 0x0010
	JOB_OBJECT_LIMIT_PRIORITY_CLASS                                    // 0x0020
	JOB_OBJECT_LIMIT_PRESERVE_JOB_TIME                                 // 0x0040
	JOB_OBJECT_LIMIT_SCHEDULING_CLASS                                  // 0x0080
	JOB_OBJECT_LIMIT_PROCESS_MEMORY                                    // 0x0100
)

// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_basic_limit_information
type JobObjectBasicLimitInformation struct {
	PerProcessUserTimeLimit int64
	PerJobUserTimeLimit     int64
	LimitFlags              JobObjectLimitFlags
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_extended_limit_information
type JobObjectExtendedLimitInformation struct {
	BasicLimitInformation JobObjectBasicLimitInformation
	IOInfo                IOCounters
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

type IOCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type JobObjectCPURateControlFlags uint32

const (
	JOB_OBJECT_CPU_RATE_CONTROL_ENABLE JobObjectCPURateControlFlags = 1 << iota // 0x00000001
	_
	JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP // 0x00000004
)

// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_cpu_rate_control_information
// Only JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP is supported
type JobObjectCPURateControlInformation struct {
	ControlFlags JobObjectCPURateControlFlags
	CPURate      uint32
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows

package winnetflow

import (
	"syscall"
	"unsafe"
)

var (
	modAdvapi32 = syscall.NewLazyDLL("advapi32.dll")
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")
	modTdh      = syscall.NewLazyDLL("tdh.dll")

	procStartTraceW         = modAdvapi32.NewProc("StartTraceW")
	procControlTraceW       = modAdvapi32.NewProc("ControlTraceW")
	procEnableTraceEx2      = modAdvapi32.NewProc("EnableTraceEx2")
	procOpenTraceW          = modAdvapi32.NewProc("OpenTraceW")
	procProcessTrace        = modAdvapi32.NewProc("ProcessTrace")
	procCloseTrace          = modAdvapi32.NewProc("CloseTrace")
	procCreateSemaphoreW    = modKernel32.NewProc("CreateSemaphoreW")
	procWaitForSingleObject = modKernel32.NewProc("WaitForSingleObject")
	procReleaseSemaphore    = modKernel32.NewProc("ReleaseSemaphore")
	procCloseHandle         = modKernel32.NewProc("CloseHandle")

	procTdhGetPropertySize = modTdh.NewProc("TdhGetPropertySize")
	procTdhGetProperty     = modTdh.NewProc("TdhGetProperty")
)

const (
	eventTraceRealTimeMode = 0x00000100

	processTraceModeRealTime    = 0x00000100
	processTraceModeEventRecord = 0x10000000

	eventControlCodeEnableProvider = 1
	eventTraceControlStop          = 1 // EVENT_TRACE_CONTROL_STOP; 0 is QUERY

	eventFilterTypeEventID = 2

	wnodeFlagTracedGUID = 0x00020000

	errorAlreadyExists = 183
	waitObject0        = 0
	waitTimeout        = 0x00000102
	waitFailed         = 0xffffffff
)

type guid [16]byte

type wnodeHeader struct {
	BufferSize        uint32
	ProviderID        uint32
	HistoricalContext uint64
	Union             int64
	Guid              guid
	ClientContext     uint32
	Flags             uint32
}

// eventTraceProperties mirrors EVENT_TRACE_PROPERTIES (evntrace.h). The session
// name follows the struct in memory; LoggerNameOffset points at it.
type eventTraceProperties struct {
	Wnode               wnodeHeader
	BufferSize          uint32
	MinimumBuffers      uint32
	MaximumBuffers      uint32
	MaximumFileSize     uint32
	LogFileMode         uint32
	FlushTimer          uint32
	EnableFlags         uint32
	AgeLimit            int32
	NumberOfBuffers     uint32
	FreeBuffers         uint32
	EventsLost          uint32
	BuffersWritten      uint32
	LogBuffersLost      uint32
	RealTimeBuffersLost uint32
	LoggerThreadID      uintptr
	LogFileNameOffset   uint32
	LoggerNameOffset    uint32
}

type eventDescriptor struct {
	ID      uint16
	Version uint8
	Channel uint8
	Level   uint8
	Opcode  uint8
	Task    uint16
	Keyword uint64
}

type eventHeader struct {
	Size          uint16
	HeaderType    uint16
	Flags         uint16
	EventProperty uint16
	ThreadID      uint32
	ProcessID     uint32
	TimeStamp     int64
	ProviderID    guid
	Descriptor    eventDescriptor
	Time          int64
	ActivityID    guid
}

type etwBufferContext struct {
	Union    uint16
	LoggerID uint16
}

// eventRecord mirrors EVENT_RECORD (evntrace.h).
type eventRecord struct {
	EventHeader       eventHeader
	BufferContext     etwBufferContext
	ExtendedDataCount uint16
	UserDataLength    uint16
	ExtendedData      unsafe.Pointer
	UserData          unsafe.Pointer
	UserContext       unsafe.Pointer
}

type eventTraceHeader struct {
	Size      uint16
	Union1    uint16
	Union2    uint32
	ThreadID  uint32
	ProcessID uint32
	TimeStamp int64
	Union3    [16]byte
	Union4    uint64
}

type eventTrace struct {
	Header           eventTraceHeader
	InstanceID       uint32
	ParentInstanceID uint32
	ParentGuid       guid
	MofData          uintptr
	MofLength        uint32
	UnionCtx         uint32
}

type systemTime struct {
	Year         uint16
	Month        uint16
	DayOfWeek    uint16
	Day          uint16
	Hour         uint16
	Minute       uint16
	Second       uint16
	Milliseconds uint16
}

type timeZoneInformation struct {
	Bias         int32
	StandardName [32]uint16
	StandardDate systemTime
	StandardBias int32
	DaylightName [32]uint16
	DaylightDate systemTime
	DaylightBias int32
}

type traceLogfileHeader struct {
	BufferSize         uint32
	VersionUnion       uint32
	ProviderVersion    uint32
	NumberOfProcessors uint32
	EndTime            int64
	TimerResolution    uint32
	MaximumFileSize    uint32
	LogFileMode        uint32
	BuffersWritten     uint32
	Union1             [16]byte
	LoggerName         *uint16
	LogFileName        *uint16
	TimeZone           timeZoneInformation
	BootTime           int64
	PerfFreq           int64
	StartTime          int64
	ReservedFlags      uint32
	BuffersLost        uint32
}

// eventTraceLogfile mirrors EVENT_TRACE_LOGFILEW (evntrace.h). Only the input
// fields (LoggerName, ProcessTraceMode, EventRecordCallback, Context) matter to
// OpenTrace; the intermediate members are output buffers whose sizes fix the
// callback offsets.
type eventTraceLogfile struct {
	LogFileName         *uint16
	LoggerName          *uint16
	CurrentTime         int64
	BuffersRead         uint32
	ProcessTraceMode    uint32
	CurrentEvent        eventTrace
	LogfileHeader       traceLogfileHeader
	BufferCallback      uintptr
	BufferSize          uint32
	Filled              uint32
	EventsLost          uint32
	EventRecordCallback uintptr
	IsKernelTrace       uint32
	Context             uintptr
}

type eventFilterDescriptor struct {
	Ptr  uint64
	Size uint32
	Type uint32
}

// eventFilterEventID mirrors EVENT_FILTER_EVENT_ID; used for kernel-side event
// ID filtering when enabling a manifest-based provider.
type eventFilterEventID struct {
	FilterIn uint8
	Reserved uint8
	Count    uint16
	Events   [64]uint16
}

type enableTraceParameters struct {
	Version          uint32
	EnableProperty   uint32
	ControlFlags     uint32
	SourceID         guid
	EnableFilterDesc *eventFilterDescriptor
	FilterDescCount  uint32
}

// propertyDataDescriptor mirrors PROPERTY_DATA_DESCRIPTOR (tdh.h).
type propertyDataDescriptor struct {
	PropertyName uint64
	ArrayIndex   uint32
	Reserved     uint32
}

func startTrace(handle *uintptr, name *uint16, props *eventTraceProperties) error {
	r, _, _ := procStartTraceW.Call(
		uintptr(unsafe.Pointer(handle)),
		uintptr(unsafe.Pointer(name)),
		uintptr(unsafe.Pointer(props)),
	)
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}

func controlTrace(handle uintptr, name *uint16, props *eventTraceProperties, code uint32) error {
	r, _, _ := procControlTraceW.Call(
		handle,
		uintptr(unsafe.Pointer(name)),
		uintptr(unsafe.Pointer(props)),
		uintptr(code),
	)
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}

func enableTraceEx2(handle uintptr, providerID *guid, controlCode uint32, level uint8,
	matchAnyKeyword, matchAllKeyword uint64, timeout uint32, params *enableTraceParameters) error {
	r, _, _ := procEnableTraceEx2.Call(
		handle,
		uintptr(unsafe.Pointer(providerID)),
		uintptr(controlCode),
		uintptr(level),
		uintptr(matchAnyKeyword),
		uintptr(matchAllKeyword),
		uintptr(timeout),
		uintptr(unsafe.Pointer(params)),
	)
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}

func openTrace(logfile *eventTraceLogfile) (uintptr, error) {
	r, _, _ := procOpenTraceW.Call(uintptr(unsafe.Pointer(logfile)))
	if r == 0 || r == ^uintptr(0) {
		if r == ^uintptr(0) {
			return 0, syscall.Errno(87) // ERROR_INVALID_PARAMETER
		}
		return 0, syscall.Errno(87)
	}
	return r, nil
}

func processTrace(handleArray *uintptr, count uint32) error {
	r, _, _ := procProcessTrace.Call(
		uintptr(unsafe.Pointer(handleArray)),
		uintptr(count),
		0,
		0,
	)
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}

func closeTrace(handle uintptr) error {
	r, _, _ := procCloseTrace.Call(handle)
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}

func acquireSessionSemaphore(name *uint16) (uintptr, error) {
	handle, _, callErr := procCreateSemaphoreW.Call(0, 1, 1, uintptr(unsafe.Pointer(name)))
	if handle == 0 {
		return 0, callErr
	}

	result, _, waitErr := procWaitForSingleObject.Call(handle, 0)
	switch result {
	case waitObject0:
		return handle, nil
	case waitTimeout:
		_, _, _ = procCloseHandle.Call(handle)
		return 0, syscall.Errno(errorAlreadyExists)
	case waitFailed:
		_, _, _ = procCloseHandle.Call(handle)
		return 0, waitErr
	default:
		_, _, _ = procCloseHandle.Call(handle)
		return 0, syscall.Errno(result)
	}
}

func releaseSessionSemaphore(handle uintptr) {
	if handle == 0 {
		return
	}
	procReleaseSemaphore.Call(handle, 1, 0)
	procCloseHandle.Call(handle)
}

func tdhGetPropertySize(rec *eventRecord, desc *propertyDataDescriptor) (uint32, error) {
	var size uint32
	r, _, _ := procTdhGetPropertySize.Call(
		uintptr(unsafe.Pointer(rec)),
		0,
		0,
		1,
		uintptr(unsafe.Pointer(desc)),
		uintptr(unsafe.Pointer(&size)),
	)
	if r != 0 {
		return 0, syscall.Errno(r)
	}
	return size, nil
}

func tdhGetProperty(rec *eventRecord, desc *propertyDataDescriptor, buf []byte) error {
	r, _, _ := procTdhGetProperty.Call(
		uintptr(unsafe.Pointer(rec)),
		0,
		0,
		1,
		uintptr(unsafe.Pointer(desc)),
		uintptr(len(buf)),
		uintptr(unsafe.Pointer(&buf[0])),
	)
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}

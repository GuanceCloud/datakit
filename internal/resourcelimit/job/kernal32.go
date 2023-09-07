// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package job

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	procCreateJobObjectA         = modkernel32.NewProc("CreateJobObjectA")
	procAssignProcessToJobObject = modkernel32.NewProc("AssignProcessToJobObject")
	procSetInformationJobObject  = modkernel32.NewProc("SetInformationJobObject")
)

func CloseHandle(handle Handle) error {
	err := syscall.CloseHandle(handle)
	if err != nil {
		return os.NewSyscallError("CloseHandle", err)
	}
	return nil
}

func OpenProcess(desiredAccess ProcessAccessFlags, inheritHandle bool, pid uint32) (Handle, error) {
	handle, err := syscall.OpenProcess(uint32(desiredAccess), inheritHandle, pid)
	if err != nil {
		return InvalidHandle, os.NewSyscallError("OpenProcess", err)
	}
	return handle, nil
}

func CreateJobObject(attr *SecurityAttributes, name string) (Handle, error) {
	r1, _, err := procCreateJobObjectA.Call(
		uintptr(unsafe.Pointer(attr)),
		uintptr(unsafe.Pointer(stringToCharPtr(name))),
	)
	if err != syscall.Errno(0) {
		return InvalidHandle, os.NewSyscallError("CreateJobObjectA", err)
	}
	return Handle(r1), nil
}

func AssignProcessToJobObject(job Handle, process Handle) error {
	_, _, err := procAssignProcessToJobObject.Call(
		uintptr(job),
		uintptr(process),
	)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("AssignProcessToJob", err)
	}
	return nil
}

func setInformationJobObject(job Handle, jobObjectInfoClass uint32, ptr uintptr, length uintptr) error {
	_, _, err := procSetInformationJobObject.Call(
		uintptr(job),
		uintptr(jobObjectInfoClass),
		ptr,
		length,
	)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("SetInformationJobObject", err)
	}
	return nil
}

func SetInformationJobObject_ExtendedLimitInformation(job Handle, info *JobObjectExtendedLimitInformation) error {
	return setInformationJobObject(
		job,
		jobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(info)),
		unsafe.Sizeof(*info),
	)
}

func SetInformationJobObject_CPURateControlInformation(job Handle, info *JobObjectCPURateControlInformation) error {
	return setInformationJobObject(
		job,
		jobObjectCpuRateControlInformation,
		uintptr(unsafe.Pointer(info)),
		unsafe.Sizeof(*info),
	)
}

// stringToCharPtr converts a Go string into pointer to a null-terminated cstring.
// This assumes the go string is already ANSI encoded.
// https://medium.com/jettech/breaking-all-the-rules-using-go-to-call-windows-api-2cbfd8c79724
func stringToCharPtr(str string) *uint8 {
	chars := append([]byte(str), 0) // null terminated
	return &chars[0]
}

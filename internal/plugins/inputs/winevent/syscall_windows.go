// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

// Package winevent Input plugin to collect Windows Event Log messages
//
//revive:disable-next-line:var-naming
package winevent

import (
	"syscall"

	"golang.org/x/sys/windows"
)

const NilHandle EvtHandle = 0

// Event log error codes.
// https://msdn.microsoft.com/en-us/library/windows/desktop/ms681382(v=vs.85).aspx
const (
	ErrorInsufficientBuffer syscall.Errno = 122
	ErrorNoMoreItems        syscall.Errno = 259
	ErrorInvalidOperation   syscall.Errno = 4317
)

// EvtSubscribeFlag defines the possible values that specify when to start subscribing to events.
type EvtSubscribeFlag uint32

// EVT_SUBSCRIBE_FLAGS enumeration
// https://msdn.microsoft.com/en-us/library/windows/desktop/aa385588(v=vs.85).aspx
const (
	EvtSubscribeToFutureEvents      EvtSubscribeFlag = 1
	EvtSubscribeStartAtOldestRecord EvtSubscribeFlag = 2
	EvtSubscribeStartAfterBookmark  EvtSubscribeFlag = 3
	EvtSubscribeOriginMask          EvtSubscribeFlag = 0x3
	EvtSubscribeTolerateQueryErrors EvtSubscribeFlag = 0x1000
	EvtSubscribeStrict              EvtSubscribeFlag = 0x10000
)

// EvtRenderFlag uint32.
type EvtRenderFlag uint32

// EVT_RENDER_FLAGS enumeration
// https://msdn.microsoft.com/en-us/library/windows/desktop/aa385563(v=vs.85).aspx
const (
	//revive:disable:var-naming
	// Render the event as an XML string. For details on the contents of the
	// XML string, see the Event schema.
	EvtRenderEventXML EvtRenderFlag = 1
	//revive:enable:var-naming
)

// EvtClearLog removes all events from the specified channel and writes them to
// the target log file.
func EvtClearLog(session EvtHandle, channelPath string, targetFilePath string) error {
	channel, err := windows.UTF16PtrFromString(channelPath)
	if err != nil {
		return err
	}

	var target *uint16
	if targetFilePath != "" {
		target, err = windows.UTF16PtrFromString(targetFilePath)
		if err != nil {
			return err
		}
	}

	return _EvtClearLog(session, channel, target, 0)
}

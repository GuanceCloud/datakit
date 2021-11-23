//go:build windows
// +build windows

// Package winevent Input plugin to collect Windows Event Log messages
//revive:disable-next-line:var-naming
package winevent

import "syscall"

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
	EvtSubscribeToFutureEvents EvtSubscribeFlag = 1
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

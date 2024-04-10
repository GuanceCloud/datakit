// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

// Package winevent Input plugin to collect Windows Event Log messages

package winevent

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _ unsafe.Pointer

// EvtHandle uintptr.
type EvtHandle uintptr

func (h EvtHandle) Close() error {
	return _EvtClose(h)
}

// Do the interface allocations only once for common
// Errno values.
const (
	ErrorIoPending = 997
)

var errIOPending error = syscall.Errno(ErrorIoPending)

// EvtFormatMessageFlag defines the values that specify the message string from
// the event to format.
type EvtFormatMessageFlag uint32

// EVT_FORMAT_MESSAGE_FLAGS enumeration
// https://msdn.microsoft.com/en-us/library/windows/desktop/aa385525(v=vs.85).aspx
const (
	// Format the event's message string.
	EvtFormatMessageEvent EvtFormatMessageFlag = iota + 1
	// Format the message string of the level specified in the event.
	EvtFormatMessageLevel
	// Format the message string of the task specified in the event.
	EvtFormatMessageTask
	// Format the message string of the task specified in the event.
	EvtFormatMessageOpcode
	// Format the message string of the keywords specified in the event. If the
	// event specifies multiple keywords, the formatted string is a list of
	// null-terminated strings. Increment through the strings until your pointer
	// points past the end of the used buffer.
	EvtFormatMessageKeyword
	// Format the message string of the channel specified in the event.
	EvtFormatMessageChannel
	// Format the provider's message string.
	EvtFormatMessageProvider
	// Format the message string associated with a resource identifier. The
	// provider's metadata contains the resource identifiers; the message
	// compiler assigns a resource identifier to each string when it compiles
	// the manifest.
	EvtFormatMessageId
	// Format all the message strings in the event. The formatted message is an
	// XML string that contains the event details and the message strings.
	EvtFormatMessageXml
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e { //nolint:exhaustive
	case 0:
		return nil
	case ErrorIoPending:
		return errIOPending
	default:
		return e
	}
}

var (
	modwevtapi = windows.NewLazySystemDLL("wevtapi.dll")

	procEvtSubscribe                    = modwevtapi.NewProc("EvtSubscribe")
	procEvtRender                       = modwevtapi.NewProc("EvtRender")
	procEvtClose                        = modwevtapi.NewProc("EvtClose")
	procEvtNext                         = modwevtapi.NewProc("EvtNext")
	procEvtFormatMessage                = modwevtapi.NewProc("EvtFormatMessage")
	procEvtOpenPublisherMetadata        = modwevtapi.NewProc("EvtOpenPublisherMetadata")
	procEvtClearLog                     = modwevtapi.NewProc("EvtClearLog")
	procEvtCreateBookmark               = modwevtapi.NewProc("EvtCreateBookmark")
	procEvtUpdateBookmark               = modwevtapi.NewProc("EvtUpdateBookmark")
	procEvtGetEventMetadataProperty     = modwevtapi.NewProc("EvtGetEventMetadataProperty")
	procEvtNextEventMetadata            = modwevtapi.NewProc("EvtNextEventMetadata")
	procEvtCreateRenderContext          = modwevtapi.NewProc("EvtCreateRenderContext")
	procEvtGetObjectArrayProperty       = modwevtapi.NewProc("EvtGetObjectArrayProperty")
	procEvtGetObjectArraySize           = modwevtapi.NewProc("EvtGetObjectArraySize")
	procEvtGetPublisherMetadataProperty = modwevtapi.NewProc("EvtGetPublisherMetadataProperty")
	procEvtNextChannelPath              = modwevtapi.NewProc("EvtNextChannelPath")
	procEvtNextPublisherId              = modwevtapi.NewProc("EvtNextPublisherId")
	procEvtOpenChannelEnum              = modwevtapi.NewProc("EvtOpenChannelEnum")
	procEvtOpenEventMetadataEnum        = modwevtapi.NewProc("EvtOpenEventMetadataEnum")
	procEvtOpenLog                      = modwevtapi.NewProc("EvtOpenLog")
	procEvtOpenPublisherEnum            = modwevtapi.NewProc("EvtOpenPublisherEnum")
	procEvtQuery                        = modwevtapi.NewProc("EvtQuery")
	procEvtSeek                         = modwevtapi.NewProc("EvtSeek")
)

func _EvtSubscribe(session EvtHandle, signalEvent uintptr,
	channelPath *uint16, query *uint16, bookmark EvtHandle, context uintptr,
	callback syscall.Handle, flags EvtSubscribeFlag,
) (handle EvtHandle, err error) {
	r0, _, e1 := syscall.Syscall9(procEvtSubscribe.Addr(), 8, uintptr(session), signalEvent,
		uintptr(unsafe.Pointer(channelPath)), uintptr(unsafe.Pointer(query)), uintptr(bookmark), // nolint:gosec
		context, uintptr(callback), uintptr(flags), 0)
	handle = EvtHandle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _EvtRender(context EvtHandle, fragment EvtHandle,
	flags EvtRenderFlag, bufferSize uint32, buffer *byte, bufferUsed *uint32, propertyCount *uint32,
) (err error) {
	r1, _, e1 := syscall.Syscall9(procEvtRender.Addr(), 7, uintptr(context), uintptr(fragment), uintptr(flags),
		uintptr(bufferSize), uintptr(unsafe.Pointer(buffer)), uintptr(unsafe.Pointer(bufferUsed)), // nolint:gosec
		uintptr(unsafe.Pointer(propertyCount)), 0, 0) // nolint:gosec
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _EvtClose(object EvtHandle) (err error) {
	r1, _, e1 := syscall.Syscall(procEvtClose.Addr(), 1, uintptr(object), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _EvtNext(resultSet EvtHandle, eventArraySize uint32, eventArray *EvtHandle,
	timeout uint32, flags uint32, numReturned *uint32,
) (err error) {
	r1, _, e1 := syscall.Syscall6(procEvtNext.Addr(), 6, uintptr(resultSet), // nolint:gosec
		uintptr(eventArraySize), uintptr(unsafe.Pointer(eventArray)), uintptr(timeout), // nolint:gosec
		uintptr(flags), uintptr(unsafe.Pointer(numReturned))) // nolint:gosec
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _EvtFormatMessage(publisherMetadata EvtHandle, event EvtHandle, messageID uint32,
	valueCount uint32, values uintptr, flags EvtFormatMessageFlag,
	bufferSize uint32, buffer *byte, bufferUsed *uint32,
) (err error) {
	r1, _, e1 := syscall.Syscall9(procEvtFormatMessage.Addr(), 9,
		uintptr(publisherMetadata), uintptr(event), uintptr(messageID),
		uintptr(valueCount), values, uintptr(flags), uintptr(bufferSize),
		uintptr(unsafe.Pointer(buffer)), uintptr(unsafe.Pointer(bufferUsed))) // nolint:gosec
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _EvtOpenPublisherMetadata(session EvtHandle, publisherIdentity *uint16, logFilePath *uint16,
	locale uint32, flags uint32,
) (handle EvtHandle, err error) {
	r0, _, e1 := syscall.Syscall6(procEvtOpenPublisherMetadata.Addr(), 5, uintptr(session),
		uintptr(unsafe.Pointer(publisherIdentity)), uintptr(unsafe.Pointer(logFilePath)), uintptr(locale), uintptr(flags), 0) // nolint:gosec
	handle = EvtHandle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _EvtClearLog(session EvtHandle, channelPath *uint16, targetFilePath *uint16, flags uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procEvtClearLog.Addr(), 4, uintptr(session), uintptr(unsafe.Pointer(channelPath)), uintptr(unsafe.Pointer(targetFilePath)), uintptr(flags), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func _EvtCreateBookmark(bookmarkXML *uint16) (EvtHandle, error) {
	//nolint:gosec // G103: Valid use of unsafe call to pass bookmarkXML
	r0, _, e1 := syscall.SyscallN(procEvtCreateBookmark.Addr(), uintptr(unsafe.Pointer(bookmarkXML)))
	handle := EvtHandle(r0)
	if handle != 0 {
		return handle, nil
	}
	if e1 != 0 {
		return handle, errnoErr(e1)
	}
	return handle, syscall.EINVAL
}

func _EvtUpdateBookmark(bookmark, event EvtHandle) error {
	r0, _, e1 := syscall.SyscallN(procEvtUpdateBookmark.Addr(), uintptr(bookmark), uintptr(event))
	if r0 != 0 {
		return nil
	}
	if e1 != 0 {
		return errnoErr(e1)
	}
	return syscall.EINVAL
}

func _EvtGetEventMetadataProperty(eventMetadata EvtHandle, propertyID EvtEventMetadataPropertyID, flags uint32, bufferSize uint32, variant *EvtVariant, bufferUsed *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procEvtGetEventMetadataProperty.Addr(), 6, uintptr(eventMetadata), uintptr(propertyID), uintptr(flags), uintptr(bufferSize), uintptr(unsafe.Pointer(variant)), uintptr(unsafe.Pointer(bufferUsed)))
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func _EvtNextEventMetadata(enumerator EvtHandle, flags uint32) (handle EvtHandle, err error) {
	r0, _, e1 := syscall.Syscall(procEvtNextEventMetadata.Addr(), 2, uintptr(enumerator), uintptr(flags), 0)
	handle = EvtHandle(r0)
	if handle == 0 {
		err = errnoErr(e1)
	}
	return
}

func _EvtOpenEventMetadataEnum(publisherMetadata EvtHandle, flags uint32) (handle EvtHandle, err error) {
	r0, _, e1 := syscall.Syscall(procEvtOpenEventMetadataEnum.Addr(), 2, uintptr(publisherMetadata), uintptr(flags), 0)
	handle = EvtHandle(r0)
	if handle == 0 {
		err = errnoErr(e1)
	}
	return
}

func _EvtGetObjectArrayProperty(objectArray EvtObjectArrayPropertyHandle, propertyID EvtPublisherMetadataPropertyID, arrayIndex uint32, flags uint32, bufferSize uint32, evtVariant *EvtVariant, bufferUsed *uint32) (err error) {
	r1, _, e1 := syscall.Syscall9(procEvtGetObjectArrayProperty.Addr(), 7, uintptr(objectArray), uintptr(propertyID), uintptr(arrayIndex), uintptr(flags), uintptr(bufferSize), uintptr(unsafe.Pointer(evtVariant)), uintptr(unsafe.Pointer(bufferUsed)), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}
func _EvtFormatMessage1(publisherMetadata EvtHandle, event EvtHandle, messageID uint32, valueCount uint32, values *EvtVariant, flags EvtFormatMessageFlag, bufferSize uint32, buffer *byte, bufferUsed *uint32) (err error) {
	r1, _, e1 := syscall.Syscall9(procEvtFormatMessage.Addr(), 9, uintptr(publisherMetadata), uintptr(event), uintptr(messageID), uintptr(valueCount), uintptr(unsafe.Pointer(values)), uintptr(flags), uintptr(bufferSize), uintptr(unsafe.Pointer(buffer)), uintptr(unsafe.Pointer(bufferUsed)))
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func _EvtGetObjectArraySize(objectArray EvtObjectArrayPropertyHandle, arraySize *uint32) (err error) {
	r1, _, e1 := syscall.Syscall(procEvtGetObjectArraySize.Addr(), 2, uintptr(objectArray), uintptr(unsafe.Pointer(arraySize)), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func _EvtGetPublisherMetadataProperty(publisherMetadata EvtHandle, propertyID EvtPublisherMetadataPropertyID, flags uint32, bufferSize uint32, variant *EvtVariant, bufferUsed *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procEvtGetPublisherMetadataProperty.Addr(), 6, uintptr(publisherMetadata), uintptr(propertyID), uintptr(flags), uintptr(bufferSize), uintptr(unsafe.Pointer(variant)), uintptr(unsafe.Pointer(bufferUsed)))
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

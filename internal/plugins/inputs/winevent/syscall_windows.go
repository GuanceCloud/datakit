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
	"fmt"
	"syscall"
	"time"
	"unsafe"

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

type EvtEventMetadataPropertyID uint32

const (
	EventMetadataEventID EvtEventMetadataPropertyID = iota
	EventMetadataEventVersion
	EventMetadataEventChannel
	EventMetadataEventLevel
	EventMetadataEventOpcode
	EventMetadataEventTask
	EventMetadataEventKeyword
	EventMetadataEventMessageID
	EventMetadataEventTemplate
)

var evtVariantTypeNames = map[EvtVariantType]string{
	EvtVarTypeNull:       "null",
	EvtVarTypeString:     "string",
	EvtVarTypeAnsiString: "ansi_string",
	EvtVarTypeSByte:      "signed_byte",
	EvtVarTypeByte:       "unsigned byte",
	EvtVarTypeInt16:      "int16",
	EvtVarTypeUInt16:     "uint16",
	EvtVarTypeInt32:      "int32",
	EvtVarTypeUInt32:     "uint32",
	EvtVarTypeInt64:      "int64",
	EvtVarTypeUInt64:     "uint64",
	EvtVarTypeSingle:     "float32",
	EvtVarTypeDouble:     "float64",
	EvtVarTypeBoolean:    "boolean",
	EvtVarTypeBinary:     "binary",
	EvtVarTypeGuid:       "guid",
	EvtVarTypeSizeT:      "size_t",
	EvtVarTypeFileTime:   "filetime",
	EvtVarTypeSysTime:    "systemtime",
	EvtVarTypeSid:        "sid",
	EvtVarTypeHexInt32:   "hex_int32",
	EvtVarTypeHexInt64:   "hex_int64",
	EvtVarTypeEvtHandle:  "evt_handle",
	EvtVarTypeEvtXml:     "evt_xml",
}

type EvtPublisherMetadataPropertyID uint32

const (
	EvtPublisherMetadataPublisherGuid EvtPublisherMetadataPropertyID = iota
	EvtPublisherMetadataResourceFilePath
	EvtPublisherMetadataParameterFilePath
	EvtPublisherMetadataMessageFilePath
	EvtPublisherMetadataHelpLink
	EvtPublisherMetadataPublisherMessageID
	EvtPublisherMetadataChannelReferences
	EvtPublisherMetadataChannelReferencePath
	EvtPublisherMetadataChannelReferenceIndex
	EvtPublisherMetadataChannelReferenceID
	EvtPublisherMetadataChannelReferenceFlags
	EvtPublisherMetadataChannelReferenceMessageID
	EvtPublisherMetadataLevels
	EvtPublisherMetadataLevelName
	EvtPublisherMetadataLevelValue
	EvtPublisherMetadataLevelMessageID
	EvtPublisherMetadataTasks
	EvtPublisherMetadataTaskName
	EvtPublisherMetadataTaskEventGuid
	EvtPublisherMetadataTaskValue
	EvtPublisherMetadataTaskMessageID
	EvtPublisherMetadataOpcodes
	EvtPublisherMetadataOpcodeName
	EvtPublisherMetadataOpcodeValue
	EvtPublisherMetadataOpcodeMessageID
	EvtPublisherMetadataKeywords
	EvtPublisherMetadataKeywordName
	EvtPublisherMetadataKeywordValue
	EvtPublisherMetadataKeywordMessageID
)

type EvtVariantType uint32

const (
	EvtVarTypeNull EvtVariantType = iota
	EvtVarTypeString
	EvtVarTypeAnsiString
	EvtVarTypeSByte
	EvtVarTypeByte
	EvtVarTypeInt16
	EvtVarTypeUInt16
	EvtVarTypeInt32
	EvtVarTypeUInt32
	EvtVarTypeInt64
	EvtVarTypeUInt64
	EvtVarTypeSingle
	EvtVarTypeDouble
	EvtVarTypeBoolean
	EvtVarTypeBinary
	EvtVarTypeGuid
	EvtVarTypeSizeT
	EvtVarTypeFileTime
	EvtVarTypeSysTime
	EvtVarTypeSid
	EvtVarTypeHexInt32
	EvtVarTypeHexInt64
	EvtVarTypeEvtHandle EvtVariantType = 32
	EvtVarTypeEvtXml    EvtVariantType = 35
)

func (t EvtVariantType) Mask() EvtVariantType {
	return t & EvtVariantTypeMask
}

func (t EvtVariantType) IsArray() bool {
	return t&EvtVariantTypeArray > 0
}

func (t EvtVariantType) String() string {
	return evtVariantTypeNames[t.Mask()]
}

const (
	EvtVariantTypeMask  = 0x7f
	EvtVariantTypeArray = 128
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

func GetEventMetadataProperty(metadataHandle EvtHandle, propertyID EvtEventMetadataPropertyID) (interface{}, error) {
	var bufferUsed uint32
	err := _EvtGetEventMetadataProperty(metadataHandle, 8, 0, 0, nil, &bufferUsed)
	if err != windows.ERROR_INSUFFICIENT_BUFFER { //nolint:errorlint // Bad linter! This is always errno or nil.
		return nil, fmt.Errorf("expected ERROR_INSUFFICIENT_BUFFER but got %w (%#v)", err, err)
	}

	buf := make([]byte, bufferUsed)
	pEvtVariant := (*EvtVariant)(unsafe.Pointer(&buf[0]))
	err = _EvtGetEventMetadataProperty(metadataHandle, propertyID, 0, uint32(len(buf)), pEvtVariant, &bufferUsed)
	if err != nil {
		return nil, fmt.Errorf("_EvtGetEventMetadataProperty: %w", err)
	}

	return pEvtVariant.Data(buf)
}

type EvtVariant struct {
	Value [8]byte // This is a union-type in the original struct.
	Count uint32
	Type  EvtVariantType
}

func (v EvtVariant) ValueAsUint64() uint64 {
	return *(*uint64)(unsafe.Pointer(&v.Value))
}

func (v EvtVariant) ValueAsUint32() uint32 {
	return *(*uint32)(unsafe.Pointer(&v.Value))
}

func (v EvtVariant) ValueAsUint16() uint16 {
	return *(*uint16)(unsafe.Pointer(&v.Value))
}

func (v EvtVariant) ValueAsUint8() uint8 {
	return *(*uint8)(unsafe.Pointer(&v.Value))
}

func (v EvtVariant) ValueAsUintPtr() uintptr {
	return *(*uintptr)(unsafe.Pointer(&v.Value))
}

func (v EvtVariant) ValueAsFloat32() float32 {
	return *(*float32)(unsafe.Pointer(&v.Value))
}

func (v EvtVariant) ValueAsFloat64() float64 {
	return *(*float64)(unsafe.Pointer(&v.Value))
}

func (v *EvtVariant) SetValue(val uintptr) {
	*(*uintptr)(unsafe.Pointer(&v.Value)) = val
}

type hexInt32 int32

func (n hexInt32) String() string {
	return fmt.Sprintf("%#x", uint32(n))
}

type hexInt64 int64

func (n hexInt64) String() string {
	return fmt.Sprintf("%#x", uint64(n))
}
func (v EvtVariant) Data(buf []byte) (interface{}, error) {
	typ := v.Type.Mask()
	switch typ {
	case EvtVarTypeNull:
		return nil, nil
	case EvtVarTypeString:
		addr := unsafe.Pointer(&buf[0])
		offset := v.ValueAsUintPtr() - uintptr(addr)
		if s, err := DecodeUTF16(buf[offset:]); err != nil {
			return nil, err
		} else {
			return string(s), nil
		}
	case EvtVarTypeSByte:
		return int8(v.ValueAsUint8()), nil
	case EvtVarTypeByte:
		return v.ValueAsUint8(), nil
	case EvtVarTypeInt16:
		return int16(v.ValueAsUint16()), nil
	case EvtVarTypeInt32:
		return int32(v.ValueAsUint32()), nil
	case EvtVarTypeHexInt32:
		return hexInt32(v.ValueAsUint32()), nil
	case EvtVarTypeInt64:
		return int64(v.ValueAsUint64()), nil
	case EvtVarTypeHexInt64:
		return hexInt64(v.ValueAsUint64()), nil
	case EvtVarTypeUInt16:
		return v.ValueAsUint16(), nil
	case EvtVarTypeUInt32:
		return v.ValueAsUint32(), nil
	case EvtVarTypeUInt64:
		return v.ValueAsUint64(), nil
	case EvtVarTypeSingle:
		return v.ValueAsFloat32(), nil
	case EvtVarTypeDouble:
		return v.ValueAsFloat64(), nil
	case EvtVarTypeBoolean:
		if v.ValueAsUint8() == 0 {
			return false, nil
		}
		return true, nil
	case EvtVarTypeGuid:
		addr := unsafe.Pointer(&buf[0])
		offset := v.ValueAsUintPtr() - uintptr(addr)
		guid := (*windows.GUID)(unsafe.Pointer(&buf[offset]))
		copy := *guid
		return copy, nil
	case EvtVarTypeFileTime:
		ft := (*windows.Filetime)(unsafe.Pointer(&v.Value))
		return time.Unix(0, ft.Nanoseconds()).UTC(), nil
	case EvtVarTypeSid:
		addr := unsafe.Pointer(&buf[0])
		offset := v.ValueAsUintPtr() - uintptr(addr)
		sidPtr := (*windows.SID)(unsafe.Pointer(&buf[offset]))
		return sidPtr.Copy()
	case EvtVarTypeEvtHandle:
		return EvtHandle(v.ValueAsUintPtr()), nil
	default:
		return nil, fmt.Errorf("unhandled type: %d", typ)
	}
}

var sizeofEvtVariant = unsafe.Sizeof(EvtVariant{})

func EvtGetObjectArrayProperty(arrayHandle EvtObjectArrayPropertyHandle, propertyID EvtPublisherMetadataPropertyID, index uint32) (interface{}, error) {
	var bufferUsed uint32
	err := _EvtGetObjectArrayProperty(arrayHandle, propertyID, index, 0, 0, nil, &bufferUsed)
	if err != windows.ERROR_INSUFFICIENT_BUFFER { //nolint:errorlint // Bad linter! This is always errno or nil.
		return nil, fmt.Errorf("failed in EvtGetObjectArrayProperty, expected ERROR_INSUFFICIENT_BUFFER: %w", err)
	}

	buf := make([]byte, bufferUsed)
	pEvtVariant := (*EvtVariant)(unsafe.Pointer(&buf[0]))
	err = _EvtGetObjectArrayProperty(arrayHandle, propertyID, index, 0, uint32(len(buf)), pEvtVariant, &bufferUsed)
	if err != nil {
		return nil, fmt.Errorf("failed in EvtGetObjectArrayProperty: %w", err)
	}

	value, err := pEvtVariant.Data(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read EVT_VARIANT value: %w", err)
	}
	return value, nil
}

type EvtObjectArrayPropertyHandle uint32

func (h EvtObjectArrayPropertyHandle) Close() error {
	return _EvtClose(EvtHandle(h))
}

func EvtGetObjectArraySize(handle EvtObjectArrayPropertyHandle) (uint32, error) {
	var arrayLen uint32
	if err := _EvtGetObjectArraySize(handle, &arrayLen); err != nil {
		return 0, err
	}
	return arrayLen, nil
}

func EvtGetPublisherMetadataProperty(publisherMetadataHandle EvtHandle, propertyID EvtPublisherMetadataPropertyID) (interface{}, error) {
	var bufferUsed uint32
	err := _EvtGetPublisherMetadataProperty(publisherMetadataHandle, propertyID, 0, 0, nil, &bufferUsed)
	if err != windows.ERROR_INSUFFICIENT_BUFFER { //nolint:errorlint // Bad linter! This is always errno or nil.
		return "", fmt.Errorf("expected ERROR_INSUFFICIENT_BUFFER but got %w (%#v)", err, err)
	}

	buf := make([]byte, bufferUsed)
	pEvtVariant := (*EvtVariant)(unsafe.Pointer(&buf[0]))
	err = _EvtGetPublisherMetadataProperty(publisherMetadataHandle, propertyID, 0, uint32(len(buf)), pEvtVariant, &bufferUsed)
	if err != nil {
		return nil, fmt.Errorf("failed in EvtGetPublisherMetadataProperty: %w", err)
	}

	v, err := pEvtVariant.Data(buf)
	if err != nil {
		return nil, err
	}

	switch t := v.(type) {
	case EvtHandle:
		return EvtObjectArrayPropertyHandle(t), nil
	default:
		return v, nil
	}
}

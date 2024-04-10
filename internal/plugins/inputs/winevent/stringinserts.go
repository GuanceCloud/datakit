// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Beats (https://github.com/elastic/beats).

//go:build windows
// +build windows

package winevent

import (
	"strconv"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	// maxInsertStrings is the maximum number of parameters supported in a
	// Windows event message.
	maxInsertStrings = 99

	leftTemplateDelim  = "[{{"
	rightTemplateDelim = "}}]"
)

// templateInserts contains EvtVariant values that can be used to substitute
// Go text/template expressions into a Windows event message.
var templateInserts = newTemplateStringInserts()

// stringsInserts holds EvtVariant values with type EvtVarTypeString.
type stringInserts struct {
	// insertStrings are slices holding the strings in the EvtVariant (this must
	// keep a reference to these to prevent GC of the strings as there is
	// an unsafe reference to them in the evtVariants).
	insertStrings [maxInsertStrings][]uint16
	evtVariants   [maxInsertStrings]EvtVariant
}

// Slice returns a slice of the full EvtVariant array.
func (si *stringInserts) Slice() []EvtVariant {
	return si.evtVariants[:]
}

// clear clears the pointers (and unsafe pointers) so that the memory can be
// garbage collected.
func (si *stringInserts) clear() {
	for i := 0; i < len(si.evtVariants); i++ {
		si.evtVariants[i] = EvtVariant{}
		si.insertStrings[i] = nil
	}
}

// newTemplateStringInserts returns a stringInserts where each value is a
// Go text/template expression that references an event data parameter.
func newTemplateStringInserts() *stringInserts {
	si := &stringInserts{}

	for i := 0; i < len(si.evtVariants); i++ {
		// Use i+1 to keep our inserts numbered the same as Window's inserts.
		templateParam := leftTemplateDelim + `eventParam $ ` + strconv.Itoa(i+1) + rightTemplateDelim
		strSlice, err := windows.UTF16FromString(templateParam)
		if err != nil {
			// This will never happen.
			panic(err)
		}

		si.insertStrings[i] = strSlice
		si.evtVariants[i] = EvtVariant{
			Count: uint32(len(strSlice)),
			Type:  EvtVarTypeString,
		}
		si.evtVariants[i].SetValue(uintptr(unsafe.Pointer(&strSlice[0])))
		si.evtVariants[i].Type = EvtVarTypeString
	}

	return si
}

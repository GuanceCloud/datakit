// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package quantity defines the unit and quantity kinds.
package quantity

var (
	UnknownKind = &Kind{}
	Count       = &Kind{}
	Duration    = &Kind{}
	Memory      = &Kind{}
)

func init() {
	// avoid loop reference
	UnknownKind.DefaultUnit = UnknownUnit
	Count.DefaultUnit = CountUnit
	Duration.DefaultUnit = MicroSecond
	Memory.DefaultUnit = Byte
}

// Kind 度量类型，数量，时间，内存...
type Kind struct {
	DefaultUnit *Unit
}

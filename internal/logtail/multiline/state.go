// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package multiline

type State uint8

const (
	NewMultiline State = iota + 1
	Written
	NoContext
	OverTime
	OverLength
)

func (state State) String() string {
	switch state {
	case NewMultiline:
		return "new-multiline"
	case Written:
		return "written"
	case NoContext:
		return "no-context"
	case OverTime:
		return "overtime"
	case OverLength:
		return "overlength"
	}
	return "unknown"
}

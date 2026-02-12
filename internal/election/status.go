// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

type ElectionStatus int

func (e ElectionStatus) String() string {
	switch e {
	case StatusDisabled:
		return "disabled"
	case StatusSuccess:
		return "success"
	case StatusFail:
		return "defeat"
	case StatusBanned:
		return "banned"
	case StatusImpeached:
		return "impeached"
	default:
		return "unknown" // should not been here
	}
}

const (
	StatusDisabled ElectionStatus = iota
	StatusSuccess
	StatusFail
	StatusBanned
	StatusImpeached
)

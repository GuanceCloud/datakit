// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

type electionStatus int

func (e electionStatus) String() string {
	switch e {
	case statusDisabled:
		return "disabled"
	case statusSuccess:
		return "success"
	case statusFail:
		return "defeat"
	default:
		return "unknown" // should not been here
	}
}

const (
	statusDisabled electionStatus = iota
	statusSuccess
	statusFail
)

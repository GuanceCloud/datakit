// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package parsing

import (
	"github.com/GuanceCloud/cliutils/pprofparser/domain/events"
)

type DisplayCtl interface {
	ShowInTrace(e events.Type) bool
	ShowInProfile(e events.Type) bool
}

type DDTrace struct{}

func (D DDTrace) ShowInTrace(e events.Type) bool {
	return e.GetShowPlaces()&events.ShowInTrace > 0
}

func (D DDTrace) ShowInProfile(e events.Type) bool {
	return e.GetShowPlaces()&events.ShowInProfile > 0
}

type PyroscopeNodejs struct{}

func (p *PyroscopeNodejs) ShowInTrace(_ events.Type) bool {
	return false
}
func (p *PyroscopeNodejs) ShowInProfile(_ events.Type) bool {
	return true
}

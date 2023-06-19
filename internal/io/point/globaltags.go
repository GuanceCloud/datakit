// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

var (
	globalHostTags     = map[string]string{}
	globalElectionTags = map[string]string{}
)

type GlobalTagger interface {
	HostTags() map[string]string
	ElectionTags() map[string]string
}

func DefaultGlobalTagger() GlobalTagger {
	return &globalTaggerImpl{}
}

type globalTaggerImpl struct{}

func (g *globalTaggerImpl) HostTags() map[string]string {
	return GlobalHostTags()
}

func (g *globalTaggerImpl) ElectionTags() map[string]string {
	return GlobalElectionTags()
}

////////////////////////////////////////////////////////////////////////////////

func SetGlobalHostTags(k, v string) {
	globalHostTags[k] = v
}

func SetGlobalHostTagsByMap(in map[string]string) {
	for k, v := range in {
		globalHostTags[k] = v
	}
}

////////////////////////////////////////////////////////////////////////////////

func SetGlobalElectionTags(k, v string) {
	globalElectionTags[k] = v
}

func SetGlobalElectionTagsByMap(in map[string]string) {
	for k, v := range in {
		globalElectionTags[k] = v
	}
}

////////////////////////////////////////////////////////////////////////////////

func GlobalHostTags() map[string]string {
	return internal.CopyMapString(globalHostTags)
}

func GlobalElectionTags() map[string]string {
	return internal.CopyMapString(globalElectionTags)
}

////////////////////////////////////////////////////////////////////////////////

func ClearGlobalTags() {
	ClearGlobalHostTags()
	ClearGlobalElectionTags()
}

func ClearGlobalHostTags() {
	globalHostTags = map[string]string{}
}

func ClearGlobalElectionTags() {
	globalElectionTags = map[string]string{}
}

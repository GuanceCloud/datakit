// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package datakit

import (
	"sync"
	"sync/atomic"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

var (
	globalHostTags     = map[string]string{}
	globalElectionTags = map[string]string{}
	tagVersion         atomic.Int64
	rw                 sync.RWMutex
)

type GlobalTagger interface {
	HostTags() map[string]string
	ElectionTags() map[string]string
	Updated() bool
	UpdateVersion()
}

// DynamicGlobalTagger get a tagger that will return tags and
// these tags may refresh dynamically.
func DynamicGlobalTagger() GlobalTagger {
	return &dynamicGlobalTaggerImpl{
		ver: &atomic.Int64{},
	}
}

type dynamicGlobalTaggerImpl struct {
	ver *atomic.Int64
}

func (g *dynamicGlobalTaggerImpl) HostTags() map[string]string     { return GlobalHostTags() }
func (g *dynamicGlobalTaggerImpl) ElectionTags() map[string]string { return GlobalElectionTags() }

func (g *dynamicGlobalTaggerImpl) Updated() bool {
	return g.ver.Load() != tagVersion.Load()
}

// UpdateVersion set tagger's current version.
func (g *dynamicGlobalTaggerImpl) UpdateVersion() {
	g.ver.Store(tagVersion.Load())
}

// DefaultGlobalTagger get a tagger that always return nothing.
func DefaultGlobalTagger() GlobalTagger {
	return &emptyGlobalTager{}
}

type emptyGlobalTager struct{}

func (g *emptyGlobalTager) HostTags() map[string]string     { return nil }
func (g *emptyGlobalTager) ElectionTags() map[string]string { return nil }
func (g *emptyGlobalTager) UpdateVersion()                  {}

// Updated should never update for emptyGlobalTager.
func (g *emptyGlobalTager) Updated() bool { return false }

func SetGlobalHostTags(k, v string) {
	rw.Lock()
	defer rw.Unlock()
	globalHostTags[k] = v
	tagVersion.Add(1)
}

func SetGlobalHostTagsByMap(in map[string]string) {
	rw.Lock()
	defer rw.Unlock()
	for k, v := range in {
		globalHostTags[k] = v
	}
	tagVersion.Add(1)
}

func SetGlobalElectionTags(k, v string) {
	rw.Lock()
	defer rw.Unlock()
	globalElectionTags[k] = v
	tagVersion.Add(1)
}

func SetGlobalElectionTagsByMap(in map[string]string) {
	rw.Lock()
	defer rw.Unlock()
	for k, v := range in {
		globalElectionTags[k] = v
	}
	tagVersion.Add(1)
}

func GlobalHostTags() map[string]string {
	rw.RLock()
	defer rw.RUnlock()
	return internal.CopyMapString(globalHostTags)
}

func GlobalElectionTags() map[string]string {
	rw.RLock()
	defer rw.RUnlock()
	return internal.CopyMapString(globalElectionTags)
}

func ClearGlobalTags() {
	ClearGlobalHostTags()
	ClearGlobalElectionTags()
	tagVersion.Add(1)
}

func ClearGlobalHostTags() {
	rw.Lock()
	defer rw.Unlock()
	globalHostTags = map[string]string{}
	tagVersion.Add(1)
}

func ClearGlobalElectionTags() {
	rw.Lock()
	defer rw.Unlock()
	globalElectionTags = map[string]string{}
	tagVersion.Add(1)
}

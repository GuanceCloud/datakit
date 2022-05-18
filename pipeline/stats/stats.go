// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package stats used to record pl metrics
package stats

import (
	"fmt"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var l = logger.DefaultSLogger("pl-stats")

var _changeEvent = ScriptChangeEvent{}

const MaxEventLen int = 100

func InitStats() {
	l = logger.SLogger("pl-stats")
}

type ScriptChangeEvent struct {
	last100event [MaxEventLen]*ChangeEvent
	pos          int
	sync.RWMutex
}

func (event *ScriptChangeEvent) Write(change *ChangeEvent) {
	if change == nil {
		return
	}
	l.Info(change)
	event.Lock()
	defer event.Unlock()
	if event.pos >= MaxEventLen {
		event.pos %= MaxEventLen
	}
	event.last100event[event.pos] = change
	event.pos += 1
}

func (event *ScriptChangeEvent) Read() []ChangeEvent {
	ret := []ChangeEvent{}
	event.RLock()
	defer event.RUnlock()
	curPos := event.pos
	if curPos >= MaxEventLen {
		curPos %= MaxEventLen
	}

	if event.last100event[curPos] == nil {
		for i := 0; i < curPos; i++ {
			ret = append(ret, *event.last100event[i])
		}
		return ret
	}

	for i := curPos; i < MaxEventLen; i++ {
		ret = append(ret, *event.last100event[i])
	}

	for i := 0; i < curPos; i++ {
		ret = append(ret, *event.last100event[i])
	}

	return ret
}

type Op string

const (
	OpAdd                Op = "ADD"
	OpUpdate             Op = "UPDATE"
	OpDelete             Op = "DELETE"
	OpIndex              Op = "INDEX"
	OpIndexUpdate        Op = "INDEX_UPDATE"
	OpIndexDelete        Op = "INDEX_DELETE"
	OpIndexDeleteAndBack Op = "INDEX_DELETE_AND_BACK"

	OpCompileError Op = "COMPILE_ERROR"
)

type ChangeEvent struct {
	Name              string
	Category          string
	NS, NSOld         string
	Script, ScriptOld string

	Op Op //

	CompileError error
	Time         time.Time
}

func (event ChangeEvent) String() string {
	ns := event.NS
	if event.NSOld != "" && event.NS != event.NSOld {
		ns = event.NSOld + "->" + ns
	}
	ret := fmt.Sprintf("ScriptStore %s %s category: %s, ns: %s, script_name: %s",
		event.Time.Format(time.RFC3339Nano), event.Op, event.Category, ns, event.Name)
	if event.CompileError != nil {
		ret += ", compile_error: " + event.CompileError.Error()
	}
	return ret
}

func WriteEvent(event *ChangeEvent) {
	_changeEvent.Write(event)
}

func ReadEvent() []ChangeEvent {
	return _changeEvent.Read()
}

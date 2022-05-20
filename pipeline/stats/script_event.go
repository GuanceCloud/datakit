package stats

import (
	"fmt"
	"sync"
	"time"
)

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

type ChangeEvent struct {
	Name              string
	Category          string
	NS, NSOld         string
	Script, ScriptOld string

	Op EventOP //

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

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/grok"
)

type GrokRunInfo struct {
	ScriptName     string
	Line           int
	Column         int
	PatternHash    uint64
	Path           string
	FallbackReason string
	WorkUnits      int
	Cost           time.Duration
}

type GrokRunObserver func(info GrokRunInfo)

var grokRunObserver atomic.Value

type grokRunObserverHolder struct {
	observer GrokRunObserver
}

func SetGrokRunObserver(observer GrokRunObserver) {
	grokRunObserver.Store(grokRunObserverHolder{observer: observer})
}

func currentGrokRunObserver() GrokRunObserver {
	v := grokRunObserver.Load()
	if v == nil {
		return nil
	}
	return v.(grokRunObserverHolder).observer
}

func observeGrokRun(observer GrokRunObserver, scriptName string, line, column int, meta grok.RunMeta, start time.Time) {
	if observer == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			l.Errorf("grok run observer panic: %v\n%s", r, debug.Stack())
		}
	}()

	observer(GrokRunInfo{
		ScriptName:     scriptName,
		Line:           line,
		Column:         column,
		PatternHash:    meta.PatternHash,
		Path:           meta.Path.String(),
		FallbackReason: meta.FallbackReason.String(),
		WorkUnits:      meta.WorkUnits,
		Cost:           time.Since(start),
	})
}

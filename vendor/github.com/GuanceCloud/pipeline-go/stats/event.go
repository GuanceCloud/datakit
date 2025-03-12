// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package stats

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

type RecEvent struct {
	labelNames []string

	size int
	ch   chan *ChangeEvent
}

func newRecEvent(size int, labelNames []string) *RecEvent {
	if size <= 0 {
		size = 256
	}
	return &RecEvent{
		labelNames: labelNames,

		size: size,
		ch:   make(chan *ChangeEvent, size),
	}
}

func (event *RecEvent) Write(change *ChangeEvent, tags map[string]string) {
	if change == nil {
		return
	}

	if len(event.labelNames) != 0 {
		if change.extraTags == nil {
			change.extraTags = make([][2]string, 0, len(event.labelNames))
		}

		for _, name := range event.labelNames {
			switch name {
			case "ns", "category", "name":
			default:
				if v, ok := tags[name]; ok {
					change.extraTags = append(change.extraTags, [2]string{name, v})
				}
			}
		}
	}

	l.Info(change)

	for {
		select {
		case event.ch <- change:
			return
		default:
			<-event.ch
		}
	}
}

func (event *RecEvent) Read(events []*ChangeEvent) []*ChangeEvent {
	if cap(events) == 0 {
		events = make([]*ChangeEvent, 0, 8)
	}

	if len(events) == cap(events) {
		return events
	}

	for {
		select {
		case e := <-event.ch:
			events = append(events, e)
			if len(events) > cap(events) || len(events) >= event.size {
				return events
			}
		default:
			return events
		}
	}
}

func (event *RecEvent) ReadChan() <-chan *ChangeEvent {
	return event.ch
}

type ChangeEvent struct {
	Name              string
	Category          point.Category
	NS, NSOld         string
	Script, ScriptOld string

	Op EventOP //

	extraTags [][2]string

	CompileError string
	Time         time.Time
}

func (event ChangeEvent) String() string {
	ns := event.NS
	if event.NSOld != "" && event.NS != event.NSOld {
		ns = event.NSOld + "->" + ns
	}
	ret := fmt.Sprintf("ScriptStore %s %s category: %s, ns: %s, script_name: %s, extraTags: %v",
		event.Time.Format(StatsTimeFormat), event.Op, event.Category, ns, event.Name, event.extraTags)

	if event.CompileError != "" {
		ret += ", compile_error: " + event.CompileError
	}
	return ret
}

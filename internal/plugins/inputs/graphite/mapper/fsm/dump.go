// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package fsm dumps fsm
package fsm

import (
	"fmt"
	"io"
)

// DumpFSM accepts a io.writer and write the current FSM into dot file format.
func (f *FSM) DumpFSM(w io.Writer) {
	idx := 0
	states := make(map[int]*mappingState)
	states[idx] = f.root

	if _, err := w.Write([]byte("digraph g {\n")); err != nil {
		return
	}
	if _, err := w.Write([]byte("rankdir=LR\n")); err != nil {
		return
	} // make it vertical
	if _, err := w.Write([]byte("node [ label=\"\",style=filled,fillcolor=white,shape=circle ]\n")); err != nil {
		return
	} // remove label of node

	for idx < len(states) {
		for field, transition := range states[idx].transitions {
			states[len(states)] = transition
			if _, err := w.Write([]byte(fmt.Sprintf("%d -> %d  [label = \"%s\"];\n", idx, len(states)-1, field))); err != nil {
				return
			}
			if idx == 0 {
				// color for metric types
				if _, err := w.Write([]byte(fmt.Sprintf("%d [color=\"#D6B656\",fillcolor=\"#FFF2CC\"];\n", len(states)-1))); err != nil {
					return
				}
			} else if transition.transitions == nil || len(transition.transitions) == 0 {
				// color for end state
				if _, err := w.Write([]byte(fmt.Sprintf("%d [color=\"#82B366\",fillcolor=\"#D5E8D4\"];\n", len(states)-1))); err != nil {
					return
				}
			}
		}
		idx++
	}
	// color for start state
	if _, err := w.Write([]byte(fmt.Sprintln("0 [color=\"#a94442\",fillcolor=\"#f2dede\"];"))); err != nil {
		return
	}
	if _, err := w.Write([]byte("}")); err != nil {
		return
	}
}

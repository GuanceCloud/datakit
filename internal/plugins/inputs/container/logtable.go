// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type logTable struct {
	table map[string]map[string]chan interface{}
	mu    sync.Mutex
}

func newLogTable() *logTable {
	return &logTable{
		table: make(map[string]map[string]chan interface{}),
	}
}

func (t *logTable) addToTable(id, path string, done chan interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.table[id] == nil {
		t.table[id] = make(map[string]chan interface{})
	}
	t.table[id][path] = done
}

func (t *logTable) closeFromTable(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, done := range t.table[id] {
		if !IsClosed(done) {
			close(done)
		}
	}
}

func (t *logTable) removeFromTable(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.table, id)
}

func (t *logTable) removePathFromTable(id, path string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	paths, ok := t.table[id]
	if !ok {
		return
	}
	if paths != nil {
		delete(paths, path)
	}
}

func (t *logTable) inTable(id, path string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.table[id] == nil {
		return false
	}
	_, ok := t.table[id][path]
	return ok
}

func (t *logTable) findDifferences(ids []string) []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	var differences []string
	for id := range t.table {
		found := false
		for _, k := range ids {
			if k == id {
				found = true
				break
			}
		}
		if !found {
			differences = append(differences, id)
		}
	}
	return differences
}

func (t *logTable) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	var str []string
	var ids []string

	for id := range t.table {
		ids = append(ids, id)
	}

	sort.Strings(ids)

	for _, id := range ids {
		paths, ok := t.table[id]
		if !ok {
			continue
		}

		var p []string

		for path := range paths {
			p = append(p, path)
		}
		sort.Strings(p)

		shortID := id
		if len(id) > 12 {
			shortID = id[:12]
		}

		str = append(str, fmt.Sprintf("{id:%s,paths:[%s]}", shortID, strings.Join(p, ",")))
	}

	return strings.Join(str, ", ")
}

func IsClosed(ch <-chan interface{}) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

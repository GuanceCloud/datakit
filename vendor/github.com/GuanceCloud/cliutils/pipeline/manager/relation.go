// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package manager for managing pipeline scripts
package manager

import (
	"sync"

	"github.com/GuanceCloud/cliutils/point"
)

// var remoteRelation = &PipelineRelation{}

type ScriptRelation struct {
	// map[<category>]: map[<source>]: <name>
	relation map[point.Category]map[string]string

	updateAt int64

	rwMutex sync.RWMutex
}

func NewPipelineRelation() *ScriptRelation {
	return &ScriptRelation{}
}

func (relation *ScriptRelation) UpdateAt() int64 {
	relation.rwMutex.RLock()
	defer relation.rwMutex.RUnlock()

	return relation.updateAt
}

func (relation *ScriptRelation) UpdateRelation(updateAt int64, rel map[point.Category]map[string]string) {
	relation.rwMutex.Lock()
	defer relation.rwMutex.Unlock()

	relation.updateAt = updateAt

	// reset relation
	relation.relation = map[point.Category]map[string]string{}

	for cat, relat := range rel {
		m := map[string]string{}
		relation.relation[cat] = m
		for source, name := range relat {
			m[source] = name
		}
	}
}

func (relation *ScriptRelation) Query(cat point.Category, source string) (string, bool) {
	relation.rwMutex.RLock()
	defer relation.rwMutex.RUnlock()

	if v, ok := relation.relation[cat]; ok {
		if name, ok := v[source]; ok {
			return name, true
		}
	}

	return "", false
}

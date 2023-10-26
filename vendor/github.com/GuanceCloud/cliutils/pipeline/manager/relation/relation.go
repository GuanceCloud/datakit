// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package relation record source-name relation
package relation

import (
	"sync"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
)

var l = logger.DefaultSLogger("pl-relation")

func InitLog() {
	l = logger.SLogger("pl-relation")
}

// var remoteRelation = &PipelineRelation{}

type ScriptRelation struct {
	// map[<category>]: map[<source>]: <name>
	relation map[point.Category]map[string]string

	// map[<category>]: <name>
	defaultScript map[point.Category]string

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

func (relation *ScriptRelation) UpdateDefaultPl(defaultPl map[string]string) {
	relation.rwMutex.Lock()
	defer relation.rwMutex.Unlock()

	if len(relation.defaultScript) > 0 || relation.defaultScript == nil {
		relation.defaultScript = map[point.Category]string{}
	}

	for categoryShort, name := range defaultPl {
		category := point.CatString(categoryShort)
		if category == point.UnknownCategory {
			l.Warnf("unknown category: %s", categoryShort)
		}
		relation.defaultScript[category] = name
	}
}

func (relation *ScriptRelation) UpdateRelation(updateAt int64, rel map[string]map[string]string) {
	relation.rwMutex.Lock()
	defer relation.rwMutex.Unlock()

	relation.updateAt = updateAt
	if len(relation.relation) > 0 || relation.relation == nil {
		relation.relation = map[point.Category]map[string]string{}
	}

	for categoryShort, relat := range rel {
		category := point.CatString(categoryShort)
		if category == point.UnknownCategory {
			l.Warnf("unknown category: %s", categoryShort)
		}
		for source, name := range relat {
			if v, ok := relation.relation[category]; !ok {
				relation.relation[category] = map[string]string{
					source: name,
				}
			} else {
				v[source] = name
			}
		}
	}
}

func (relation *ScriptRelation) Query(category point.Category, source string) (string, bool) {
	relation.rwMutex.RLock()
	defer relation.rwMutex.RUnlock()

	if v, ok := relation.relation[category]; ok {
		if name, ok := v[source]; ok {
			return name, true
		}
	}

	// defaultPl
	if v, ok := relation.defaultScript[category]; ok {
		return v, true
	}

	return "", false
}

func ScriptName(relation *ScriptRelation, cat point.Category, pt *point.Point, scriptMap map[string]string) (string, bool) {
	if pt == nil {
		return "", false
	}

	var scriptName string

	// built-in rules last
	switch cat { //nolint:exhaustive
	case point.RUM:
		scriptName = _rumSName(pt)
	case point.Security:
		scriptName = _securitySName(pt)
	case point.Tracing, point.Profiling:
		scriptName = _apmSName(pt)
	default:
		scriptName = _defaultCatSName(pt)
	}

	if scriptName == "" {
		return "", false
	}

	// configuration first
	if sName, ok := scriptMap[scriptName]; ok {
		switch sName {
		case "-":
			return "", false
		case "":
		default:
			return sName, true
		}
	}

	if relation != nil {
		// remote relation sencond
		if sName, ok := relation.Query(cat, scriptName); ok {
			return sName, true
		}
	}

	return scriptName + ".p", true
}

// func QueryRemoteRelation(category point.Category, source string) (string, bool) {
// 	return remoteRelation.query(category, source)
// }

// func RelationRemoteUpdateAt() int64 {
// 	return remoteRelation.UpdateAt()
// }

// func UpdateRemoteDefaultPl(defaultPl map[string]string) {
// 	remoteRelation.UpdateDefaultPl(defaultPl)
// }

// func UpdateRemoteRelation(updateAt int64, rel map[string]map[string]string) {
// 	remoteRelation.UpdateRelation(updateAt, rel)
// }

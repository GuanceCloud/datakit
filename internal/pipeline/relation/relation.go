// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package relation record source-name relation
package relation

import (
	"sync"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/convertutil"
)

var l = logger.DefaultSLogger("pl-relation")

func InitRelationLog() {
	l = logger.SLogger("pl-relation")
}

var remoteRelation = &PipelineRelation{}

type PipelineRelation struct {
	// map[<category>]: map[<source>]: <name>
	relation map[string]map[string]string

	// map[<category>]: <name>
	defaultScript map[string]string

	updateAt int64

	rwMutex sync.RWMutex
}

func (relation *PipelineRelation) UpdateAt() int64 {
	relation.rwMutex.RLock()
	defer relation.rwMutex.RUnlock()

	return relation.updateAt
}

func (relation *PipelineRelation) UpdateDefaultPl(defaultPl map[string]string) {
	relation.rwMutex.Lock()
	defer relation.rwMutex.Unlock()

	if len(relation.defaultScript) > 0 || relation.defaultScript == nil {
		relation.defaultScript = map[string]string{}
	}

	for categoryShort, name := range defaultPl {
		category, err := convertutil.GetMapCategoryShortToFull(categoryShort)
		if err != nil {
			l.Warnf("GetMapCategoryShortToFull failed: err = %s, categoryShort = %s", err, categoryShort)
			continue
		}
		relation.defaultScript[category] = name
	}
}

func (relation *PipelineRelation) UpdateRelation(updateAt int64, rel map[string]map[string]string) {
	relation.rwMutex.Lock()
	defer relation.rwMutex.Unlock()

	relation.updateAt = updateAt
	if len(relation.relation) > 0 || relation.relation == nil {
		relation.relation = map[string]map[string]string{}
	}

	for categoryShort, relat := range rel {
		category, err := convertutil.GetMapCategoryShortToFull(categoryShort)
		if err != nil {
			l.Warnf("GetMapCategoryShortToFull failed: err = %s, categoryShort = %s", err, categoryShort)
			continue
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

func (relation *PipelineRelation) query(category, source string) (string, bool) {
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

func QueryRemoteRelation(category, source string) (string, bool) {
	return remoteRelation.query(category, source)
}

func RelationRemoteUpdateAt() int64 {
	return remoteRelation.UpdateAt()
}

func UpdateRemoteDefaultPl(defaultPl map[string]string) {
	remoteRelation.UpdateDefaultPl(defaultPl)
}

func UpdateRemoteRelation(updateAt int64, rel map[string]map[string]string) {
	remoteRelation.UpdateRelation(updateAt, rel)
}

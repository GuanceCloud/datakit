// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package neo4j

import (
	"fmt"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
)

// formatPoint modify the point who have suffix and suffix.
// A point can have multiple fields, some of which may match.
//
// If a metric name not be "neo4j":
// - rename metric name to "neo4j".
// - old metric name put to the prefix of field name.
//
// like: vm_memory_pool_metaspace gauge -> meo4j_vm_memory_pool_metaspace gauge
//
// If a field(KV) matches:
// - find the key words.
// - Delete this field(KV).
// - Add a New field(KV), but the K needs to be cut off key words.
// - The key words will be a new tag's value.
//
// like: database_system_pool_transaction_system_used_heap -> database_pool_used_heap
// tags: db->system, pool->transaction, database->system
func formatPoint(pt *point.Point) error {
	willRenameFields := map[string]string{}
	willAddTags := map[string]string{}

	// Step 1: rename metric name.
	if ptName := pt.Name(); ptName != inputName {
		newMetricName := inputName
		for _, kv := range pt.Fields() {
			key := ptName + "_" + kv.Key
			willRenameFields[kv.Key] = key
		}

		rebuildPoint(pt, newMetricName, willRenameFields, willAddTags)

		willRenameFields = map[string]string{}
	}

	// Step 2: need add tag, need rename field(KV).
	for _, kv := range pt.Fields() {
		key := kv.Key

		tempField, tempTags, err := formatFieldName(key)
		if err != nil {
			return err
		}

		for k, v := range tempField {
			willRenameFields[k] = v
		}
		for k, v := range tempTags {
			willAddTags[k] = v
		}
	}

	if len(willAddTags) == 0 {
		return nil
	}

	// Step 3: do add tag, do rename field(KV).
	rebuildPoint(pt, "", willRenameFields, willAddTags)

	return nil
}

// formatFieldName convert field name -> new name + tags
//
//   - 1. no additional tags
//     name: "dbms_vm_pause_time_total", prefix: "dbms_vm_pause_time_total",
//     dbms_vm_pause_time_total -> dbms_vm_pause_time_total
//
//   - 2. one additional tag have "_", with "total" suffix
//     name: "dbms_vm_gc_time_total", prefix: "dbms_vm_gc_time_", tag: []string{"gc"}, suffix: "total",
//     dbms_vm_gc_time_g1_old_generation_total->dbms_vm_gc_time_total. tags: gc->g1_old_generation
//
//   - 3. one additional tag have "_", without suffix
//     name: "dbms_vm_memory_pool", prefix: "dbms_vm_memory_pool_", tag: []string{"pool"}, suffix: "",
//     dbms_vm_memory_pool_g1_eden_space->dbms_vm_memory_pool. tags: pool->g1_eden_space
//
//   - 4. with some additional tags, with any suffix
//     name: "database_pool_used_heap", prefix: "database_", tag: []string{"db", "", "pool", "database"}, suffix: "_used_heap",
//     database_system_pool_transaction_system_used_heap -> database_pool_used_heap. tags: db->system, pool->transaction, database->system
func formatFieldName(name string) (map[string]string, map[string]string, error) {
	willRenameFields := map[string]string{}
	willAddTags := map[string]string{}

	findIdx := 0
	for findIdx = 0; findIdx < len(fieldNames); findIdx++ {
		switch {
		// 1. no additional tags
		case name == fieldNames[findIdx].prefix:
			return willRenameFields, willAddTags, nil

		// 2. one additional tag have "_", with "total" suffix
		case strings.HasPrefix(name, fieldNames[findIdx].prefix) &&
			fieldNames[findIdx].suffix == "total" &&
			len(fieldNames[findIdx].tag) == 1 &&
			strings.HasSuffix(name, fieldNames[findIdx].suffix):

			temp := strings.TrimPrefix(name, fieldNames[findIdx].prefix)
			temp = strings.TrimSuffix(temp, "_total")
			if temp == "" {
				continue
			}
			willRenameFields[name] = fieldNames[findIdx].name
			willAddTags[fieldNames[findIdx].tag[0]] = temp
			return willRenameFields, willAddTags, nil

		// 3. one additional tag have "_", without suffix
		case strings.HasPrefix(name, fieldNames[findIdx].prefix) &&
			fieldNames[findIdx].suffix == "" &&
			len(fieldNames[findIdx].tag) == 1:

			temp := strings.TrimPrefix(name, fieldNames[findIdx].prefix)
			if temp == "" {
				continue
			}
			willRenameFields[name] = fieldNames[findIdx].name
			willAddTags[fieldNames[findIdx].tag[0]] = temp
			return willRenameFields, willAddTags, nil

		// 4. with some additional tags, with any suffix
		case strings.HasPrefix(name, fieldNames[findIdx].prefix) &&
			strings.HasSuffix(name, fieldNames[findIdx].suffix):

			modifyName(willAddTags, willRenameFields, name, findIdx)
			if len(willAddTags) == 0 {
				continue
			}
			return willRenameFields, willAddTags, nil
		}
	}
	if findIdx >= len(fieldNames) {
		return willRenameFields, willAddTags, fmt.Errorf("not real error, additional field name: %s", name)
	}

	return willRenameFields, willAddTags, nil
}

func modifyName(willAddTags, willRenameFields map[string]string, key string, findIdx int) {
	if strings.HasPrefix(key, fieldNames[findIdx].prefix) && strings.HasSuffix(key, fieldNames[findIdx].suffix) {
		if len(fieldNames[findIdx].tag) == 0 {
			return
		}

		temp := strings.TrimPrefix(key, fieldNames[findIdx].prefix)
		temp = strings.TrimSuffix(temp, fieldNames[findIdx].suffix)
		keys := strings.Split(temp, "_")
		if len(keys) != len(fieldNames[findIdx].tag) {
			return
		}

		willRenameFields[key] = fieldNames[findIdx].name
		for j := 0; j < len(keys); j++ {
			if fieldNames[findIdx].tag[j] != "" {
				willAddTags[fieldNames[findIdx].tag[j]] = keys[j]
			}
		}
	}
}

// rebuildPoint rename (add/delete) metric name, add tag, rename (add/delete) field(KV).
func rebuildPoint(pt *point.Point, newMetricName string, willRenameFields, willAddTags map[string]string) {
	if newMetricName != "" {
		kvs := pt.KVs()
		ts := pt.Time()
		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ts))
		*pt = *point.NewPointV2(newMetricName, kvs, opts...)
	}

	for k, v := range willAddTags {
		pt.MustAddTag(k, v)
	}
	for oldKey, newKey := range willRenameFields {
		// Add field(KV).
		kv := &point.Field{
			Key: newKey,
			Val: pt.Fields().Get(oldKey).Val,
		}
		pt.MustAddKVs(kv)

		// Delete old field(KV).
		pt.Del(oldKey)
	}
}

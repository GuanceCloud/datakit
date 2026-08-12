// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package jsonfields converts a JSON log object into logging Point fields.
package jsonfields

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
)

var reservedKeys = [...]string{"time", "source", "date", "storage_index"}

// These limits mirror the cliutils Point defaults. File tailers disable Point
// precheck, so JSON fields must enforce them before the Point is created.
const (
	maxPointFields      = 1024
	maxPointFieldKeyLen = 256
)

// Conversion is the result of converting one log body.
type Conversion struct {
	Fields    map[string]any
	Converted bool
}

// Apply overlays converted JSON fields on existing Point KVs. Set replaces an
// existing tag or field with the JSON field, so each key remains unique. Apply
// must be the final KVs mutation before creating the Point because it enforces
// the Point field-count limit.
func (c Conversion) Apply(kvs point.KVs) point.KVs {
	collectorFields := make(map[string]struct{}, kvs.FieldCount())
	for _, kv := range kvs {
		if !kv.IsTag {
			collectorFields[kv.Key] = struct{}{}
		}
	}

	keys := make([]string, 0, len(c.Fields))
	for key := range c.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := c.Fields[key]
		kvs = kvs.Set(key, value)
	}

	// Collector fields keep their existing priority. A JSON field that replaces
	// a tag joins new JSON fields in lexical order because the tag consumed no
	// field capacity.
	ordered := make(point.KVs, 0, len(kvs))
	jsonFields := make(map[string]*point.Field, len(c.Fields))
	for _, kv := range kvs {
		if kv.IsTag {
			ordered = append(ordered, kv)
			continue
		}
		if _, ok := collectorFields[kv.Key]; ok {
			ordered = append(ordered, kv)
			continue
		}
		jsonFields[kv.Key] = kv
	}
	for _, key := range keys {
		if kv, ok := jsonFields[key]; ok {
			ordered = append(ordered, kv)
		}
	}
	return ordered.TrimFields(maxPointFields)
}

// Convert converts a JSON root object into Point fields. Invalid JSON, a
// non-object root, or an object without usable fields returns Converted=false.
func Convert(data []byte) Conversion {
	root, result := decodeObject(data)
	if result != resultConverted {
		observe(result)
		return Conversion{}
	}

	type fieldCandidate struct {
		rawKey   string
		fieldKey string
	}

	candidates := make([]fieldCandidate, 0, len(root))
	for rawKey := range root {
		fieldKey, ok := normalizeFieldKey(rawKey)
		if !ok {
			continue
		}
		candidates = append(candidates, fieldCandidate{
			rawKey:   rawKey,
			fieldKey: fieldKey,
		})
	}
	// Prefer a key that already satisfies Point rules. For other collisions,
	// the lexical raw-key order makes the winner deterministic.
	sort.Slice(candidates, func(i, j int) bool {
		iUnchanged := candidates[i].fieldKey == candidates[i].rawKey
		jUnchanged := candidates[j].fieldKey == candidates[j].rawKey
		if iUnchanged != jUnchanged {
			return iUnchanged
		}
		return candidates[i].rawKey < candidates[j].rawKey
	})

	fields := make(map[string]any, len(root))
	reservedFields := make(map[string]any, len(reservedKeys))
	selectedKeys := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if _, exists := selectedKeys[candidate.fieldKey]; exists {
			continue
		}
		normalized, ok := normalize(root[candidate.rawKey])
		if !ok {
			continue
		}
		selectedKeys[candidate.fieldKey] = struct{}{}
		if isReservedKey(candidate.fieldKey) {
			reservedFields[candidate.fieldKey] = normalized
		} else {
			fields[candidate.fieldKey] = normalized
		}
	}

	for _, key := range reservedKeys {
		normalized, ok := reservedFields[key]
		if !ok {
			continue
		}
		fields[firstAvailableName(fields, "json_"+key)] = normalized
	}

	if len(fields) == 0 {
		observe(resultNoFields)
		return Conversion{}
	}

	observe(resultConverted)
	return Conversion{
		Fields:    fields,
		Converted: true,
	}
}

func normalizeFieldKey(key string) (string, bool) {
	if len(key) > maxPointFieldKeyLen {
		key = key[:maxPointFieldKeyLen]
	}
	key = pointutil.ReplaceLabelKey(key)
	key = strings.TrimRight(key, `\`)
	key = strings.ReplaceAll(key, "\n", " ")
	return key, key != ""
}

func decodeObject(data []byte) (map[string]any, string) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, resultInvalidJSON
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, resultInvalidJSON
	}

	root, ok := value.(map[string]any)
	if !ok {
		return nil, resultNonObject
	}
	return root, resultConverted
}

func normalize(value any) (any, bool) {
	switch value := value.(type) {
	case string, bool:
		return value, true
	case json.Number:
		return normalizeNumber(value)
	case map[string]any, []any:
		data, err := json.Marshal(value)
		if err != nil {
			return nil, false
		}
		return string(data), true
	case nil:
		return nil, false
	default:
		return nil, false
	}
}

func normalizeNumber(number json.Number) (any, bool) {
	value := number.String()
	if strings.ContainsAny(value, ".eE") {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsInf(parsed, 0) || math.IsNaN(parsed) {
			return nil, false
		}
		return parsed, true
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, false
	}
	return parsed, true
}

func isReservedKey(key string) bool {
	for _, reserved := range reservedKeys {
		if key == reserved {
			return true
		}
	}
	return false
}

func firstAvailableName(fields map[string]any, base string) string {
	if _, ok := fields[base]; !ok {
		return base
	}
	for index := 1; ; index++ {
		name := base + "_" + strconv.Itoa(index)
		if _, ok := fields[name]; !ok {
			return name
		}
	}
}

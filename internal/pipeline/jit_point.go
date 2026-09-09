// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/ptinput/utils"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

const (
	// Keep the Go-side projection, wire buffer, native output, and any replay
	// bounded independently of the size of the collector input batch.
	jitChunkMaxRecords  = 4_096
	jitChunkTargetBytes = 8 << 20
	jitChunkRecordBytes = 64
	jitChunkEntryBytes  = 16
)

func nextJITChunkEnd(
	category point.Category,
	indexes []int,
	runs []pointRun,
	projection pljit.InputProjection,
	start int,
) int {
	end := start
	estimated := 16
	for end < len(indexes) && end-start < jitChunkMaxRecords {
		pt := runs[indexes[end]].point
		next := estimateProjectedJITPointBytes(category, pt, projection)
		if end > start && estimated+next > jitChunkTargetBytes {
			break
		}
		estimated += next
		end++
	}
	// Always make progress. A single record that exceeds the target is sent on
	// its own; the wire hard limit can then reject it without replaying peers.
	if end == start {
		return min(start+1, len(indexes))
	}
	return end
}

func estimateProjectedJITPointBytes(
	category point.Category,
	pt *point.Point,
	projection pljit.InputProjection,
) int {
	if pt == nil {
		return jitChunkRecordBytes
	}
	estimated := jitChunkRecordBytes + len(category.String()) + len(pt.Name())
	for _, kv := range pt.KVs() {
		if !projection.Includes(kv.Key) {
			continue
		}
		estimated += jitChunkEntryBytes + len(kv.Key)
		estimated += estimateJITFieldBytes(kv)
	}
	return estimated
}

func estimateJITFieldBytes(kv *point.Field) int {
	switch value := kv.Val.(type) {
	case *point.Field_S:
		return len(value.S)
	case *point.Field_A:
		// The encoded-size check still catches composite expansion.
		if value.A != nil {
			return max(len(value.A.Value), 64)
		}
		return 64
	default:
		return 8
	}
}

func encodeJITPoints(category point.Category, points []*point.Point) ([]byte, error) {
	values := make([]pljit.Point, len(points))
	for index, pt := range points {
		if pt == nil {
			return nil, fmt.Errorf("point %d is nil", index)
		}
		values[index] = pljit.Point{
			Version:      pljit.FlatPointABIVersion,
			Category:     category.String(),
			Measurement:  pt.Name(),
			Tags:         pt.MapTags(),
			Fields:       pt.InfluxFields(),
			TimeUnixNano: pt.Time().UnixNano(),
		}
	}
	return pljit.EncodeFlatPoints(values)
}

func encodeProjectedJITPoints(
	category point.Category,
	points []*point.Point,
	projection pljit.InputProjection,
) ([]byte, error) {
	if projection.IsAll() && !projection.AllowsRawStringValues() && !projection.AllowsRawTagValues() {
		return encodeJITPoints(category, points)
	}
	records := make([]pljit.FlatPointRecord, len(points))
	entries := make([]pljit.FlatPointEntry, 0, len(points)*projection.KeyCount())
	categoryName := category.String()
	for index, pt := range points {
		if pt == nil {
			return nil, fmt.Errorf("point %d is nil", index)
		}
		start := len(entries)
		for _, kv := range pt.KVs() {
			if !projection.Includes(kv.Key) {
				continue
			}
			if kv.IsTag {
				entries = append(entries, pljit.FlatPointEntry{
					Key: kv.Key, Value: jitPointTagWireValue(kv), IsTag: true,
				})
				continue
			}
			value, err := jitPointFieldWireValue(kv)
			if err != nil {
				return nil, fmt.Errorf("encode projected JIT field %q: %w", kv.Key, err)
			}
			entries = append(entries, pljit.FlatPointEntry{Key: kv.Key, Value: value})
		}
		sortFlatPointEntries(entries[start:])
		records[index] = pljit.FlatPointRecord{
			Version:      pljit.FlatPointABIVersion,
			Category:     categoryName,
			Measurement:  pt.Name(),
			TimeUnixNano: pt.Time().UnixNano(),
			Entries:      entries[start:len(entries):len(entries)],
		}
	}
	return pljit.EncodeProjectedPointRecords(records, projection)
}

type projectedJITEncoder struct {
	preparedError      error
	preparedErrorIndex int
	records            []pljit.FlatPointRecord
	entries            []pljit.FlatPointEntry
	buffer             []byte
	batch              pljit.Batch
}

// Keep reuse local to the adapter and bounded after unusually large batches.
// Each borrower owns the scratch until the synchronous native call completes.
var projectedJITEncoders = sync.Pool{New: func() any { return new(projectedJITEncoder) }}

func releaseProjectedJITEncoder(encoder *projectedJITEncoder) {
	encoder.batch.ResetForReuse()
	encoder.preparedError = nil
	if cap(encoder.buffer) > 1<<20 || cap(encoder.records) > jitChunkMaxRecords || cap(encoder.entries) > 65_536 {
		return
	}
	clear(encoder.records)
	clear(encoder.entries)
	encoder.records = encoder.records[:0]
	encoder.entries = encoder.entries[:0]
	encoder.buffer = encoder.buffer[:0]
	projectedJITEncoders.Put(encoder)
}

// prepareChunk shares the projection scan with the existing conservative size
// estimate. No native execution occurs here. Any payloads are decoded only when
// the selected prefix is encoded, including when an error causes a split retry.
func (encoder *projectedJITEncoder) prepareChunk(
	category point.Category, indexes []int, runs []pointRun,
	projection pljit.InputProjection, start int,
) int {
	clear(encoder.records)
	clear(encoder.entries)
	encoder.records = encoder.records[:0]
	encoder.entries = encoder.entries[:0]
	encoder.preparedError = nil
	limit := min(len(indexes)-start, jitChunkMaxRecords)
	if cap(encoder.records) < limit {
		encoder.records = make([]pljit.FlatPointRecord, 0, limit)
	}
	needed := min(limit*min(projection.KeyCount(), 65_536), 65_536)
	if cap(encoder.entries) < needed {
		encoder.entries = make([]pljit.FlatPointEntry, 0, needed)
	}
	categoryName := category.String()
	estimated, end := 16, start
	for end-start < limit {
		pt := runs[indexes[end]].point
		entryStart := len(encoder.entries)
		next := jitChunkRecordBytes
		record := pljit.FlatPointRecord{}
		var recordError error
		if pt == nil {
			recordError = fmt.Errorf("point %d is nil", end-start)
		} else {
			next += len(categoryName) + len(pt.Name())
			for _, kv := range pt.KVs() {
				if !projection.Includes(kv.Key) {
					continue
				}
				next += jitChunkEntryBytes + len(kv.Key) + estimateJITFieldBytes(kv)
				// A Field_A pointer is a deferred conversion, never a wire value.
				var value any = kv
				if kv.IsTag {
					value = jitPointTagWireValue(kv)
				} else if _, deferred := kv.Val.(*point.Field_A); !deferred {
					var err error
					value, err = jitPointFieldWireValue(kv)
					if err != nil && recordError == nil {
						recordError = fmt.Errorf("encode projected JIT field %q: %w", kv.Key, err)
					}
				}
				encoder.entries = append(encoder.entries, pljit.FlatPointEntry{Key: kv.Key, Value: value, IsTag: kv.IsTag})
			}
			record = pljit.FlatPointRecord{
				Version: pljit.FlatPointABIVersion, Category: categoryName,
				Measurement: pt.Name(), TimeUnixNano: pt.Time().UnixNano(),
			}
		}
		if end > start && estimated+next > jitChunkTargetBytes {
			clear(encoder.entries[entryStart:])
			encoder.entries = encoder.entries[:entryStart]
			break
		}
		if recordError != nil && encoder.preparedError == nil {
			encoder.preparedError, encoder.preparedErrorIndex = recordError, end-start
		}
		record.Entries = encoder.entries[entryStart:len(encoder.entries):len(encoder.entries)]
		sortFlatPointEntries(record.Entries)
		encoder.records = append(encoder.records, record)
		estimated += next
		end++
	}
	return end
}

func (encoder *projectedJITEncoder) encodePrepared(count int, projection pljit.InputProjection) ([]byte, error) {
	if count < 1 || count > len(encoder.records) {
		return nil, errors.New("invalid prepared JIT chunk prefix")
	}
	if count < len(encoder.records) {
		entryCount := 0
		for _, record := range encoder.records[:count] {
			entryCount += len(record.Entries)
		}
		// Record slices may use older backing arrays after entries grows. Clear
		// each discarded view as well, so retries cannot retain decoded payloads.
		for _, record := range encoder.records[count:] {
			clear(record.Entries)
		}
		clear(encoder.records[count:])
		encoder.records = encoder.records[:count]
		clear(encoder.entries[entryCount:])
		encoder.entries = encoder.entries[:entryCount]
	}
	if encoder.preparedError != nil && encoder.preparedErrorIndex < count {
		return nil, encoder.preparedError
	}
	for _, record := range encoder.records {
		for index := range record.Entries {
			if field, deferred := record.Entries[index].Value.(*point.Field); deferred {
				record.Entries[index].Value = jitPointAnyValue(field)
			}
		}
	}
	encoded, err := pljit.EncodeProjectedPointRecordsInto(encoder.buffer[:0], encoder.records, projection)
	if err == nil {
		encoder.buffer = encoded
	}
	return encoded, err
}

func (encoder *projectedJITEncoder) encode(
	category point.Category,
	points []*point.Point,
	projection pljit.InputProjection,
) ([]byte, error) {
	if projection.IsAll() && !projection.AllowsRawStringValues() && !projection.AllowsRawTagValues() {
		return encodeJITPoints(category, points)
	}
	if cap(encoder.records) < len(points) {
		encoder.records = make([]pljit.FlatPointRecord, len(points))
	} else {
		encoder.records = encoder.records[:cap(encoder.records)]
		clear(encoder.records)
		encoder.records = encoder.records[:len(points)]
	}
	encoder.entries = encoder.entries[:cap(encoder.entries)]
	clear(encoder.entries)
	encoder.entries = encoder.entries[:0]
	needed := len(points) * projection.KeyCount()
	if cap(encoder.entries) < needed {
		encoder.entries = make([]pljit.FlatPointEntry, 0, needed)
	}
	categoryName := category.String()
	for index, pt := range points {
		if pt == nil {
			return nil, fmt.Errorf("point %d is nil", index)
		}
		start := len(encoder.entries)
		for _, kv := range pt.KVs() {
			if !projection.Includes(kv.Key) {
				continue
			}
			if kv.IsTag {
				encoder.entries = append(encoder.entries, pljit.FlatPointEntry{Key: kv.Key, Value: jitPointTagWireValue(kv), IsTag: true})
				continue
			}
			value, err := jitPointFieldWireValue(kv)
			if err != nil {
				return nil, fmt.Errorf("encode projected JIT field %q: %w", kv.Key, err)
			}
			encoder.entries = append(encoder.entries, pljit.FlatPointEntry{Key: kv.Key, Value: value})
		}
		sortFlatPointEntries(encoder.entries[start:])
		encoder.records[index] = pljit.FlatPointRecord{
			Version: pljit.FlatPointABIVersion, Category: categoryName,
			Measurement: pt.Name(), TimeUnixNano: pt.Time().UnixNano(),
			Entries: encoder.entries[start:len(encoder.entries):len(encoder.entries)],
		}
	}
	encoded, err := pljit.EncodeProjectedPointRecordsInto(encoder.buffer[:0], encoder.records, projection)
	if err == nil {
		encoder.buffer = encoded
	}
	return encoded, err
}

func sortFlatPointEntries(entries []pljit.FlatPointEntry) {
	slices.SortFunc(entries, func(left, right pljit.FlatPointEntry) int {
		if left.IsTag != right.IsTag {
			if left.IsTag {
				return -1
			}
			return 1
		}
		return strings.Compare(left.Key, right.Key)
	})
}

// Views are read only while encoding. The output buffer contains copied bytes;
// no Go pointer is passed to Rust or retained after the encoder is released.
func jitPointTagWireValue(kv *point.Field) any {
	if value, ok := kv.Val.(*point.Field_S); ok && value != nil {
		return &value.S
	}
	return kv.GetS()
}

func jitPointFieldWireValue(kv *point.Field) (any, error) {
	switch value := kv.Val.(type) {
	case *point.Field_I:
		return &value.I, nil
	case *point.Field_U:
		return &value.U, nil
	case *point.Field_F:
		return &value.F, nil
	case *point.Field_B:
		return &value.B, nil
	case *point.Field_D:
		return &value.D, nil
	case *point.Field_S:
		return &value.S, nil
	case *point.Field_A:
		return jitPointAnyValue(kv), nil
	default:
		return nil, fmt.Errorf("unsupported point field type %T", kv.Val)
	}
}

// jitPointAnyValue mirrors pipeline-go's PlPt.getVal boundary without
// mutating cliutils' process-global Dict/MixedArray switches. Invalid Any
// payloads are a readable nil value in pipeline-go, not a whole-batch input
// encoding failure.
func jitPointAnyValue(field *point.Field) any {
	payload := field.GetA()
	if payload == nil {
		return nil
	}
	switch payload.TypeUrl {
	case point.ArrayFieldType:
		var array point.Array
		if err := array.Unmarshal(payload.Value); err != nil {
			return nil
		}
		values := make([]any, 0, len(array.Arr))
		var firstType any
		for index, item := range array.Arr {
			value, kind, ok := jitPointBasicValue(item, false)
			if !ok || (!point.EnableMixedArrayField && index > 0 && kind != firstType) {
				return nil
			}
			if index == 0 {
				firstType = kind
			}
			values = append(values, value)
		}
		if len(values) == 0 && !point.EnableMixedArrayField {
			return nil
		}
		return values
	case point.DictFieldType:
		var dictionary point.Map
		if err := dictionary.Unmarshal(payload.Value); err != nil {
			return nil
		}
		values := make(map[string]any, len(dictionary.Map))
		for key, item := range dictionary.Map {
			value, _, ok := jitPointBasicValue(item, true)
			if !ok {
				return nil
			}
			values[key] = value
		}
		return values
	default:
		return nil
	}
}

func jitPointBasicValue(value *point.BasicTypes, preserveUnsigned bool) (any, any, bool) {
	if value == nil {
		return nil, nil, false
	}
	switch typed := value.X.(type) {
	case *point.BasicTypes_I:
		return typed.I, (*point.BasicTypes_I)(nil), true
	case *point.BasicTypes_U:
		if preserveUnsigned {
			return pljit.OpaqueGoUint(typed.U), (*point.BasicTypes_U)(nil), true
		}
		return int64(typed.U), (*point.BasicTypes_U)(nil), true
	case *point.BasicTypes_F:
		return typed.F, (*point.BasicTypes_F)(nil), true
	case *point.BasicTypes_B:
		return typed.B, (*point.BasicTypes_B)(nil), true
	case *point.BasicTypes_D:
		if !preserveUnsigned {
			return pljit.RawString(string(typed.D)), (*point.BasicTypes_D)(nil), true
		}
		return pljit.OpaqueGoBytes(string(typed.D)), (*point.BasicTypes_D)(nil), true
	case *point.BasicTypes_S:
		return typed.S, (*point.BasicTypes_S)(nil), true
	default:
		return nil, nil, false
	}
}

func applyJITDelta(
	category point.Category,
	target *point.Point,
	encoded []byte,
	recordIndex uint64,
	option *lang.LogOption,
) (map[point.Category][]*point.Point, bool, error) {
	deltas, err := pljit.DecodeMutationDeltas(encoded)
	if err != nil {
		return nil, false, fmt.Errorf("decode JIT mutation delta: %w", err)
	}
	if len(deltas) != 1 || deltas[0].RecordIndex != recordIndex {
		return nil, false, fmt.Errorf("JIT mutation delta identity mismatch: got %d deltas for record %d", len(deltas), recordIndex)
	}

	return applyJITOperations(category, target, deltas[0].Operations, option)
}

func applyJITRecord(category point.Category, target *point.Point, record pljit.Record, recordIndex uint64, option *lang.LogOption) (map[point.Category][]*point.Point, bool, error) {
	if record.Mutations != nil {
		return applyJITOperations(category, target, record.Mutations, option)
	}
	return applyJITDelta(category, target, record.Delta, recordIndex, option)
}

func applyJITOperations(category point.Category, target *point.Point, operations []pljit.MutationOp, option *lang.LogOption) (map[point.Category][]*point.Point, bool, error) {
	var created map[point.Category][]*point.Point
	dropped := false
	type preparedMutation struct {
		value       any
		setValue    bool
		subCategory point.Category
		subpoint    *point.Point
	}
	prepared := make([]preparedMutation, len(operations))
	for index, operation := range operations {
		switch operation.Kind {
		case pljit.MutationSetField:
			value, ok, err := jitFieldValue(operation.Value)
			if err != nil {
				return nil, false, fmt.Errorf("convert JIT field %q: %w", operation.Key, err)
			}
			prepared[index].value, prepared[index].setValue = value, ok
		case pljit.MutationSetCategory:
			if !strings.EqualFold(operation.String, category.String()) {
				return nil, false, fmt.Errorf("JIT point category changed from %q to %q", category.String(), operation.String)
			}
		case pljit.MutationAddSubpoint:
			if operation.Point == nil {
				return nil, false, fmt.Errorf("JIT add-subpoint mutation has no point")
			}
			prepared[index].subCategory = point.CatString(operation.Point.Category)
			if prepared[index].subCategory == point.UnknownCategory {
				return nil, false, fmt.Errorf("JIT subpoint has unknown category %q", operation.Point.Category)
			}
			subpoint, err := pointFromJIT(operation.Point)
			if err != nil {
				return nil, false, fmt.Errorf("convert JIT subpoint: %w", err)
			}
			prepared[index].subpoint = subpoint
		case pljit.MutationDeleteField, pljit.MutationSetTag, pljit.MutationSetRawTag, pljit.MutationDeleteTag,
			pljit.MutationSetMeasurement, pljit.MutationSetRawMeasurement, pljit.MutationSetTime, pljit.MutationSetStatus,
			pljit.MutationClearStatus, pljit.MutationSetDropped, pljit.MutationClearSubpoints:
		default:
			return nil, false, fmt.Errorf("unknown JIT mutation operation %d", operation.Kind)
		}
	}

	// Validate the entire delta before committing any mutation. Invalid native
	// output is reported as a failure without replaying the script.
	for index, operation := range operations {
		switch operation.Kind {
		case pljit.MutationSetField:
			if prepared[index].setValue {
				target.Set(operation.Key, prepared[index].value)
			} else {
				target.Del(operation.Key)
			}
		case pljit.MutationDeleteField, pljit.MutationDeleteTag:
			// Native tags and fields are separate namespaces. A preceding
			// SetTag may already have replaced the old field with this name.
			for _, kv := range target.KVs() {
				if kv.Key == operation.Key && kv.IsTag == (operation.Kind == pljit.MutationDeleteTag) {
					target.Del(operation.Key)
					break
				}
			}
		case pljit.MutationSetTag, pljit.MutationSetRawTag:
			target.SetTag(operation.Key, operation.String)
		case pljit.MutationSetCategory:
		case pljit.MutationSetMeasurement, pljit.MutationSetRawMeasurement:
			target.SetName(operation.String)
		case pljit.MutationSetTime:
			target.SetTime(time.Unix(0, operation.Time))
		case pljit.MutationSetStatus, pljit.MutationClearStatus:
			// The status field is emitted as a separate field mutation. The ABI status
			// property is retained for runtimes that consume FlatPoint directly.
		case pljit.MutationSetDropped:
			dropped = operation.Dropped
		case pljit.MutationClearSubpoints:
			created = nil
		case pljit.MutationAddSubpoint:
			if created == nil {
				created = make(map[point.Category][]*point.Point)
			}
			created[prepared[index].subCategory] = append(created[prepared[index].subCategory], prepared[index].subpoint)
		}
	}

	if category == point.Logging && option != nil {
		status, _ := target.Get("status").(string)
		for _, ignored := range option.IgnoreStatus {
			if strings.ToLower(ignored) == status {
				dropped = true
				break
			}
		}
	}
	return created, dropped, nil
}

func decodeJITEmitted(encoded [][]byte) (map[point.Category][]*point.Point, error) {
	var created map[point.Category][]*point.Point
	for index, payload := range encoded {
		var value pljit.Point
		if len(payload) >= 4 && string(payload[:4]) == "PPF1" {
			points, err := pljit.DecodeFlatPoints(payload)
			if err != nil || len(points) != 1 {
				return nil, fmt.Errorf("decode emitted JIT point %d binary envelope: points=%d: %w", index, len(points), err)
			}
			value = points[0]
		} else {
			decoder := json.NewDecoder(bytes.NewReader(payload))
			decoder.UseNumber()
			if err := decoder.Decode(&value); err != nil {
				return nil, fmt.Errorf("decode emitted JIT point %d: %w", index, err)
			}
			if value.Version != pljit.FlatPointABIVersion || !json.Valid(payload) {
				return nil, fmt.Errorf("invalid emitted JIT point %d version or JSON envelope", index)
			}
		}
		category := point.CatString(value.Category)
		if category == point.UnknownCategory {
			return nil, fmt.Errorf("emitted JIT point %d has unknown category %q", index, value.Category)
		}
		converted, err := pointFromJIT(&value)
		if err != nil {
			return nil, fmt.Errorf("convert emitted JIT point %d: %w", index, err)
		}
		if created == nil {
			created = make(map[point.Category][]*point.Point)
		}
		created[category] = append(created[category], converted)
	}
	return created, nil
}

func applyJITStatic(
	category point.Category,
	target *point.Point,
	batch *pljit.StaticBatch,
	recordIndex int,
	option *lang.LogOption,
) (map[point.Category][]*point.Point, bool, error) {
	if batch == nil || recordIndex < 0 || recordIndex+1 >= len(batch.StateOffsets) ||
		recordIndex+1 >= len(batch.ValueOffsets) {
		return nil, false, errors.New("JIT static batch record index is invalid")
	}
	slotCount := len(batch.Schema.Keys)
	stateStart := int(batch.StateOffsets[recordIndex])
	stateEnd := int(batch.StateOffsets[recordIndex+1])
	if slotCount == 0 || stateStart < 0 || stateEnd-stateStart != slotCount ||
		stateEnd > len(batch.States) {
		return nil, false, errors.New("JIT static batch state range is invalid")
	}
	valueIndex := int(batch.ValueOffsets[recordIndex])
	valueEnd := int(batch.ValueOffsets[recordIndex+1])
	valueCount := len(batch.Values)
	preparedFields := batch.Validated() && batch.PointFields != nil
	if preparedFields {
		valueCount = len(batch.PointFields)
	}
	if valueIndex < 0 || valueEnd < valueIndex || valueEnd > valueCount {
		return nil, false, errors.New("JIT static batch value range is invalid")
	}
	if batch.Validated() {
		// decodeStaticBatch already validated and normalized every state/value.
		// Apply directly so the common static ABI does not allocate a temporary
		// mutation slice for every Point.
		for slotIndex, key := range batch.Schema.Keys {
			state := batch.States[stateStart+slotIndex]
			if preparedFields && (state == pljit.StaticSetField || state == pljit.StaticSetTag) {
				target.MustAddKVs(batch.PointFields[valueIndex])
				valueIndex++
				continue
			}
			switch batch.States[stateStart+slotIndex] {
			case pljit.StaticNoop:
			case pljit.StaticSetField:
				target.Set(key, batch.Values[valueIndex])
				valueIndex++
			case pljit.StaticDeleteField, pljit.StaticDeleteTag:
				target.Del(key)
			case pljit.StaticSetTag:
				target.SetTag(key, batch.Values[valueIndex].(string))
				valueIndex++
			}
		}
	} else {
		type preparedMutation struct {
			value    any
			setValue bool
		}
		prepared := make([]preparedMutation, slotCount)
		for slotIndex := range slotCount {
			state := batch.States[stateStart+slotIndex]
			switch state {
			case pljit.StaticNoop, pljit.StaticDeleteField, pljit.StaticDeleteTag:
			case pljit.StaticSetField:
				if valueIndex >= valueEnd {
					return nil, false, errors.New("JIT static field value is missing")
				}
				value, ok, err := jitFieldValue(batch.Values[valueIndex])
				if err != nil {
					return nil, false, fmt.Errorf("convert JIT static field %q: %w", batch.Schema.Keys[slotIndex], err)
				}
				prepared[slotIndex] = preparedMutation{value: value, setValue: ok}
				valueIndex++
			case pljit.StaticSetTag:
				if valueIndex >= valueEnd {
					return nil, false, errors.New("JIT static tag value is missing")
				}
				value, ok := batch.Values[valueIndex].(string)
				if !ok {
					return nil, false, fmt.Errorf("JIT static tag %q has value type %T", batch.Schema.Keys[slotIndex], batch.Values[valueIndex])
				}
				prepared[slotIndex] = preparedMutation{value: value, setValue: true}
				valueIndex++
			default:
				return nil, false, fmt.Errorf("unknown JIT static mutation state %d", state)
			}
		}
		if valueIndex != valueEnd {
			return nil, false, errors.New("JIT static record contains unused values")
		}

		for slotIndex, key := range batch.Schema.Keys {
			switch batch.States[stateStart+slotIndex] {
			case pljit.StaticNoop:
			case pljit.StaticSetField:
				if prepared[slotIndex].setValue {
					target.Set(key, prepared[slotIndex].value)
				} else {
					target.Del(key)
				}
			case pljit.StaticDeleteField, pljit.StaticDeleteTag:
				target.Del(key)
			case pljit.StaticSetTag:
				target.SetTag(key, prepared[slotIndex].value.(string))
			}
		}
	}
	dropped := false
	if category == point.Logging && option != nil {
		status, _ := target.Get("status").(string)
		for _, ignored := range option.IgnoreStatus {
			if strings.ToLower(ignored) == status {
				dropped = true
				break
			}
		}
	}
	return nil, dropped, nil
}

func pointFromJIT(value *pljit.Point) (*point.Point, error) {
	kvs := make(point.KVs, 0, len(value.Tags)+len(value.Fields))
	for key, tag := range value.Tags {
		kvs = append(kvs, point.NewKV(key, tag, point.WithKVTagSet(true)))
	}
	for key, raw := range value.Fields {
		converted, ok, err := jitFieldValue(raw)
		if err != nil {
			return nil, fmt.Errorf("convert field %q: %w", key, err)
		}
		if ok {
			kvs = append(kvs, point.NewKV(key, converted))
		}
	}
	timestamp := time.Unix(0, value.TimeUnixNano)
	if value.Time != "" {
		parsed, err := time.Parse(time.RFC3339Nano, value.Time)
		if err != nil {
			return nil, fmt.Errorf("parse time %q: %w", value.Time, err)
		}
		timestamp = parsed
	}
	// Match NewPlPt: category policies supply required fields and enforce
	// category-specific key/value rules when publishing newly created points.
	options := utils.PtCatOption(point.CatString(value.Category))
	options = append(options, point.WithTime(timestamp))
	return point.NewPoint(value.Measurement, kvs, options...), nil
}

func jitFieldValue(value any) (any, bool, error) {
	switch value := value.(type) {
	case json.Number:
		if integer, err := value.Int64(); err == nil {
			return integer, true, nil
		}
		floating, err := value.Float64()
		return floating, err == nil, err
	case nil:
		return nil, true, nil
	case pljit.RawString:
		// Go strings preserve arbitrary bytes. Do not UTF-8-normalize native
		// field values when applying either static output or mutation deltas.
		return string(value), true, nil
	case pljit.OpaqueGoBytes:
		return nil, true, nil
	case pljit.OpaqueGoUint:
		return int64(value), true, nil
	case string, bool, int64, float64:
		return value, true, nil
	case []any:
		converted, err := pljit.NormalizeRawPointArrayValue(value)
		return converted, err == nil, err
	case map[string]any:
		converted, err := pljit.NormalizeRawPointMap(value)
		return converted, err == nil, err
	default:
		return nil, false, fmt.Errorf("unsupported value type %T", value)
	}
}

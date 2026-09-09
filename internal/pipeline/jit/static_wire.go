// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/GuanceCloud/cliutils/point"
)

const (
	staticBatchABIVersion         = 2
	staticBatchHeaderSize         = 56
	staticBatchHasErrors          = 1
	staticRecordCommitPrefixError = 1
)

func decodeStaticBatch(input []byte, schema StaticOutputSchema) (Batch, error) {
	return decodeStaticBatchInto(input, schema, nil)
}

func decodeStaticBatchInto(input []byte, schema StaticOutputSchema, scratch *Batch) (Batch, error) {
	if schema.Dynamic {
		return decodeDynamicBatchInto(input, schema, scratch)
	}
	if len(input) < staticBatchHeaderSize || len(input) > staticBatchHeaderSize+maxResultBytes {
		return Batch{}, errors.New("invalid static batch result length")
	}
	if string(input[:4]) != "PPS2" || binary.LittleEndian.Uint16(input[4:6]) != staticBatchABIVersion {
		return Batch{}, errors.New("invalid static batch header")
	}
	flags := binary.LittleEndian.Uint16(input[6:8])
	if flags&^uint16(staticBatchHasErrors) != 0 {
		return Batch{}, errors.New("unknown static batch flags")
	}
	recordCount := binary.LittleEndian.Uint32(input[8:12])
	slotCount := binary.LittleEndian.Uint32(input[12:16])
	payloadLength := binary.LittleEndian.Uint32(input[16:20])
	if recordCount > pointWireMaxItems || slotCount == 0 || slotCount > pointWireMaxItems ||
		int(slotCount) != len(schema.Keys) || int(payloadLength) != len(input)-staticBatchHeaderSize ||
		binary.LittleEndian.Uint32(input[20:24]) != 0 ||
		recordCount > payloadLength/7 {
		return Batch{}, errors.New("invalid static batch counts")
	}
	if !bytes.Equal(input[24:56], schema.Hash[:]) {
		return Batch{}, errors.New("static batch schema hash mismatch")
	}

	var batch Batch
	if scratch != nil {
		scratch.ResetForReuse()
		batch = *scratch
	}
	if cap(batch.Records) < int(recordCount) {
		batch.Records = make([]Record, recordCount)
	} else {
		batch.Records = batch.Records[:recordCount]
	}
	batch.Protocol = "static-v2"
	batch.EncodedBytes = len(input)
	if batch.Static == nil {
		batch.Static = &StaticBatch{}
	}
	batch.Static.Schema = schema
	if cap(batch.Static.StateOffsets) < int(recordCount)+1 {
		batch.Static.StateOffsets = make([]uint32, 1, int(recordCount)+1)
	} else {
		batch.Static.StateOffsets = batch.Static.StateOffsets[:1]
		batch.Static.StateOffsets[0] = 0
	}
	if cap(batch.Static.ValueOffsets) < int(recordCount)+1 {
		batch.Static.ValueOffsets = make([]uint32, 1, int(recordCount)+1)
	} else {
		batch.Static.ValueOffsets = batch.Static.ValueOffsets[:1]
		batch.Static.ValueOffsets[0] = 0
	}

	cursor := &pointCursor{data: input[staticBatchHeaderSize:]}
	stateBytes := (int(slotCount) + 1) / 2
	nodes := 0
	hasErrors := false
	for recordIndex := range int(recordCount) {
		length, err := cursor.u32()
		if err != nil {
			return Batch{}, fmt.Errorf("decode static record %d: %w", recordIndex, err)
		}
		data, err := cursor.take(int(length))
		if err != nil {
			return Batch{}, fmt.Errorf("decode static record %d: %w", recordIndex, err)
		}
		record := pointCursor{data: data}
		var ownedRecord string
		statusByte, err := record.u8()
		if err != nil {
			return Batch{}, fmt.Errorf("decode static record %d status: %w", recordIndex, err)
		}
		status := TerminalStatus(statusByte)
		if status < TerminalOK || status > TerminalCancelled {
			return Batch{}, fmt.Errorf("static record %d has invalid terminal %d", recordIndex, status)
		}
		recordFlags, err := record.u16()
		if err != nil {
			return Batch{}, fmt.Errorf("decode static record %d flags: %w", recordIndex, err)
		}
		if recordFlags&^uint16(staticRecordCommitPrefixError) != 0 ||
			(recordFlags != 0 && status != TerminalError) {
			return Batch{}, fmt.Errorf("static record %d has invalid flags %d", recordIndex, recordFlags)
		}
		batch.Records[recordIndex].Status = status
		batch.Records[recordIndex].CommitPrefixError = recordFlags&staticRecordCommitPrefixError != 0
		hasMutations := status == TerminalOK || batch.Records[recordIndex].CommitPrefixError
		if hasMutations {
			if cap(batch.Static.States) == 0 {
				// Two slots fit in one payload byte. Bound speculative capacity
				// by actual input and allocate only for records with mutations.
				batch.Static.States = make([]StaticMutationState, 0, min(uint64(recordCount)*uint64(slotCount), uint64(payloadLength)*2))
				if batch.preparePointFields {
					batch.Static.PointFields = make([]*point.Field, 0, recordCount)
				} else {
					batch.Static.Values = make([]any, 0, recordCount)
				}
			}
			statesStart := len(batch.Static.States)
			batch.Static.States = append(batch.Static.States, make([]StaticMutationState, slotCount)...)
			packed, err := record.take(stateBytes)
			if err != nil {
				return Batch{}, fmt.Errorf("decode static record %d states: %w", recordIndex, err)
			}
			if slotCount%2 != 0 && packed[len(packed)-1]&0xf0 != 0 {
				return Batch{}, fmt.Errorf("static record %d has non-zero unused state bits", recordIndex)
			}
			for slotIndex := range int(slotCount) {
				state := StaticMutationState((packed[slotIndex/2] >> ((slotIndex % 2) * 4)) & 0x0f)
				if state > StaticDeleteTag {
					return Batch{}, fmt.Errorf("static record %d slot %d has invalid state %d", recordIndex, slotIndex, state)
				}
				batch.Static.States[statesStart+slotIndex] = state
				if state != StaticSetField && state != StaticSetTag {
					continue
				}
				if batch.preparePointFields {
					field, err := record.preparedPointField(schema.Keys[slotIndex], state == StaticSetTag, &ownedRecord, &nodes)
					if err != nil {
						return Batch{}, fmt.Errorf("decode static record %d slot %d value: %w", recordIndex, slotIndex, err)
					}
					batch.Static.PointFields = append(batch.Static.PointFields, field)
					continue
				}
				value, err := record.value(0, &nodes)
				if err != nil {
					return Batch{}, fmt.Errorf("decode static record %d slot %d value: %w", recordIndex, slotIndex, err)
				}
				if state == StaticSetTag {
					if raw, ok := value.(RawString); ok {
						// Value tag 8 is explicit and unambiguous; preserve bytes
						// while removing the internal wire type before Point apply.
						value = string(raw)
					}
					if _, ok := value.(string); !ok {
						return Batch{}, fmt.Errorf("static record %d slot %d tag value is %T", recordIndex, slotIndex, value)
					}
				} else {
					value, err = normalizeStaticFieldValue(value)
					if err != nil {
						return Batch{}, fmt.Errorf("normalize static record %d slot %d field: %w", recordIndex, slotIndex, err)
					}
				}
				batch.Static.Values = append(batch.Static.Values, value)
			}
		}
		switch status { //nolint:exhaustive // The wire status range was validated before decoding.
		case TerminalOK, TerminalDropped:
		case TerminalError, TerminalCancelled:
			hasErrors = true
			remaining, err := record.take(len(record.data) - record.offset)
			if err != nil {
				return Batch{}, fmt.Errorf("decode static record %d error: %w", recordIndex, err)
			}
			// The native fast path decodes directly from a Rust-owned buffer and
			// frees it on return. Error payloads are the only decoded values that
			// otherwise retain a view into that buffer.
			batch.Records[recordIndex].Error = append([]byte(nil), remaining...)
		}
		if !record.done() {
			return Batch{}, fmt.Errorf("static record %d contains trailing bytes", recordIndex)
		}
		batch.Static.StateOffsets = append(batch.Static.StateOffsets, uint32(len(batch.Static.States)))
		values := len(batch.Static.Values)
		if batch.preparePointFields {
			values = len(batch.Static.PointFields)
		}
		batch.Static.ValueOffsets = append(batch.Static.ValueOffsets, uint32(values))
	}
	if !cursor.done() {
		return Batch{}, errors.New("static batch contains trailing bytes")
	}
	if hasErrors != (flags&staticBatchHasErrors != 0) {
		return Batch{}, errors.New("static batch error flag does not match terminals")
	}
	batch.Static.validated = true
	if scratch != nil {
		*scratch = batch
	}
	return batch, nil
}

// preparedPointField avoids boxing strings and converting them back into Point
// fields during apply. Each record's text is copied once into immutable Go
// storage; transferred fields never retain native memory or reusable scratch.
func (cursor *pointCursor) preparedPointField(key string, tag bool, ownedRecord *string, nodes *int) (*point.Field, error) {
	if cursor.offset < len(cursor.data) && (cursor.data[cursor.offset] == 5 || cursor.data[cursor.offset] == 8) {
		if err := pointWireNode(0, nodes); err != nil {
			return nil, err
		}
		kind, err := cursor.u8()
		if err != nil {
			return nil, err
		}
		length, err := cursor.u32()
		if err != nil {
			return nil, err
		}
		start := cursor.offset
		value, err := cursor.take(int(length))
		if err != nil {
			return nil, err
		}
		if kind == 5 && !utf8.Valid(value) {
			return nil, errors.New("point wire string is not valid UTF-8")
		}
		if *ownedRecord == "" {
			*ownedRecord = string(cursor.data)
		}
		text := (*ownedRecord)[start:cursor.offset]
		var field *point.Field
		if point.GetPointPool() != nil {
			// Preserve the configured Point pool without boxing the dynamic text.
			field = point.NewKV(key, "")
			field.Val.(*point.Field_S).S = text
		} else {
			owned := new(struct {
				field point.Field
				value point.Field_S
			})
			owned.value.S = text
			owned.field.Key, owned.field.Val = key, &owned.value
			field = &owned.field
		}
		field.IsTag = tag
		return field, nil
	}
	value, err := cursor.value(0, nodes)
	if err != nil {
		return nil, err
	}
	if tag {
		return nil, fmt.Errorf("static tag value is %T", value)
	}
	value, err = normalizeStaticFieldValue(value)
	if err != nil {
		return nil, err
	}
	return point.NewKV(key, value), nil
}

func normalizeStaticFieldValue(value any) (any, error) {
	switch value := value.(type) {
	case RawString:
		return string(value), nil
	case nil, string, bool, int64, float64:
		return value, nil
	case []any:
		return NormalizeRawPointArrayValue(value)
	case map[string]any:
		return NormalizeRawPointMap(value)
	default:
		return nil, fmt.Errorf("unsupported value type %T", value)
	}
}

// Point arrays contain scalar entries only. pipeline-go serializes an array
// containing a map or another array as JSON when NewAnyArray rejects it.
// Preserve ordinary typed scalar arrays; never pass composite entries to
// NewKV, which otherwise silently produces a field without a value.
func NormalizeRawPointArrayValue(values []any) (any, error) {
	for _, value := range values {
		switch value.(type) {
		case []any, map[string]any:
			encoded, err := json.Marshal(values)
			return string(encoded), err
		}
	}
	return NormalizeRawPointList(values), nil
}

// NormalizeRawPointList removes wire-only scalar wrappers before Point builds
// its array protobuf. The input has already passed wire depth/node limits.
func NormalizeRawPointList(values []any) []any {
	var out []any
	for i, value := range values {
		var converted any
		switch v := value.(type) {
		case RawString:
			converted = string(v)
		case OpaqueGoBytes:
			converted = []byte(v)
		case OpaqueGoUint:
			converted = uint64(v)
		default:
			continue
		}
		if out == nil {
			out = append([]any(nil), values...)
		}
		out[i] = converted
	}
	if out == nil {
		return values
	}
	return out
}

// NormalizeRawPointMap builds the protobuf directly, avoiding changes to the
// process-global EnableDictField setting during concurrent native application.
func NormalizeRawPointMap(values map[string]any) (any, error) {
	dict := &point.Map{Map: make(map[string]*point.BasicTypes, len(values))}
	for key, value := range values {
		var entry *point.BasicTypes
		switch v := value.(type) {
		case nil:
			// pipeline-go's point.NewMap retains the key with a nil protobuf map
			// value. Its value field is omitted on the wire (not encoded as an
			// explicit empty BasicTypes message), and reading the Map yields nil.
			entry = nil
		case int64:
			entry = &point.BasicTypes{X: &point.BasicTypes_I{I: v}}
		case float64:
			entry = &point.BasicTypes{X: &point.BasicTypes_F{F: v}}
		case bool:
			entry = &point.BasicTypes{X: &point.BasicTypes_B{B: v}}
		case string:
			entry = &point.BasicTypes{X: &point.BasicTypes_S{S: v}}
		case RawString:
			entry = &point.BasicTypes{X: &point.BasicTypes_S{S: string(v)}}
		case OpaqueGoBytes:
			entry = &point.BasicTypes{X: &point.BasicTypes_D{D: []byte(v)}}
		case OpaqueGoUint:
			entry = &point.BasicTypes{X: &point.BasicTypes_U{U: uint64(v)}}
		default:
			encoded, err := json.Marshal(values)
			return string(encoded), err
		}
		dict.Map[key] = entry
	}
	return point.NewAny(dict)
}

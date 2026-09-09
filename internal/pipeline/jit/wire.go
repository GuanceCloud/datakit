// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Portions of this file are based on GuanceCloud/platypus-plus-go.
// Copyright 2026-present Guance, Inc. Licensed under the MIT License.

package jit

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"
	"unicode/utf8"
)

const (
	FlatPointABIVersion     = 1
	MutationDeltaABIVersion = 1
)

// RawString is the lossless field-value representation for the planned
// byte-string wire extension (value tag 8). It is not a JSON representation.
// Callers must negotiate support before sending it to a native runtime;
// ordinary string encoding deliberately continues to require UTF-8.
type RawString string

type OpaqueGoBytes string
type OpaqueGoUint uint64

type Point struct {
	Version      uint16
	Category     string
	Measurement  string
	Tags         map[string]string
	Fields       map[string]any
	Time         string
	TimeUnixNano int64
	Status       *string
	Subpoints    []Point
	Dropped      bool
}

// FlatPointRecord is the allocation-light subset used by DataKit's proven
// input projection. Entries must be grouped with tags first and sorted by key
// within tags and fields when byte-for-byte compatibility with EncodeFlatPoints
// is required.
type FlatPointRecord struct {
	Version      uint16
	Category     string
	Measurement  string
	TimeUnixNano int64
	Entries      []FlatPointEntry
}

// FlatPointEntry is one projected tag or field in a FlatPointRecord.
// Value may also be a non-nil *int64, *uint64, *float64, *bool, *string, or
// *[]byte for allocation-free scalar reads. The source must remain unchanged
// during encoding; the returned wire buffer owns a copy of its bytes.
type FlatPointEntry struct {
	Key   string
	Value any
	IsTag bool
}

const (
	pointWireHeaderBytes = 16
	pointWireMaxBytes    = 64 << 20
	pointWireMaxPrealloc = 4 << 20
	pointWireMaxItems    = 65_536
	pointWireMaxDepth    = 64
	pointWireMaxNodes    = 1_000_000
)

type MutationOpKind uint8

const (
	MutationSetField          MutationOpKind = 1
	MutationDeleteField       MutationOpKind = 2
	MutationSetTag            MutationOpKind = 3
	MutationDeleteTag         MutationOpKind = 4
	MutationSetCategory       MutationOpKind = 5
	MutationSetMeasurement    MutationOpKind = 6
	MutationSetTime           MutationOpKind = 7
	MutationSetStatus         MutationOpKind = 8
	MutationClearStatus       MutationOpKind = 9
	MutationSetDropped        MutationOpKind = 10
	MutationClearSubpoints    MutationOpKind = 11
	MutationAddSubpoint       MutationOpKind = 12
	MutationSetRawTag         MutationOpKind = 13
	MutationSetRawMeasurement MutationOpKind = 14
)

type MutationOp struct {
	Kind    MutationOpKind
	Key     string
	String  string
	Value   any
	Time    int64
	Dropped bool
	Point   *Point
}

type MutationDelta struct {
	RecordIndex uint64
	Operations  []MutationOp
}

func (delta MutationDelta) Apply(point *Point) error {
	if point == nil {
		return errors.New("cannot apply a mutation delta to a nil point")
	}
	for _, operation := range delta.Operations {
		switch operation.Kind {
		case MutationSetField:
			if point.Fields == nil {
				point.Fields = make(map[string]any)
			}
			point.Fields[operation.Key] = operation.Value
		case MutationDeleteField:
			delete(point.Fields, operation.Key)
		case MutationSetTag, MutationSetRawTag:
			if point.Tags == nil {
				point.Tags = make(map[string]string)
			}
			point.Tags[operation.Key] = operation.String
		case MutationDeleteTag:
			delete(point.Tags, operation.Key)
		case MutationSetCategory:
			point.Category = operation.String
		case MutationSetMeasurement, MutationSetRawMeasurement:
			point.Measurement = operation.String
		case MutationSetTime:
			point.Time = time.Unix(0, operation.Time).UTC().Format(time.RFC3339Nano)
		case MutationSetStatus:
			value := operation.String
			point.Status = &value
		case MutationClearStatus:
			point.Status = nil
		case MutationSetDropped:
			point.Dropped = operation.Dropped
		case MutationClearSubpoints:
			point.Subpoints = nil
		case MutationAddSubpoint:
			if operation.Point == nil {
				return errors.New("add-subpoint mutation has no point")
			}
			point.Subpoints = append(point.Subpoints, *operation.Point)
		default:
			return fmt.Errorf("unknown mutation operation %d", operation.Kind)
		}
	}
	return nil
}

func EncodeFlatPoints(points []Point) ([]byte, error) {
	writer, err := newPointWriter("PPF1", FlatPointABIVersion, len(points))
	if err != nil {
		return nil, err
	}
	nodes := 0
	for index := range points {
		if err := writer.lengthPrefixed(func() error { return writer.point(&points[index], 0, &nodes) }); err != nil {
			return nil, err
		}
	}
	return writer.finish()
}

// EncodeFlatPointRecords encodes projected records without constructing and
// sorting intermediate tag and field maps.
func EncodeFlatPointRecords(records []FlatPointRecord) ([]byte, error) {
	return encodeFlatPointRecords(records, false, false, false)
}

// EncodeProjectedPointRecords uses only the byte-value capability negotiated
// with the program that owns projection. Metadata remains strict UTF-8.
func EncodeProjectedPointRecords(records []FlatPointRecord, projection InputProjection) ([]byte, error) {
	return EncodeProjectedPointRecordsInto(nil, records, projection)
}

// EncodeProjectedPointRecordsInto reuses dst's storage. The returned slice may
// alias dst and remains owned by the caller. This avoids allocating an input
// wire buffer for every persistent native batch.
func EncodeProjectedPointRecordsInto(dst []byte, records []FlatPointRecord, projection InputProjection) ([]byte, error) {
	return encodeFlatPointRecordsInto(dst, records, projection.AllowsRawStringValues(), projection.AllowsRawTagValues(), projection.AllowsRawPointKeys())
}

func encodeFlatPointRecords(records []FlatPointRecord, rawStrings, rawTags, rawKeys bool) ([]byte, error) {
	return encodeFlatPointRecordsInto(nil, records, rawStrings, rawTags, rawKeys)
}

func encodeFlatPointRecordsInto(dst []byte, records []FlatPointRecord, rawStrings, rawTags, rawKeys bool) ([]byte, error) {
	writer, err := newPointWriterBuffer(
		"PPF1",
		FlatPointABIVersion,
		len(records),
		flatPointRecordCapacity(records),
		dst,
	)
	if err != nil {
		return nil, err
	}
	writer.rawStringValues = rawStrings
	writer.rawTagValues = rawTags
	writer.rawPointKeys = rawKeys
	nodes := 0
	for index := range records {
		if err := writer.lengthPrefixed(func() error {
			return writer.flatPointRecord(&records[index], &nodes)
		}); err != nil {
			return nil, err
		}
	}
	return writer.finish()
}

func DecodeFlatPoints(input []byte) ([]Point, error) {
	count, cursor, err := pointEnvelope(input, "PPF1", FlatPointABIVersion)
	if err != nil {
		return nil, err
	}
	points := make([]Point, 0, count)
	nodes := 0
	for index := 0; index < count; index++ {
		record, err := cursor.lengthPrefixed()
		if err != nil {
			return nil, err
		}
		point, err := record.point(0, &nodes)
		if err != nil {
			return nil, err
		}
		if !record.done() {
			return nil, errors.New("flat point record contains trailing bytes")
		}
		points = append(points, point)
	}
	if !cursor.done() {
		return nil, errors.New("flat point envelope contains trailing bytes")
	}
	return points, nil
}

func EncodeMutationDeltas(deltas []MutationDelta) ([]byte, error) {
	writer, err := newPointWriter("PPD1", MutationDeltaABIVersion, len(deltas))
	if err != nil {
		return nil, err
	}
	nodes := 0
	for _, delta := range deltas {
		delta := delta
		if err := writer.lengthPrefixed(func() error {
			writer.u64(delta.RecordIndex)
			if err := writer.count(len(delta.Operations)); err != nil {
				return err
			}
			for _, operation := range delta.Operations {
				if err := writer.mutation(operation, &nodes); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return writer.finish()
}

func DecodeMutationDeltas(input []byte) ([]MutationDelta, error) {
	count, cursor, err := pointEnvelope(input, "PPD1", MutationDeltaABIVersion)
	if err != nil {
		return nil, err
	}
	deltas := make([]MutationDelta, 0, count)
	nodes := 0
	for index := 0; index < count; index++ {
		record, err := cursor.lengthPrefixed()
		if err != nil {
			return nil, err
		}
		recordIndex, err := record.u64()
		if err != nil {
			return nil, err
		}
		operationCount, err := record.count()
		if err != nil {
			return nil, err
		}
		delta := MutationDelta{RecordIndex: recordIndex, Operations: make([]MutationOp, 0, operationCount)}
		for operationIndex := 0; operationIndex < operationCount; operationIndex++ {
			operation, err := record.mutation(&nodes)
			if err != nil {
				return nil, err
			}
			delta.Operations = append(delta.Operations, operation)
		}
		if !record.done() {
			return nil, errors.New("mutation delta contains trailing bytes")
		}
		deltas = append(deltas, delta)
	}
	if !cursor.done() {
		return nil, errors.New("mutation envelope contains trailing bytes")
	}
	return deltas, nil
}

type pointWriter struct {
	data            []byte
	rawStringValues bool
	rawTagValues    bool
	rawPointKeys    bool
}

func newPointWriter(magic string, version uint16, count int) (pointWriter, error) {
	return newPointWriterCapacity(magic, version, count, pointWireHeaderBytes)
}

func newPointWriterCapacity(magic string, version uint16, count, capacity int) (pointWriter, error) {
	return newPointWriterBuffer(magic, version, count, capacity, nil)
}

func newPointWriterBuffer(magic string, version uint16, count, capacity int, dst []byte) (pointWriter, error) {
	if count < 0 || count > pointWireMaxItems {
		return pointWriter{}, fmt.Errorf("point wire count exceeds %d", pointWireMaxItems)
	}
	capacity = max(capacity, pointWireHeaderBytes)
	capacity = min(capacity, pointWireHeaderBytes+pointWireMaxPrealloc)
	var data []byte
	if cap(dst) >= capacity {
		data = dst[:pointWireHeaderBytes]
		clear(data)
	} else {
		data = make([]byte, pointWireHeaderBytes, capacity)
	}
	writer := pointWriter{data: data}
	copy(writer.data[:4], magic)
	binary.LittleEndian.PutUint16(writer.data[4:6], version)
	binary.LittleEndian.PutUint32(writer.data[8:12], uint32(count))
	return writer, nil
}

func flatPointRecordCapacity(records []FlatPointRecord) int {
	capacity := pointWireHeaderBytes
	add := func(length int) {
		maximum := pointWireHeaderBytes + pointWireMaxPrealloc
		if length <= 0 || capacity >= maximum {
			return
		}
		capacity += min(length, maximum-capacity)
	}
	for _, record := range records {
		// Record length, fixed header, three item counts, and two string lengths.
		add(4 + 12 + 12 + 8 + len(record.Category) + len(record.Measurement))
		for _, entry := range record.Entries {
			// Key length and value kind.
			add(5 + len(entry.Key))
			switch value := entry.Value.(type) {
			case string:
				add(4 + len(value))
			case *string:
				if value != nil {
					add(4 + len(*value))
				}
			case *[]byte:
				if value != nil {
					add(4 + len(*value))
				}
			case *int64, *uint64, *float64:
				add(8)
			case *bool:
			case nil, bool:
			case int, int8, int16, int32, int64,
				uint, uint8, uint16, uint32, uint64, float32, float64:
				add(8)
			default:
				// Composite values can still grow the buffer normally; avoid walking
				// them twice merely to obtain an exact capacity estimate.
				add(16)
			}
		}
	}
	return capacity
}

func (writer *pointWriter) finish() ([]byte, error) {
	if len(writer.data)-pointWireHeaderBytes > pointWireMaxBytes {
		return nil, fmt.Errorf("point wire payload exceeds %d bytes", pointWireMaxBytes)
	}
	binary.LittleEndian.PutUint32(writer.data[12:16], uint32(len(writer.data)-pointWireHeaderBytes))
	return writer.data, nil
}

func (writer *pointWriter) reserve(length int) error {
	if length < 0 || len(writer.data)-pointWireHeaderBytes > pointWireMaxBytes-length {
		return fmt.Errorf("point wire payload exceeds %d bytes", pointWireMaxBytes)
	}
	return nil
}

func (writer *pointWriter) u8(value byte) error { writer.data = append(writer.data, value); return nil }
func (writer *pointWriter) u16(value uint16) {
	var data [2]byte
	binary.LittleEndian.PutUint16(data[:], value)
	writer.data = append(writer.data, data[:]...)
}
func (writer *pointWriter) u32(value uint32) {
	var data [4]byte
	binary.LittleEndian.PutUint32(data[:], value)
	writer.data = append(writer.data, data[:]...)
}
func (writer *pointWriter) u64(value uint64) {
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], value)
	writer.data = append(writer.data, data[:]...)
}
func (writer *pointWriter) i64(value int64) { writer.u64(uint64(value)) }
func (writer *pointWriter) count(value int) error {
	if value < 0 || value > pointWireMaxItems {
		return errors.New("point wire item count exceeds its limit")
	}
	writer.u32(uint32(value))
	return nil
}
func (writer *pointWriter) string(value string) error {
	if !utf8.ValidString(value) {
		return errors.New("point wire string is not valid UTF-8")
	}
	return writer.rawString(value)
}

func (writer *pointWriter) rawString(value string) error {
	if uint64(len(value)) > math.MaxUint32 {
		return errors.New("point wire string is too large")
	}
	writer.u32(uint32(len(value)))
	if err := writer.reserve(len(value)); err != nil {
		return err
	}
	writer.data = append(writer.data, value...)
	return nil
}
func (writer *pointWriter) lengthPrefixed(encode func() error) error {
	if err := writer.reserve(4); err != nil {
		return err
	}
	offset := len(writer.data)
	writer.data = append(writer.data, 0, 0, 0, 0)
	start := len(writer.data)
	if err := encode(); err != nil {
		writer.data = writer.data[:offset]
		return err
	}
	length := len(writer.data) - start
	if uint64(length) > math.MaxUint32 {
		writer.data = writer.data[:offset]
		return errors.New("point wire record is too large")
	}
	binary.LittleEndian.PutUint32(writer.data[offset:offset+4], uint32(length))
	return nil
}

func pointWireNode(depth int, nodes *int) error {
	if depth > pointWireMaxDepth {
		return errors.New("point wire nesting exceeds its limit")
	}
	(*nodes)++
	if *nodes > pointWireMaxNodes {
		return errors.New("point wire node count exceeds its limit")
	}
	return nil
}

func (writer *pointWriter) point(point *Point, depth int, nodes *int) error {
	if err := pointWireNode(depth, nodes); err != nil {
		return err
	}
	timestampNanos := point.TimeUnixNano
	if point.Time != "" {
		timestamp, err := time.Parse(time.RFC3339Nano, point.Time)
		if err != nil {
			return fmt.Errorf("invalid point timestamp: %w", err)
		}
		timestampNanos = timestamp.UnixNano()
	}
	flags := uint16(0)
	if point.Dropped {
		flags |= 1
	}
	if point.Status != nil {
		flags |= 2
	}
	if !utf8.ValidString(point.Measurement) {
		flags |= 16
	}
	writer.u16(point.Version)
	writer.u16(flags)
	writer.i64(timestampNanos)
	if err := writer.string(point.Category); err != nil {
		return err
	}
	writeMeasurement := writer.string
	if flags&16 != 0 {
		writeMeasurement = writer.rawString
	}
	if err := writeMeasurement(point.Measurement); err != nil {
		return err
	}
	if point.Status != nil {
		if err := writer.string(*point.Status); err != nil {
			return err
		}
	}
	tagKeys := sortedStringKeys(point.Tags)
	if err := writer.count(len(tagKeys)); err != nil {
		return err
	}
	for _, key := range tagKeys {
		if err := pointWireNode(depth+1, nodes); err != nil {
			return err
		}
		if err := writer.string(key); err != nil {
			return err
		}
		if err := writer.string(point.Tags[key]); err != nil {
			return err
		}
	}
	fieldKeys := sortedAnyKeys(point.Fields)
	if err := writer.count(len(fieldKeys)); err != nil {
		return err
	}
	for _, key := range fieldKeys {
		if err := writer.string(key); err != nil {
			return err
		}
		if err := writer.value(point.Fields[key], depth+1, nodes); err != nil {
			return err
		}
	}
	if err := writer.count(len(point.Subpoints)); err != nil {
		return err
	}
	for index := range point.Subpoints {
		if err := writer.lengthPrefixed(func() error { return writer.point(&point.Subpoints[index], depth+1, nodes) }); err != nil {
			return err
		}
	}
	return nil
}

// Scalar references are an in-process encoding view. They never enter the wire.
func flatPointString(value any) (string, bool) {
	switch value := value.(type) {
	case string:
		return value, true
	case *string:
		if value != nil {
			return *value, true
		}
	}
	return "", false
}

func (writer *pointWriter) flatPointRecord(record *FlatPointRecord, nodes *int) error {
	if err := pointWireNode(0, nodes); err != nil {
		return err
	}
	writer.u16(record.Version)
	rawTags := false
	tagCount := 0
	flags := uint16(0)
	writeKey := writer.string
	if writer.rawPointKeys {
		// The scan below proves keys valid or selects raw-key mode. Both use
		// the same representation, so emission need not validate twice.
		writeKey = writer.rawString
	}
	for _, entry := range record.Entries {
		if entry.IsTag {
			tagCount++
			if writer.rawTagValues && !rawTags {
				if value, ok := flatPointString(entry.Value); ok && !utf8.ValidString(value) {
					rawTags = true
				}
			}
		}
		if writer.rawPointKeys && flags&8 == 0 && !utf8.ValidString(entry.Key) {
			flags |= 8
		}
	}
	if rawTags {
		flags |= 4
	}
	if !utf8.ValidString(record.Measurement) {
		flags |= 16
	}
	writer.u16(flags)
	writer.i64(record.TimeUnixNano)
	if err := writer.string(record.Category); err != nil {
		return err
	}
	// Measurement validity was checked when computing its wire flag.
	if err := writer.rawString(record.Measurement); err != nil {
		return err
	}
	if err := writer.count(tagCount); err != nil {
		return err
	}
	for _, entry := range record.Entries {
		if !entry.IsTag {
			continue
		}
		if err := pointWireNode(1, nodes); err != nil {
			return err
		}
		value, ok := flatPointString(entry.Value)
		if !ok {
			return fmt.Errorf("flat point tag %q is not a string", entry.Key)
		}
		if err := writeKey(entry.Key); err != nil {
			return err
		}
		writeTag := writer.string
		if writer.rawTagValues {
			writeTag = writer.rawString
		}
		if err := writeTag(value); err != nil {
			return err
		}
	}
	if err := writer.count(len(record.Entries) - tagCount); err != nil {
		return err
	}
	for _, entry := range record.Entries {
		if entry.IsTag {
			continue
		}
		if err := writeKey(entry.Key); err != nil {
			return err
		}
		if err := writer.value(entry.Value, 1, nodes); err != nil {
			return err
		}
	}
	return writer.count(0)
}

//nolint:funlen // Keep scalar encoding paths together in this hot type dispatcher.
func (writer *pointWriter) value(value any, depth int, nodes *int) error {
	if err := pointWireNode(depth, nodes); err != nil {
		return err
	}
	switch value := value.(type) {
	case nil:
		return writer.u8(0)
	case bool:
		if value {
			return writer.u8(2)
		}
		return writer.u8(1)
	case *int64:
		if value == nil {
			return fmt.Errorf("unsupported flat point value type %T", value)
		}
		writer.u8(3)
		writer.i64(*value)
		return nil
	case *uint64:
		if value == nil {
			return fmt.Errorf("unsupported flat point value type %T", value)
		}
		writer.u8(3)
		writer.i64(int64(*value))
		return nil
	case *float64:
		if value == nil {
			return fmt.Errorf("unsupported flat point value type %T", value)
		}
		writer.u8(4)
		writer.u64(math.Float64bits(*value))
		return nil
	case *bool:
		if value == nil {
			return fmt.Errorf("unsupported flat point value type %T", value)
		}
		if *value {
			return writer.u8(2)
		}
		return writer.u8(1)
	case *string:
		if value == nil {
			return fmt.Errorf("unsupported flat point value type %T", value)
		}
		text := *value
		if writer.rawStringValues {
			if utf8.ValidString(text) {
				writer.u8(5)
			} else {
				writer.u8(8)
			}
			return writer.rawString(text)
		}
		writer.u8(5)
		return writer.string(text)
	case *[]byte:
		if value == nil {
			return fmt.Errorf("unsupported flat point value type %T", value)
		}
		data := *value
		if writer.rawStringValues && !utf8.Valid(data) {
			writer.u8(8)
			return writer.rawString(string(data))
		}
		writer.u8(5)
		return writer.string(string(data))

	case int:
		writer.u8(3)
		writer.i64(int64(value))
		return nil
	case int8:
		writer.u8(3)
		writer.i64(int64(value))
		return nil
	case int16:
		writer.u8(3)
		writer.i64(int64(value))
		return nil
	case int32:
		writer.u8(3)
		writer.i64(int64(value))
		return nil
	case int64:
		writer.u8(3)
		writer.i64(value)
		return nil
	case uint:
		writer.u8(3)
		writer.i64(int64(value))
		return nil
	case uint8:
		writer.u8(3)
		writer.i64(int64(value))
		return nil
	case uint16:
		writer.u8(3)
		writer.i64(int64(value))
		return nil
	case uint32:
		writer.u8(3)
		writer.i64(int64(value))
		return nil
	case uint64:
		writer.u8(3)
		writer.i64(int64(value))
		return nil
	case float32:
		writer.u8(4)
		writer.u64(math.Float64bits(float64(value)))
		return nil
	case float64:
		writer.u8(4)
		writer.u64(math.Float64bits(value))
		return nil
	case string:
		if writer.rawStringValues {
			if utf8.ValidString(value) {
				writer.u8(5)
			} else {
				writer.u8(8)
			}
			return writer.rawString(value)
		}
		writer.u8(5)
		return writer.string(value)
	case RawString:
		writer.u8(8)
		return writer.rawString(string(value))
	case OpaqueGoBytes:
		writer.u8(9)
		return writer.rawString(string(value))
	case OpaqueGoUint:
		writer.u8(10)
		writer.u64(uint64(value))
		return nil
	case []byte:
		if writer.rawStringValues && !utf8.Valid(value) {
			writer.u8(8)
			return writer.rawString(string(value))
		}
		writer.u8(5)
		return writer.string(string(value))
	case []any:
		writer.u8(6)
		if err := writer.count(len(value)); err != nil {
			return err
		}
		for _, item := range value {
			if err := writer.value(item, depth+1, nodes); err != nil {
				return err
			}
		}
		return nil
	case []int64:
		return writePointSlice(writer, value, depth, nodes)
	case []int:
		return writeSignedPointSlice(writer, value, depth, nodes)
	case []int8:
		return writeSignedPointSlice(writer, value, depth, nodes)
	case []int16:
		return writeSignedPointSlice(writer, value, depth, nodes)
	case []int32:
		return writeSignedPointSlice(writer, value, depth, nodes)
	case []uint:
		return writeUnsignedPointSlice(writer, value, depth, nodes)
	case []uint16:
		return writeUnsignedPointSlice(writer, value, depth, nodes)
	case []uint32:
		return writeUnsignedPointSlice(writer, value, depth, nodes)
	case []uint64:
		return writeUnsignedPointSlice(writer, value, depth, nodes)
	case []float32:
		return writeFloatPointSlice(writer, value, depth, nodes)
	case []float64:
		return writePointSlice(writer, value, depth, nodes)
	case []bool:
		return writePointSlice(writer, value, depth, nodes)
	case []string:
		return writePointSlice(writer, value, depth, nodes)
	case [][]byte:
		return writeBytesPointSlice(writer, value, depth, nodes)
	case map[string]any:
		writer.u8(7)
		keys := sortedAnyKeys(value)
		if err := writer.count(len(keys)); err != nil {
			return err
		}
		for _, key := range keys {
			if err := writer.string(key); err != nil {
				return err
			}
			if err := writer.value(value[key], depth+1, nodes); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported flat point value type %T", value)
	}
}

// Point stores native typed arrays. Encode the same list representation as
// []any without allocating an intermediate converted slice. Each element still
// passes through value's scalar checks and the shared resource counters.
func writePointSlice[T int64 | float64 | bool | string](writer *pointWriter, values []T, depth int, nodes *int) error {
	writer.u8(6)
	if err := writer.count(len(values)); err != nil {
		return err
	}
	for _, value := range values {
		if err := writer.value(value, depth+1, nodes); err != nil {
			return err
		}
	}
	return nil
}

func writeSignedPointSlice[T int | int8 | int16 | int32](writer *pointWriter, values []T, depth int, nodes *int) error {
	writer.u8(6)
	if err := writer.count(len(values)); err != nil {
		return err
	}
	for _, value := range values {
		if err := writer.value(int64(value), depth+1, nodes); err != nil {
			return err
		}
	}
	return nil
}

func writeUnsignedPointSlice[T uint | uint16 | uint32 | uint64](writer *pointWriter, values []T, depth int, nodes *int) error {
	writer.u8(6)
	if err := writer.count(len(values)); err != nil {
		return err
	}
	for _, value := range values {
		// pipeline-go's normalizeListValue applies the same Go uint64 ->
		// int64 conversion, including two's-complement wrapping.
		if err := writer.value(int64(value), depth+1, nodes); err != nil {
			return err
		}
	}
	return nil
}

func writeFloatPointSlice(writer *pointWriter, values []float32, depth int, nodes *int) error {
	writer.u8(6)
	if err := writer.count(len(values)); err != nil {
		return err
	}
	for _, value := range values {
		if err := writer.value(float64(value), depth+1, nodes); err != nil {
			return err
		}
	}
	return nil
}

func writeBytesPointSlice(writer *pointWriter, values [][]byte, depth int, nodes *int) error {
	writer.u8(6)
	if err := writer.count(len(values)); err != nil {
		return err
	}
	for _, value := range values {
		if err := writer.value(value, depth+1, nodes); err != nil {
			return err
		}
	}
	return nil
}

func (writer *pointWriter) mutation(operation MutationOp, nodes *int) error {
	if err := pointWireNode(0, nodes); err != nil {
		return err
	}
	if operation.Kind < MutationSetField || operation.Kind > MutationSetRawMeasurement {
		return errors.New("unknown mutation operation")
	}
	writer.u8(byte(operation.Kind))
	switch operation.Kind {
	case MutationSetField:
		if err := writer.string(operation.Key); err != nil {
			return err
		}
		return writer.value(operation.Value, 0, nodes)
	case MutationDeleteField, MutationDeleteTag:
		return writer.string(operation.Key)
	case MutationSetTag:
		if err := writer.string(operation.Key); err != nil {
			return err
		}
		return writer.string(operation.String)
	case MutationSetRawTag:
		if err := writer.string(operation.Key); err != nil {
			return err
		}
		return writer.rawString(operation.String)
	case MutationSetCategory, MutationSetMeasurement, MutationSetStatus:
		return writer.string(operation.String)
	case MutationSetRawMeasurement:
		return writer.rawString(operation.String)
	case MutationSetTime:
		writer.i64(operation.Time)
		return nil
	case MutationClearStatus, MutationClearSubpoints:
		return nil
	case MutationSetDropped:
		if operation.Dropped {
			return writer.u8(1)
		}
		return writer.u8(0)
	case MutationAddSubpoint:
		if operation.Point == nil {
			return errors.New("add-subpoint mutation has no point")
		}
		return writer.lengthPrefixed(func() error { return writer.point(operation.Point, 0, nodes) })
	}
	return nil
}

type pointCursor struct {
	data   []byte
	offset int
}

func (cursor *pointCursor) take(length int) ([]byte, error) {
	if length < 0 || cursor.offset > len(cursor.data)-length {
		return nil, errors.New("point wire value is truncated")
	}
	value := cursor.data[cursor.offset : cursor.offset+length]
	cursor.offset += length
	return value, nil
}
func (cursor *pointCursor) u8() (byte, error) {
	data, err := cursor.take(1)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}
func (cursor *pointCursor) u16() (uint16, error) {
	data, err := cursor.take(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(data), nil
}
func (cursor *pointCursor) u32() (uint32, error) {
	data, err := cursor.take(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(data), nil
}
func (cursor *pointCursor) u64() (uint64, error) {
	data, err := cursor.take(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(data), nil
}
func (cursor *pointCursor) i64() (int64, error) { value, err := cursor.u64(); return int64(value), err }
func (cursor *pointCursor) count() (int, error) {
	value, err := cursor.u32()
	if err != nil {
		return 0, err
	}
	if value > pointWireMaxItems {
		return 0, errors.New("point wire item count exceeds its limit")
	}
	return int(value), nil
}
func (cursor *pointCursor) key(raw bool) (string, error) {
	if !raw {
		return cursor.string()
	}
	length, err := cursor.u32()
	if err != nil {
		return "", err
	}
	value, err := cursor.take(int(length))
	return string(value), err
}

func (cursor *pointCursor) string() (string, error) {
	length, err := cursor.u32()
	if err != nil {
		return "", err
	}
	value, err := cursor.take(int(length))
	if err != nil {
		return "", err
	}
	if !utf8.Valid(value) {
		return "", errors.New("point wire string is not valid UTF-8")
	}
	return string(value), nil
}
func (cursor *pointCursor) lengthPrefixed() (*pointCursor, error) {
	length, err := cursor.u32()
	if err != nil {
		return nil, err
	}
	value, err := cursor.take(int(length))
	if err != nil {
		return nil, err
	}
	return &pointCursor{data: value}, nil
}
func (cursor *pointCursor) done() bool { return cursor.offset == len(cursor.data) }

func (cursor *pointCursor) point(depth int, nodes *int) (Point, error) {
	if err := pointWireNode(depth, nodes); err != nil {
		return Point{}, err
	}
	version, err := cursor.u16()
	if err != nil {
		return Point{}, err
	}
	flags, err := cursor.u16()
	if err != nil || flags & ^uint16(31) != 0 {
		return Point{}, errors.New("flat point flags are invalid")
	}
	timestamp, err := cursor.i64()
	if err != nil {
		return Point{}, err
	}
	category, err := cursor.string()
	if err != nil {
		return Point{}, err
	}
	var measurement string
	if flags&16 != 0 {
		var length uint32
		length, err = cursor.u32()
		if err == nil {
			var bytes []byte
			bytes, err = cursor.take(int(length))
			measurement = string(bytes)
		}
	} else {
		measurement, err = cursor.string()
	}
	if err != nil {
		return Point{}, err
	}
	point := Point{Version: version, Category: category, Measurement: measurement, Time: time.Unix(0, timestamp).UTC().Format(time.RFC3339Nano), TimeUnixNano: timestamp, Tags: make(map[string]string), Fields: make(map[string]any), Dropped: flags&1 != 0}
	if flags&2 != 0 {
		status, err := cursor.string()
		if err != nil {
			return Point{}, err
		}
		point.Status = &status
	}
	tags, err := cursor.count()
	if err != nil {
		return Point{}, err
	}
	for index := 0; index < tags; index++ {
		if err := pointWireNode(depth+1, nodes); err != nil {
			return Point{}, err
		}
		key, err := cursor.key(flags&8 != 0)
		if err != nil {
			return Point{}, err
		}
		var value string
		if flags&4 != 0 {
			var length uint32
			length, err = cursor.u32()
			if err == nil {
				var bytes []byte
				bytes, err = cursor.take(int(length))
				value = string(bytes)
			}
		} else {
			value, err = cursor.string()
		}
		if err != nil {
			return Point{}, err
		}
		point.Tags[key] = value
	}
	fields, err := cursor.count()
	if err != nil {
		return Point{}, err
	}
	for index := 0; index < fields; index++ {
		key, err := cursor.key(flags&8 != 0)
		if err != nil {
			return Point{}, err
		}
		value, err := cursor.value(depth+1, nodes)
		if err != nil {
			return Point{}, err
		}
		point.Fields[key] = value
	}
	subpoints, err := cursor.count()
	if err != nil {
		return Point{}, err
	}
	point.Subpoints = make([]Point, 0, subpoints)
	for index := 0; index < subpoints; index++ {
		subpointCursor, err := cursor.lengthPrefixed()
		if err != nil {
			return Point{}, err
		}
		subpoint, err := subpointCursor.point(depth+1, nodes)
		if err != nil {
			return Point{}, err
		}
		if !subpointCursor.done() {
			return Point{}, errors.New("flat subpoint contains trailing bytes")
		}
		point.Subpoints = append(point.Subpoints, subpoint)
	}
	return point, nil
}

func (cursor *pointCursor) value(depth int, nodes *int) (any, error) {
	if err := pointWireNode(depth, nodes); err != nil {
		return nil, err
	}
	kind, err := cursor.u8()
	if err != nil {
		return nil, err
	}
	switch kind {
	case 0:
		return nil, nil
	case 1:
		return false, nil
	case 2:
		return true, nil
	case 3:
		return cursor.i64()
	case 4:
		value, err := cursor.u64()
		return math.Float64frombits(value), err
	case 5:
		return cursor.string()
	case 8:
		length, err := cursor.u32()
		if err != nil {
			return nil, err
		}
		value, err := cursor.take(int(length))
		if err != nil {
			return nil, err
		}
		return RawString(string(value)), nil
	case 9:
		length, err := cursor.u32()
		if err != nil {
			return nil, err
		}
		value, err := cursor.take(int(length))
		if err != nil {
			return nil, err
		}
		return OpaqueGoBytes(string(value)), nil
	case 10:
		value, err := cursor.u64()
		return OpaqueGoUint(value), err
	case 6:
		count, err := cursor.count()
		if err != nil {
			return nil, err
		}
		values := make([]any, 0, count)
		for index := 0; index < count; index++ {
			value, err := cursor.value(depth+1, nodes)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return values, nil
	case 7:
		count, err := cursor.count()
		if err != nil {
			return nil, err
		}
		values := make(map[string]any, count)
		for index := 0; index < count; index++ {
			key, err := cursor.string()
			if err != nil {
				return nil, err
			}
			value, err := cursor.value(depth+1, nodes)
			if err != nil {
				return nil, err
			}
			values[key] = value
		}
		return values, nil
	default:
		return nil, errors.New("unknown flat point value type")
	}
}

func (cursor *pointCursor) mutation(nodes *int) (MutationOp, error) {
	return cursor.mutationWithDictionary(nodes, nil)
}

func (cursor *pointCursor) mutationWithDictionary(nodes *int, dictionary []string) (MutationOp, error) {
	readKey := func(raw bool) (string, error) {
		if dictionary == nil {
			return cursor.key(raw)
		}
		id, err := cursor.u32()
		if err != nil || uint64(id) >= uint64(len(dictionary)) {
			return "", errors.New("invalid dynamic dictionary key ID")
		}
		key := dictionary[id]
		if !raw && !utf8.ValidString(key) {
			return "", errors.New("invalid UTF-8 dynamic key")
		}
		return key, nil
	}
	if err := pointWireNode(0, nodes); err != nil {
		return MutationOp{}, err
	}
	kind, err := cursor.u8()
	if err != nil {
		return MutationOp{}, err
	}
	rawKey := kind&128 != 0
	kind &= 127
	if rawKey && kind != 1 && kind != 2 && kind != 3 && kind != 4 && kind != 13 {
		return MutationOp{}, errors.New("raw key bit on keyless mutation")
	}
	operation := MutationOp{Kind: MutationOpKind(kind)}
	switch operation.Kind {
	case MutationSetField:
		operation.Key, err = readKey(rawKey)
		if err == nil {
			operation.Value, err = cursor.value(0, nodes)
		}
	case MutationDeleteField, MutationDeleteTag:
		operation.Key, err = readKey(rawKey)
	case MutationSetTag:
		operation.Key, err = readKey(rawKey)
		if err == nil {
			operation.String, err = cursor.string()
		}
	case MutationSetCategory, MutationSetMeasurement, MutationSetStatus:
		operation.String, err = cursor.string()
	case MutationSetRawMeasurement:
		var length uint32
		length, err = cursor.u32()
		if err == nil {
			var bytes []byte
			bytes, err = cursor.take(int(length))
			operation.String = string(bytes)
		}
	case MutationSetRawTag:
		operation.Key, err = readKey(rawKey)
		if err == nil {
			var length uint32
			length, err = cursor.u32()
			if err == nil {
				var bytes []byte
				bytes, err = cursor.take(int(length))
				operation.String = string(bytes)
			}
		}
	case MutationSetTime:
		operation.Time, err = cursor.i64()
	case MutationClearStatus, MutationClearSubpoints:
	case MutationSetDropped:
		var value byte
		value, err = cursor.u8()
		if err == nil {
			if value > 1 {
				return MutationOp{}, errors.New("invalid dropped mutation value")
			}
			operation.Dropped = value == 1
		}
	case MutationAddSubpoint:
		var nested *pointCursor
		nested, err = cursor.lengthPrefixed()
		if err == nil {
			var point Point
			point, err = nested.point(0, nodes)
			if err == nil && !nested.done() {
				err = errors.New("mutation subpoint contains trailing bytes")
			}
			operation.Point = &point
		}
	default:
		err = errors.New("unknown mutation operation")
	}
	return operation, err
}

func pointEnvelope(input []byte, magic string, version uint16) (int, *pointCursor, error) {
	if len(input) < pointWireHeaderBytes || len(input) > pointWireHeaderBytes+pointWireMaxBytes {
		return 0, nil, errors.New("point wire envelope length is invalid")
	}
	if string(input[:4]) != magic || binary.LittleEndian.Uint16(input[4:6]) != version || binary.LittleEndian.Uint16(input[6:8]) != 0 {
		return 0, nil, errors.New("point wire envelope identity is invalid")
	}
	count := binary.LittleEndian.Uint32(input[8:12])
	length := binary.LittleEndian.Uint32(input[12:16])
	if count > pointWireMaxItems || int(length) != len(input)-pointWireHeaderBytes {
		return 0, nil, errors.New("point wire envelope counts are invalid")
	}
	return int(count), &pointCursor{data: input[pointWireHeaderBytes:]}, nil
}

func sortedStringKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
func sortedAnyKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

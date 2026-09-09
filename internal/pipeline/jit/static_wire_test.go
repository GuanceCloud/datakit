// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
)

func TestStaticPreparedPointFieldsMatchValuesAndOwnStorage(t *testing.T) {
	originalPool := point.GetPointPool()
	t.Cleanup(func() { point.SetPointPool(originalPool) })
	for _, pooled := range []bool{false, true} {
		if pooled {
			point.SetPointPool(point.NewReservedCapPointPool(8))
		} else {
			point.ClearPointPool()
		}
		for _, value := range []any{"text", "", "中文", RawString("a\xffb"), int64(-42), 1.25, true, nil, []any{int64(1), "two"}, map[string]any{"nested": []any{true}}} {
			for _, tag := range []bool{false, true} {
				state := StaticSetField
				if tag {
					state = StaticSetTag
				}
				writer := pointWriter{data: []byte{byte(TerminalOK), 0, 0, byte(state)}}
				nodes := 0
				if err := writer.value(value, 0, &nodes); err != nil {
					t.Fatal(err)
				}
				schema := StaticOutputSchema{Keys: []string{"field"}}
				input := encodeStaticTestBatch(schema.Hash, 0, writer.data)
				reference, referenceErr := decodeStaticBatch(input, schema)
				var scratch Batch
				scratch.PreparePointFields()
				got, err := decodeStaticBatchInto(input, schema, &scratch)
				if (err == nil) != (referenceErr == nil) {
					t.Fatalf("pooled=%v tag=%v value=%#v: acceptance differs: %v / %v", pooled, tag, value, err, referenceErr)
				}
				if err != nil {
					continue
				}
				if len(got.Static.Values) != 0 || len(got.Static.PointFields) != 1 || !got.Static.Validated() {
					t.Fatal("missing prepared field or populated legacy values")
				}
				field := got.Static.PointFields[0]
				want := point.NewKV("field", reference.Static.Values[0], point.WithKVTagSet(tag))
				for range 2 {
					if _, err := decodeStaticBatchInto(input, schema, &scratch); err != nil {
						t.Fatal(err)
					}
				}
				clear(input)
				scratch.ResetForReuse()
				if !reflect.DeepEqual(field, want) {
					t.Fatalf("pooled=%v tag=%v value=%#v: field changed after wire/scratch reuse: %#v / %#v", pooled, tag, value, field, want)
				}
				if !scratch.preparePointFields {
					t.Fatal("reset lost adapter preference")
				}
			}
		}
	}
}

func TestStaticPreparedPointFieldsRejectMalformedResults(t *testing.T) {
	schema := StaticOutputSchema{Keys: []string{"field"}}
	writer := pointWriter{data: []byte{byte(TerminalOK), 0, 0, byte(StaticSetField)}}
	nodes := 0
	if err := writer.value("valid", 0, &nodes); err != nil {
		t.Fatal(err)
	}
	input := encodeStaticTestBatch(schema.Hash, 0, writer.data)
	check := func(data []byte) {
		t.Helper()
		_, referenceErr := decodeStaticBatch(data, schema)
		var scratch Batch
		scratch.PreparePointFields()
		got, err := decodeStaticBatchInto(data, schema, &scratch)
		if (err == nil) != (referenceErr == nil) {
			t.Fatalf("decoder acceptance differs: %v / %v", err, referenceErr)
		}
		if err != nil && got.Static != nil {
			t.Fatal("malformed result published partial fields")
		}
	}
	for i := range len(input) {
		check(input[:i])
		changed := append([]byte(nil), input...)
		changed[i] ^= 0xff
		check(changed)
	}
}

func TestStaticPreparedPointFieldsPrefixErrorsAndLimits(t *testing.T) {
	schema := StaticOutputSchema{Keys: []string{"field"}}
	writer := pointWriter{data: []byte{byte(TerminalError), byte(staticRecordCommitPrefixError), 0, byte(StaticSetField)}}
	nodes := 0
	if err := writer.value("committed", 0, &nodes); err != nil {
		t.Fatal(err)
	}
	writer.data = append(writer.data, []byte(`{"code":"E_TEST"}`)...)
	input := encodeStaticTestBatch(schema.Hash, staticBatchHasErrors, writer.data)
	var scratch Batch
	scratch.PreparePointFields()
	batch, err := decodeStaticBatchInto(input, schema, &scratch)
	if err != nil || !batch.Records[0].CommitPrefixError || len(batch.Static.PointFields) != 1 {
		t.Fatalf("prefix error lost mutations: %v, %+v", err, batch)
	}
	field, payload := batch.Static.PointFields[0], batch.Records[0].Error
	clear(input)
	scratch.ResetForReuse()
	if field.GetS() != "committed" || string(payload) != `{"code":"E_TEST"}` {
		t.Fatal("prefix-error output retained borrowed wire storage")
	}
	for _, kind := range []byte{5, 8} {
		cursor := pointCursor{data: []byte{kind, 0, 0, 0, 0}}
		nodes, owned := pointWireMaxNodes, ""
		if _, err := cursor.preparedPointField("key", false, &owned, &nodes); err == nil {
			t.Fatal("prepared string bypassed node budget")
		}
	}
	scratch.Static.PointFields = make([]*point.Field, 0, 65537)
	scratch.ResetForReuse()
	if scratch.Static != nil || !scratch.preparePointFields {
		t.Fatal("oversized prepared result retained or lost adapter preference")
	}
}

func TestDecodeStaticBatchStrictValidation(t *testing.T) {
	var hash [32]byte
	hash[0] = 0x5a
	schema := StaticOutputSchema{Hash: hash, Keys: []string{"value"}}
	valid := encodeStaticTestBatch(hash, 0, []byte{byte(TerminalOK), 0, 0, byte(StaticNoop)})
	batch, err := decodeStaticBatch(valid, schema)
	if err != nil {
		t.Fatalf("decode valid static batch: %v", err)
	}
	if len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK || batch.Static == nil || len(batch.Static.States) != 1 {
		t.Fatalf("unexpected static batch: %#v", batch)
	}
	if len(batch.Static.StateOffsets) != 2 || batch.Static.StateOffsets[1] != 1 {
		t.Fatalf("unexpected static state offsets: %v", batch.Static.StateOffsets)
	}
	if !batch.Static.Validated() {
		t.Fatal("wire-decoded static batch is not marked validated")
	}

	tests := map[string]func([]byte){
		"schema-hash": func(value []byte) { value[24] ^= 1 },
		"unknown-state": func(value []byte) {
			value[staticBatchHeaderSize+4+3] = 5
		},
		"unused-state-bits": func(value []byte) {
			value[staticBatchHeaderSize+4+3] = 0x10
		},
		"error-flag": func(value []byte) {
			binary.LittleEndian.PutUint16(value[6:8], staticBatchHasErrors)
		},
		"record-length": func(value []byte) {
			binary.LittleEndian.PutUint32(value[staticBatchHeaderSize:staticBatchHeaderSize+4], 1)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			encoded := append([]byte(nil), valid...)
			mutate(encoded)
			if _, err := decodeStaticBatch(encoded, schema); err == nil {
				t.Fatal("tampered static batch was accepted")
			}
		})
	}
}

func TestDecodeStaticRawTagBytes(t *testing.T) {
	var hash [32]byte
	schema := StaticOutputSchema{Hash: hash, Keys: []string{"tag"}}
	for _, bytes := range [][]byte{{}, {0}, {0xff}, {0xe2, 0x82}, []byte("valid")} {
		record := []byte{byte(TerminalOK), 0, 0, byte(StaticSetTag), 8}
		record = binary.LittleEndian.AppendUint32(record, uint32(len(bytes)))
		record = append(record, bytes...)
		encoded := encodeStaticTestBatch(hash, 0, record)
		batch, err := decodeStaticBatch(encoded, schema)
		if err != nil {
			t.Fatal(err)
		}
		for i := range encoded {
			encoded[i] = 0
		}
		if got := batch.Static.Values[0]; got != string(bytes) {
			t.Fatalf("tag bytes changed: %#v", got)
		}
	}
	truncated := []byte{byte(TerminalOK), 0, 0, byte(StaticSetTag), 8, 2, 0, 0, 0, 0xff}
	if _, err := decodeStaticBatch(encodeStaticTestBatch(hash, 0, truncated), schema); err == nil {
		t.Fatal("accepted truncated raw tag")
	}
}

func TestDecodeStaticBatchDetachesErrorFromWireBuffer(t *testing.T) {
	var hash [32]byte
	schema := StaticOutputSchema{Hash: hash, Keys: []string{"value"}}
	errorPayload := []byte(`{"code":"E_TEST"}`)
	record := append([]byte{byte(TerminalError), 0, 0}, errorPayload...)
	encoded := encodeStaticTestBatch(hash, staticBatchHasErrors, record)
	batch, err := decodeStaticBatch(encoded, schema)
	if err != nil {
		t.Fatalf("decode static error: %v", err)
	}
	for index := range encoded {
		encoded[index] = 0
	}
	if got := string(batch.Records[0].Error); got != string(errorPayload) {
		t.Fatalf("error payload retained wire storage: got %q", got)
	}
}

func TestDecodeStaticBatchCommittedPrefixError(t *testing.T) {
	var hash [32]byte
	schema := StaticOutputSchema{Hash: hash, Keys: []string{"value"}}
	errorPayload := []byte(`{"code":"E_TEST"}`)
	record := []byte{byte(TerminalError), staticRecordCommitPrefixError, 0, byte(StaticSetField), 3}
	record = binary.LittleEndian.AppendUint64(record, 42)
	record = append(record, errorPayload...)
	batch, err := decodeStaticBatch(encodeStaticTestBatch(hash, staticBatchHasErrors, record), schema)
	if err != nil {
		t.Fatalf("decode committed static error: %v", err)
	}
	if !batch.Records[0].CommitPrefixError || string(batch.Records[0].Error) != string(errorPayload) {
		t.Fatalf("unexpected committed static error: %#v", batch.Records[0])
	}
	if len(batch.Static.States) != 1 || batch.Static.States[0] != StaticSetField ||
		len(batch.Static.Values) != 1 || batch.Static.Values[0] != int64(42) {
		t.Fatalf("unexpected committed static mutations: %#v", batch.Static)
	}
}

func TestDecodeStaticTerminalDoesNotAllocateSlotStates(t *testing.T) {
	var hash [32]byte
	schema := StaticOutputSchema{Hash: hash, Keys: []string{"value"}}
	batch, err := decodeStaticBatch(
		encodeStaticTestBatch(hash, 0, []byte{byte(TerminalDropped), 0, 0}),
		schema,
	)
	if err != nil {
		t.Fatalf("decode dropped static record: %v", err)
	}
	if cap(batch.Static.States) != 0 || cap(batch.Static.Values) != 0 || len(batch.Static.StateOffsets) != 2 ||
		batch.Static.StateOffsets[1] != 0 {
		t.Fatalf("terminal record allocated slot states: %#v", batch.Static)
	}
}

func TestDecodeStaticBatchRejectsNonStringTag(t *testing.T) {
	var hash [32]byte
	schema := StaticOutputSchema{Hash: hash, Keys: []string{"value"}}
	payload := []byte{byte(TerminalOK), 0, 0, byte(StaticSetTag), 3}
	payload = binary.LittleEndian.AppendUint64(payload, 42)
	if _, err := decodeStaticBatch(encodeStaticTestBatch(hash, 0, payload), schema); err == nil {
		t.Fatal("integer static tag value was accepted")
	}
}

//nolint:makezero // The initialized wire header must precede appended payloads.
func encodeStaticTestBatch(hash [32]byte, flags uint16, record []byte) []byte {
	payload := binary.LittleEndian.AppendUint32(nil, uint32(len(record)))
	payload = append(payload, record...)
	encoded := make([]byte, staticBatchHeaderSize)
	copy(encoded[:4], "PPS2")
	binary.LittleEndian.PutUint16(encoded[4:6], staticBatchABIVersion)
	binary.LittleEndian.PutUint16(encoded[6:8], flags)
	binary.LittleEndian.PutUint32(encoded[8:12], 1)
	binary.LittleEndian.PutUint32(encoded[12:16], 1)
	binary.LittleEndian.PutUint32(encoded[16:20], uint32(len(payload)))
	copy(encoded[24:56], hash[:])
	return append(encoded, payload...)
}

func TestStaticBatchReuseClearsPreviousResults(t *testing.T) {
	schema := StaticOutputSchema{Keys: []string{"field"}}
	var scratch Batch
	field := []byte{byte(TerminalOK), 0, 0, byte(StaticSetField), 5}
	field = binary.LittleEndian.AppendUint32(field, 5)
	field = append(field, "owned"...)
	input := encodeStaticTestBatch(schema.Hash, 0, field)
	independent, err := decodeStaticBatch(input, schema)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		batch, err := decodeStaticBatchInto(input, schema, &scratch)
		if err != nil || len(batch.Static.Values) != 1 || batch.Static.Values[0] != "owned" {
			t.Fatalf("reuse: %+v %v", batch, err)
		}
		dropped := encodeStaticTestBatch(schema.Hash, 0, []byte{byte(TerminalDropped), 0, 0})
		batch, err = decodeStaticBatchInto(dropped, schema, &scratch)
		if err != nil || len(batch.Static.Values) != 0 || len(batch.Static.States) != 0 || batch.Records[0].Status != TerminalDropped {
			t.Fatalf("stale mutation: %+v %v", batch, err)
		}
		for _, value := range scratch.Static.Values[:cap(scratch.Static.Values)] {
			if value != nil {
				t.Fatal("result pool retains previous field value")
			}
		}
		broken := append([]byte(nil), input...)
		broken[len(broken)-1] = 0xff
		if _, err := decodeStaticBatchInto(broken, schema, &scratch); err == nil {
			t.Fatal("invalid UTF-8 accepted after reuse")
		}
	}
	if independent.Static.Values[0] != "owned" {
		t.Fatal("reusing scratch changed independent result")
	}
	scratch.ResetForReuse()
	if scratch.Static.Validated() || len(scratch.Records) != 0 || scratch.Static.Schema.Keys != nil {
		t.Fatal("reset retained result metadata")
	}
	scratch.Records = make([]Record, 1, 4097)
	scratch.ResetForReuse()
	if scratch.Records != nil || scratch.Static != nil {
		t.Fatal("oversized result retained in pool")
	}
}

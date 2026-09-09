// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"encoding/binary"
	"reflect"
	"testing"
)

//nolint:makezero // The initialized wire header must precede appended payloads.
func dynamicTestBatch(tb testing.TB, status TerminalStatus, flags uint16, keys []string, frames ...indexedTestFrame) []byte {
	tb.Helper()
	frames = append(frames, indexedTestFrame{kind: 6, status: status, flags: flags})
	indexed := encodeIndexedTestBatch(tb, frames...)
	output := make([]byte, 56)
	copy(output, indexed[:20])
	copy(output[:4], "PPS2")
	binary.LittleEndian.PutUint16(output[6:8], 0)
	if status == TerminalError || status == TerminalCancelled {
		binary.LittleEndian.PutUint16(output[6:8], 1)
	}
	output = append(output, indexed[20:]...)
	binary.LittleEndian.PutUint32(output[20:24], uint32(len(output)))
	output = binary.LittleEndian.AppendUint32(output, uint32(len(keys)))
	for _, key := range keys {
		output = binary.LittleEndian.AppendUint32(output, uint32(len(key)))
		output = append(output, key...)
	}
	binary.LittleEndian.PutUint32(output[16:20], uint32(len(output)-56))
	return output
}

func TestDynamicBatchTypedMutationsAndOwnership(t *testing.T) {
	for _, key := range []string{"field", "", "raw\xffkey"} {
		value := map[string]any{"nested": []any{int64(42), "owned text"}}
		writer := pointWriter{data: []byte{1, 0, 0, 0, 1}}
		if key == "raw\xffkey" {
			writer.data[4] |= 128
		}
		writer.data = binary.LittleEndian.AppendUint32(writer.data, 0)
		nodes := 0
		if err := writer.value(value, 0, &nodes); err != nil {
			t.Fatal(err)
		}
		encoded := dynamicTestBatch(t, TerminalError, 1, []string{key},
			indexedTestFrame{kind: 7, category: 1, codec: 7, payload: writer.data},
			indexedTestFrame{kind: 5, category: 3, codec: 5, payload: []byte(`{"code":"E_TEST"}`)},
			indexedTestFrame{kind: 4, category: 3, codec: 4, payload: []byte("diagnostic")})
		batch, err := decodeDynamicBatchInto(encoded, StaticOutputSchema{Dynamic: true}, nil)
		if err != nil {
			t.Fatal(err)
		}
		clear(encoded)
		record := batch.Records[0]
		if !record.CommitPrefixError || record.Mutations[0].Key != key || !reflect.DeepEqual(record.Mutations[0].Value, value) || string(record.Error) != `{"code":"E_TEST"}` || string(batch.Diagnostics[0]) != "diagnostic" {
			t.Fatalf("lost owned output: %#v", batch)
		}
	}
}

func TestDynamicBatchRejectsMalformedAndTruncatedOutput(t *testing.T) {
	payload := []byte{1, 0, 0, 0, 2, 0, 0, 0, 0} // DeleteField, dictionary ID 0
	valid := dynamicTestBatch(t, TerminalOK, 0, []string{"key"}, indexedTestFrame{kind: 7, category: 1, codec: 7, payload: payload})
	schema := StaticOutputSchema{Dynamic: true}
	if _, err := decodeDynamicBatchInto(valid, schema, nil); err != nil {
		t.Fatal(err)
	}
	for n := 0; n < len(valid); n++ {
		if _, err := decodeDynamicBatchInto(valid[:n], schema, nil); err == nil {
			t.Fatalf("accepted truncation %d", n)
		}
	}
	changes := map[string]func([]byte){
		"version": func(b []byte) { b[4] = 2 }, "flags": func(b []byte) { b[6] = 2 },
		"hash": func(b []byte) { b[24] = 1 }, "dictionary offset": func(b []byte) { binary.LittleEndian.PutUint32(b[20:24], 55) },
		"key ID": func(b []byte) { b[56+20+5] = 1 }, "operation": func(b []byte) { b[56+20+4] = 127 },
		"reserved": func(b []byte) { b[56+14] = 1 }, "terminal": func(b []byte) { b[56+20+len(payload)+9] = 0 },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			bad := append([]byte(nil), valid...)
			change(bad)
			if _, err := decodeDynamicBatchInto(bad, schema, nil); err == nil {
				t.Fatal("accepted malformed output")
			}
		})
	}
	duplicate := dynamicTestBatch(t, TerminalOK, 0, []string{"key", "key"}, indexedTestFrame{kind: 7, category: 1, codec: 7, payload: payload})
	if _, err := decodeDynamicBatchInto(duplicate, schema, nil); err == nil {
		t.Fatal("accepted duplicate dictionary")
	}
	empty := dynamicTestBatch(t, TerminalOK, 0, nil, indexedTestFrame{kind: 7, category: 1, codec: 7, payload: []byte{0, 0, 0, 0}})
	var scratch Batch
	batch, err := decodeDynamicBatchInto(empty, schema, &scratch)
	if err != nil || !batch.Records[0].HasMutations() {
		t.Fatalf("empty valid mutation: %v", err)
	}
	dropped := dynamicTestBatch(t, TerminalDropped, 0, nil)
	batch, err = decodeDynamicBatchInto(dropped, schema, &scratch)
	if err != nil || batch.Records[0].HasMutations() {
		t.Fatalf("stale mutations after scratch reuse: %v", err)
	}
}

func FuzzDynamicBatchDecode(f *testing.F) {
	f.Add(dynamicTestBatch(f, TerminalOK, 0, []string{"key"}, indexedTestFrame{kind: 7, category: 1, codec: 7, payload: []byte{1, 0, 0, 0, 2, 0, 0, 0, 0}}))
	f.Fuzz(func(t *testing.T, input []byte) {
		_, _ = decodeDynamicBatchInto(input, StaticOutputSchema{Dynamic: true}, nil)
	})
}

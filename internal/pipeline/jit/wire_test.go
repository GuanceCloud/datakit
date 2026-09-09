// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"bytes"
	"math"
	"testing"
)

func TestEncodeFlatPointRecordsMatchesFlatPoints(t *testing.T) {
	fields := map[string]any{
		"active": true,
		"count":  int64(42),
		"meta":   map[string]any{"name": "value"},
	}
	want, err := EncodeFlatPoints([]Point{{
		Version:      FlatPointABIVersion,
		Category:     "logging",
		Measurement:  "projected",
		Tags:         map[string]string{"env": "prod", "host": "node-1"},
		Fields:       fields,
		TimeUnixNano: 123456789,
	}})
	if err != nil {
		t.Fatalf("encode flat point: %v", err)
	}
	got, err := EncodeFlatPointRecords([]FlatPointRecord{{
		Version:      FlatPointABIVersion,
		Category:     "logging",
		Measurement:  "projected",
		TimeUnixNano: 123456789,
		Entries: []FlatPointEntry{
			{Key: "env", Value: "prod", IsTag: true},
			{Key: "host", Value: "node-1", IsTag: true},
			{Key: "active", Value: true},
			{Key: "count", Value: int64(42)},
			{Key: "meta", Value: map[string]any{"name": "value"}},
		},
	}})
	if err != nil {
		t.Fatalf("encode projected flat point: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("projected flat point encoding differs from the general encoder")
	}
}

func TestEncodeProjectedPointRecordsIntoReusesBuffer(t *testing.T) {
	records := []FlatPointRecord{{
		Version: FlatPointABIVersion, Category: "logging", Measurement: "reuse",
		Entries: []FlatPointEntry{{Key: "message", Value: "first"}},
	}}
	projection := newKeyProjection([]string{"message"})
	fresh, err := EncodeProjectedPointRecords(records, projection)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := EncodeProjectedPointRecordsInto(make([]byte, 0, cap(fresh)+64), records, projection)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, fresh) {
		t.Fatal("reused encoding differs from fresh encoding")
	}
	start := &encoded[0]
	records[0].Entries[0].Value = "second"
	reused, err := EncodeProjectedPointRecordsInto(encoded[:0], records, projection)
	if err != nil {
		t.Fatal(err)
	}
	if &reused[0] != start {
		t.Fatal("encoder did not reuse caller-owned storage")
	}
	decoded, err := DecodeFlatPoints(reused)
	if err != nil {
		t.Fatal(err)
	}
	if got := decoded[0].Fields["message"]; got != "second" {
		t.Fatalf("decoded reused value = %#v", got)
	}
}

func TestTypedPointArraysMatchGenericListEncoding(t *testing.T) {
	for _, tc := range []struct {
		name    string
		typed   any
		generic []any
	}{
		{"int", []int64{-1, 0, 42}, []any{int64(-1), int64(0), int64(42)}},
		{"float", []float64{1.5, -2}, []any{1.5, float64(-2)}},
		{"bool", []bool{true, false}, []any{true, false}},
		{"string", []string{"中", "", "a\x00b"}, []any{"中", "", "a\x00b"}},
		{"nil", []int64(nil), []any{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encode := func(value any) ([]byte, error) {
				return EncodeFlatPoints([]Point{{Version: FlatPointABIVersion, Category: "logging", Measurement: "arrays", Fields: map[string]any{"array": value}}})
			}
			got, err := encode(tc.typed)
			if err != nil {
				t.Fatal(err)
			}
			want, err := encode(tc.generic)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatal("typed array changed wire representation")
			}
		})
	}
	writer := &pointWriter{}
	nodes := 0
	if err := writer.value([]string{string([]byte{0xff})}, 0, &nodes); err == nil {
		t.Fatal("invalid UTF8 bypassed scalar validation")
	}
	nodes = pointWireMaxNodes - 1
	if err := (&pointWriter{}).value([]int64{1}, 0, &nodes); err == nil {
		t.Fatal("typed element bypassed node budget")
	}
	nodes = 0
	if err := (&pointWriter{}).value([]bool{true}, pointWireMaxDepth, &nodes); err == nil {
		t.Fatal("typed element bypassed depth budget")
	}
}

func TestEncodeFlatPointRecordsRejectsNonStringTag(t *testing.T) {
	_, err := EncodeFlatPointRecords([]FlatPointRecord{{
		Version: FlatPointABIVersion,
		Entries: []FlatPointEntry{{Key: "invalid", Value: int64(1), IsTag: true}},
	}})
	if err == nil {
		t.Fatal("non-string projected tag was accepted")
	}
}

func TestFlatRecordMetadataMatchesGeneralEncoder(t *testing.T) {
	for mask := 0; mask < 32; mask++ {
		rawKeys, rawTags := mask&1 != 0, mask&2 != 0
		key, tag, measurement := "b", "tag value", "measurement"
		if mask&4 != 0 {
			key += "\xff"
		}
		if mask&8 != 0 {
			tag += "\xff"
		}
		if mask&16 != 0 {
			measurement += "\xff"
		}
		record := FlatPointRecord{Version: 1, Category: "logging", Measurement: measurement,
			Entries: []FlatPointEntry{
				{Key: key, Value: int64(42)},
				{Key: "a", Value: tag, IsTag: true},
				{Key: "c", Value: "field value"},
				{Key: "z", Value: "last tag", IsTag: true},
			},
		}
		projection := InputProjection{rawPointKeys: rawKeys, rawTagValues: rawTags}
		got, gotErr := EncodeProjectedPointRecords([]FlatPointRecord{record}, projection)
		rejected := (mask&4 != 0 && !rawKeys) || (mask&8 != 0 && !rawTags)
		if (gotErr != nil) != rejected {
			t.Fatalf("mask=%d: error=%v", mask, gotErr)
		}
		if gotErr != nil {
			continue
		}
		decoded, err := DecodeFlatPoints(got)
		if err != nil || len(decoded) != 1 {
			t.Fatalf("mask=%d: decode=%v", mask, err)
		}
		actual := decoded[0]
		if len(actual.Fields) != 2 || len(actual.Tags) != 2 || actual.Fields[key] != int64(42) || actual.Fields["c"] != "field value" || actual.Tags["a"] != tag || actual.Tags["z"] != "last tag" || actual.Measurement != measurement {
			t.Fatalf("mask=%d: decoded Point differs", mask)
		}
		expectedFlags := byte(0)
		if mask&4 != 0 {
			expectedFlags |= 8
		}
		if mask&8 != 0 {
			expectedFlags |= 4
		}
		if mask&16 != 0 {
			expectedFlags |= 16
		}
		if got[pointWireHeaderBytes+4+2] != expectedFlags {
			t.Fatalf("mask=%d: wrong metadata flags", mask)
		}
		// The general map encoder is intentionally strict for keys and tags.
		// Compare its exact bytes only on inputs it can represent.
		if mask&12 != 0 {
			continue
		}
		writer, err := newPointWriter("PPF1", FlatPointABIVersion, 1)
		if err != nil {
			t.Fatal(err)
		}
		writer.rawPointKeys, writer.rawTagValues = rawKeys, rawTags
		point := Point{Version: 1, Category: "logging", Measurement: measurement,
			Fields: map[string]any{key: int64(42), "c": "field value"},
			Tags:   map[string]string{"a": tag, "z": "last tag"},
		}
		nodes := 0
		wantErr := writer.lengthPrefixed(func() error { return writer.point(&point, 0, &nodes) })
		if (gotErr == nil) != (wantErr == nil) {
			t.Fatalf("mask=%d: flat=%v general=%v", mask, gotErr, wantErr)
		}
		if gotErr != nil {
			if gotErr.Error() != wantErr.Error() {
				t.Fatalf("mask=%d: flat=%v general=%v", mask, gotErr, wantErr)
			}
			continue
		}
		want, err := writer.finish()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("mask=%d: wire mismatch", mask)
		}
	}
}

func TestScalarWireViewsPreserveBytesAndSnapshot(t *testing.T) {
	signed, unsigned := int64(-9223372036854775808), ^uint64(0)
	number := math.Float64frombits(0x7ff800000000002a)
	truth, text, tag := true, "raw\xfftext", "raw\xfftag"
	data := []byte{0xff, 0x00, 'x'}
	values := []any{signed, unsigned, number, truth, text, data}
	views := []any{&signed, &unsigned, &number, &truth, &text, &data}
	makeRecord := func(values []any, tag any) FlatPointRecord {
		entries := make([]FlatPointEntry, 0, len(values)+1)
		for i, value := range values {
			entries = append(entries, FlatPointEntry{Key: string(rune('a' + i)), Value: value})
		}
		entries = append(entries, FlatPointEntry{Key: "tag", Value: tag, IsTag: true})
		return FlatPointRecord{Version: 1, Category: "logging", Measurement: "snapshot", Entries: entries}
	}
	for mask := 0; mask < 4; mask++ {
		projection := InputProjection{rawStringValues: mask&1 != 0, rawTagValues: mask&2 != 0, rawPointKeys: true}
		want, wantErr := EncodeProjectedPointRecords([]FlatPointRecord{makeRecord(values, tag)}, projection)
		got, gotErr := EncodeProjectedPointRecords([]FlatPointRecord{makeRecord(views, &tag)}, projection)
		if (gotErr == nil) != (wantErr == nil) {
			t.Fatalf("mask=%d: view=%v value=%v", mask, gotErr, wantErr)
		}
		if gotErr != nil {
			if gotErr.Error() != wantErr.Error() {
				t.Fatalf("mask=%d: view=%v value=%v", mask, gotErr, wantErr)
			}
			continue
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("mask=%d: scalar view changed wire", mask)
		}
		want = append([]byte(nil), want...) // Freeze the expected snapshot before mutation.
		signed, unsigned, number, truth, text, tag, data[0] = 1, 2, 3, false, "changed", "changed", 0
		if !bytes.Equal(got, want) {
			t.Fatal("wire aliases scalar sources")
		}
		if _, err := DecodeFlatPoints(got); err != nil {
			t.Fatal(err)
		}
	}
}

func TestScalarWireViewsCountOnceAndRejectNil(t *testing.T) {
	number, unsigned, decimal, truth, text, data := int64(1), uint64(2), 1.25, true, "text", []byte("bytes")
	for _, view := range []any{&number, &unsigned, &decimal, &truth, &text, &data} {
		var writer pointWriter
		nodes := pointWireMaxNodes - 1
		if err := writer.value(view, 1, &nodes); err != nil {
			t.Fatal(err)
		}
		if nodes != pointWireMaxNodes {
			t.Fatal("scalar view was not counted exactly once")
		}
		if err := writer.value(view, 1, &nodes); err == nil {
			t.Fatal("scalar view bypassed node ceiling")
		}
	}
	for _, view := range []any{(*int64)(nil), (*uint64)(nil), (*float64)(nil), (*bool)(nil), (*string)(nil), (*[]byte)(nil)} {
		var writer pointWriter
		nodes := 0
		if err := writer.value(view, 1, &nodes); err == nil {
			t.Fatal("nil scalar view accepted")
		}
	}
}

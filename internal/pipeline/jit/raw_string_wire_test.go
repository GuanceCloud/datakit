// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRawStringCapabilityDefaultsToDisabled(t *testing.T) {
	for _, tc := range []struct {
		descriptor string
		want       bool
	}{
		{`{"version":1}`, false},
		{`{"version":1,"raw_string_values":false}`, false},
		{`{"version":1,"raw_string_values":true}`, true},
	} {
		var capabilities ProgramCapabilities
		if err := json.Unmarshal([]byte(tc.descriptor), &capabilities); err != nil {
			t.Fatal(err)
		}
		if capabilities.RawStringValues != tc.want {
			t.Fatalf("%s: got %v", tc.descriptor, capabilities.RawStringValues)
		}
	}
}

func TestRawStringWireRoundTrip(t *testing.T) {
	original := Point{Version: 1, Category: "logging", Measurement: "raw",
		Fields: map[string]any{
			"bytes":  RawString(string([]byte{0, 0xff, 0xb2, 0xe2})),
			"empty":  RawString(""),
			"text":   "测试",
			"nested": []any{map[string]any{"raw": RawString(string([]byte{0x80}))}},
		},
	}
	wire, err := EncodeFlatPoints([]Point{original})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeFlatPoints(wire)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 1 || !reflect.DeepEqual(decoded[0].Fields, original.Fields) {
		t.Fatalf("raw values changed: %#v", decoded)
	}
	deltas := []MutationDelta{{RecordIndex: 0, Operations: []MutationOp{{Kind: MutationSetField, Key: "raw", Value: original.Fields["bytes"]}}}}
	wire, err = EncodeMutationDeltas(deltas)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := DecodeMutationDeltas(wire)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, deltas) {
		t.Fatalf("raw mutation changed: %#v", actual)
	}
}

func TestRawTagMutationRoundTrip(t *testing.T) {
	for _, value := range []string{"", "\x00", "\xff", "\xe2\x82", "valid"} {
		want := []MutationDelta{{RecordIndex: 0, Operations: []MutationOp{{Kind: MutationSetRawTag, Key: "tag", String: value}}}}
		encoded, err := EncodeMutationDeltas(want)
		if err != nil {
			t.Fatal(err)
		}
		got, err := DecodeMutationDeltas(encoded)
		if err != nil {
			t.Fatal(err)
		}
		for i := range encoded {
			encoded[i] = 0
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got=%#v want=%#v", got, want)
		}
		pt := Point{}
		if err := got[0].Apply(&pt); err != nil || pt.Tags["tag"] != value {
			t.Fatalf("apply=%#v err=%v", pt, err)
		}
	}
	if _, err := EncodeMutationDeltas([]MutationDelta{{Operations: []MutationOp{{Kind: MutationSetRawTag, Key: "\xff", String: "value"}}}}); err == nil {
		t.Fatal("invalid metadata accepted")
	}
	for _, input := range [][]byte{{13}, {13, 0, 0, 0, 0, 2, 0, 0, 0, 255}} {
		cursor := pointCursor{data: input}
		nodes := 0
		if _, err := cursor.mutation(&nodes); err == nil {
			t.Fatal("truncated raw tag accepted")
		}
	}
}

func TestRawTagPointNegotiation(t *testing.T) {
	records := []FlatPointRecord{{Version: 1, Category: "logging", Measurement: "test", Entries: []FlatPointEntry{{Key: "tag", IsTag: true, Value: "a\xff\x00"}}}}
	for _, projection := range []InputProjection{{}, {rawStringValues: true}} {
		if _, err := EncodeProjectedPointRecords(records, projection); err == nil {
			t.Fatal("raw field capability authorized raw tags")
		}
	}
	wire, err := EncodeProjectedPointRecords(records, InputProjection{rawTagValues: true})
	if err != nil {
		t.Fatal(err)
	}
	points, err := DecodeFlatPoints(wire)
	if err != nil || len(points) != 1 || points[0].Tags["tag"] != "a\xff\x00" {
		t.Fatalf("points=%#v error=%v", points, err)
	}
	// Without the explicit flag, the same bytes must still fail UTF-8 validation.
	wire[pointWireHeaderBytes+4+2] = 0
	if _, err := DecodeFlatPoints(wire); err == nil {
		t.Fatal("untagged raw bytes accepted")
	}
}

func TestNegotiatedNestedRawStrings(t *testing.T) {
	bad := string([]byte{0xff, 0, 0xe2})
	records := []FlatPointRecord{{Version: 1, Category: "logging", Measurement: "nested", Entries: []FlatPointEntry{
		{Key: "values", Value: []any{[]string{"valid", bad}, map[string]any{"inner": bad}}},
	}}}
	if _, err := EncodeProjectedPointRecords(records, InputProjection{}); err == nil {
		t.Fatal("legacy encoding accepted invalid string")
	}
	encoded, err := EncodeProjectedPointRecords(records, InputProjection{rawStringValues: true})
	if err != nil {
		t.Fatal(err)
	}
	points, err := DecodeFlatPoints(encoded)
	if err != nil {
		t.Fatal(err)
	}
	want := []any{[]any{"valid", RawString(bad)}, map[string]any{"inner": RawString(bad)}}
	if !reflect.DeepEqual(points[0].Fields["values"], want) {
		t.Fatalf("nested value=%#v", points[0].Fields["values"])
	}
	if records[0].Entries[0].Value.([]any)[0].([]string)[1] != bad {
		t.Fatal("input mutated")
	}
	for _, value := range []any{map[string]any{bad: "value"}, func() any {
		var deep any = "leaf"
		for i := 0; i < pointWireMaxDepth+2; i++ {
			deep = []any{deep}
		}
		return deep
	}()} {
		records[0].Entries[0].Value = value
		if _, err := EncodeProjectedPointRecords(records, InputProjection{rawStringValues: true}); err == nil {
			t.Fatal("metadata/depth validation bypassed")
		}
	}
}

func TestRawMeasurementRoundTripsWithoutRelaxingOtherUTF8Metadata(t *testing.T) {
	bad := string([]byte{0xff})
	measurement := Point{Version: 1, Category: "logging", Measurement: bad}
	wire, err := EncodeFlatPoints([]Point{measurement})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeFlatPoints(wire)
	if err != nil || len(decoded) != 1 || decoded[0].Measurement != bad {
		t.Fatalf("raw measurement roundtrip=%#v error=%v", decoded, err)
	}
	deltas := []MutationDelta{{RecordIndex: 0, Operations: []MutationOp{{Kind: MutationSetRawMeasurement, String: bad}}}}
	deltaWire, err := EncodeMutationDeltas(deltas)
	if err != nil {
		t.Fatal(err)
	}
	decodedDeltas, err := DecodeMutationDeltas(deltaWire)
	if err != nil || len(decodedDeltas) != 1 || len(decodedDeltas[0].Operations) != 1 ||
		decodedDeltas[0].Operations[0].Kind != MutationSetRawMeasurement || decodedDeltas[0].Operations[0].String != bad {
		t.Fatalf("raw measurement delta=%#v error=%v", decodedDeltas, err)
	}
	applied := Point{Version: 1, Category: "logging", Measurement: "before"}
	if err := decodedDeltas[0].Apply(&applied); err != nil || applied.Measurement != bad {
		t.Fatalf("raw measurement apply=%#v error=%v", applied, err)
	}
	for _, pt := range []Point{
		{Version: 1, Category: bad, Measurement: "valid"},
		{Version: 1, Category: "logging", Fields: map[string]any{bad: RawString(bad)}},
		{Version: 1, Category: "logging", Tags: map[string]string{"tag": bad}},
		{Version: 1, Category: "logging", Fields: map[string]any{"ordinary": bad}},
	} {
		if _, err := EncodeFlatPoints([]Point{pt}); err == nil {
			t.Fatalf("accepted invalid UTF-8 outside explicit raw value: %#v", pt)
		}
	}
}

func TestRawStringRejectsTruncatedPayload(t *testing.T) {
	for _, data := range [][]byte{{8}, {8, 1, 0, 0, 0}, {8, 0xff, 0xff, 0xff, 0xff}} {
		cursor := pointCursor{data: data}
		nodes := 0
		if _, err := cursor.value(0, &nodes); err == nil {
			t.Fatalf("accepted truncated raw value: %x", data)
		}
	}
}

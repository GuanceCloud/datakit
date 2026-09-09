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

func TestRawPointKeyCapabilityAndWire(t *testing.T) {
	for _, descriptor := range []string{`{"version":1}`, `{"version":1,"raw_string_values":true,"raw_tag_values":true}`} {
		var capability ProgramCapabilities
		if err := json.Unmarshal([]byte(descriptor), &capability); err != nil {
			t.Fatal(err)
		}
		if capability.RawPointKeys {
			t.Fatal("other byte capabilities authorized Point keys")
		}
	}
	records := []FlatPointRecord{{Version: 1, Category: "logging", Measurement: "keys", Entries: []FlatPointEntry{
		{Key: "\xff", Value: int64(1)}, {Key: "�", Value: int64(2)}, {Key: "\xe2\x82", Value: "tag", IsTag: true},
	}}}
	for _, projection := range []InputProjection{{}, {rawStringValues: true, rawTagValues: true}} {
		if _, err := EncodeProjectedPointRecords(records, projection); err == nil {
			t.Fatal("unnegotiated key accepted")
		}
	}
	wire, err := EncodeProjectedPointRecords(records, InputProjection{rawPointKeys: true})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeFlatPoints(wire)
	if err != nil {
		t.Fatal(err)
	}
	if decoded[0].Fields["\xff"] != int64(1) || decoded[0].Fields["�"] != int64(2) || decoded[0].Tags["\xe2\x82"] != "tag" {
		t.Fatalf("keys changed: %#v", decoded)
	}
	for i := 0; i < len(wire); i++ {
		if _, err := DecodeFlatPoints(wire[:i]); err == nil {
			t.Fatalf("accepted truncated wire at %d", i)
		}
	}
	wire[16+4+2] &^= 8
	if _, err := DecodeFlatPoints(wire); err == nil {
		t.Fatal("legacy text layout accepted raw key")
	}
	for i := range wire {
		wire[i] = 0
	}
	if decoded[0].Fields["\xff"] != int64(1) {
		t.Fatal("decoded key aliases input memory")
	}
}

func TestRawKeyMutationOpcodeDecode(t *testing.T) {
	for _, kind := range []MutationOpKind{MutationSetField, MutationDeleteField, MutationSetTag, MutationDeleteTag, MutationSetRawTag} {
		writer := &pointWriter{}
		writer.u8(byte(kind) | 128)
		if err := writer.rawString("\xff"); err != nil {
			t.Fatal(err)
		}
		want := MutationOp{Kind: kind, Key: "\xff"}
		switch kind { //nolint:exhaustive // This fixture only writes values for set operations.
		case MutationSetField:
			want.Value = int64(7)
			nodes := 0
			if err := writer.value(want.Value, 0, &nodes); err != nil {
				t.Fatal(err)
			}
		case MutationSetTag, MutationSetRawTag:
			want.String = "value"
			if kind == MutationSetRawTag {
				want.String = "\xfe"
			}
			if err := writer.rawString(want.String); err != nil {
				t.Fatal(err)
			}
		}
		for i := 0; i < len(writer.data); i++ {
			cursor := pointCursor{data: writer.data[:i]}
			nodes := 0
			if _, err := cursor.mutation(&nodes); err == nil {
				t.Fatalf("accepted truncated opcode %d at %d", kind, i)
			}
		}
		cursor := pointCursor{data: writer.data}
		nodes := 0
		got, err := cursor.mutation(&nodes)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("kind=%d got=%#v err=%v", kind, got, err)
		}
	}
	for _, kind := range []byte{0, 5, 6, 7, 8, 9, 10, 11, 12, 127} {
		cursor := pointCursor{data: []byte{kind | 128}}
		nodes := 0
		if _, err := cursor.mutation(&nodes); err == nil {
			t.Fatalf("keyless opcode accepted: %d", kind)
		}
	}
}

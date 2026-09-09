// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// PPS2 v3 retains indexed lifecycle frames, but replaces each PPD1 payload with
// count + typed operations keyed by a batch-local dictionary. No Go Point layout
// crosses the ABI. Dictionary offsets are absolute from the start of the batch.
func decodeDynamicBatchInto(input []byte, schema StaticOutputSchema, scratch *Batch) (Batch, error) {
	if len(input) < staticBatchHeaderSize+4 || len(input) > staticBatchHeaderSize+maxResultBytes ||
		string(input[:4]) != "PPS2" || binary.LittleEndian.Uint16(input[4:6]) != 3 ||
		!bytes.Equal(input[24:56], schema.Hash[:]) {
		return Batch{}, errors.New("invalid dynamic batch header")
	}
	flags := binary.LittleEndian.Uint16(input[6:8])
	records := binary.LittleEndian.Uint32(input[8:12])
	frames := binary.LittleEndian.Uint32(input[12:16])
	length := binary.LittleEndian.Uint32(input[16:20])
	offset := uint64(binary.LittleEndian.Uint32(input[20:24]))
	if flags & ^uint16(1) != 0 || records > pointWireMaxItems || frames > maxFrames || records > frames ||
		uint64(length) != uint64(len(input)-staticBatchHeaderSize) || offset < staticBatchHeaderSize || offset > uint64(len(input)-4) ||
		uint64(frames) > (offset-staticBatchHeaderSize)/frameHeaderBytes {
		return Batch{}, errors.New("invalid dynamic batch counts")
	}
	cursor := pointCursor{data: input[offset:]}
	count, err := cursor.u32()
	if err != nil || count > pointWireMaxItems || uint64(count) > uint64(len(cursor.data)-4)/4 {
		return Batch{}, errors.New("invalid dynamic dictionary count")
	}
	dictionary := make([]string, count) // non-nil also for the empty dictionary
	seen := make(map[string]struct{}, count)
	for i := range dictionary {
		key, err := cursor.key(true)
		if err != nil {
			return Batch{}, err
		}
		if _, exists := seen[key]; exists {
			return Batch{}, errors.New("duplicate dynamic dictionary key")
		}
		seen[key] = struct{}{}
		dictionary[i] = key
	}
	if !cursor.done() {
		return Batch{}, errors.New("trailing dynamic dictionary bytes")
	}
	return decodeIndexedFrames(input[staticBatchHeaderSize:offset], records, frames, flags, scratch, dictionary, true, len(input))
}

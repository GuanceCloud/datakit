// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package msgpack

import (
	"bytes"
	"errors"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/tinylib/msgp/msgp"
)

// RepairUTF8 ensures all characters in s are UTF-8 by replacing non-UTF-8 characters
// with the replacement char �.
func RepairUTF8(s string) string {
	in := strings.NewReader(s)
	var out bytes.Buffer
	out.Grow(len(s))

	for {
		r, _, err := in.ReadRune()
		if err != nil {
			// note: by contract, if `in` contains non-valid utf-8, no error is returned. Rather the utf-8 replacement
			// character is returned. Therefore, the only error should usually be io.EOF indicating end of string.
			// If any other error is returned by chance, we quit as well, outputting whatever part of the string we
			// had already constructed.
			return out.String()
		}
		out.WriteRune(r)
	}
}

// ParseStringBytes reads the next type in the msgpack payload and
// converts the BinType or the StrType in a valid string.
func ParseStringBytes(data []byte) (str string, bts []byte, err error) {
	bts = data
	if msgp.IsNil(bts) {
		bts, err = msgp.ReadNilBytes(bts)
		return
	}
	// read the generic representation type without decoding
	t := msgp.NextType(bts)

	var i []byte
	switch t {
	case msgp.BinType:
		i, bts, err = msgp.ReadBytesZC(bts)
	case msgp.StrType:
		i, bts, err = msgp.ReadStringZC(bts)
	default:
		err = msgp.TypeError{Encoded: t, Method: msgp.StrType}
		return
	}

	if err != nil {
		return
	}

	if utf8.Valid(i) {
		str = string(i)
		return
	}

	str = RepairUTF8(msgp.UnsafeString(i))
	return str, bts, nil
}

// ParseFloat64Bytes parses a float64 even if the sent value is an int64 or an uint64;
// this is required because the encoding library could remove bytes from the encoded
// payload to reduce the size, if they're not needed.
func ParseFloat64Bytes(data []byte) (f float64, bts []byte, err error) {
	bts = data
	if msgp.IsNil(bts) {
		bts, err = msgp.ReadNilBytes(bts)
		return
	}
	// read the generic representation type without decoding
	t := msgp.NextType(bts)

	switch t {
	case msgp.IntType:
		var i int64
		i, bts, err = msgp.ReadInt64Bytes(bts)
		if err != nil {
			return 0, bts, err
		}

		f = float64(i)
		return

	case msgp.UintType:
		var i uint64
		i, bts, err = msgp.ReadUint64Bytes(bts)
		if err != nil {
			return 0, bts, err
		}
		f = float64(i)
		return

	case msgp.Float64Type:
		f, bts, err = msgp.ReadFloat64Bytes(bts)
		return

	default:
		err = msgp.TypeError{Encoded: t, Method: msgp.Float64Type}
		return
	}
}

// cast to int64 values that are int64 but that are sent in uint64
// over the wire. Set to 0 if they overflow the MaxInt64 size. This
// cast should be used ONLY while decoding int64 values that are
// sent as uint64 to reduce the payload size, otherwise the approach
// is not correct in the general sense.
func CastInt64(v uint64) (int64, bool) {
	if v > math.MaxInt64 {
		return 0, false
	}
	return int64(v), true
}

// ParseInt64Bytes parses an int64 even if the sent value is an uint64;
// this is required because the encoding library could remove bytes from the encoded
// payload to reduce the size, if they're not needed.
//nolint:dupl
func ParseInt64Bytes(data []byte) (i int64, bts []byte, err error) { //nolint:gocritic
	bts = data

	if msgp.IsNil(bts) {
		bts, err = msgp.ReadNilBytes(bts)
		return
	}
	// read the generic representation type without decoding
	t := msgp.NextType(bts)

	switch t {
	case msgp.IntType:
		i, bts, err = msgp.ReadInt64Bytes(bts)
		return
	case msgp.UintType:
		var u uint64
		u, bts, err = msgp.ReadUint64Bytes(bts)
		if err != nil {
			return
		}

		// force-cast
		v, ok := CastInt64(u)
		if !ok {
			err = errors.New("found uint64, overflows int64")
			return
		}
		i = v

		return
	default:
		err = msgp.TypeError{Encoded: t, Method: msgp.IntType}
		return
	}
}

// ParseUint64Bytes parses an uint64 even if the sent value is an int64;
// this is required because the language used for the encoding library
// may not have unsigned types. An example is early version of Java
// (and so JRuby interpreter) that encodes uint64 as int64:
// http://docs.oracle.com/javase/tutorial/java/nutsandbolts/datatypes.html
func ParseUint64Bytes(data []byte) (u uint64, bts []byte, err error) { //nolint:gocritic
	bts = data
	if msgp.IsNil(bts) {
		bts, err = msgp.ReadNilBytes(bts)
		return
	}
	// read the generic representation type without decoding
	t := msgp.NextType(bts)

	switch t {
	case msgp.UintType:
		u, bts, err = msgp.ReadUint64Bytes(bts)

	case msgp.IntType:
		var i int64
		i, bts, err = msgp.ReadInt64Bytes(bts)
		if err != nil {
			return
		}
		u = uint64(i)

	default:
		err = msgp.TypeError{Encoded: t, Method: msgp.IntType}
	}

	return
}

// cast to int32 values that are int32 but that are sent in uint32
// over the wire. Set to 0 if they overflow the MaxInt32 size. This
// cast should be used ONLY while decoding int32 values that are
// sent as uint32 to reduce the payload size, otherwise the approach
// is not correct in the general sense.
func CastInt32(v uint32) (int32, bool) {
	if v > math.MaxInt32 {
		return 0, false
	}
	return int32(v), true
}

// ParseInt32Bytes parses an int32 even if the sent value is an uint32;
// this is required because the encoding library could remove bytes from the encoded
// payload to reduce the size, if they're not needed.
//nolint:dupl,gocritic
func ParseInt32Bytes(data []byte) (i int32, bts []byte, err error) {
	bts = data
	if msgp.IsNil(bts) {
		bts, err = msgp.ReadNilBytes(bts)
		return
	}

	// read the generic representation type without decoding
	t := msgp.NextType(bts)

	switch t {
	case msgp.IntType:
		i, bts, err = msgp.ReadInt32Bytes(bts)
		return

	case msgp.UintType:
		var u uint32

		u, bts, err = msgp.ReadUint32Bytes(bts)
		if err != nil {
			return
		}

		// force-cast
		val, ok := CastInt32(u)
		if !ok {
			err = errors.New("found uint32, overflows int32")
		}
		i = val
		return
	default:
		err = msgp.TypeError{Encoded: t, Method: msgp.IntType}
		return
	}
}

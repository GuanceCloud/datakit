//go:build linux
// +build linux

package protodec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
)

func TestDecodeCompressIntExactLengths(t *testing.T) {
	if got := decodeCompressInt([]byte{intFlags2, 0x34, 0x12}); got != 0x1234 {
		t.Fatalf("2-byte encoded int = %#x, want %#x", got, uint64(0x1234))
	}
	if got := decodeCompressInt([]byte{intFlags3, 0x56, 0x34, 0x12}); got != 0x123456 {
		t.Fatalf("3-byte encoded int = %#x, want %#x", got, uint64(0x123456))
	}

	payload := make([]byte, 1+8)
	payload[0] = intFlags8
	binary.LittleEndian.PutUint64(payload[1:], 0x1122334455667788)
	if got := decodeCompressInt(payload); got != 0x1122334455667788 {
		t.Fatalf("8-byte encoded int = %#x, want %#x", got, uint64(0x1122334455667788))
	}
}

func TestMysqlStatementIDUsesFourBytes(t *testing.T) {
	var payload [4]byte
	binary.LittleEndian.PutUint32(payload[:], 0x01020304)

	if got := (&mysqlInfo{}).getStatementID(payload[:]); got != 0x01020304 {
		t.Fatalf("statement id = %#x, want %#x", got, 0x01020304)
	}
}

func TestMysqlPendingBufferLimitResetsReader(t *testing.T) {
	info := &mysqlInfo{reader: bytes.NewBuffer(make([]byte, 0, 8))}
	info.reader.Write([]byte("1234"))

	_, err := info.parse([]byte("56789"), 0, 8)
	if !errors.Is(err, errMysqlPendingBufferLimit) {
		t.Fatalf("parse err = %v, want %v", err, errMysqlPendingBufferLimit)
	}
	if got := info.reader.Len(); got != 0 {
		t.Fatalf("reader len = %d, want 0 after pending limit", got)
	}
}

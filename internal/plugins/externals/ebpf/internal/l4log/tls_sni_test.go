//go:build linux
// +build linux

package l4log

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestParseTLSClientHelloServerName(t *testing.T) {
	clientHello := mustBuildTLSClientHello("example.com")

	got, ok, incomplete := parseTLSClientHelloServerName(clientHello)
	if incomplete {
		t.Fatal("unexpected incomplete parse")
	}
	if !ok {
		t.Fatal("expected SNI to be parsed")
	}
	if got != "example.com" {
		t.Fatalf("unexpected SNI %q", got)
	}
}

func TestParseTLSClientHelloServerNameIncomplete(t *testing.T) {
	clientHello := []byte{0x16, 0x03, 0x01, 0x00}

	_, ok, incomplete := parseTLSClientHelloServerName(clientHello)
	if ok {
		t.Fatal("expected no SNI on incomplete payload")
	}
	if !incomplete {
		t.Fatal("expected incomplete parse")
	}
}

func mustBuildTLSClientHello(serverName string) []byte {
	var ext bytes.Buffer
	binary.Write(&ext, binary.BigEndian, uint16(tlsExtensionServerName))

	listLen := 1 + 2 + len(serverName)
	binary.Write(&ext, binary.BigEndian, uint16(2+listLen))
	binary.Write(&ext, binary.BigEndian, uint16(listLen))
	ext.WriteByte(0)
	binary.Write(&ext, binary.BigEndian, uint16(len(serverName)))
	ext.WriteString(serverName)

	var body bytes.Buffer
	body.Write([]byte{0x03, 0x03})
	body.Write(make([]byte, 32))
	body.WriteByte(0)
	binary.Write(&body, binary.BigEndian, uint16(2))
	body.Write([]byte{0x13, 0x01})
	body.WriteByte(1)
	body.WriteByte(0)
	binary.Write(&body, binary.BigEndian, uint16(ext.Len()))
	body.Write(ext.Bytes())

	var hs bytes.Buffer
	hs.WriteByte(tlsHandshakeTypeClientHello)
	hs.Write([]byte{0, 0, byte(body.Len())})
	hs.Write(body.Bytes())

	var record bytes.Buffer
	record.WriteByte(tlsRecordTypeHandshake)
	record.Write([]byte{0x03, 0x01})
	binary.Write(&record, binary.BigEndian, uint16(hs.Len()))
	record.Write(hs.Bytes())
	return record.Bytes()
}

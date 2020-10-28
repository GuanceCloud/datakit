package tailf

import (
	"testing"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func TestEncoding(t *testing.T) {
	var str = "ABCD"

	utf16BE, _, _ := transform.String(unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewEncoder(), str)

	t.Logf("utf16BE: %b\n", []byte(utf16BE))
	t.Logf("utf16BE: % x\n", []byte(utf16BE))
	t.Logf("utf16BE: %s\n", utf16BE)

	decoder, _ := NewDecoder("utf-16be")

	b, err := decoder.String(utf16BE)
	if err != nil {
		panic(err)
	}

	t.Logf("end: %b\n", []byte(b))
	t.Logf("end: % x\n", []byte(b))
	t.Logf("end: %s\n", b)
}

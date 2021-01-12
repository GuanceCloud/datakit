package tailf

import (
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func TestEncodingUTF16(t *testing.T) {
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

func TestEncodingGBK(t *testing.T) {
	var str = "你好，世界"

	gbk, _, _ := transform.String(simplifiedchinese.GBK.NewEncoder(), str)

	t.Logf("GBK: %b\n", []byte(gbk))
	t.Logf("GBK: % x\n", []byte(gbk))
	t.Logf("GBK: %s\n", gbk)

	decoder, _ := NewDecoder("gbk")

	b, err := decoder.String(gbk)
	if err != nil {
		panic(err)
	}

	t.Logf("end: %b\n", []byte(b))
	t.Logf("end: % x\n", []byte(b))
	t.Logf("end: %s\n", b)
}

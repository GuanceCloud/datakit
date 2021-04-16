package inputs

import (
	"math/rand"
	"testing"
	"unsafe"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func TestCheckFieldsLength(t *testing.T) {

	randString := func(n int) string {
		const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		b := make([]byte, n)
		for i := range b {
			b[i] = letterBytes[rand.Intn(len(letterBytes))]
		}
		return *(*string)(unsafe.Pointer(&b))
	}

	const maxlength = 30

	testcase := []map[string]interface{}{
		// 有效
		{
			"message": randString(10),
		},
		{
			"message": randString(maxlength),
		},
		{
			"message": randString(maxlength + 1),
		},
		{
			"message": randString(maxlength + 2),
		},
		{
			"message": randString(10),
			"waste":   randString(maxlength),
		},

		// 无效，报错，非 message 字段超过最大限定
		{
			"message": randString(10),
			"invalid": randString(maxlength + 1),
		},
	}

	for idx, tc := range testcase {
		t.Logf("[%d] source: %v message_length: %d\n", idx, tc, len(tc["message"].(string)))

		if err := checkFieldsLength(tc, maxlength); err != nil {
			t.Logf("[%d] ending: error %s\n\n", idx, err)

		} else {
			t.Logf("[%d] ending truncated: %v message_length: %d\n\n", idx, tc, len(tc["message"].(string)))
		}
	}
}

func TestAddStatus(t *testing.T) {
	testcase := []map[string]interface{}{
		// 有效
		{"status": "i"},
		// 有效，大写
		{"status": "DEBUG"},
		// 无效，非枚举status
		{"status": "invalid"},
		// 无status
		{"invalidKey": "XX"},
		// status不是str类型
		{"status": 123},
		// nil
		{},
	}

	for idx, tc := range testcase {
		t.Logf("[%d] source: %v\n", idx, tc)
		addStatus(tc)
		t.Logf("[%d] ending: %v\n\n", idx, tc)
	}
}

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

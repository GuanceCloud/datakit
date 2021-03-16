// +build !solaris

package tailf

import (
	"math/rand"
	"testing"
	"unsafe"
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

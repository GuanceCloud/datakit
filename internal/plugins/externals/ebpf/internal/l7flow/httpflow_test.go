//go:build linux
// +build linux

package httpflow

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindURI(t *testing.T) {
	cases := []string{
		"GET https://ffff.fvvv?22",
		"GET https://ffff.fvvv/",
		"GET http://ffff.fvvv/sfsf?22",
		"GET https://ffff.fvvv",
		"GET http://ffff.fvvv?1",
		"GET /localhosttest/FFF?1",
		"GET /abccom/?1",
		"POST /abccom",
		"POST /abccom/ ",
		"POST / ",
	}

	result := []string{
		"/",
		"/",
		"/sfsf",
		"/",
		"/",
		"/localhosttest/FFF",
		"/abccom/",
		"/abccom",
		"/abccom/",
		"/",
	}

	trunc_result := []bool{
		false,
		true,
		false,
		true,
		false,
		false,
		false,
		true,
		false,
		false,
	}

	for k, v := range cases {
		path, trunc_path := FindHTTPURI(v)
		assert.Equal(t, path, result[k])
		assert.Equal(t, trunc_result[k], trunc_path, strconv.FormatInt(int64(k), 10)+
			":"+path+": "+result[k])
	}
}

func TestHTTPVersion(t *testing.T) {
	cases := []uint32{
		1<<16 + 0,
		1<<16 + 1,
		2<<16 + 0,
		3<<16 + 0,
	}
	result := []string{
		"1.0", "1.1", "2.0", "3.0",
	}
	for k, v := range cases {
		if ParseHTTPVersion(v) != result[k] {
			t.Error(v, " ", result[k])
		}
	}
}

func TestPayld2Thr(t *testing.T) {
	r := TransPayloadToThrID(&CPayloadID{ktime: 12222222222, prandom: 1})
	t.Log(r)
}

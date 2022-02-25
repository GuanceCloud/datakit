//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

package httpflow

import "testing"

func TestFindURI(t *testing.T) {
	cases := []string{
		"GET http://ffff.fvvv/sfsf?22",
		"GET https://ffff.fvvv",
		"GET http://ffff.fvvv?1",
		"GET /localhosttest/FFF?1",
		"GET /abccom/?1",
		"POST /abccom",
		"POST /abccom/",
	}

	result := []string{
		"/sfsf",
		"/",
		"/",
		"/localhosttest/FFF",
		"/abccom/",
		"/abccom",
		"/abccom/",
	}

	for k, v := range cases {
		if FindHTTPURI(v) != result[k] {
			t.Error(FindHTTPURI(v), " ", result[k])
		}
	}
}

func TestHTTPCode(t *testing.T) {
	cases := []uint32{
		201,
		300,
		299,
	}
	result := []int{
		200, 300, 200,
	}
	for k, v := range cases {
		if ParseHTTPCode(v) != result[k] {
			t.Error(ParseHTTPCode(v), " ", result[k])
		}
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
			t.Error(ParseHTTPCode(v), " ", result[k])
		}
	}
}

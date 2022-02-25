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
		if FindHttpURI(v) != result[k] {
			t.Error(FindHttpURI(v), " ", result[k])
		}
	}
}

func TestHttpCode(t *testing.T) {
	cases := []uint32{
		201,
		300,
		299,
	}
	result := []int{
		200, 300, 200,
	}
	for k, v := range cases {
		if ParseHttpCode(v) != result[k] {
			t.Error(ParseHttpCode(v), " ", result[k])
		}
	}
}

func TestHttpVersion(t *testing.T) {
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
		if ParseHttpVersion(v) != result[k] {
			t.Error(ParseHttpCode(v), " ", result[k])
		}
	}
}

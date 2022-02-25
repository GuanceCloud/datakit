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
		if HttpCode(v) != result[k] {
			t.Error(HttpCode(v), " ", result[k])
		}
	}
}

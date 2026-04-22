//go:build linux
// +build linux

package l4log

import "testing"

func TestNormalizeHTTPHostString(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  string
	}{
		{name: "domain", in: "Example.COM", out: "example.com"},
		{name: "strip port", in: "example.com:8443", out: "example.com"},
		{name: "trim suffix dot", in: "example.com.", out: "example.com"},
		{name: "ipv4", in: "127.0.0.1:80", out: ""},
		{name: "ipv6", in: "[2001:db8::1]:443", out: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeHTTPHostString(tc.in); got != tc.out {
				t.Fatalf("normalizeHTTPHostString(%q) = %q, want %q", tc.in, got, tc.out)
			}
		})
	}
}

func TestExtractAbsoluteFormHost(t *testing.T) {
	target := []byte("http://Example.COM:8080/path?q=1")
	if got := extractAbsoluteFormHost(target); got != "example.com" {
		t.Fatalf("extractAbsoluteFormHost() = %q, want %q", got, "example.com")
	}
}

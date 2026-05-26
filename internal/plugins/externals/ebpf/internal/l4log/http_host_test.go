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
		{name: "service domain", in: "api-service.default.svc.cluster.local", out: "api-service.default.svc.cluster.local"},
		{name: "truncated label", in: "openapi.excp-", out: ""},
		{name: "label starts with hyphen", in: "openapi.-excp.example", out: ""},
		{name: "empty label", in: "openapi..example", out: ""},
		{name: "invalid port", in: "api.example.com:http", out: ""},
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
	cases := []struct {
		name   string
		target string
		want   string
	}{
		{name: "absolute form", target: "http://Example.COM:8080/path?q=1", want: "example.com"},
		{name: "specific domain", target: "http://openapi.excp-can-dao.epay/api", want: "openapi.excp-can-dao.epay"},
		{name: "truncated label", target: "http://openapi.excp-", want: ""},
		{name: "no path boundary", target: "http://api.example.com", want: ""},
		{name: "query boundary", target: "http://api.example.com?q=1", want: "api.example.com"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractAbsoluteFormHost([]byte(tc.target)); got != tc.want {
				t.Fatalf("extractAbsoluteFormHost() = %q, want %q", got, tc.want)
			}
		})
	}
}

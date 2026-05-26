//go:build linux
// +build linux

package protodec

import "testing"

func TestNormalizeHTTPHost(t *testing.T) {
	tcases := []struct {
		name string
		in   string
		want string
	}{
		{name: "domain", in: "API.Example.COM", want: "api.example.com"},
		{name: "domain with port", in: "api.example.com:8443", want: "api.example.com"},
		{name: "domain with trailing dot", in: "api.example.com.", want: "api.example.com"},
		{name: "service domain", in: "api-service.default.svc.cluster.local", want: "api-service.default.svc.cluster.local"},
		{name: "truncated label", in: "openapi.excp-", want: ""},
		{name: "label starts with hyphen", in: "openapi.-excp.example", want: ""},
		{name: "empty label", in: "openapi..example", want: ""},
		{name: "invalid port", in: "api.example.com:http", want: ""},
		{name: "ipv4", in: "10.0.0.1:8080", want: ""},
		{name: "ipv6", in: "[2001:db8::1]:443", want: ""},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeHTTPHost(tc.in); got != tc.want {
				t.Fatalf("normalizeHTTPHost(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestExtractHTTPRequestHost(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		headers map[string]string
		want    string
	}{
		{
			name:    "absolute form",
			payload: "GET http://api.example.com:8080/demo?q=1 HTTP/1.1\r\nUser-Agent: curl/8.0\r\n\r\n",
			want:    "api.example.com",
		},
		{
			name:    "target captured without crlf",
			payload: "GET http://openapi.excp-can-dao.epay/api HTTP/1.",
			want:    "openapi.excp-can-dao.epay",
		},
		{
			name:    "truncated absolute form",
			payload: "GET http://openapi.excp-",
			want:    "",
		},
		{
			name:    "truncated label in complete line",
			payload: "GET http://api.example HTTP/1.1\r\n\r\n",
			want:    "",
		},
		{
			name:    "absolute form without path falls back to host header",
			payload: "GET http://api.example.com HTTP/1.1\r\nHost: openapi.excp-can-dao.epay\r\n\r\n",
			headers: map[string]string{"Host": "openapi.excp-can-dao.epay"},
			want:    "openapi.excp-can-dao.epay",
		},
		{
			name:    "host header fallback",
			payload: "GET /api HTTP/1.1\r\nHost: ignored-by-test\r\n\r\n",
			headers: map[string]string{"Host": "openapi.excp-can-dao.epay"},
			want:    "openapi.excp-can-dao.epay",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractHTTPRequestHost([]byte(tc.payload), tc.headers); got != tc.want {
				t.Fatalf("extractHTTPRequestHost() = %q, want %q", got, tc.want)
			}
		})
	}
}

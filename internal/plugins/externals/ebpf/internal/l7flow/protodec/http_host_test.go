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
	payload := []byte("GET http://api.example.com:8080/demo?q=1 HTTP/1.1\r\nUser-Agent: curl/8.0\r\n\r\n")

	if got := extractHTTPRequestHost(payload, nil); got != "api.example.com" {
		t.Fatalf("extractHTTPRequestHost() = %q, want %q", got, "api.example.com")
	}
}

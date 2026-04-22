//go:build linux
// +build linux

package l4log

import (
	"bytes"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

func extractAbsoluteFormHost(target []byte) string {
	if !bytes.HasPrefix(target, []byte("http://")) && !bytes.HasPrefix(target, []byte("https://")) {
		return ""
	}

	host := target
	if bytes.HasPrefix(host, []byte("http://")) {
		host = host[len("http://"):]
	} else {
		host = host[len("https://"):]
	}

	if idx := bytes.IndexByte(host, '/'); idx >= 0 {
		host = host[:idx]
	}
	if idx := bytes.IndexByte(host, '?'); idx >= 0 {
		host = host[:idx]
	}

	return normalizeHTTPHostBytes(host)
}

func normalizeHTTPHostBytes(host []byte) string {
	return normalizeHTTPHostString(string(bytes.TrimSpace(host)))
}

func normalizeHTTPHostString(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return ""
	}

	switch parsedHost, _, err := net.SplitHostPort(host); {
	case err == nil:
		host = parsedHost
	case strings.Count(host, ":") == 1:
		idx := strings.LastIndex(host, ":")
		if idx > 0 && idx+1 < len(host) {
			if _, err := strconv.ParseUint(host[idx+1:], 10, 16); err == nil {
				host = host[:idx]
			}
		}
	case strings.Count(host, ":") > 1:
		host = strings.Trim(host, "[]")
		if _, err := netip.ParseAddr(host); err == nil {
			return ""
		}
	}

	host = strings.Trim(host, "[]")
	if host == "" {
		return ""
	}

	if _, err := netip.ParseAddr(host); err == nil {
		return ""
	}

	return host
}

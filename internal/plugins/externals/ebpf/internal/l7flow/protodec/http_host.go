//go:build linux
// +build linux

package protodec

import (
	"bytes"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

func extractHTTPRequestHost(payload []byte, headers map[string]string) string {
	if host := normalizeHTTPHost(findAbsoluteFormHost(payload)); host != "" {
		return host
	}
	if headers == nil {
		return ""
	}
	return normalizeHTTPHost(headers["Host"])
}

func findAbsoluteFormHost(payload []byte) string {
	idx := bytes.Index(payload, []byte("\r\n"))
	if idx <= 0 {
		return ""
	}

	parts := strings.SplitN(string(payload[:idx]), " ", 3)
	if len(parts) < 2 {
		return ""
	}

	target := parts[1]
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		return ""
	}

	target = strings.TrimPrefix(strings.TrimPrefix(target, "http://"), "https://")
	if off := strings.IndexByte(target, '/'); off >= 0 {
		target = target[:off]
	}

	return target
}

func normalizeHTTPHost(host string) string {
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

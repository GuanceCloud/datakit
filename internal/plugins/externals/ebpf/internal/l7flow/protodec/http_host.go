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
	if host := findAbsoluteFormHost(payload); host != "" {
		return host
	}
	if headers == nil {
		return ""
	}
	return normalizeHTTPHost(headers["Host"])
}

func findAbsoluteFormHost(payload []byte) string {
	idx := bytes.Index(payload, []byte("\r\n"))
	line := payload
	if idx >= 0 {
		if idx == 0 {
			return ""
		}
		line = payload[:idx]
	}

	firstSpace := bytes.IndexByte(line, ' ')
	if firstSpace <= 0 || firstSpace+1 >= len(line) {
		return ""
	}

	target := string(line[firstSpace+1:])
	if secondSpace := strings.IndexByte(target, ' '); secondSpace >= 0 {
		target = target[:secondSpace]
	}

	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		return ""
	}

	target = strings.TrimPrefix(strings.TrimPrefix(target, "http://"), "https://")
	pathOff := strings.IndexByte(target, '/')
	queryOff := strings.IndexByte(target, '?')
	switch {
	case pathOff >= 0 && (queryOff < 0 || pathOff < queryOff):
		target = target[:pathOff]
	case queryOff >= 0:
		target = target[:queryOff]
	default:
		return ""
	}

	return normalizeHTTPHost(target)
}

func normalizeHTTPHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return ""
	}

	switch parsedHost, port, err := net.SplitHostPort(host); {
	case err == nil:
		if _, err := strconv.ParseUint(port, 10, 16); err != nil {
			return ""
		}
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
	if !isValidHTTPHostName(host) {
		return ""
	}

	return host
}

func isValidHTTPHostName(host string) bool {
	if host == "" || len(host) > 253 {
		return false
	}

	labelStart := 0
	for i := 0; i <= len(host); i++ {
		if i < len(host) && host[i] != '.' {
			c := host[i]
			switch {
			case c >= 'a' && c <= 'z':
			case c >= '0' && c <= '9':
			case c == '-' || c == '_':
			default:
				return false
			}
			continue
		}

		if i == labelStart || i-labelStart > 63 {
			return false
		}
		if host[labelStart] == '-' || host[i-1] == '-' {
			return false
		}
		labelStart = i + 1
	}

	return true
}

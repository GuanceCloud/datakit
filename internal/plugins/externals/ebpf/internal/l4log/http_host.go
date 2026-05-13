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

	pathOff := bytes.IndexByte(host, '/')
	queryOff := bytes.IndexByte(host, '?')
	switch {
	case pathOff >= 0 && (queryOff < 0 || pathOff < queryOff):
		host = host[:pathOff]
	case queryOff >= 0:
		host = host[:queryOff]
	default:
		return ""
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

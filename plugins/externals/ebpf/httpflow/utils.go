//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

package httpflow

import (
	"fmt"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/k8sinfo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/netflow"
)

var k8sNetInfo *k8sinfo.K8sNetInfo

func SetK8sNetInfo(n *k8sinfo.K8sNetInfo) {
	k8sNetInfo = n
}

const (
	HTTP_METHOD_UNKNOWN = 0x00 + iota
	HTTP_METHOD_GET
	HTTP_METHOD_POST
	HTTP_METHOD_PUT
	HTTP_METHOD_DELETE
	HTTP_METHOD_HEAD
	HTTP_METHOD_OPTIONS
	HTTP_METHOD_PATCH

	// TODO 解析此类 HTTP 数据
	HTTP_METHOD_CONNECT
	HTTP_METHOD_TRACE
)

func HTTPMethodInt(method int) string {
	switch method {
	case HTTP_METHOD_GET:
		return "GET"
	case HTTP_METHOD_POST:
		return "POST"
	case HTTP_METHOD_PUT:
		return "PUT"
	case HTTP_METHOD_DELETE:
		return "DELETE"
	case HTTP_METHOD_HEAD:
		return "HEAD"
	case HTTP_METHOD_OPTIONS:
		return "OPTIONS"
	case HTTP_METHOD_PATCH:
		return "PATCH"
	default:
		return ""
	}
}

func HTTPMethodString(method string) int {
	switch method {
	case "GET":
		return HTTP_METHOD_GET
	case "POST":
		return HTTP_METHOD_POST
	case "PUT":
		return HTTP_METHOD_PUT
	case "DELETE":
		return HTTP_METHOD_DELETE
	case "HEAD":
		return HTTP_METHOD_HEAD
	case "OPTIONS":
		return HTTP_METHOD_OPTIONS
	case "PATCH":
		return HTTP_METHOD_PATCH
	default:
		return HTTP_METHOD_UNKNOWN
	}
}

func FindHTTPURI(payload string) string {
	split := strings.Split(payload, " ")
	if len(split) < 2 {
		return ""
	}
	if HTTPMethodString(split[0]) == HTTP_METHOD_UNKNOWN {
		return ""
	}
	uri := split[1]
	startOffset := -1
	if (len(uri) > 0) && (uri[:1] == "/") {
		startOffset = 0
	} else if len(uri) > 8 && (uri[:8] == "https://") {
		off := strings.Index(uri[8:], "/")
		if off == -1 {
			return "/"
		}
		startOffset = 8 + off
	} else if len(uri) > 7 && (uri[:7] == "http://") {
		off := strings.Index(uri[7:], "/")
		if off == -1 {
			return "/"
		}
		startOffset = 7 + off
	}
	if startOffset == -1 {
		return ""
	}

	endOffset := strings.Index(uri, "?")
	if endOffset > 0 && startOffset < endOffset {
		return uri[startOffset:endOffset]
	}
	return uri[startOffset:]
}

func ParseHTTPVersion(v uint32) string {
	return fmt.Sprintf("%d.%d", v>>16, v&0xFFFF)
}

func ConnNotNeedToFilter(conn ConnectionInfo) bool {
	if (conn.Saddr[0]|conn.Saddr[1]|conn.Saddr[2]|conn.Saddr[3]) == 0 ||
		(conn.Daddr[0]|conn.Daddr[1]|conn.Daddr[2]|conn.Daddr[3]) == 0 ||
		conn.Sport == 0 || conn.Dport == 0 {
		return false
	}
	if netflow.ConnAddrIsIPv4(conn.Meta) { // IPv4
		if (conn.Saddr[3]&0xff) == 127 && (conn.Daddr[3]&0xff) == 127 {
			return false
		}
	} else { // IPv6
		if conn.Saddr[2] == 0xffff0000 && conn.Daddr[2] == 0xffff0000 {
			if (conn.Saddr[3]&0xff) == 127 && (conn.Daddr[3]&0xff) == 127 {
				return false
			}
		} else if (conn.Saddr[0]|conn.Saddr[1]|conn.Saddr[2]) == 0 && conn.Saddr[3] == 1 &&
			(conn.Daddr[0]|conn.Daddr[1]|conn.Daddr[2]) == 0 && conn.Daddr[3] == 1 {
			return false
		}
	}
	return true
}

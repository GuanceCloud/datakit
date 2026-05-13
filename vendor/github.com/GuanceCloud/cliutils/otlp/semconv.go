// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package otlp

const (
	AttrServiceName = "service.name"

	AttrDBSystem        = "db.system"
	AttrMessagingSystem = "messaging.system"
	AttrRPCSystem       = "rpc.system"

	EventException = "exception"

	AttrExceptionType       = "exception.type"
	AttrExceptionMessage    = "exception.message"
	AttrExceptionStacktrace = "exception.stacktrace"
)

const (
	MetricTagUnit = "unit"
	MetricTagLE   = "le"

	MetricSuffixBucket = "_bucket"
	MetricSuffixSum    = "_sum"
	MetricSuffixCount  = "_count"
	MetricSuffixAvg    = "_avg"
	MetricSuffixMin    = "_min"
	MetricSuffixMax    = "_max"
	MetricSuffixInf    = "+Inf"
)

var SpanKindNames = map[int32]string{
	0: "unspecified",
	1: "internal",
	2: "server",
	3: "client",
	4: "producer",
	5: "consumer",
}

var PublicAttributeAliases = map[string]string{
	"db.system":    "db_system",
	"db.operation": "db_operation",
	"db.name":      "db_name",
	"db.statement": "db_statement",

	"server.address":       "server_address",
	"net.host.name":        "net_host_name",
	"server.port":          "server_port",
	"net.host.port":        "net_host_port",
	"network.peer.address": "network_peer_address",
	"network.peer.port":    "network_peer_port",
	"network.transport":    "network_transport",

	"http.request.method":       "http_method",
	"http.method":               "http_method",
	"error.type":                "error_type",
	"http.response.status_code": "http_status_code",
	"http.status_code":          "http_status_code",
	"http.route":                "http_route",
	"http.target":               "http_target",
	"http.scheme":               "http_scheme",
	"http.url":                  "http_url",
	"url.full":                  "url_full",
	"url.scheme":                "url_scheme",
	"url.path":                  "url_path",
	"url.query":                 "url_query",
	"client.address":            "client_address",
	"client.port":               "client_port",

	"messaging.system":           "messaging_system",
	"messaging.operation":        "messaging_operation",
	"messaging.message.id":       "messaging_message.id",
	"messaging.destination.name": "messaging_destination.name",

	"rpc.service": "rpc_service",
	"rpc.system":  "rpc_system",

	"exception":            "exception",
	"exception.type":       "exception_type",
	"exception.message":    "exception_message",
	"exception.stacktrace": "exception_stacktrace",

	"container.name": "container_name",
	"process.pid":    "process_pid",
	"project":        "project",
	"version":        "version",
	"env":            "env",
	"host":           "host",
	"pod_name":       "pod_name",
	"pod_namespace":  "pod_namespace",

	"telemetry.sdk.language": "sdk_language",
	"telemetry.sdk.name":     "sdk_name",
	"telemetry.sdk.version":  "sdk_version",
}

var droppedMetricAttributeKeys = map[string]struct{}{
	"process.command_line":        {},
	"process.executable.path":     {},
	"process.runtime.description": {},
	"process.runtime.name":        {},
	"process.runtime.version":     {},
	"telemetry.distro.name":       {},
	"telemetry.distro.version":    {},
	"telemetry.sdk.language":      {},
	"telemetry.sdk.name":          {},
	"telemetry.sdk.version":       {},
}

// SpanKindName returns the readable span kind. Unknown values fallback to "unspecified".
func SpanKindName(kind int32) string {
	if name, ok := SpanKindNames[kind]; ok {
		return name
	}

	return SpanKindNames[0]
}

// PublicAttributeAlias returns the normalized alias of a semantic-convention attribute key.
func PublicAttributeAlias(key string) (string, bool) {
	alias, ok := PublicAttributeAliases[key]

	return alias, ok
}

// ShouldDropMetricAttribute reports whether a resource or datapoint attribute is ignored for metrics.
func ShouldDropMetricAttribute(key string) bool {
	_, ok := droppedMetricAttributeKeys[key]

	return ok
}

// SystemNameForService returns db/rpc/messaging system value for service-name fallback.
func SystemNameForService(attrs map[string]string) string {
	for _, key := range []string{AttrDBSystem, AttrRPCSystem, AttrMessagingSystem} {
		if system := attrs[key]; system != "" {
			return system
		}
	}

	return ""
}

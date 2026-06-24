// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package coredns

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type (
	aCLMeasurement      struct{}
	cacheMeasurement    struct{}
	dnsSecMeasurement   struct{}
	forwardMeasurement  struct{}
	grpcMeasurement     struct{}
	hostsMeasurement    struct{}
	templateMeasurement struct{}
	promMeasurement     struct{}
)

func (*aCLMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "coredns_acl",
		Cat:    point.Metric,
		Desc:   "CoreDNS acl plugin metrics scraped from Prometheus output.",
		DescZh: "从 Prometheus 输出采集的 CoreDNS ACL 插件指标，记录被阻止、过滤、允许和丢弃的 DNS 请求。",
		Fields: map[string]interface{}{
			"blocked_requests_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of DNS requests being blocked",
			},
			"filtered_requests_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of DNS requests being filtered",
			},
			"allowed_requests_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of DNS requests being allowed",
			},
			"dropped_requests_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of DNS requests being dropped",
			},
		},
		Tags: map[string]interface{}{
			"server":   inputs.NewTagInfo("Server responsible for the request."),
			"zone":     inputs.NewTagInfo("Zone name used for the request/response."),
			"instance": inputs.NewTagInfo("Instance endpoint"),
			"host":     inputs.NewTagInfo("Host name"),
		},
	}
}

func (*cacheMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "coredns_cache",
		Cat:    point.Metric,
		Desc:   "CoreDNS cache plugin metrics scraped from Prometheus output.",
		DescZh: "从 Prometheus 输出采集的 CoreDNS 缓存插件指标，涵盖缓存条目、请求、命中、未命中、预取、丢弃和淘汰。",
		Fields: map[string]interface{}{
			"entries": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of elements in the cache",
			},
			"requests_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The count of cache requests",
			},
			"hits_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The count of cache hits",
			},
			"misses_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The count of cache misses. Deprecated, derive misses from cache hits/requests counters",
			},
			"prefetch_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The number of times the cache has prefetched a cached item.",
			},
			"drops_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Number of responses that were not cached because the reply was malformed.",
			},
			"served_stale_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The number of requests served from stale cache entries",
			},
			"evictions_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The count of cache evictions",
			},
		},
		Tags: map[string]interface{}{
			"server":   inputs.NewTagInfo("Server responsible for the request"),
			"zones":    inputs.NewTagInfo("Zone name used for the request/response"),
			"type":     inputs.NewTagInfo("Cache type"),
			"instance": inputs.NewTagInfo("Instance endpoint"),
			"host":     inputs.NewTagInfo("Host name"),
		},
	}
}

func (*dnsSecMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "coredns_dnssec",
		Cat:    point.Metric,
		Desc:   "CoreDNS dnssec plugin metrics scraped from Prometheus output.",
		DescZh: "从 Prometheus 输出采集的 CoreDNS DNSSEC 插件指标，记录 DNSSEC 缓存条目、命中和未命中。",
		Fields: map[string]interface{}{
			"cache_entries": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of elements in the dnssec cache",
			},
			"cache_hits_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The count of cache hits",
			},
			"cache_misses_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The count of cache misses",
			},
		},
		Tags: map[string]interface{}{
			"server":   inputs.NewTagInfo("Server responsible for the request"),
			"type":     inputs.NewTagInfo("signature"),
			"instance": inputs.NewTagInfo("Instance endpoint"),
			"host":     inputs.NewTagInfo("Host name"),
		},
	}
}

func (*forwardMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "coredns_forward",
		Cat:    point.Metric,
		Desc:   "CoreDNS forward plugin metrics scraped from Prometheus output.",
		DescZh: "从 Prometheus 输出采集的 CoreDNS 转发插件指标，涵盖上游请求、响应、耗时、健康检查和连接缓存。",
		Fields: map[string]interface{}{
			"requests_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of requests made per upstream",
			},
			"responses_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of responses received per upstream",
			},
			"request_duration_seconds": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Histogram,
				Unit:     inputs.DurationSecond,
				Desc:     "Histogram of the time each request took",
			},
			"healthcheck_failures_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of the number of failed health checks",
			},
			"healthcheck_broken_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of the number of complete failures of the health checks",
			},
			"max_concurrent_rejects_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of the number of queries rejected because the concurrent queries were at maximum",
			},
			"conn_cache_hits_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of connection cache hits per upstream and protocol",
			},
			"conn_cache_misses_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of connection cache misses per upstream and protocol",
			},
		},
		Tags: map[string]interface{}{
			"to":       inputs.NewTagInfo("Upstream server"),
			"rcode":    inputs.NewTagInfo("Upstream returned `RCODE`"),
			"proto":    inputs.NewTagInfo("Transport protocol like `udp`, `tcp`, `tcp-tls`"),
			"instance": inputs.NewTagInfo("Instance endpoint"),
			"host":     inputs.NewTagInfo("Host name"),
		},
	}
}

func (*grpcMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "coredns_grpc",
		Cat:    point.Metric,
		Desc:   "CoreDNS grpc plugin metrics scraped from Prometheus output.",
		DescZh: "从 Prometheus 输出采集的 CoreDNS gRPC 插件指标，记录上游请求、响应和请求耗时。",
		Fields: map[string]interface{}{
			"requests_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of requests made per upstream.",
			},
			"responses_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of responses received per upstream.",
			},
			"request_duration_seconds": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Histogram,
				Unit:     inputs.DurationSecond,
				Desc:     "Histogram of the time each request took",
			},
		},
		Tags: map[string]interface{}{
			"to":       inputs.NewTagInfo("Upstream server"),
			"rcode":    inputs.NewTagInfo("Upstream returned `RCODE`"),
			"instance": inputs.NewTagInfo("Instance endpoint"),
			"host":     inputs.NewTagInfo("Host name"),
		},
	}
}

func (*hostsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "coredns_hosts",
		Cat:    point.Metric,
		Desc:   "CoreDNS hosts plugin metrics scraped from Prometheus output.",
		DescZh: "从 Prometheus 输出采集的 CoreDNS hosts 插件指标，记录主机条目数量和 hosts 文件最近重载时间。",
		Fields: map[string]interface{}{
			"entries": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The combined number of entries in hosts and Corefile",
			},
			"reload_timestamp_seconds": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.TimestampSec,
				Desc:     "The timestamp of the last reload of hosts file",
			},
		},
		Tags: map[string]interface{}{
			"instance": inputs.NewTagInfo("Instance endpoint"),
			"host":     inputs.NewTagInfo("Host name"),
		},
	}
}

func (*templateMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "coredns_template",
		Cat:    point.Metric,
		Desc:   "CoreDNS template plugin metrics scraped from Prometheus output.",
		DescZh: "从 Prometheus 输出采集的 CoreDNS template 插件指标，记录模板匹配、模板执行失败和资源记录构造失败。",
		Fields: map[string]interface{}{
			"matches_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of template regex matches",
			},
			"failures_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of go template failures",
			},
			"rr_failures_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of mis-templated RRs",
			},
		},
		Tags: map[string]interface{}{
			"server":   inputs.NewTagInfo("Server responsible for the request"),
			"zone":     inputs.NewTagInfo("Zone name"),
			"view":     inputs.NewTagInfo("View name"),
			"class":    inputs.NewTagInfo("The query class (usually `IN`)"),
			"type":     inputs.NewTagInfo("The DNS resource record type requested, for example `PTR`."),
			"section":  inputs.NewTagInfo("Section label"),
			"template": inputs.NewTagInfo("Template label"),
			"instance": inputs.NewTagInfo("Instance endpoint"),
			"host":     inputs.NewTagInfo("Host name"),
		},
	}
}

func (*promMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "coredns",
		Cat:    point.Metric,
		Desc:   "CoreDNS server and plugin metrics scraped from Prometheus output.",
		DescZh: "从 Prometheus 输出采集的 CoreDNS 服务端和插件综合指标，涵盖 DNS 请求、响应、耗时、健康检查和插件运行状态。",
		Fields: map[string]interface{}{
			"dns64_requests_translated_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of DNS requests translated by dns64",
			},
			"health_request_duration_seconds": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Histogram,
				Unit:     inputs.DurationSecond,
				Desc:     "Histogram of the time (in seconds) each request took",
			},
			"health_request_failures_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "The number of times the health checks failed",
			},
			"kubernetes_dns_programming_duration_seconds": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Histogram,
				Unit:     inputs.DurationSecond,
				Desc:     "Histogram of the time (in seconds) it took to program a dns instance",
			},
			"local_localhost_requests_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of localhost. `domain` requests",
			},
			"build_info": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Bool,
				Desc:     "A metric with a constant '1' value labeled by version, revision, and Go version from which CoreDNS was built",
			},
			"dns_requests_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of DNS requests made per zone, protocol and family",
			},
			"dns_request_duration_seconds": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Histogram,
				Unit:     inputs.DurationSecond,
				Desc:     "Histogram of the time (in seconds) each request took per zone",
			},
			"dns_request_size_bytes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Histogram,
				Unit:     inputs.SizeByte,
				Desc:     "Size of the `EDNS0` UDP buffer in bytes (64K for TCP) per zone and protocol",
			},
			"dns_do_requests_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of DNS requests with DO bit set per zone",
			},
			"dns_response_size_bytes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Histogram,
				Unit:     inputs.SizeByte,
				Desc:     "Size of the returned response in bytes",
			},
			"dns_responses_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of response status codes",
			},
			"dns_panics_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of CoreDNS panics.",
			},
			"dns_plugin_enabled": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Bool,
				Desc:     "A metric that indicates whether a plugin is enabled on per server and zone basis",
			},
			"dns_https_responses_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of DoH responses per server and http status code",
			},
			"reload_failed_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of the number of failed reload attempts",
			},
			"reload_version_info": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Bool,
				Desc:     "Constant value of 1 labeled with the reload hash type and hash value.",
			},
			"autopath_success_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Counter of requests that did auto path",
			},
		},
		Tags: map[string]interface{}{
			"server":       inputs.NewTagInfo("Server responsible for the request"),
			"service_kind": inputs.NewTagInfo("Service kind"),
			"version":      inputs.NewTagInfo("CoreDNS version"),
			"revision":     inputs.NewTagInfo("Git commit revision used to build CoreDNS."),
			"goversion":    inputs.NewTagInfo("Golang version"),
			"zone":         inputs.NewTagInfo("Zone name used for the request/response"),
			"view":         inputs.NewTagInfo("View name"),
			"proto":        inputs.NewTagInfo("Transport protocol like `udp`, `tcp`, `tcp-tls`"),
			"rcode":        inputs.NewTagInfo("Upstream returned `RCODE`"),
			"plugin":       inputs.NewTagInfo("The name of the plugin that made the write to the client"),
			"name":         inputs.NewTagInfo("Handler name"),
			"status":       inputs.NewTagInfo("HTTPS response status code."),
			"hash":         inputs.NewTagInfo("Hash algorithm label for reload_version_info, such as `sha512`."),
			"value":        inputs.NewTagInfo("Hash value label for reload_version_info."),
			"instance":     inputs.NewTagInfo("Instance endpoint"),
			"host":         inputs.NewTagInfo("Host name"),
		},
	}
}

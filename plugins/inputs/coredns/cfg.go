package coredns

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

type (
	ACLMeasurement      struct{ measurement }
	CacheMeasurement    struct{ measurement }
	DNSSecMeasurement   struct{ measurement }
	ForwardMeasurement  struct{ measurement }
	GrpcMeasurement     struct{ measurement }
	HostsMeasurement    struct{ measurement }
	TemplateMeasurement struct{ measurement }
	PromMeasurement     struct{ measurement }
)

func (m *measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *ACLMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "coredns_acl",
		Fields: map[string]interface{}{
			"acl_blocked_requests_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "被拦截的DNS请求个数",
			},
			"acl_allowed_requests_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "被放行的DNS请求个数",
			},
		},
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("监听服务地址"),
			"zone":   inputs.NewTagInfo("请求所属区域"),
		},
	}
}

func (m *CacheMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "coredns_cache",
		Fields: map[string]interface{}{
			"cache_entries": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "缓存总数",
			},
			"cache_hits_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "缓存命中个数",
			},
			"cache_misses_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "缓存miss个数",
			},
			"cache_prefetch_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "缓存预读取个数",
			},
			"cache_drops_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "被排除在缓存外的响应个数",
			},
			"cache_served_stale_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "提供过时缓存的请求个数",
			},
		},
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("监听服务地址"),
			"type":   inputs.NewTagInfo("缓存类型"),
		},
	}
}

func (m *DNSSecMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "coredns_dnssec",
		Fields: map[string]interface{}{
			"dnssec_cache_entries": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "dnssec缓存总数",
			},
			"dnssec_cache_hits_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "dnssec缓存命中个数",
			},
			"dnssec_cache_misses_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "dnssec缓存miss个数",
			},
		},
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("监听服务地址"),
			"type":   inputs.NewTagInfo("签名"),
		},
	}
}

func (m *ForwardMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "coredns_forward",
		Fields: map[string]interface{}{
			"forward_requests_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "转发给每个上游的请求个数",
			},
			"forward_responses_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "从每个上游得到的RCODE响应个数",
			},
			"forward_request_duration_seconds": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "请求时长",
			},
			"forward_healthcheck_failures_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "每个上游健康检查失败个数",
			},
			"forward_healthcheck_broken_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "所有上游均不健康次数",
			},
			"forward_max_concurrent_rejects_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "由于并发达到峰值而被拒绝的查询个数",
			},
		},
		Tags: map[string]interface{}{
			"to":    inputs.NewTagInfo("上游服务器"),
			"rcode": inputs.NewTagInfo("上游返回的RCODE"),
			"proto": inputs.NewTagInfo("传输协议"),
		},
	}
}

func (m *GrpcMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "coredns_grpc",
		Fields: map[string]interface{}{
			"grpc_request_duration_seconds": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "grpc与上游交互时长",
			},
			"grpc_requests_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "grpc在每个上游查询个数",
			},
			"grpc_responses_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "grpc在每个上游得到的RCODE响应个数",
			},
		},
		Tags: map[string]interface{}{
			"to":    inputs.NewTagInfo("上游服务器"),
			"rcode": inputs.NewTagInfo("上游返回的RCODE"),
		},
	}
}

func (m *HostsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "coredns_hosts",
		Fields: map[string]interface{}{
			"hosts_entries": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "hosts总条数",
			},
			"hosts_reload_timestamp_seconds": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.TimestampSec,
				Desc:     "最后一次重载hosts文件的时间戳",
			},
		},
		Tags: map[string]interface{}{},
	}
}

func (m *TemplateMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "coredns_template",
		Fields: map[string]interface{}{
			"template_matches_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "正则匹配的请求总数",
			},
			"template_failures_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "Go模板失败次数",
			},
			"template_rr_failures_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "因模板资源记录无效而无法处理的次数",
			},
		},
		Tags: map[string]interface{}{
			"server":   inputs.NewTagInfo("监听服务地址"),
			"regex":    inputs.NewTagInfo("正则表达式"),
			"section":  inputs.NewTagInfo("所属板块"),
			"template": inputs.NewTagInfo("模板"),
		},
	}
}

func (m *PromMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "coredns",
		Fields: map[string]interface{}{
			"dns_requests_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "查询总数",
			},
			"dns_request_duration_seconds": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "处理每个查询的时长",
			},
			"dns_request_size_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "请求大小(以byte计)",
			},
			"dns_responses_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "对每个zone和RCODE的响应总数",
			},
			"dns_response_size_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "响应大小(以byte计)",
			},
		},
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("监听服务地址"),
			"zone":   inputs.NewTagInfo("请求所属区域"),
			"type":   inputs.NewTagInfo("查询类型"),
			"proto":  inputs.NewTagInfo("传输协议"),
			"family": inputs.NewTagInfo("IP地址家族"),
			"rcode":  inputs.NewTagInfo("上游返回的RCODE"),
		},
	}
}

package ddtrace

//go:generate msgp -io=false

import itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"

var ddtraceSpanType = map[string]string{
	"consul":        itrace.SPAN_SERVICE_APP,
	"cache":         itrace.SPAN_SERVICE_CACHE,
	"memcached":     itrace.SPAN_SERVICE_CACHE,
	"redis":         itrace.SPAN_SERVICE_CACHE,
	"aerospike":     itrace.SPAN_SERVICE_DB,
	"cassandra":     itrace.SPAN_SERVICE_DB,
	"db":            itrace.SPAN_SERVICE_DB,
	"elasticsearch": itrace.SPAN_SERVICE_DB,
	"leveldb":       itrace.SPAN_SERVICE_DB,
	"mongodb":       itrace.SPAN_SERVICE_DB,
	"sql":           itrace.SPAN_SERVICE_DB,
	"http":          itrace.SPAN_SERVICE_WEB,
	"web":           itrace.SPAN_SERVICE_WEB,
	"":              itrace.SPAN_SERVICE_CUSTOM,
	"benchmark":     itrace.SPAN_SERVICE_CUSTOM,
	"build":         itrace.SPAN_SERVICE_CUSTOM,
	"custom":        itrace.SPAN_SERVICE_CUSTOM,
	"datanucleus":   itrace.SPAN_SERVICE_CUSTOM,
	"dns":           itrace.SPAN_SERVICE_CUSTOM,
	"graphql":       itrace.SPAN_SERVICE_CUSTOM,
	"grpc":          itrace.SPAN_SERVICE_CUSTOM,
	"hibernate":     itrace.SPAN_SERVICE_CUSTOM,
	"queue":         itrace.SPAN_SERVICE_CUSTOM,
	"rpc":           itrace.SPAN_SERVICE_CUSTOM,
	"soap":          itrace.SPAN_SERVICE_CUSTOM,
	"template":      itrace.SPAN_SERVICE_CUSTOM,
	"test":          itrace.SPAN_SERVICE_CUSTOM,
	"worker":        itrace.SPAN_SERVICE_CUSTOM,
}

//nolint:lll
type DDSpan struct {
	Service  string             `protobuf:"bytes,1,opt,name=service,proto3" json:"service" msg:"service" codec:"service"`
	Name     string             `protobuf:"bytes,2,opt,name=name,proto3" json:"name" msg:"name" codec:"name"`
	Resource string             `protobuf:"bytes,3,opt,name=resource,proto3" json:"resource" msg:"resource" codec:"resource"`
	TraceID  uint64             `protobuf:"varint,4,opt,name=traceID,proto3" json:"trace_id" msg:"trace_id" codec:"trace_id"`
	SpanID   uint64             `protobuf:"varint,5,opt,name=spanID,proto3" json:"span_id" msg:"span_id" codec:"span_id"`
	ParentID uint64             `protobuf:"varint,6,opt,name=parentID,proto3" json:"parent_id" msg:"parent_id" codec:"parent_id"`
	Start    int64              `protobuf:"varint,7,opt,name=start,proto3" json:"start" msg:"start" codec:"start"`
	Duration int64              `protobuf:"varint,8,opt,name=duration,proto3" json:"duration" msg:"duration" codec:"duration"`
	Error    int32              `protobuf:"varint,9,opt,name=error,proto3" json:"error" msg:"error" codec:"error"`
	Meta     map[string]string  `protobuf:"bytes,10,rep,name=meta" json:"meta" msg:"meta" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3" codec:"meta"`
	Metrics  map[string]float64 `protobuf:"bytes,11,rep,name=metrics" json:"metrics" msg:"metrics" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"fixed64,2,opt,name=value,proto3" codec:"metrics"`
	Type     string             `protobuf:"bytes,12,opt,name=type,proto3" json:"type" msg:"type" codec:"type"`
}

type DDTrace []*DDSpan

type DDTraces []DDTrace

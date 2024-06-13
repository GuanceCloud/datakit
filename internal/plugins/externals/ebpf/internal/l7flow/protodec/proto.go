//go:build linux
// +build linux

package protodec

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/k8sinfo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/spanid"
)

type L7Protocol int

const (
	DirectionOutgoing = "outgoing"
	DirectionIncoming = "incoming"
	DirectionUnknown  = "unknown"

	NoValue = "N/A"
)

const (
	ProtoUnknown L7Protocol = iota
	ProtoHTTP
	ProtoHTTP2
	ProtoGRPC
	ProtoMySQL
	ProtoRedis
)

var _protoSet = &ProtoSet{
	protoDect: make([]protoDectFn, 0, 8),
	protoAgg:  make(map[L7Protocol](func(L7Protocol) AggPool)),
}

type AggPool interface {
	Proto() L7Protocol
	Obs(conn *comm.ConnectionInfo, data *ProtoData)
	Export(tags map[string]string, k8sInfo *k8sinfo.K8sNetInfo) []*point.Point
	Cleanup()
}

type protoDectFn func([]byte, int) (L7Protocol, ProtoDecPipe, bool)

type ProtoSet struct {
	protoDect []protoDectFn
	protoAgg  map[L7Protocol](func(L7Protocol) AggPool)
}

func (p *ProtoSet) RegisterDetector(fnDect protoDectFn) {
	p.protoDect = append(p.protoDect, fnDect)
}

func (p *ProtoSet) RegisterAggregator(proto L7Protocol, fn func(L7Protocol) AggPool) {
	p.protoAgg[proto] = fn
}

func (p L7Protocol) StringLower() string {
	switch p { //nolint:exhaustive
	case ProtoHTTP:
		return "http"
	case ProtoHTTP2:
		return "http2"
	case ProtoGRPC:
		return "grpc"
	case ProtoMySQL:
		return "mysql"
	case ProtoRedis:
		return "redis"
	default:
		return "unknown"
	}
}

func (p L7Protocol) String() string {
	switch p { //nolint:exhaustive
	case ProtoHTTP:
		return "HTTP"
	case ProtoHTTP2:
		return "HTTP2"
	case ProtoGRPC:
		return "gRPC"
	case ProtoMySQL:
		return "MySQL"
	case ProtoRedis:
		return "Redis"
	default:
		return "unknown"
	}
}

type ProtoMeta struct {
	TraceID      spanid.ID128
	ParentSpanID spanid.ID64
	SpanHexEnc   bool
	SampledSpan  bool

	ReqTCPSeq  uint32
	RespTCPSeq uint32
	InnerID    int64

	// req, resp ;; sys_thread, user_thread
	Threads [2][2]int32
}

type ProtoData struct {
	Meta      ProtoMeta
	KVs       point.KVs
	Cost      int64
	Duration  int64
	Direction comm.Direcion
	L7Proto   L7Protocol
	Time      int64
	KTime     uint64
}

type threadTrace interface {
	Insert(d comm.Direcion, thrID [2]int32, ts0_1 uint64) (id int64)
	GetInnerID(thrID [2]int32, ts uint64) int64
}

type ProtoDecPipe interface {
	Proto() L7Protocol

	// TODO: 线程跟踪放到外面，由函数返回当前数据是 req/resp/data/error/... 状态
	Decode(txRx comm.NICDirection, data *comm.NetwrkData, ts int64, thrTr threadTrace)

	Export(force bool) []*ProtoData
	ConnClose()
}

func ProtoDetector(data []byte, actSize int) (L7Protocol, ProtoDecPipe, bool) {
	for _, fn := range _protoSet.protoDect {
		proto, pipe, ok := fn(data, actSize)
		if ok {
			return proto, pipe, ok
		}
	}
	return ProtoUnknown, nil, false
}

func NewProtoAggregators() map[L7Protocol]AggPool {
	r := make(map[L7Protocol]AggPool)
	for k, v := range _protoSet.protoAgg {
		r[k] = v(k)
	}
	return r
}

func Init() {
	// HTTP
	_protoSet.RegisterDetector(HTTPProtoDetect)
	// HTTP2 and gRPC
	_protoSet.RegisterDetector(H2ProtoDetect)

	_protoSet.RegisterDetector(RedisProtoDetect)

	_protoSet.RegisterDetector(MysqlProtoDetect)

	_protoSet.RegisterAggregator(ProtoHTTP, newHTTPAggP)
}

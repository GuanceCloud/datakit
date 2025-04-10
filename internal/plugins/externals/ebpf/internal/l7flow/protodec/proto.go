//go:build linux
// +build linux

package protodec

import (
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
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
	ProtoAMQP
	ProtoPgsql
)

var _protoSet = &ProtoSet{
	protoDect: map[L7Protocol]protoDectFn{},
	protoAgg:  map[L7Protocol](func(L7Protocol) AggPool){},
}

type AggPool interface {
	Proto() L7Protocol
	Obs(conn *comm.ConnectionInfo, data *ProtoData)
	Export(tags map[string]string, k8sInfo *cli.K8sInfo) []*point.Point
	Cleanup()
}

type protoDectFn func([]byte, int) (L7Protocol, ProtoDecPipe, bool)

type ProtoSet struct {
	protoDect map[L7Protocol]protoDectFn
	protoAgg  map[L7Protocol](func(L7Protocol) AggPool)
}

func (p *ProtoSet) RegisterDetector(proto L7Protocol, fnDect protoDectFn) {
	p.protoDect[proto] = fnDect
}

func (p *ProtoSet) RegisterAggregator(proto L7Protocol, fn func(L7Protocol) AggPool) {
	p.protoAgg[proto] = fn
}

func (p *ProtoSet) ProtoDetector(data []byte, actSize int) (L7Protocol, ProtoDecPipe, bool) {
	for _, fn := range p.protoDect {
		proto, pipe, ok := fn(data, actSize)
		if ok {
			return proto, pipe, ok
		}
	}
	return ProtoUnknown, nil, false
}

func (p *ProtoSet) NewProtoAggregators() map[L7Protocol]AggPool {
	r := make(map[L7Protocol]AggPool)
	for k, v := range p.protoAgg {
		r[k] = v(k)
	}
	return r
}

func SubProtoSet(protos ...L7Protocol) *ProtoSet {
	newP := &ProtoSet{
		protoDect: map[L7Protocol]protoDectFn{},
		protoAgg:  map[L7Protocol](func(L7Protocol) AggPool){},
	}
	for _, proto := range protos {
		if d, ok := _protoSet.protoDect[proto]; ok {
			newP.protoDect[proto] = d
		}
		if a, ok := _protoSet.protoAgg[proto]; ok {
			newP.protoAgg[proto] = a
		}
	}
	return newP
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
	case ProtoAMQP:
		return "amqp"
	case ProtoPgsql:
		return "postgresql"
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
	case ProtoAMQP:
		return "AMQP"
	case ProtoPgsql:
		return "PostgreSQL"
	default:
		return "unknown"
	}
}

func ProtocalNum(p string) L7Protocol {
	p = strings.ToLower(p)
	switch p { //nolint:exhaustive
	case "http":
		return ProtoHTTP
	case "http2":
		return ProtoHTTP2
	case "grpc":
		return ProtoGRPC
	case "mysql":
		return ProtoMySQL
	case "redis":
		return ProtoRedis
	case "amqp":
		return ProtoAMQP
	case "postgresql":
		return ProtoPgsql
	default:
		return ProtoUnknown
	}
}

var AllProtos = []L7Protocol{ProtoHTTP, ProtoHTTP2, ProtoGRPC, ProtoMySQL, ProtoRedis, ProtoPgsql, ProtoAMQP}

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
	Insert(d comm.Direcion, pid int32, thrID [2]int32, ts0_1 uint64) (id int64)
	GetInnerID(pid int32, thrID [2]int32, ts uint64) int64
}

type ProtoDecPipe interface {
	Proto() L7Protocol

	// TODO: 线程跟踪放到外面，由函数返回当前数据是 req/resp/data/error/... 状态
	Decode(txRx comm.NICDirection, data *comm.NetwrkData, ts int64, thrTr threadTrace)

	Export(force bool) []*ProtoData
	ConnClose()
}

func Init() {
	// HTTP
	_protoSet.RegisterDetector(ProtoHTTP, HTTPProtoDetect)
	// HTTP2 and gRPC
	_protoSet.RegisterDetector(ProtoHTTP2, H2ProtoDetect)
	_protoSet.RegisterDetector(ProtoGRPC, H2ProtoDetect)

	_protoSet.RegisterDetector(ProtoAMQP, AMQPProtoDetect)

	_protoSet.RegisterDetector(ProtoRedis, RedisProtoDetect)

	_protoSet.RegisterDetector(ProtoMySQL, MysqlProtoDetect)

	_protoSet.RegisterDetector(ProtoPgsql, PgsqlProtoDetect)

	_protoSet.RegisterAggregator(ProtoHTTP, newHTTPAggP)
}

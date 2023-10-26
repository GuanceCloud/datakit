// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package spans connect span
package spans

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"io"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/oklog/ulid"
	expRand "golang.org/x/exp/rand"
)

var l = logger.DefaultSLogger("ebpftrace-span")

const (
	// 原始数据中必须存在.
	Direction = "direction"

	// NetTraceID     = "net_trace_id".

	AppTraceIDL    = "app_trace_id_l"
	AppTraceIDH    = "app_trace_id_h"
	AppParentIDL   = "app_parent_id_l"
	AppSpanSampled = "app_span_sampled"
	AppTraceEncode = "app_trace_encode"

	ThrTraceID = "thread_trace_id"
	ReqSeq     = "req_seq"
	RespSeq    = "resp_seq"

	EBPFSpanType = "ebpf_span_type"

	// 插入时生成.
	SpanID = "span_id"

	ParentID = "parent_id"
	TraceID  = "trace_id"

	EBPFTraceID  = "ebpf_trace_id"
	EBPFParentID = "ebpf_parent_id"

	_SynTime = "trace_gen_syn_time_ns"

	RejectTrace     = -1
	AutoRejectRrace = 0
	KeepTrace       = 1
)

const (
	OUTG  = "outgoing"
	INCO  = "incoming"
	OUTGB = false
	INCOB = true

	ENCHEX = "hex" // default
	ENCDEC = "dec"

	SpanTypEntry  = "entry"
	SpanTypEntryB = true
)

type EBPFSpan struct {
	pre, next *EBPFSpan
	childs    []*EBPFSpan

	directionIn bool
	espanEntry  bool

	traceID  ID128
	spanID   ID64
	parentID ID64

	// req_seq, resp_seq
	netTraceID ID128
	thrTraceID ID64

	eTraceID  ID128
	eParentID ID64

	// w3c have high 64bit
	aTraceID    ID128
	aParentID   ID64
	aSampled    int
	idEncodeDec bool

	used bool
	// base16 bool # auto
}
type ULID struct {
	// 使用 ulid 是期望其作为 id 时能尽可能的按时间索引
	_rand io.Reader
	sync.Mutex
}

func NewULID() (*ULID, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	v := binary.LittleEndian.Uint64(b)
	r := expRand.New(expRand.NewSource(v))
	r.Seed(v)

	return &ULID{
		_rand: r,
	}, nil
}

func (dkid *ULID) ID() (*ID128, bool) {
	dkid.Lock()
	defer dkid.Unlock()
	if dkid._rand != nil {
		id, err := ulid.New(uint64(time.Now().UnixMilli()), dkid._rand)
		if err != nil {
			return nil, false
		}
		i := [16]byte(id)
		return &ID128{
			Low:  binary.LittleEndian.Uint64(i[:8]),
			High: binary.BigEndian.Uint64(i[8:]),
		}, true
	} else {
		b := make([]byte, 25)
		_, err := rand.Read(b)
		if err != nil {
			return nil, false
		}

		id, err := ulid.New(uint64(time.Now().UnixMilli()), bytes.NewReader(b))
		if err != nil {
			return nil, false
		}
		i := [16]byte(id)
		return &ID128{
			Low:  binary.LittleEndian.Uint64(i[:8]),
			High: binary.BigEndian.Uint64(i[8:]),
		}, true
	}
}

type ID64 uint64

type ID128 struct {
	Low  uint64
	High uint64
}

func (id ID64) StringHex() string {
	if id == 0 {
		return "0"
	}

	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return hex.EncodeToString(b)
}

func (id ID64) StringDec() string {
	return strconv.FormatUint(uint64(id), 10)
}

func (id ID64) Byte() []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(id))
	return b
}

func (id ID64) Zero() bool {
	return id == 0
}

func (id ID128) StringHex() string {
	if id.Low == 0 && id.High == 0 {
		return "0"
	}
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[8:], id.Low)
	binary.BigEndian.PutUint64(b[:8], id.High)
	return hex.EncodeToString(b)
}

func (id ID128) StringDec() string {
	return strconv.FormatUint(id.Low, 10)
}

func (id ID128) Byte() []byte {
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, id.Low)
	binary.LittleEndian.PutUint64(b[8:], id.High)
	return b
}

func (id ID128) Zero() bool {
	return id.Low == 0 && id.High == 0
}

// constants used for the Knuth hashing, same as dd-agent.
const knuthFactor = uint64(1111111111111111111)

func (id ID128) Sampled(rate float64) bool {
	if rate < 1 {
		return (id.Low+id.High)*knuthFactor < uint64(math.MaxUint64*rate)
	}

	return true
}

func ptSetMeta(pt *point.Point, sp *EBPFSpan) {
	if pt == nil || sp == nil {
		return
	}

	// span id is always hex
	pt.MustAdd(SpanID, sp.spanID.StringHex())

	// ebpf: traceid, parenid
	pt.MustAdd(EBPFTraceID, sp.eTraceID.StringHex())
	pt.MustAdd(EBPFParentID, sp.eParentID.StringHex())

	// set spanid, traceid and parentid
	if sp.idEncodeDec { // only for datadog trace id
		// traceid, parenid
		pt.MustAdd(TraceID, sp.traceID.StringDec())

		if sp.parentID != sp.eParentID { // parent not ebpf span
			pt.MustAdd(ParentID, sp.parentID.StringDec())
		} else {
			pt.MustAdd(ParentID, sp.parentID.StringHex())
		}
	} else {
		pt.MustAdd(TraceID, sp.traceID.StringHex())
		pt.MustAdd(ParentID, sp.parentID.StringHex())
	}

	// always hex

	pt.Del(AppTraceIDL)
	pt.Del(AppTraceIDH)
	pt.Del(AppParentIDL)
}

func spanMetaData(pt *point.Point) (*EBPFSpan, bool) {
	if pt == nil {
		return nil, false
	}

	spanid, ok := getID64(pt, SpanID)
	if !ok {
		return nil, false
	}

	var spanEntry bool
	espanType, ok := getPtValStr(pt, EBPFSpanType)
	if !ok {
		return nil, false
	}
	if espanType == SpanTypEntry {
		spanEntry = true
	}

	netraceid, ok := getID128(pt, ReqSeq, RespSeq)
	if !ok {
		return nil, false
	}

	thrTraceid, ok := getID64(pt, ThrTraceID)
	if !ok {
		return nil, false
	}

	directionStr, ok := getPtValStr(pt, Direction)
	if !ok {
		return nil, false
	}

	var direction bool
	if directionStr == INCO {
		direction = true
	}

	espan := &EBPFSpan{
		directionIn: direction,
		spanID:      spanid,
		netTraceID:  netraceid,
		thrTraceID:  thrTraceid,
		espanEntry:  spanEntry,
	}

	if appTraceid, ok := getID128(pt, AppTraceIDL, AppTraceIDH); ok {
		espan.aTraceID = appTraceid
		espan.aParentID, _ = getID64(pt, AppParentIDL)
		if aSpanSampled, ok := getPtValInt(pt, AppSpanSampled); ok {
			switch {
			case aSpanSampled > 0:
				espan.aSampled = KeepTrace
			case aSpanSampled < 0:
				espan.aSampled = RejectTrace
			}
		}
		if aEncode, ok := getPtValStr(pt, AppTraceEncode); ok {
			switch aEncode {
			case ENCHEX:
				espan.idEncodeDec = false
			case ENCDEC:
				espan.idEncodeDec = true
			}
		}
	}

	return espan, true
}

func getPtValStr(pt *point.Point, name string) (string, bool) {
	if pt == nil {
		return "", false
	}

	if val := pt.Get(name); val != nil {
		if val, ok := val.(string); ok {
			return val, ok
		}
	}

	return "", false
}

func getID128(pt *point.Point, low, high string) (ID128, bool) {
	if pt == nil {
		return ID128{}, false
	}

	id := ID128{}
	if val := pt.Get(low); val != nil {
		if val, ok := val.(int64); ok {
			id.Low = uint64(val)
		} else {
			return ID128{}, false
		}
	} else {
		return ID128{}, false
	}

	if val := pt.Get(high); val != nil {
		if val, ok := val.(int64); ok {
			id.High = uint64(val)
		} else {
			return ID128{}, false
		}
	} else {
		return ID128{}, false
	}

	return id, true
}

func getID64(pt *point.Point, name string) (ID64, bool) {
	if pt == nil {
		return 0, false
	}

	if val := pt.Get(name); val != nil {
		if val, ok := val.(int64); ok {
			return ID64(val), true
		}
	}
	return 0, false
}

func getPtValInt(pt *point.Point, name string) (int64, bool) {
	if pt == nil {
		return 0, false
	}

	if val := pt.Get(name); val != nil {
		if val, ok := val.(int64); ok {
			return val, ok
		}
	}

	return 0, false
}

func Init() {
	l = logger.SLogger("ebpftrace-span")
}

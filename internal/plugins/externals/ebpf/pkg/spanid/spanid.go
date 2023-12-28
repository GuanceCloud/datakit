// Package spanid define spanid
package spanid

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"io"
	"strconv"
	"sync"
	"time"

	expRand "golang.org/x/exp/rand"

	"github.com/oklog/ulid"
)

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
)

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

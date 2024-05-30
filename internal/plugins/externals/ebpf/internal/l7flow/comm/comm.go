//go:build linux
// +build linux

// Package comm stores connection information
package comm

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	expRand "golang.org/x/exp/rand"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
)

type FnID int

const (
	FnUnknown FnID = iota
	FnSysWrite
	FnSysRead
	FnSysSendto
	FnSysRecvfrom
	FnSysWritev
	FnSysReadv
	FnSysSendfile
	FnSysClose
	FnSSLWrite
	FnSSLRead
)

type Direcion uint8

const (
	DUnknown Direcion = iota
	DIn
	DOut
)

const (
	DirectionOutgoing = "outgoing"
	DirectionIncoming = "incoming"
	DirectionUnknown  = "unknown"

	FieldUserThread   = "tid_usr"
	FieldKernelThread = "tid"
	FieldKernelTime   = "ktime"

	FieldBytesWritten = "bytes_written"
	FieldBytesRead    = "bytes_read"

	FieldHTTPRoute      = "http_route"
	FieldHTTPMethod     = "http_method"
	FieldHTTPStatusCode = "http_status_code"
	FieldHTTPVersion    = "http_version"

	FieldGRPCStatusCode = "grpc_status_code"
	FieldGRPCMessage    = "grpc_message"

	FieldMysqlStatusCode = "mysql_status_code"
	FieldMysqlErrMsg     = "mysql_err_msg"

	FieldStatus    = "status"
	FieldOperation = "operation"
	FieldResource  = "resource"

	NoValue = "N/A"
)

func (d Direcion) String() string {
	switch d { //nolint:exhaustive
	case DIn:
		return DirectionIncoming
	case DOut:
		return DirectionOutgoing
	default:
		return DirectionUnknown
	}
}

func (fn FnID) String() string {
	switch fn { //nolint:exhaustive
	case FnSysWrite:
		return "write"
	case FnSysRead:
		return "read"
	case FnSysSendto:
		return "sendto"
	case FnSysRecvfrom:
		return "recvfrom"
	case FnSysWritev:
		return "writev"
	case FnSysReadv:
		return "readv"
	case FnSysSendfile:
		return "sendfile"
	case FnSSLWrite:
		return "ssl_write"
	case FnSSLRead:
		return "ssl_read"
	case FnSysClose:
		return "close"
	default:
		return "unknown"
	}
}

type ConnectionInfo struct {
	Saddr [4]uint32
	Daddr [4]uint32
	Sport uint32
	Dport uint32
	Pid   uint32
	Netns uint32
	Meta  uint32

	NATDaddr [4]uint32
	NATDport uint32

	ProcessName string
	TaskName    string
	ServiceName string
}

func (conn ConnectionInfo) String() string {
	return fmt.Sprintf("%s:%d -> %s:%d, pid:%d, tcp:%t", netflow.U32BEToIP(conn.Saddr,
		!netflow.ConnAddrIsIPv4(conn.Meta)), conn.Sport,
		netflow.U32BEToIP(conn.Daddr, !netflow.ConnAddrIsIPv4(conn.Meta)),
		conn.Dport, conn.Pid, netflow.ConnProtocolIsTCP(conn.Meta))
}

type NICDirection int

const (
	NICDUnknown NICDirection = iota
	NICDIngress
	NICDEgress
)

func FnInOut(fn FnID) NICDirection {
	switch fn { //nolint:exhaustive
	case FnSysWrite:
		return NICDEgress
	case FnSysRead:
		return NICDIngress
	case FnSysSendto:
		return NICDEgress
	case FnSysRecvfrom:
		return NICDIngress
	case FnSysWritev:
		return NICDEgress
	case FnSysReadv:
		return NICDIngress
	case FnSysSendfile:
		return NICDEgress
	case FnSSLWrite:
		return NICDEgress
	case FnSSLRead:
		return NICDIngress
	default:
		return NICDUnknown
	}
}

type ConnUniID struct {
	SkPtr uint64
	Ktime uint32
	Rand  uint32
}
type NetwrkData struct {
	Conn      ConnectionInfo
	ConnUniID ConnUniID
	ActSize   int
	TCPSeq    uint32
	Thread    [2]int32
	TS        uint64
	TSTail    uint64
	Index     uint32
	Fn        FnID
	Payload   []byte
}

func (d NetwrkData) String() string {
	str := fmt.Sprintf("\tconn %s\n", d.Conn.String())
	str += fmt.Sprintf("\nptr: %x, thread %d, user thread %d, idx %d\n",
		d.ConnUniID, d.Thread[0], d.Thread[1], d.Index)
	str += fmt.Sprintf("\tfn %s, size %d, tcp seq: %d\n", d.Fn.String(),
		d.ActSize, d.TCPSeq)

	ts := d.TS
	tsNano := ts % uint64(time.Second)
	ts -= tsNano

	str += fmt.Sprintf("\tts %s %s %s\n", time.Duration(ts).String(),
		time.Duration(tsNano).String(),
		time.Duration(tsNano+d.TSTail).String())
	if len(d.Payload) > 10 {
		str += fmt.Sprintf("\t%s\n", string(d.Payload[:16]))
	} else {
		str += fmt.Sprintf("\t%s\n", string(d.Payload))
	}
	return str
}

var randInnerID func() int64

func newRandFunc() func() int64 {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err == nil {
		v := binary.LittleEndian.Uint64(b)
		r := expRand.New(expRand.NewSource(v))
		r.Seed(v)
		return func() int64 {
			return r.Int63()
		}
	}
	return func() int64 {
		return -1
	}
}

type ThrEntry struct {
	ts      uint64
	innerID int64

	prv *ThrEntry
}

type ThreadTrace struct {
	sync.RWMutex

	// only for incoming requests
	Threads map[int32]*ThrEntry

	lastTS uint64

	delCount int
}

func (thrTr *ThreadTrace) Insert(d Direcion, thrID2 [2]int32, ts0_1 uint64) (id int64) {
	switch d { //nolint:exhaustive
	case DIn:
	default:
		return -1
	}

	var thrID int32
	if thrID2[1] != 0 {
		thrID = thrID2[1]
	} else {
		thrID = thrID2[0]
	}

	thrTr.Lock()
	defer thrTr.Unlock()

	if thrTr.lastTS < ts0_1 {
		thrTr.lastTS = ts0_1
	}

	if thrTr.Threads == nil {
		thrTr.Threads = make(map[int32]*ThrEntry)
	}

	id = randInnerID()
	insertEntry := &ThrEntry{
		ts:      ts0_1,
		innerID: id,
	}

	if tailTr, ok := thrTr.Threads[thrID]; ok {
		if ts0_1 >= tailTr.ts {
			insertEntry.prv = tailTr
			thrTr.Threads[thrID] = insertEntry
			return
		}

		for cur := tailTr; cur != nil; cur = cur.prv {
			prv := cur.prv
			if prv != nil && ts0_1 < prv.ts {
				continue
			}
			insertEntry.prv = prv
			cur.prv = insertEntry
			return
		}
	} else {
		thrTr.Threads[thrID] = insertEntry
	}
	return id
}

func (thrTr *ThreadTrace) GetInnerID(thrID2 [2]int32, ts uint64) int64 {
	thrTr.RLock()
	defer thrTr.RUnlock()

	var thrID int32
	if thrID2[1] != 0 {
		thrID = thrID2[1]
	} else {
		thrID = thrID2[0]
	}

	if tailTr, ok := thrTr.Threads[thrID]; ok {
		if ts >= tailTr.ts {
			return tailTr.innerID
		}

		for cur := tailTr; cur != nil; cur = cur.prv {
			prv := cur.prv
			if prv != nil && ts >= prv.ts {
				return prv.innerID
			}
		}
	}
	return randInnerID()
}

func (thrTr *ThreadTrace) Cleanup() {
	thrTr.Lock()
	defer thrTr.Unlock()

	lastTS := thrTr.lastTS
	var del []int32
	for k, v := range thrTr.Threads {
		if v == nil {
			continue
		} else if lastTS-v.ts > uint64(time.Minute*10) {
			del = append(del, k)
			continue
		}

		for ptr := v; ptr != nil && ptr.prv != nil; ptr = ptr.prv {
			ts := ptr.prv.ts
			if (lastTS > ts) && (lastTS-ts > uint64(time.Minute*10)) {
				ptr.prv = nil
				break
			}
		}
	}

	for _, k := range del {
		delete(thrTr.Threads, k)
	}

	thrTr.delCount += len(del)
	if thrTr.delCount > 1e3 && thrTr.delCount >= len(thrTr.Threads) {
		mp := make(map[int32]*ThrEntry)
		for k, v := range thrTr.Threads {
			mp[k] = v
		}
		thrTr.Threads = mp
		thrTr.delCount = 0
	}
}

var log = logger.DefaultSLogger("tracer-comm")

func Init(nl *logger.Logger) {
	randInnerID = newRandFunc()
	log = nl
	_ = log
}

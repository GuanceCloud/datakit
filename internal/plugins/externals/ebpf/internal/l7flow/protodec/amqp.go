//go:build linux
// +build linux

package protodec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

func AMQPProtoDetect(data []byte, actSize int) (L7Protocol, ProtoDecPipe, bool) {
	if checkAMQP(data) {
		return ProtoAMQP, newAMQPDecPipe(ProtoAMQP), true
	}
	return ProtoUnknown, nil, false
}

func checkAMQP(payload []byte) bool {
	if len(payload) < AMQPMinHeaderLen {
		return false
	}
	// New Connection should start with protocol header of AMQP.
	header := payload[:AMQPMinHeaderLen]
	if bytes.Equal(header, []byte(AMQPHeader)) {
		return true
	}

	if len(payload) < AMQPMinLen {
		return false
	}

	if payload[0] != AMQPFrameMethodType {
		return false
	}

	// big endian read
	header = payload[AMQPClassOffset:]
	classID := binary.BigEndian.Uint16(header[:2])
	methodID := binary.BigEndian.Uint16(header[2:4])

	isAMQP, _ := checkAMQPClassMethod(classID, methodID)
	return isAMQP
}

type amqpInfo struct {
	meta ProtoMeta

	frameType byte
	channelID uint16
	size      uint32
	classID   uint16
	methodID  uint16
	bodySize  uint64

	reqMethod  string
	respMethod string

	vhost      string
	queue      string
	exchange   string
	routingKey string
	isHeader   bool

	readyToExport bool

	reqBytes  int
	respBytes int

	reqResp int
	ktime   [4]uint64
	ts      int64
	dur     [2]uint64
}

func (dec *amqpDecPipe) decode(txRx comm.NICDirection, data *comm.NetwrkData,
	ts int64, thrTr threadTrace,
) error {
	payload := data.Payload
	if len(payload) < AMQPMinHeaderLen {
		return errors.New("payload len must gt AMQPMinHeaderLen")
	}

	header := payload[:AMQPMinHeaderLen]
	if bytes.Equal(header, []byte(AMQPHeader)) {
		inf := &amqpInfo{}
		inf.reqResp = AMQPPacketSession
		inf.meta.ReqTCPSeq = data.TCPSeq
		inf.reqMethod = AMQP091
		inf.classID = AMQPVersion
		inf.isHeader = true
		inf.readyToExport = true
		inf.ts = ts

		inf.reqBytes += data.FnCallSize

		if dec.direction == comm.DUnknown {
			switch txRx { //nolint:exhaustive
			case comm.NICDIngress:
				dec.direction = comm.DIn
			case comm.NICDEgress:
				dec.direction = comm.DOut
			}
		}

		inf.ktime[3], inf.ktime[2] = data.TSTail, data.TSTail
		inf.ktime[0], inf.ktime[1] = data.TS, data.TS

		dec.inf = inf

		return nil
	}

	infArr := []*amqpInfo{}
	for {
		if len(payload) < AMQPMinLen {
			break
		}
		inf := &amqpInfo{}
		inf.frameType = payload[0]
		if !checkFrameType(inf.frameType) {
			return errors.New("amqp frameType unknown")
		}

		inf.channelID = binary.BigEndian.Uint16(payload[1:3])
		inf.size = binary.BigEndian.Uint32(payload[3:7])

		payload = payload[AMQPClassOffset:]
		if payload[inf.size] != AMQPEnd {
			return errors.New("payload should end with 'ce'")
		}

		switch inf.frameType {
		case AMQPFrameMethodType:
			if inf.size < 4 {
				return errors.New("amqp method size should ge 4")
			}
			inf.classID = binary.BigEndian.Uint16(payload[:2])
			inf.methodID = binary.BigEndian.Uint16(payload[2:4])
			if ok, _ := checkAMQPClassMethod(inf.classID, inf.methodID); !ok {
				return errors.New("classID methodID should match")
			}
			inf.queue = inf.parseQueue(payload[AMQPArgumentOffset:])
			inf.exchange = inf.parseExchange(payload[AMQPArgumentOffset:])
			inf.routingKey = inf.parseRoutingKey(payload[AMQPArgumentOffset:])
		case AMQPFrameHeaderType:
			if inf.size < AMQPMinFrameHeaderLen {
				return errors.New("amqp header size should ge 14")
			}
			inf.classID = binary.BigEndian.Uint16(payload[:2])
			// usually weight equals 0
			weight := binary.BigEndian.Uint16(payload[2:4])
			if weight != 0 {
				return errors.New("weight should eq 0")
			}
			inf.bodySize = binary.BigEndian.Uint64(payload[4:12])
		case AMQPFrameBodyType:
		case AMQPFrameHeartBeatType:
		default:
		}

		// Connection Open for vhost
		if inf.classID == AMQPConnectionClass && inf.methodID == 40 {
			if str, _, err := readStr(payload, AMQPArgumentOffset); err == nil {
				dec.vhost = str
			}
		}
		inf.vhost = dec.vhost
		if inf.classID == AMQPConnectionClass && inf.methodID == 51 {
			dec.vhost = ""
		}

		payload = payload[inf.size:]
		if payload[0] != AMQPEnd {
			break
		}
		payload = payload[1:]

		infArr = append(infArr, inf)
	}
	if len(infArr) == 0 {
		return nil
	}
	for _, inf := range infArr[:1] {
		reqResp := inf.checkReqResp()
		_, name := checkAMQPClassMethod(inf.classID, inf.methodID)
		switch reqResp {
		case AMQPPacketRequest:
			inf.reqMethod = name

			inf.meta.Threads[0] = data.Thread
			inf.ts = ts

			if dec.direction == comm.DIn {
				inf.meta.InnerID = thrTr.Insert(dec.direction, int32(data.Conn.Pid),
					data.Thread, data.TSTail)
			}

			inf.meta.ReqTCPSeq = data.TCPSeq
			inf.dur[0] = data.TS
			inf.ktime[0] = data.TSTail

			dec.inf = inf
		case AMQPPacketResponse:
			dec.inf.respMethod = name
			dec.inf.meta.RespTCPSeq = data.TCPSeq
			dec.inf.meta.Threads[1] = data.Thread

			dec.inf.readyToExport = true

			dec.inf.dur[1] = data.TS
			dec.inf.ktime[2] = data.TSTail

		case AMQPPacketSession:
			inf.meta.ReqTCPSeq = data.TCPSeq
			inf.reqMethod = name
			inf.ts = ts

			inf.dur[0] = data.TS
			inf.dur[1] = data.TSTail

			inf.ktime[3], inf.ktime[2] = data.TSTail, data.TSTail
			inf.ktime[0], inf.ktime[1] = data.TS, data.TS
			inf.readyToExport = true
			dec.inf = inf
		}

		switch dec.direction { //nolint:exhaustive
		case comm.DIn:
			switch txRx { //nolint:exhaustive
			case comm.NICDIngress:
				dec.inf.reqBytes += data.FnCallSize
			case comm.NICDEgress:
				dec.inf.respBytes += data.FnCallSize
			}
		case comm.DOut:
			switch txRx { //nolint:exhaustive
			case comm.NICDIngress:
				dec.inf.respBytes += data.FnCallSize
			case comm.NICDEgress:
				dec.inf.reqBytes += data.FnCallSize
			}
		}

		switch reqResp { //nolint:exhaustive
		case AMQPPacketRequest:
			dec.inf.ktime[1] = data.TSTail
		case AMQPPacketResponse:
			dec.inf.ktime[3] = data.TSTail
		}
	}
	return nil
}

func (a *amqpInfo) parseQueue(payload []byte) string {
	switch a.classID {
	case AMQPQueueClass, AMQPBasicClass:
		_, name := checkAMQPClassMethod(a.classID, a.methodID)
		switch name {
		case Declare, Bind, Unbind, Purge, Delete, Consume, Get:
			// reserved-1 queue
			if str, _, err := readStr(payload, 2); err == nil {
				return str
			}
		case DeclareOk:
			if str, _, err := readShortStr(payload); err == nil {
				return str
			}
		}
	default:
		return ""
	}
	return ""
}

func (a *amqpInfo) parseExchange(payload []byte) string {
	_, name := checkAMQPClassMethod(a.classID, a.methodID)
	var (
		str string
		err error
	)

	switch a.classID {
	case AMQPExchangeClass:
		if name == Declare || name == Delete {
			if str, _, err = readStr(payload, 2); err == nil {
				return str
			}
		}
	case AMQPQueueClass:
		if name == Bind {
			if _, payload, err = readStr(payload, 2); err == nil {
				if str, _, err = readShortStr(payload); err == nil {
					return str
				}
			}
		}
	case AMQPBasicClass:
		switch name {
		case Publish:
			if str, _, err = readStr(payload, 2); err == nil {
				return str
			}
		case Return:
			if _, payload, err = readStr(payload, 2); err == nil {
				if str, _, err = readShortStr(payload); err == nil {
					return str
				}
			}
		case Deliver:
			if _, payload, err = readShortStr(payload); err == nil {
				if str, _, err = readStr(payload, 9); err == nil {
					return str
				}
			}
		case GetOk:
			if str, _, err = readStr(payload, 9); err == nil {
				return str
			}
		}
	}
	return ""
}

func (a *amqpInfo) parseRoutingKey(payload []byte) string {
	_, name := checkAMQPClassMethod(a.classID, a.methodID)
	var (
		str string
		err error
	)
	switch a.classID {
	case AMQPExchangeClass:
		if name == Bind || name == Unbind {
			if _, payload, err = readStr(payload, 2); err == nil {
				if _, payload, err = readShortStr(payload); err == nil {
					if str, _, err = readShortStr(payload); err == nil {
						return str
					}
				}
			}
		}
	case AMQPQueueClass:
		if name == Bind {
			if _, payload, err = readStr(payload, 2); err == nil {
				if _, payload, err = readShortStr(payload); err == nil {
					if str, _, err = readShortStr(payload); err == nil {
						return str
					}
				}
			}
		}
	case AMQPBasicClass:
		switch name {
		case Publish:
			if _, payload, err = readStr(payload, 2); err == nil {
				if str, _, err = readShortStr(payload); err == nil {
					return str
				}
			}
		case Return:
			if _, payload, err = readStr(payload, 2); err == nil {
				if _, payload, err = readShortStr(payload); err == nil {
					if str, _, err = readShortStr(payload); err == nil {
						return str
					}
				}
			}
		case Deliver:
			if _, payload, err = readShortStr(payload); err == nil {
				if _, payload, err = readStr(payload, 9); err == nil {
					if str, _, err = readShortStr(payload); err == nil {
						return str
					}
				}
			}
		case GetOk:
			if _, payload, err = readStr(payload, 9); err == nil {
				if str, _, err = readShortStr(payload); err == nil {
					return str
				}
			}
		}
	}
	return ""
}

type amqpDecPipe struct {
	direction  comm.Direcion
	connClosed bool
	vhost      string
	inf        *amqpInfo
	infCache   []*amqpInfo
}

func (dec *amqpDecPipe) Decode(txRx comm.NICDirection, data *comm.NetwrkData,
	ts int64, thrTr threadTrace,
) {
	if len(data.Payload) == 0 {
		return
	}

	if dec.inf == nil {
		dec.inf = &amqpInfo{}
	}

	inf := dec.inf
	if inf.readyToExport {
		dec.infCache = append(dec.infCache, inf)
		inf = &amqpInfo{}
		dec.inf = inf
	}

	if err := dec.decode(txRx, data, ts, thrTr); err != nil {
		return
	}
}

func (dec *amqpDecPipe) Proto() L7Protocol {
	return ProtoAMQP
}

func (dec *amqpDecPipe) Export(force bool) []*ProtoData {
	if force {
		dec.infCache = append(dec.infCache, dec.inf)
		dec.inf = nil
	}

	var result []*ProtoData

	for _, inf := range dec.infCache {
		if inf == nil {
			continue
		}

		kvs := make(point.KVs, 0, 20)

		switch dec.direction { //nolint:exhaustive
		case comm.DIn:
			kvs = kvs.Add(comm.FieldBytesRead, int64(inf.reqBytes), false, true)
			kvs = kvs.Add(comm.FieldBytesWritten, int64(inf.respBytes), false, true)
		default:
			kvs = kvs.Add(comm.FieldBytesRead, int64(inf.respBytes), false, true)
			kvs = kvs.Add(comm.FieldBytesWritten, int64(inf.reqBytes), false, true)
		}
		kvs = kvs.Add(comm.FieldResource, inf.reqMethod, false, true)
		kvs = kvs.Add(comm.FieldAMQPRespMethod, inf.respMethod, false, true)
		kvs = kvs.Add(comm.FieldAMQPClass, toClassName(inf.classID), false, true)
		kvs = kvs.Add(comm.FieldAMQPQueue, inf.queue, false, true)
		kvs = kvs.Add(comm.FieldAMQPExchange, inf.exchange, false, true)
		kvs = kvs.Add(comm.FieldAMQPRoutingKey, inf.routingKey, false, true)
		kvs = kvs.Add(comm.FieldAMQPVhost, inf.vhost, false, true)

		dur := int64(inf.ktime[3] - inf.ktime[0])
		cost := int64(inf.ktime[2] - inf.ktime[1])
		result = append(result, &ProtoData{
			Meta:      inf.meta,
			Time:      inf.ts,
			KVs:       kvs,
			Cost:      cost,
			Duration:  dur,
			Direction: dec.direction,
			L7Proto:   ProtoAMQP,
			KTime:     inf.ktime[0],
		})
	}

	dec.infCache = dec.infCache[:0]

	return result
}

func (dec *amqpDecPipe) ConnClose() {
	dec.connClosed = true
}

func newAMQPDecPipe(L7Protocol) ProtoDecPipe {
	return &amqpDecPipe{}
}

// checkReqResp 0: unknown 1: req 2: resp 3: pending.
func (a *amqpInfo) checkReqResp() int {
	if a.isHeader {
		return AMQPPacketSession
	}

	switch a.frameType {
	case AMQPFrameMethodType:
	case AMQPFrameHeaderType, AMQPFrameBodyType, AMQPFrameHeartBeatType:
		return AMQPPacketSession
	case AMQPFrameUnknown:
		return AMQPPacketUnknown
	}

	method := amqpClassMethods[a.classID][a.methodID]

	switch method {
	case Start, Secure, Tune, Open, Close, Flow, Declare, Delete, Bind, Unbind, QOS, Consume, Cancel, Get, Recover, Select, Commit, Rollback:
		return AMQPPacketRequest
	case StartOk, SecureOk, TuneOk, OpenOk, CloseOk, FlowOk, DeclareOk, DeleteOk, BindOk, UnbindOk, QOSOk, ConsumeOk, CancelOk, GetOk,
		GetEmpty, RecoverOk, SelectOk, CommitOk, RollbackOk:
		return AMQPPacketResponse
	case Publish, Return, Deliver, ACK, Reject, RecoverAsync:
		return AMQPPacketSession
	default:
		return AMQPPacketUnknown
	}
}

func checkAMQPClassMethod(classID, methodID uint16) (bool, string) {
	if methods, ok := amqpClassMethods[classID]; ok {
		if name, exist := methods[methodID]; exist {
			return true, name
		}
	}
	return false, ""
}

func checkFrameType(b byte) bool {
	switch b {
	case AMQPFrameMethodType, AMQPFrameHeaderType, AMQPFrameBodyType, AMQPFrameHeartBeatType:
		return true
	default:
		return false
	}
}

func toClassName(classID uint16) string {
	switch classID {
	case AMQPConnectionClass:
		return "Connection"
	case AMQPChannelClass:
		return "Channel"
	case AMQPExchangeClass:
		return "Exchange"
	case AMQPQueueClass:
		return "Queue"
	case AMQPBasicClass:
		return "Basic"
	case AMQPTXClass:
		return "Tx"
	case AMQPVersion:
		return "AMQP 0-9-1"
	default:
		return "Unknown"
	}
}

func readStr(payload []byte, offset int) (string, []byte, error) {
	if len(payload) <= offset {
		return "", nil, errors.New("payload too short")
	}
	return readShortStr(payload[offset:])
}

func readShortStr(payload []byte) (string, []byte, error) {
	if len(payload) < 1 {
		return "", nil, fmt.Errorf("payload too short to contain string size")
	}
	size := int(payload[0])
	if len(payload) < size+1 {
		return "", nil, fmt.Errorf("payload too short for the specified string size")
	}
	return string(payload[1 : 1+size]), payload[1+size:], nil
}

const (
	AMQPMinHeaderLen      = 8
	AMQPMinLen            = 11
	AMQPMinFrameMethodLen = 4
	AMQPMinFrameHeaderLen = 14
	AMQPClassOffset       = 7
	AMQPArgumentOffset    = 4
	AMQPHeader            = "\x41\x4d\x51\x50\x00\x00\x09\x01" // AMQP0-9-1
	AMQP091               = "AMQP 0-9-1"
	AMQPEnd               = '\xce'
)

const (
	AMQPFrameUnknown       = 0
	AMQPFrameMethodType    = 1
	AMQPFrameHeaderType    = 2
	AMQPFrameBodyType      = 3
	AMQPFrameHeartBeatType = 4
)

const (
	AMQPConnectionClass = 10
	AMQPChannelClass    = 20
	AMQPExchangeClass   = 40
	AMQPQueueClass      = 50
	AMQPBasicClass      = 60
	AMQPTXClass         = 90
	AMQPVersion         = 100
)

const (
	AMQPPacketUnknown = iota
	AMQPPacketRequest
	AMQPPacketResponse
	AMQPPacketSession
)

const (
	Start        = "Start"
	StartOk      = "Start-Ok"
	Secure       = "Secure"
	SecureOk     = "Secure-Ok"
	Tune         = "Tune"
	TuneOk       = "Tune-Ok"
	Open         = "Open"
	OpenOk       = "Open-Ok"
	Close        = "Close"
	CloseOk      = "Close-Ok"
	Flow         = "Flow"
	FlowOk       = "Flow-Ok"
	Declare      = "Declare"
	DeclareOk    = "Declare-Ok"
	Delete       = "Delete"
	DeleteOk     = "Delete-Ok"
	Bind         = "Bind"
	BindOk       = "Bind-Ok"
	Purge        = "Purge"
	PurgeOk      = "Purge-Ok"
	Unbind       = "Unbind"
	UnbindOk     = "Unbind-Ok"
	QOS          = "QOS"
	QOSOk        = "QOS-Ok"
	Consume      = "Consume"
	ConsumeOk    = "Consume-Ok"
	Cancel       = "Cancel"
	CancelOk     = "Cancel-Ok"
	Publish      = "Publish"
	Return       = "Return"
	Deliver      = "Deliver"
	Get          = "Get"
	GetOk        = "Get-Ok"
	GetEmpty     = "Get-Empty"
	ACK          = "ACK"
	Reject       = "Reject"
	RecoverAsync = "Recover-Async"
	Recover      = "Recover"
	RecoverOk    = "Recover-Ok"
	Select       = "Select"
	SelectOk     = "Select-Ok"
	Commit       = "Commit"
	CommitOk     = "Commit-Ok"
	Rollback     = "Rollback"
	RollbackOk   = "Rollback-Ok"
)

// TODO: 优化这个.
var amqpClassMethods = map[uint16]map[uint16]string{
	AMQPConnectionClass: {
		10: Start, 11: StartOk, 20: Secure, 21: SecureOk,
		30: Tune, 31: TuneOk, 40: Open, 41: OpenOk,
		50: Close, 51: CloseOk,
	},
	AMQPChannelClass: {
		10: Open, 11: OpenOk, 20: Flow, 21: FlowOk,
		40: Close, 41: CloseOk,
	},
	AMQPExchangeClass: {
		10: Declare, 11: DeclareOk, 20: Delete, 21: DeleteOk,
	},
	AMQPQueueClass: {
		10: Declare, 11: DeclareOk, 20: Bind, 21: BindOk,
		30: Purge, 31: PurgeOk, 40: Delete, 41: DeleteOk,
		50: Unbind, 51: UnbindOk,
	},
	AMQPBasicClass: {
		10: QOS, 11: QOSOk, 20: Consume, 21: ConsumeOk,
		30: Cancel, 31: CancelOk, 40: Publish, 50: Return,
		60: Deliver, 70: Get, 71: GetOk, 72: GetEmpty,
		80: ACK, 90: Reject, 100: RecoverAsync, 110: Recover, 111: RecoverOk,
	},
	AMQPTXClass: {
		10: Select, 11: SelectOk, 20: Commit, 21: CommitOk,
		30: Rollback, 31: RollbackOk,
	},
}

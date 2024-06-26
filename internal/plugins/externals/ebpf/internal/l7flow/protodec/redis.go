//go:build linux
// +build linux

package protodec

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

// RedisProtoDetect 只检测请求的RESP的体，以 '*' 开头，之后再去进一步解析关联请求和响应.
func RedisProtoDetect(data []byte, actSize int) (L7Protocol, ProtoDecPipe, bool) {
	r := &redisInfo{}
	if err := r.parseRequest(data); err != nil {
		return ProtoUnknown, nil, false
	}
	return ProtoRedis, newRedisDecPipe(ProtoRedis), true
}

type redisInfo struct {
	meta ProtoMeta

	resource     string // set key value
	resourceType string // set
	statusMsg    string
	errMsg       string
	statusCode   int

	payload []byte
	cmd     string
	length  int

	reqBytes  int
	respBytes int

	ktime [4]uint64
	ts    int64
	dur   [2]uint64
}

type redisDecPipe struct {
	direction    comm.Direcion
	fisrtDstPort uint32
	connClosed   bool
	reqResp      int // 0, 1, 2
	infCache     []*redisInfo
	inf          *redisInfo
}

func (dec *redisDecPipe) Decode(txRx comm.NICDirection, data *comm.NetwrkData,
	ts int64, thrTr threadTrace,
) {
	if len(data.Payload) == 0 {
		return
	}

	if dec.fisrtDstPort == 0 {
		switch txRx { //nolint:exhaustive
		case comm.NICDIngress:
			dec.fisrtDstPort = data.Conn.Sport
		case comm.NICDEgress:
			dec.fisrtDstPort = data.Conn.Dport
		}
	}
	if dec.inf == nil {
		dec.inf = &redisInfo{}
		dec.reqResp = 0
	}

	direction := directionUnknown
	switch txRx { //nolint:exhaustive
	case comm.NICDIngress:
		direction = dec.determineDirection(data.Conn.Sport, dec.inf)
	case comm.NICDEgress:
		direction = dec.determineDirection(data.Conn.Dport, dec.inf)
	}

	inf := dec.inf
	err := inf.parse(data.Payload, direction)
	if err != nil {
		log.Debugf("error occur when decode redis: %v", err)
		return
	}
	switch direction { //nolint:exhaustive
	case clientToServer:
		dec.reqResp = 1
		resource := inf.stringify()
		inf.resource = string(resource)
		inf.resourceType = inf.cmd

		inf.meta.Threads[0] = data.Thread
		inf.ts = ts

		if dec.direction == comm.DUnknown {
			switch txRx { //nolint:exhaustive
			case comm.NICDIngress:
				dec.direction = comm.DIn
			case comm.NICDEgress:
				dec.direction = comm.DOut
			}
		}

		if dec.direction == comm.DIn {
			inf.meta.InnerID = thrTr.Insert(dec.direction, data.Thread, data.TSTail)
		}

		inf.meta.ReqTCPSeq = data.TCPSeq
		inf.dur[0] = data.TS
		inf.ktime[0] = data.TSTail

	case serverToClient:
		dec.reqResp = 2

		inf.statusCode = statusOK
		inf.meta.RespTCPSeq = data.TCPSeq
		inf.meta.Threads[1] = data.Thread

		if len(inf.payload) != 0 {
			switch inf.payload[0] {
			case '+':
				inf.statusMsg = string(inf.payload)
			case '-', '|':
				inf.errMsg = string(inf.payload)
				inf.statusCode = statusError
			default:
			}
		}

		inf.dur[1] = data.TS
		inf.ktime[2] = data.TSTail
	}

	switch dec.direction { //nolint:exhaustive
	case comm.DIn:
		switch txRx { //nolint:exhaustive
		case comm.NICDIngress:
			inf.reqBytes += data.ActSize
		case comm.NICDEgress:
			inf.respBytes += data.ActSize
		}
	case comm.DOut:
		switch txRx { //nolint:exhaustive
		case comm.NICDIngress:
			inf.respBytes += data.ActSize
		case comm.NICDEgress:
			inf.reqBytes += data.ActSize
		}
	}

	switch dec.reqResp {
	case 1:
		inf.ktime[1] = data.TSTail
	case 2:
		inf.ktime[3] = data.TSTail
	}
}

func (dec *redisDecPipe) Proto() L7Protocol {
	return ProtoRedis
}

func (dec *redisDecPipe) Export(force bool) []*ProtoData {
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
		kvs = kvs.Add(comm.FieldResource, inf.resource, false, true)
		kvs = kvs.Add(comm.FieldResourceType, inf.resourceType, false, true)
		kvs = kvs.Add(comm.FieldStatusMsg, inf.statusMsg, false, true)
		kvs = kvs.Add(comm.FieldErrMsg, inf.errMsg, false, true)
		kvs = kvs.Add(comm.FieldStatusCode, inf.statusCode, false, true)
		kvs = kvs.Add(comm.FieldStatus, redisCode2Status(inf.statusCode), false, true)

		kvs = kvs.Add(comm.FieldOperation, fmt.Sprintf("Redis %s", inf.cmd), false, true)

		dur := int64(inf.ktime[3] - inf.ktime[0])
		cost := int64(inf.ktime[2] - inf.ktime[1])
		result = append(result, &ProtoData{
			Meta:      inf.meta,
			Time:      inf.ts,
			KVs:       kvs,
			Cost:      cost,
			Duration:  dur,
			Direction: dec.direction,
			L7Proto:   ProtoRedis,
			KTime:     inf.ktime[0],
		})
	}

	dec.infCache = dec.infCache[:0]
	return result
}

func (dec *redisDecPipe) ConnClose() {
	dec.connClosed = true
}

func newRedisDecPipe(L7Protocol) ProtoDecPipe {
	return &redisDecPipe{}
}

func (r *redisInfo) parse(payload []byte, direction Direction) error {
	if len(payload) == 0 {
		return errors.New("payload empty")
	}

	switch direction { //nolint:exhaustive
	case clientToServer:
		return r.parseRequest(payload)
	case serverToClient:
		return r.parseResponse(payload)
	default:
		return errors.New("unknown protocol direction")
	}
}

func (r *redisInfo) parseResponse(payload []byte) error {
	if len(payload) == 0 {
		return errors.New("parse redis response error: payload too short")
	}
	var output []byte
	if payload[0] == '+' || payload[0] == '-' || payload[0] == '!' {
		output = make([]byte, 0, len(payload))
	} else {
		output = []byte{}
	}

	_, err := r.decodeRespType(&output, payload)
	if err != nil {
		return err
	}
	r.payload = output
	return nil
}

func (r *redisInfo) parseRequest(payload []byte) error {
	if len(payload) < len(minReqPayload) || payload[0] != '*' {
		return errors.New("parse redis request error: payload too short or dismatched")
	}

	payload, length, err := readLength(payload[1:])
	if err != nil {
		return err
	}
	payloadIter, command, err := r.decodeBulkString(payload)
	if err != nil {
		return err
	}

	var cmd string
	if len(command) <= maxCommandLength && isASCII(command) {
		cmd = strings.ToUpper(string(command))
	}

	for i := 1; i < length; i++ {
		payloadIter, _, err = r.decodeBulkString(payloadIter)
		if err != nil {
			return err
		}
	}
	r.payload = payload
	r.length = length
	r.cmd = cmd
	return nil
}

func (r *redisInfo) decodeBulkString(payload []byte) ([]byte, []byte, error) {
	if len(payload) < len(minBulkString) || payload[0] != '$' {
		return nil, nil, errors.New("parse bulk string error, payload too short or dismatched")
	}

	var (
		length int
		err    error
	)
	payload, length, err = readLength(payload[1:])
	if err != nil {
		return nil, nil, err
	}
	if length < 0 {
		return payload, []byte{}, err
	}

	if length+2 > len(payload) || string(payload[length:length+2]) != "\r\n" {
		return nil, nil, errors.New("parse bulk string error")
	}

	return payload[length+2:], payload[:length], nil
}

func (r *redisInfo) decodeRespType(output *[]byte, payload []byte) ([]byte, error) {
	if len(payload) == 0 {
		return nil, errors.New("decode resp type error: payload too short")
	}

	// referred: https://redis.io/docs/latest/develop/reference/protocol-spec/#resp-protocol-description
	switch payload[0] {
	case '+':
		return assertDecErrNil(r.decodeSimpleString, output, payload)
	case '-':
		return assertDecErrNil(r.decodeSimpleError, output, payload)
	case ':':
		return assertValErrNil(r.validateInteger, payload)
	case '$':
		return assertValErrNil(r.validateBulkString, payload)
	case '*':
		return assertValErrNil(r.validateArray, payload)
	case '_':
		return assertValErrNil(r.validateNull, payload)
	case '#':
		return assertValErrNil(r.validateBoolean, payload)
	case ',':
		return assertValErrNil(r.validateDouble, payload)
	case '(':
		return assertValErrNil(r.validateBigNumber, payload)
	case '!':
		return assertDecErrNil(r.decodeBulkError, output, payload)
	case '=':
		return assertValErrNil(r.validateVerbatimString, payload)
	case '%':
		return assertValErrNil(r.validateMap, payload)
	case '~':
		return assertValErrNil(r.validateSet, payload)
	case '>':
		return assertValErrNil(r.validatePush, payload)
	default:
		return nil, errors.New("decode resp type error: unknown resp data type")
	}
}

func (r *redisInfo) decodeSimpleType(output *[]byte, payload []byte, cond func(int, byte) bool, limit int) ([]byte, error) {
	if len(payload) < 1 {
		return nil, errors.New("decode simple type error")
	}
	payload = payload[1:]
	i := 0
	end := bytes.IndexFunc(payload, func(r rune) bool {
		defer func() { i++ }()
		return !cond(i, byte(r)) || byte(r) == '\r'
	})

	if end == -1 {
		return nil, errors.New("decode simple type error")
	}

	if end+2 > len(payload) || string(payload[end:end+2]) != customReturn {
		return nil, errors.New("decode simple type error")
	}

	if output != nil {
		*output = append(*output, payload[:min(end, limit)]...)
		if end > 1+limit {
			*output = append(*output, "..."...)
		}
	}

	return payload[end+2:], nil
}

func (r *redisInfo) decodeSimpleString(output *[]byte, payload []byte) ([]byte, error) {
	if payload[0] != '+' {
		return nil, errors.New("decode simple string error: payload[0] should eq '+'")
	}
	if output != nil {
		*output = append(*output, payload[0])
	}
	return r.decodeSimpleType(output, payload, func(_ int, b byte) bool {
		return b >= 32 || b <= 126
	}, 32)
}

func (r *redisInfo) decodeSimpleError(output *[]byte, payload []byte) ([]byte, error) {
	if payload[0] != '-' {
		return nil, errors.New("decode simple error error: payload[0] should eq '-'")
	}
	if output != nil {
		*output = append(*output, payload[0])
	}
	return r.decodeSimpleType(output, payload, func(_ int, b byte) bool {
		return b >= 32 || b <= 126
	}, 256)
}

func (r *redisInfo) decodeBulkError(output *[]byte, payload []byte) ([]byte, error) {
	if payload[0] != '!' {
		return nil, errors.New("decode bulk error error: payload[0] should eq '!'")
	}
	if output != nil {
		*output = append(*output, payload[0])
	}
	return r.decodeBulkType(output, payload)
}

func (r *redisInfo) validateSimpleType(payload []byte, cond func(int, byte) bool) ([]byte, error) {
	return r.decodeSimpleType(nil, payload, cond, 0)
}

func (r *redisInfo) validateInteger(payload []byte) ([]byte, error) {
	if payload[0] != ':' {
		return nil, errors.New("decode integer error: payload[0] should eq ':'")
	}
	return r.validateSimpleType(payload, func(i int, b byte) bool {
		return b >= '0' && b <= '9' || (i == 0 && (b == '+' || b == '-'))
	})
}

func (r *redisInfo) validateBulkString(payload []byte) ([]byte, error) {
	if payload[0] != '$' {
		return nil, errors.New("validate bulk string error: payload[0] should eq '$'")
	}
	return r.validateBulkType(payload)
}

func (r *redisInfo) validateArray(payload []byte) ([]byte, error) {
	if payload[0] != '*' {
		return nil, errors.New("validate array error: payload[0] should eq '*'")
	}
	return r.validateArrayType(payload)
}

func (r *redisInfo) validateNull(payload []byte) ([]byte, error) {
	if len(payload) < 3 || payload[0] != '_' {
		return nil, errors.New("validate null error: payload[0] should eq '_' or too short")
	}

	if string(payload[1:3]) == customReturn {
		return payload[3:], nil
	} else {
		return nil, errors.New("validate null error")
	}
}

func (r *redisInfo) validateBoolean(payload []byte) ([]byte, error) {
	if len(payload) < 4 || payload[0] != '#' {
		return nil, errors.New("validate boolean error: payload[0] should eq '#' or too short")
	}
	if string(payload[1:4]) == trueReturn || string(payload[1:4]) == falseReturn {
		return payload[4:], nil
	} else {
		return nil, errors.New("validate boolean error")
	}
}

func (r *redisInfo) validateDouble(payload []byte) ([]byte, error) {
	if payload[0] != ',' {
		return nil, errors.New("validate double error: payload[0] should eq ','")
	}
	return r.validateSimpleType(payload, func(_ int, b byte) bool {
		return b >= 32 && b <= 126
	})
}

func (r *redisInfo) validateBigNumber(payload []byte) ([]byte, error) {
	if payload[0] != '(' {
		return nil, errors.New("validate big number error: payload[0] should eq '('")
	}
	return r.validateSimpleType(payload, func(i int, b byte) bool {
		return b >= '0' && b <= '9' || (i == 0 && (b == '+' || b == '-'))
	})
}

func (r *redisInfo) validateBulkType(payload []byte) ([]byte, error) {
	return r.decodeBulkType(nil, payload)
}

func (r *redisInfo) validateVerbatimString(payload []byte) ([]byte, error) {
	if payload[0] != '=' {
		return nil, errors.New("validate verbatim string error: payload[0] should eq '=")
	}
	return r.validateBulkType(payload)
}

func (r *redisInfo) validateMap(payload []byte) ([]byte, error) {
	if payload[0] != '%' {
		return nil, errors.New("validate map error: payload[0] should eq '%")
	}
	var (
		length int
		err    error
	)
	payload, length, err = readLength(payload[1:])
	if err != nil {
		return nil, err
	}
	if length < 0 {
		return payload, nil
	}

	for i := 0; i < length; i++ {
		p, err := r.decodeRespType(nil, payload)
		if err != nil {
			return nil, err
		}
		payload = p
		p, err = r.decodeRespType(nil, payload)
		if err != nil {
			return nil, err
		}
		payload = p
	}

	return payload, nil
}

func (r *redisInfo) validateSet(payload []byte) ([]byte, error) {
	if payload[0] != '~' {
		return nil, errors.New("validate set error: payload[0] should eq '~")
	}

	return r.validateArrayType(payload)
}

func (r *redisInfo) validatePush(payload []byte) ([]byte, error) {
	if payload[0] != '>' {
		return nil, errors.New("validate push error: payload[0] should eq '>")
	}

	return r.validateArrayType(payload)
}

func (r *redisInfo) validateArrayType(payload []byte) ([]byte, error) {
	var (
		length int
		err    error
	)
	payload, length, err = readLength(payload[1:])
	if err != nil {
		return nil, err
	}
	if length < 0 {
		return payload, nil
	}

	for i := 0; i < length; i++ {
		p, err := r.decodeRespType(nil, payload)
		if err != nil {
			return nil, err
		}
		payload = p
	}

	return payload, nil
}

func (r *redisInfo) decodeBulkType(output *[]byte, payload []byte) ([]byte, error) {
	var (
		length int
		err    error
	)
	payload, length, err = readLength(payload[1:])
	if err != nil {
		return nil, err
	}
	if length < 0 {
		if output != nil {
			*output = append(*output, ""...)
		}
		return payload, nil
	}

	if length+2 > len(payload) || string(payload[length:length+2]) != customReturn {
		if output != nil {
			*output = append(*output, payload[:min(length, len(payload))]...)
		}
		return nil, errors.New("decode bulk type err")
	}

	if output != nil {
		*output = append(*output, payload[:length]...)
	}
	return payload[length+2:], nil
}

func (r *redisInfo) stringify() []byte {
	output := make([]byte, 0, len(r.payload))

	if len(r.cmd) == 0 {
		newIterator(r.payload, r.length).stringfyIn(&output)
		return output
	}

	iter := newIterator(r.payload, r.length)
	if cmd, ok := iter.next(); ok {
		output = append(output, cmd...)
	}

	switch r.cmd {
	case cmdAuth:
		if cmd, ok := iter.next(); ok && len(cmd) > 0 {
			output = append(output, " ?"...)
		}
	case cmdHello:
		for {
			cmd, ok := iter.next()
			if !ok {
				break
			}
			output = append(output, ' ')
			output = append(output, cmd...)
			if bytes.EqualFold(cmd, []byte(cmdAuth)) {
				if cmd, ok = iter.next(); ok && len(cmd) > 0 {
					output = append(output, " ?"...)
				}
				break
			}
		}
	case cmdAppend, cmdGetSet, cmdLPushX, cmdGeoRadiusByMemeber, cmdRpushX,
		cmdSet, cmdSetNX, cmdSisMember, cmdZRank, cmdZRevRank, cmdZScore:
		iter.obfuscateNthIn(&output, 1)
	case cmdHSetNX, cmdLRem, cmdLSet, cmdSetBit, cmdSetEX, cmdPSetEX, cmdSetRange,
		cmdZIncBy, cmdSMove, cmdReStore:
		iter.obfuscateNthIn(&output, 2)
	case cmdLInsert:
		iter.obfuscateNthIn(&output, 3)
	case cmdGeoHash, cmdGeoPos, cmdGeoDist, cmdLPush, cmdRPush, cmdSRem, cmdZRem,
		cmdSAdd:
		if cmd, ok := iter.next(); ok && len(cmd) > 0 {
			output = append(output, ' ')
			output = append(output, cmd...)
			if cmd, ok = iter.next(); ok && len(cmd) > 0 {
				output = append(output, " ?"...)
			}
		}
	case cmdGeoAdd:
		if cmd, ok := iter.next(); ok && len(cmd) > 0 {
			output = append(output, ' ')
			output = append(output, cmd...)
			iter.obfuscateEveryNthIn(&output, 3)
		}
	case cmdHSet, cmdHMSet:
		if cmd, ok := iter.next(); ok && len(cmd) > 0 {
			output = append(output, ' ')
			output = append(output, cmd...)
			iter.obfuscateEveryNthIn(&output, 2)
		}
	case cmdMSet, cmdMSetNX:
		iter.obfuscateEveryNthIn(&output, 2)
	case cmdConfig:
		for {
			cmd, ok := iter.next()
			if !ok {
				break
			}
			output = append(output, ' ')
			output = append(output, cmd...)
			if bytes.EqualFold(cmd, []byte(cmdSet)) {
				iter.obfuscateEveryNthIn(&output, 2)
				break
			}
		}
	case cmdBitField:
		var indexAfterSet *int
		for {
			cmd, ok := iter.next()
			if !ok {
				break
			}
			output = append(output, ' ')

			if indexAfterSet != nil {
				*indexAfterSet++
				if *indexAfterSet == 3 {
					output = append(output, '?')
					indexAfterSet = nil
				} else {
					output = append(output, cmd...)
				}
			} else {
				output = append(output, cmd...)
			}
			if bytes.EqualFold(cmd, []byte(cmdSet)) {
				tmp := 0
				indexAfterSet = &tmp
			}
		}
	case cmdZAdd:
		if cmd, ok := iter.next(); ok && len(cmd) > 0 {
			output = append(output, ' ')
			output = append(output, cmd...)

			for {
				cmd, ok = iter.next()
				if !ok {
					break
				}
				output = append(output, ' ')
				output = append(output, cmd...)
				if len(cmd) > 4 || !isASCII(cmd) {
					break
				}
				cmdUpper := strings.ToUpper(string(cmd))
				if cmdUpper == "NX" || cmdUpper == "XX" || cmdUpper == "GT" ||
					cmdUpper == "LT" || cmdUpper == "CH" || cmdUpper == "INCR" {
					continue
				} else {
					break
				}
			}
			if cmd, ok = iter.next(); ok && len(cmd) > 0 {
				output = append(output, " ?"...)
			}

			iter.obfuscateEveryNthIn(&output, 2)
		}
	default:
		iter.stringfyIn(&output)
	}

	return output
}

// iterator.
type iterator struct {
	payload []byte
	index   int
	size    int
}

func newIterator(payload []byte, size int) *iterator {
	return &iterator{
		index:   0,
		payload: payload,
		size:    size,
	}
}

func (i *iterator) stringfyIn(output *[]byte) {
	for {
		cmd, ok := i.next()
		if !ok {
			break
		}
		if len(*output) > 0 && len(cmd) > 0 {
			*output = append(*output, ' ')
		}
		*output = append(*output, cmd...)
	}
}

func (i *iterator) obfuscateNthIn(output *[]byte, n int) {
	count := 0
	for {
		cmd, ok := i.next()
		if !ok {
			break
		}
		if len(*output) > 0 && len(cmd) > 0 {
			*output = append(*output, ' ')
		}
		if count == n {
			*output = append(*output, '?')
		} else {
			*output = append(*output, cmd...)
		}
		count++
	}
}

func (i *iterator) obfuscateEveryNthIn(output *[]byte, n int) {
	count := 0
	for {
		cmd, ok := i.next()
		if !ok {
			break
		}
		if len(*output) > 0 && len(cmd) > 0 {
			*output = append(*output, ' ')
		}
		if (count+1)%n == 0 {
			*output = append(*output, '?')
		} else {
			*output = append(*output, cmd...)
		}
		count++
	}
}

func (i *iterator) next() ([]byte, bool) {
	if i.index >= i.size {
		return nil, false
	}

	r := &redisInfo{}
	payload, command, err := r.decodeBulkString(i.payload)
	if err != nil {
		return nil, false
	}

	i.payload = payload
	i.index += 1

	return command, true
}

func readLength(payload []byte) ([]byte, int, error) {
	end := bytes.IndexFunc(payload, func(r rune) bool {
		return !(r == '+' || r == '-' || (r >= '0' && r <= '9')) || r == '\r'
	})
	if end == -1 {
		return nil, 0, errors.New("parse redis request error")
	}

	if end+2 > len(payload) || string(payload[end:end+2]) != customReturn {
		return nil, 0, errors.New("parse redis request error")
	}

	s := string(payload[:end])
	length, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, 0, errors.New("parse redis request error")
	}
	return payload[end+2:], int(length), nil
}

func isASCII(data []byte) bool {
	for _, c := range data {
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}

func (dec *redisDecPipe) determineDirection(port uint32, inf *redisInfo) (direction Direction) {
	if port == dec.fisrtDstPort {
		direction = clientToServer
		if dec.reqResp == 2 {
			dec.infCache = append(dec.infCache, inf)
			inf = &redisInfo{}
			dec.inf = inf
		}
	} else {
		direction = serverToClient
	}
	return direction
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type (
	decodeFunc func(*[]byte, []byte) ([]byte, error)
	valiteFunc func([]byte) ([]byte, error)
)

func assertDecErrNil(fn decodeFunc, output *[]byte, payload []byte) ([]byte, error) {
	if res, err := fn(output, payload); err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func assertValErrNil(fn valiteFunc, payload []byte) ([]byte, error) {
	if res, err := fn(payload); err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func redisCode2Status(code int) string {
	if code == 200 {
		return "ok"
	} else if code == 201 {
		return "error"
	}
	return ""
}

type Direction int

const (
	directionUnknown Direction = 0
	clientToServer   Direction = 1
	serverToClient   Direction = 2
)

const (
	customReturn     = "\r\n"
	trueReturn       = "t\r\n"
	falseReturn      = "f\r\n"
	minReqPayload    = "*0\r\n"
	minBulkString    = "$0\r\n"
	maxCommandLength = 17
	statusOK         = 200
	statusError      = 201
)

const (
	cmdAuth               = "AUTH"
	cmdHello              = "HELLO"
	cmdAppend             = "APPEND"
	cmdGetSet             = "GETSET"
	cmdLPushX             = "LPUSHX"
	cmdGeoRadiusByMemeber = "GEORADIUSBYMEMBER"
	cmdRpushX             = "RPUSHX"
	cmdSet                = "SET"
	cmdSetNX              = "SETNX"
	cmdSisMember          = "SISMEMBER"
	cmdZRank              = "ZRANK"
	cmdZRevRank           = "ZREVRANK"
	cmdZScore             = "ZSCORE"
	cmdHSetNX             = "HSETNX"
	cmdLRem               = "LREM"
	cmdLSet               = "LSET"
	cmdSetBit             = "SETBIT"
	cmdSetEX              = "SETEX"
	cmdPSetEX             = "PSETEX"
	cmdSetRange           = "SETRANGE"
	cmdZIncBy             = "ZINCBY"
	cmdSMove              = "SMOVE"
	cmdReStore            = "RESTORE"
	cmdLInsert            = "LINSERT"
	cmdGeoHash            = "GEOHASH"
	cmdGeoPos             = "GEOPOS"
	cmdGeoDist            = "GEODIST"
	cmdLPush              = "LPUSH"
	cmdRPush              = "RPUSH"
	cmdSRem               = "SREM"
	cmdZRem               = "ZRem"
	cmdSAdd               = "SADD"
	cmdGeoAdd             = "GEOADD"
	cmdHSet               = "HSET"
	cmdHMSet              = "HMSET"
	cmdMSet               = "MSET"
	cmdMSetNX             = "MSETNX"
	cmdConfig             = "CONFIG"
	cmdBitField           = "BITFIELD"
	cmdZAdd               = "ZADD"
)

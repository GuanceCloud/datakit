//go:build linux
// +build linux

package protodec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

func MysqlProtoDetect(data []byte, actSize int) (L7Protocol, ProtoDecPipe, bool) {
	if _, err := detectMysql(data, actSize); err != nil {
		return ProtoUnknown, nil, false
	}
	return ProtoMySQL, newMysqlDecPipe(ProtoMySQL), true
}

func detectMysql(data []byte, actSize int) (uint8, error) {
	if len(data) < 4 {
		return 0, errors.New("data payload too short")
	}

	header := data[:4]
	length := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)
	if length == 0 || length > actSize {
		return 0, errors.New("length should not eq 0 or gt actSize")
	}
	seq := header[3]

	var cnt bytes.Buffer
	reader := bytes.NewReader(data)
	_, err := reader.Seek(4, io.SeekStart)
	if err != nil {
		return 0, errors.New("unknown err")
	}

	if n, err := io.CopyN(&cnt, reader, int64(length)); err != nil {
		return 0, errors.New("unknown stream")
	} else if n != int64(length) {
		return 0, errors.New("unknown stream")
	}

	return seq, nil
}

type mysqlInfo struct {
	meta ProtoMeta

	stmtID   int
	inserted bool

	resource   string
	errMsg     string
	statusCode int

	reqBytes  int
	respBytes int

	ktime [4]uint64
	ts    int64
	dur   [2]uint64
}

type mysqlDecPipe struct {
	direction  comm.Direcion
	elems      []*mysqlInfo
	inf        *mysqlInfo
	reqResp    int // 0, 1, 2 1是请求 2是响应
	connClosed bool
}

func (dec *mysqlDecPipe) Decode(txRx comm.NICDirection, data *comm.NetwrkData,
	ts int64, thrTr threadTrace,
) {
	var (
		req, resp          bool
		resource           string
		statusCode, stmtID int
		err                error
	)
	if len(data.Payload) <= 4 {
		return
	}

	if dec.inf == nil {
		dec.inf = &mysqlInfo{}
		dec.reqResp = 0
	}
	inf := dec.inf
	if inf.resource != "" && inf.errMsg != "" {
		inf.resource = ""
		inf.errMsg = ""
	}

	if resource, stmtID, err = inf.isClientMsg(data.Payload, data.Fn); err == nil {
		req = true
	} else if resource, statusCode, err = inf.isServerMsg(data.Payload, data.Fn); err == nil {
		resp = true
	} else {
		return
	}

	switch {
	case req:
		if resource == StmtExec || resource == StmtClose {
			if len(dec.elems) > 0 { // 以防未知panic出现
				dec.inf = dec.elems[len(dec.elems)-1]
			}
			inf = dec.inf
			if resource == StmtExec {
				inf.errMsg = ""
			}
			if resource == StmtClose {
				inf.stmtID = stmtID
			}
			break
		}
		if dec.reqResp == 2 {
			inf = &mysqlInfo{}
			dec.inf = inf
		}
		dec.reqResp = 1
		inf.resource = resource
		inf.stmtID = stmtID
		inf.meta.Threads[0] = data.Thread
		inf.ts = ts

		// 这里确定这次连接到底是出还是入
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

	case resp:
		dec.reqResp = 2
		inf.meta.RespTCPSeq = data.TCPSeq
		inf.meta.Threads[1] = data.Thread

		inf.statusCode = statusCode
		inf.errMsg = resource

		if dec.direction == comm.DUnknown {
			switch txRx { //nolint:exhaustive
			case comm.NICDIngress:
				dec.direction = comm.DOut
			case comm.NICDEgress:
				dec.direction = comm.DIn
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

	if inf.resource != "" && inf.errMsg != "" {
		infCopy := *inf
		if inf.inserted && len(dec.elems) > 0 {
			dec.elems = dec.elems[:len(dec.elems)-1]
		}
		infCopy.inserted = true
		dec.elems = append(dec.elems, &infCopy)
	}
}

func (dec *mysqlDecPipe) Proto() L7Protocol {
	return ProtoMySQL
}

func (dec *mysqlDecPipe) Export(force bool) []*ProtoData {
	var result []*ProtoData
	var keep []*mysqlInfo
	for _, inf := range dec.elems {
		if inf == nil {
			continue
		}
		if inf.stmtID == 0 {
			kvs := make(point.KVs, 0, 20)

			switch dec.direction { //nolint:exhaustive
			case comm.DIn:
				kvs = kvs.Add(comm.FieldBytesRead, int64(inf.reqBytes), false, true)
				kvs = kvs.Add(comm.FieldBytesWritten, int64(inf.respBytes), false, true)
			default:
				kvs = kvs.Add(comm.FieldBytesRead, int64(inf.respBytes), false, true)
				kvs = kvs.Add(comm.FieldBytesWritten, int64(inf.reqBytes), false, true)
			}
			kvs = kvs.Add(comm.FieldResource, inf.resource, false, false)
			kvs = kvs.Add(comm.FieldMysqlStatusCode, inf.statusCode, false, false)
			kvs = kvs.Add(comm.FieldMysqlErrMsg, inf.errMsg, false, false)
			kvs = kvs.Add(comm.FieldStatus, code2Status(inf.statusCode), false, false)

			kvs = kvs.Add(comm.FieldOperation, ProtoMySQL.String(), false, true)

			dur := int64(inf.ktime[3] - inf.ktime[0])
			cost := int64(inf.ktime[2] - inf.ktime[1])
			result = append(result, &ProtoData{
				Meta:      inf.meta,
				Time:      inf.ts,
				KVs:       kvs,
				Cost:      cost,
				Duration:  dur,
				Direction: dec.direction,
				L7Proto:   ProtoMySQL,
				KTime:     inf.ktime[0],
			})
		} else {
			keep = append(keep, inf)
		}
	}
	dec.elems = keep
	return result
}

func (dec *mysqlDecPipe) ConnClose() {
	dec.connClosed = true
}

func newMysqlDecPipe(L7Protocol) ProtoDecPipe {
	return &mysqlDecPipe{}
}

func (m *mysqlInfo) isClientMsg(data []byte, fn comm.FnID) (string, int, error) {
	var (
		msg    string
		stmtID int
	)
	if len(data) <= 4 {
		return "", 0, errors.New("data length must gt 4")
	}
	switch fn { //nolint:exhaustive
	case comm.FnSysWrite, comm.FnSysWritev:
		data = data[4:]
	case comm.FnSysRecvfrom:
	default:
		return "", 0, errors.New("unsupported fn")
	}

	if len(data) < 1 {
		return "", 0, errors.New("data length too short")
	}

	switch data[0] {
	case ComInitDB:
		msg = fmt.Sprintf("USE %s;\n", data[1:])
	case ComDropDB:
		msg = fmt.Sprintf("Drop DB %s;\n", data[1:])
	case ComCreateDB, ComQuery:
		msg = string(data[1:])
	case ComStmtPrepare:
		msg = string(data[1:])
		stmtID = -1
	case ComStmtExecute:
		msg = StmtClose
		if len(data) < 5 {
			return "", 0, errors.New("data length should ge 5")
		}
		stmtID = int(binary.LittleEndian.Uint32(data[1:5]))
		if stmtID != m.stmtID {
			return "", 0, errors.New("not stmt exec")
		}
	case ComStmtSendLongData, ComStmtReset:
		msg = string(data[1:])
		if len(data) < 5 {
			return "", 0, errors.New("data length should ge 5")
		}
		stmtID = int(binary.LittleEndian.Uint32(data[1:5]))
	case ComStmtClose:
		msg = StmtClose
		stmtID = 0
	default:
		return "", 0, errors.New("unknown type")
	}

	return msg, stmtID, nil
}

func (m *mysqlInfo) isServerMsg(data []byte, fn comm.FnID) (string, int, error) {
	var (
		msg  string
		code int
	)
	if len(data) <= 4 {
		return msg, 0, errors.New("data length must gt 4")
	}

	switch fn { //nolint:exhaustive
	case comm.FnSysSendto:
		data = data[4:]
	case comm.FnSysRead, comm.FnSysReadv:
		data = data[4:]
	default:
	}

	if len(data) < 1 {
		return "", 0, errors.New("data length too short")
	}

	switch data[0] {
	case 0xff:
		if len(data) < 3 {
			return msg, 0, errors.New("len(data) should ge 3")
		}
		code = int(binary.LittleEndian.Uint16(data[1:3]))
		var errorMsg string
		if data[3] == 0x23 {
			if len(data) <= 9 {
				return msg, 0, errors.New("error message should ge 10")
			}
			errorMsg = readMsgByte(data[9:])
		} else {
			if len(data) <= 4 {
				return msg, 0, errors.New("error message should ge 5")
			}
			errorMsg = readMsgByte(data[4:])
		}
		msg = fmt.Sprintf("Error Code: %s, Error Msg: %s", strconv.Itoa(code), strings.TrimSpace(errorMsg))
	case 0x00:
		code = int(data[0])
		msg = MsgOk
		if m.stmtID == -1 && len(data) >= 5 {
			m.stmtID = int(binary.LittleEndian.Uint32(data[1:5]))
		}
	case 0x0a: // 现在 mysql protocol 默认为 10
		var l int
		l = bytes.IndexByte(data, 0x00)
		if l == -1 {
			return "", 0, errors.New("not found version")
		}
		version := data[1:l]
		msg = fmt.Sprintf("Server Greeting, Mysql Version: %s", version)
	default:
		return msg, 0, errors.New("unknown type")
	}

	return msg, code, nil
}

func readMsgByte(cnt []byte) string {
	var l int
	l = bytes.IndexByte(cnt, 0x00)
	if l == -1 {
		l = len(cnt)
	}
	return string(cnt[0:l])
}

func code2Status(code int) string {
	switch {
	case code == 0:
		return "ok"
	case code >= 1000:
		return "error"
	default:
		return ""
	}
}

//nolint:staticcheck
const (
	ComSleep            byte = 0
	ComQuit                  = 1
	ComInitDB                = 2
	ComQuery                 = 3
	ComFieldList             = 4
	ComCreateDB              = 5
	ComDropDB                = 6
	ComRefresh               = 7
	ComShutdown              = 8
	ComStatistics            = 9
	ComProcessInfo           = 10
	ComConnect               = 11
	ComProcessKill           = 12
	ComDebug                 = 13
	ComPing                  = 14
	ComTime                  = 15
	ComDelayedInsert         = 16
	ComChangeUser            = 17
	ComBinlogDump            = 18
	ComTableDump             = 19
	ComConnectOut            = 20
	ComRegisterSlave         = 21
	ComStmtPrepare           = 22
	ComStmtExecute           = 23
	ComStmtSendLongData      = 24
	ComStmtClose             = 25
	ComStmtReset             = 26
	ComSetOption             = 27
	ComStmtFetch             = 28
	ComDaemon                = 29
	ComBinlogDumpGtid        = 30
	ComResetConnection       = 31
)

const (
	MsgOk     = "OK"
	StmtExec  = "STMT EXEC"
	StmtClose = "STMT CLOSE"
)

//go:build linux
// +build linux

package protodec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/obfuscate"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

func MysqlProtoDetect(data []byte, actSize int) (L7Protocol, ProtoDecPipe, bool) {
	if checkMysql(data) {
		return ProtoMySQL, newMysqlDecPipe(ProtoMySQL), true
	}
	return ProtoUnknown, nil, false
}

func checkMysql(payload []byte) bool {
	hd := &headerDecoder{}
	offset, err := hd.decode(payload, 0)
	if err != nil {
		return false
	}
	if offset < 0 {
		return false
	}

	if hd.number != 0 || offset+hd.length > len(payload) {
		return false
	}

	protoVersionOrQueryType := payload[offset]
	switch protoVersionOrQueryType {
	// 只对Query、Statement过程进行检测，其余丢弃
	case ComQuery, ComStmtPrepare:
		payload = readMysql(payload[offset+1:])
		m := &mysqlInfo{}
		return isASCII(payload) && m.isMysql(payload)
	case protoVersionFix:
		return true
	default:
		return false
	}
}

type headerDecoder struct {
	number uint8
	length int
}

func (d *headerDecoder) decode(payload []byte, offset int) (int, error) {
	if len(payload) < 5 {
		return -1, errors.New("payload too short")
	}
	lastPacketNumber := 0
	for offset < len(payload) {
		if len(payload[offset:]) < headerLen {
			return offset, errContinue
		}

		if len(payload[offset:]) > compressHeaderLen+headerLen {
			compressedLen := int(binary.LittleEndian.Uint32(payload[offset:]) & 0xffffff)
			unCompressedLen := int(binary.LittleEndian.Uint32(payload[offset+compressHeaderUncompressOffset:]) & 0xffffff)
			packetLen := int(binary.LittleEndian.Uint32(payload[offset+compressHeaderLen:]) & 0xffffff)
			if unCompressedLen == 0 && compressedLen == packetLen+4 {
				offset += compressHeaderLen
			}
		}

		length := int(binary.LittleEndian.Uint32(payload[offset:]) & 0xffffff)
		if offset+headerLen+responseCodeOffset >= len(payload) {
			// 当前包没有读完，应该读完等下次继续
			return offset, errContinue
		}

		responseCode := payload[offset+headerLen+responseCodeOffset]
		if responseCode == mysqlResponseCodeOK || responseCode == mysqlResponseCodeERR ||
			responseCode == mysqlResponseCodeEOF || payload[offset+numberOffset] == 0 && lastPacketNumber != 255 {
			d.length = length
			d.number = payload[offset+numberOffset]
			return offset + headerLen, nil
		}
		// 当packetNum超过255,则会归0，这里需要特别判断一下
		lastPacketNumber = int(payload[offset+numberOffset])
		offset += headerLen + length
	}
	return offset, errContinue
}

// checkHeader return 0: unknown 1: request 2: response 3: greeting.
func (d *headerDecoder) checkHeader(payload []byte, offset int) (mysqlPacketType, error) {
	if offset >= len(payload) || len(payload) == 0 {
		return packetUnknown, errors.New("check mysql header error: payload length error")
	}

	if d.number == 0 {
		payload = payload[offset:]
		if len(payload) < protoVersionLen {
			return packetUnknown, errors.New("check mysql header error: payload length error")
		}
		protoVersionOrCommand := payload[protoVersionOffset]
		end := bytes.IndexFunc(payload[serverVersionOffset:], func(r rune) bool {
			return byte(r) == serverVersionEOF
		})
		switch {
		case end != -1 && protoVersionOrCommand == protoVersionFix:
			return packetGreeting, nil
		case int(protoVersionOrCommand) > 0 && int(protoVersionOrCommand) < 36:
			return packetRequest, nil
		default:
			return packetUnknown, errors.New("check mysql header error")
		}
	} else {
		return packetResponse, nil
	}
}

type mysqlInfo struct {
	meta ProtoMeta

	statementID int

	resource     string
	resourceType string
	comment      string
	errMsg       string

	protoVersion byte
	responseCode byte
	command      byte
	errCode      int
	affectedRow  uint64

	packetType    mysqlPacketType
	hasRequest    bool
	readyToExport bool

	reader *bytes.Buffer
	offset int

	reqBytes  int
	respBytes int

	ktime [4]uint64
	ts    int64
	dur   [2]uint64
}

type mysqlDecPipe struct {
	direction comm.Direcion
	// fisrtDstPort uint32
	protoVersion byte
	infCache     []*mysqlInfo
	inf          *mysqlInfo
	reqResp      int // 0, 1, 2 1是请求 2是响应
	connClosed   bool

	lastFn             comm.FnID // 用于判断是否是头四个字节丢失的问题
	lastDur, lastKtime uint64
}

func (m *mysqlInfo) parse(payload []byte, seq uint32) (uint32, error) {
	var (
		packetType  mysqlPacketType
		offset      int
		err         error
		shouldReset = true
	)

	// 重置缓冲区
	defer func() {
		if shouldReset {
			m.offset = 0
			m.reader.Reset()
		}
	}()

	// 用于计算seq偏移量
	length := m.reader.Len()
	m.reader.Write(payload)
	m.packetType = packetUnknown
	hd := &headerDecoder{}
	offset, err = hd.decode(m.reader.Bytes(), m.offset)
	m.offset = offset
	if errors.Is(err, errContinue) {
		shouldReset = false // 不重置缓冲区
		return 0, errContinue
	} else if err != nil {
		return 0, errors.New("parse mysql error")
	}
	if offset < 0 {
		return 0, errors.New("parse mysql error")
	}

	packetType, err = hd.checkHeader(m.reader.Bytes(), offset)
	if err != nil {
		return 0, err
	}

	// 跳过offset部分，用于下面解析
	m.reader.Next(offset)

	switch packetType { //nolint:exhaustive
	case packetRequest:
		var err error
		packetType, err = m.parseRequest(m.reader.Bytes())
		if err != nil {
			return 0, err
		}
		if packetType == packetRequest {
			m.hasRequest = true
		}
	case packetResponse:
		err = m.parseResponse(m.reader.Bytes())
		if err != nil {
			return 0, err
		}
		m.hasRequest = false
	default:
		return 0, errors.New("unknown packet")
	}
	m.packetType = packetType
	return seq - uint32(length), nil
}

func (m *mysqlInfo) parseRequest(payload []byte) (mysqlPacketType, error) {
	if len(payload) < commandLen {
		return packetUnknown, errors.New("parse mysql request error: payload too short")
	}
	m.command = payload[commandOffset]
	packetType := packetRequest
	switch m.command {
	case ComQuit, ComStmtClose:
		packetType = packetSession
	case ComFieldList, ComStmtFetch:
	case ComInitDB, ComQuery:
		if err := m.decodeRequestString(payload[commandOffset+commandLen:]); err != nil {
			return packetUnknown, errors.New("parse mysql request error")
		}
	case ComStmtPrepare:
		if err := m.decodeRequestString(payload[commandOffset+commandLen:]); err != nil {
			return packetUnknown, errors.New("parse mysql request error")
		}
	case ComStmtExecute:
		statementID := m.getStatementID(payload[statementIDOffset:])
		if statementID == -1 || (m.statementID != 0 && m.statementID != statementID) {
			return packetUnknown, errors.New("parse mysql request error")
		}
		m.statementID = statementID
	// if len(payload) > executeStatementParamsOffset {

	// }
	case ComPing:
		m.resource = "PING"
	default:
		return packetUnknown, errors.New("unknown")
	}
	return packetType, nil
}

func (m *mysqlInfo) parseResponse(payload []byte) error {
	length := len(payload)
	if length < responseCodeLen {
		return errors.New("parse mysql response error: payload too short")
	}
	m.responseCode = payload[responseCodeOffset]
	length -= responseCodeLen
	switch m.responseCode {
	case mysqlResponseCodeERR:
		if length > errorCodeLen {
			code := binary.LittleEndian.Uint16(payload[errorCodeOffset:3])
			if code < serverStatusCodeMin || code > clientStatusCodeMax {
				return errors.New("parse mysql response error: code error")
			}
			m.errCode = int(code)
			length -= errorCodeLen
		}
		var errorMsgOffset int
		if length > sqlStateLen && payload[sqlStateOffset] == sqlStateMarker {
			errorMsgOffset = sqlStateOffset + sqlStateLen
		} else {
			errorMsgOffset = sqlStateOffset
		}
		if len(payload) > errorMsgOffset {
			context := readMysql(payload[errorMsgOffset:])
			if !isASCII(context) {
				return errors.New("parse mysql response error")
			}
			m.errMsg = string(bytes.Runes(context))
		}
	case mysqlResponseCodeOK, mysqlResponseCodeEOF:
		m.affectedRow = decodeCompressInt(payload)
		statementID := m.getStatementID(payload[statementIDOffset:])
		if statementID == -1 {
			return errors.New("parse mysql response error")
		}
		m.statementID = statementID
	default:
		return errors.New("parse mysql response error")
	}
	return nil
}

/*
func (m *mysqlInfo) parseGreeting(payload []byte) error {
	length := len(payload)
	if length < protoVersionLen {
		return errors.New("parse greeting error: payload too short")
	}
	m.protoVersion = payload[protoVersionOffset]
	length -= protoVersionLen
	end := bytes.IndexFunc(payload[serverVersionOffset:], func(r rune) bool {
		return byte(r) == serverVersionEOF
	})
	if end == -1 {
		return errors.New("parse greeting error")
	}
	length -= end
	if length < threadIDLen+1 {
		return errors.New("parse greeting error")
	}
	return nil
}
*/

func decodeCompressInt(payload []byte) uint64 {
	if length := len(payload); length == 0 {
		return 0
	}
	value := payload[0]
	switch value {
	case intFlags2:
		if len(payload) > intBaseLen+2 {
			return uint64(binary.LittleEndian.Uint16(payload[intBaseLen:]))
		}
	case intFlags3:
		if len(payload) > intBaseLen+3 {
			return uint64(binary.LittleEndian.Uint16(payload[intBaseLen:])) |
				(uint64(payload[intBaseLen+2]) << 16)
		}
	case intFlags8:
		if len(payload) > intBaseLen+8 {
			return binary.LittleEndian.Uint64(payload[intBaseLen:])
		}
	default:
		return uint64(value)
	}
	return uint64(value)
}

func (dec *mysqlDecPipe) Decode(txRx comm.NICDirection, data *comm.NetwrkData,
	ts int64, thrTr threadTrace,
) {
	if len(data.Payload) == 0 {
		return
	}

	// TODO: fix sshd
	if data.Conn.Dport == 22 || data.Conn.Sport == 22 {
		return
	}

	if dec.inf == nil {
		dec.inf = &mysqlInfo{
			reader: bytes.NewBuffer(nil),
		}
		dec.reqResp = 0
	}

	inf := dec.inf
	if inf.readyToExport {
		dec.infCache = append(dec.infCache, inf)
		inf = &mysqlInfo{
			reader: bytes.NewBuffer(nil),
		}
		dec.inf = inf
	}
	inf.readyToExport = false

	// 采集的数据可能会出现头四个字节分开的问题，如果是这种问题，先读入reader,下一步再做检测
	if (data.Fn == comm.FnSysRecvfrom || data.Fn == comm.FnSysRead) && len(data.Payload) == 4 {
		dec.lastFn = data.Fn
		inf.reader.Write(data.Payload)
		dec.lastDur = data.TS
		dec.lastKtime = data.TSTail
		return
	} else {
		dec.lastDur = 0
		dec.lastKtime = 0
	}

	firstSeq, err := inf.parse(data.Payload, data.TCPSeq)
	if err != nil {
		return
	}

	// 恢复lastFn
	dec.lastFn = comm.FnUnknown

	if dec.lastDur != 0 {
		data.TS = dec.lastDur
	}
	if dec.lastKtime != 0 {
		data.TSTail = dec.lastKtime
	}

	switch inf.packetType { // nolint:exhaustive
	case packetRequest, packetSession:
		// 对于 Statement 过程，当执行到 Close Statement 请求就完成
		if inf.packetType == packetSession {
			inf.readyToExport = true
			dec.reqResp = 1
			inf.dur[1] = data.TS
			inf.ktime[2] = data.TSTail
			break
		}
		dec.reqResp = 1

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

		inf.meta.ReqTCPSeq = firstSeq
		if inf.dur[0] == 0 {
			inf.dur[0] = data.TS
		}
		if inf.ktime[0] == 0 {
			inf.ktime[0] = data.TSTail
		}

	case packetResponse:
		if inf.packetType == packetGreeting {
			dec.protoVersion = inf.protoVersion
		} else if inf.statementID == 0 {
			// 对于 Query 过程，不存在StatementID, 此时请求就完成
			inf.readyToExport = true
		}

		if dec.direction == comm.DUnknown {
			switch txRx { //nolint:exhaustive
			case comm.NICDIngress:
				dec.direction = comm.DOut
			case comm.NICDEgress:
				dec.direction = comm.DIn
			}
		}

		dec.reqResp = 2
		inf.meta.RespTCPSeq = firstSeq
		inf.meta.Threads[1] = data.Thread

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

func (dec *mysqlDecPipe) Proto() L7Protocol {
	return ProtoMySQL
}

func (dec *mysqlDecPipe) Export(force bool) []*ProtoData {
	if force {
		dec.infCache = append(dec.infCache, dec.inf)
		dec.inf = nil
	}
	var result []*ProtoData

	for _, inf := range dec.infCache {
		if inf != nil && inf.readyToExport && inf.resource != "" {
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
			kvs = kvs.Add(comm.FiledMysqlComment, inf.comment, false, true)
			kvs = kvs.Add(comm.FieldStatusCode, inf.responseCode, false, true)
			kvs = kvs.Add(comm.FieldErrMsg, inf.errMsg, false, true)
			kvs = kvs.Add(comm.FieldErrCode, inf.errCode, false, true)
			kvs = kvs.Add(comm.FieldStatus, code2Status(inf.errCode), false, true)

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
		}
	}
	dec.infCache = dec.infCache[:0]
	return result
}

func (dec *mysqlDecPipe) ConnClose() {
	dec.connClosed = true
}

func newMysqlDecPipe(L7Protocol) ProtoDecPipe {
	return &mysqlDecPipe{}
}

func (m *mysqlInfo) decodeRequestString(payload []byte) error {
	payload = readMysql(payload)
	if (m.command == ComQuery || m.command == ComStmtPrepare) && !m.isMysql(payload) {
		return errors.New("parse mysql request error")
	}
	comment, command, clean := trimCommentGetFirst(payload, 8)

	var resource []byte
	if output, err := obfuscate.NewObfuscator(nil).Obfuscate("sql", string(clean)); err != nil && output != nil {
		o := []byte(output.Query)
		validLen := utf8ValidLength(o)
		resource = o[:validLen]
	} else {
		resource = []byte(string(bytes.Runes(clean)))
	}

	m.resource = string(bytes.Runes(resource))
	m.comment = strings.TrimSpace(string(bytes.Runes(comment)))
	m.resourceType = string(bytes.Runes(command))
	return nil
}

func (m *mysqlInfo) getStatementID(payload []byte) int {
	if len(payload) > statementIDLen {
		return int(binary.LittleEndian.Uint16(payload[:4]))
	}
	return -1
}

func (m *mysqlInfo) isMysql(sql []byte) bool {
	_, command, _ := trimCommentGetFirst(sql, 8)
	return m.isValidSQL(command)
}

func (m *mysqlInfo) isValidSQL(word []byte) bool {
	if _, ok := sqlCommand[strings.ToUpper(string(word))]; ok {
		return true
	}
	if _, ok := mysqlStart[strings.ToUpper(string(word))]; ok {
		return true
	}
	return false
}

func trimCommentGetFirst(sql []byte, firstMaxLen int) ([]byte, []byte, []byte) {
	length := len(sql)
	sqlIter := newByteIterator(sql, length)
	next := 0
	for {
		_, _, ok := sqlIter.peek()
		if !ok {
			break
		}
		if !(bytes.HasPrefix(sql[next:], []byte("/*")) || bytes.HasPrefix(sql[next:], []byte(" ")) ||
			bytes.HasPrefix(sql[next:], []byte("\n")) || bytes.HasPrefix(sql[next:], []byte("\t")) ||
			bytes.HasPrefix(sql[next:], []byte("\r"))) {
			break
		}

		for {
			if _, ok := sqlIter.nextIf(func(_ int, c byte) bool {
				return c == ' ' || c == '\n' || c == '\t' || c == '\r'
			}); !ok {
				break
			}
		}

		next = sqlIter.nextIndex()
		if next >= length {
			break
		}
		if bytes.HasPrefix(sql[next:], []byte("/*")) {
			// skip comment
			for {
				if _, ok := sqlIter.nextIf(func(j int, _ byte) bool {
					return !bytes.HasPrefix(sql[j:], []byte("*/"))
				}); !ok {
					break
				}
			}
			next = sqlIter.nextIndex()
			if next >= length {
				break
			} else {
				for i := 0; i < 2; i++ {
					sqlIter.next()
				}
				next = sqlIter.nextIndex()
			}
		}
	}

	start := sqlIter.nextIndex()
	for {
		if _, ok := sqlIter.nextIf(func(_ int, c byte) bool {
			return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
		}); !ok {
			break
		}
	}
	end := sqlIter.nextIndex()
	if start < end && end-start <= firstMaxLen {
		return sql[:start], sql[start:end], sql[start:]
	}
	return nil, nil, nil
}

type byteIterator struct {
	data  []byte
	index int
	size  int
}

func newByteIterator(data []byte, size int) *byteIterator {
	return &byteIterator{
		data:  data,
		index: 0,
		size:  size,
	}
}

func (it *byteIterator) next() (byte, bool) {
	if it.index >= it.size {
		return 0, false
	}
	current := it.data[it.index]
	it.index += 1

	return current, true
}

func (it *byteIterator) peek() (byte, int, bool) {
	if it.index >= it.size {
		return 0, 0, false
	}
	return it.data[it.index], it.index, true
}

func (it *byteIterator) nextIf(cond func(int, byte) bool) (byte, bool) {
	if it.index >= it.size {
		return 0, false
	}
	if cond(it.index, it.data[it.index]) {
		current := it.data[it.index]
		it.index += 1
		return current, true
	}
	return 0, false
}

func (it *byteIterator) nextIndex() int {
	if _, _, ok := it.peek(); ok {
		return it.index
	}
	return it.size
}

/*
func (it *byteIterator) forward(n int) {
	for i := 0; i < n; i++ {
		it.next()
	}
}
*/

func utf8ValidLength(data []byte) int {
	validLen := 0
	for len(data) > 0 {
		_, size := utf8.DecodeRune(data)
		if size == 0 {
			break
		}
		validLen += size
		data = data[size:]
	}
	return validLen
}

func readMysql(payload []byte) []byte {
	if len(payload) > 2 && payload[0] == 0 && payload[1] == 1 {
		return payload[2:]
	}
	return payload
}

func code2Status(code int) string {
	switch {
	case code == 0:
		return "ok"
	case code >= int(clientStatusCodeMin) && code <= int(clientStatusCodeMax):
		return "client error"
	default:
		return "server error"
	}
}

//nolint:staticcheck
const (
	ComSleep            byte = 0
	ComQuit             byte = 1
	ComInitDB           byte = 2
	ComQuery            byte = 3
	ComFieldList        byte = 4
	ComCreateDB         byte = 5
	ComDropDB           byte = 6
	ComRefresh          byte = 7
	ComShutdown         byte = 8
	ComStatistics       byte = 9
	ComProcessInfo      byte = 10
	ComConnect          byte = 11
	ComProcessKill      byte = 12
	ComDebug            byte = 13
	ComPing             byte = 14
	ComTime             byte = 15
	ComDelayedInsert    byte = 16
	ComChangeUser       byte = 17
	ComBinlogDump       byte = 18
	ComTableDump        byte = 19
	ComConnectOut       byte = 20
	ComRegisterSlave    byte = 21
	ComStmtPrepare      byte = 22
	ComStmtExecute      byte = 23
	ComStmtSendLongData byte = 24
	ComStmtClose        byte = 25
	ComStmtReset        byte = 26
	ComSetOption        byte = 27
	ComStmtFetch        byte = 28
	ComDaemon           byte = 29
	ComBinlogDumpGtid   byte = 30
	ComResetConnection  byte = 31
)

const (
	MsgOk     = "OK"
	StmtExec  = "STMT EXEC"
	StmtClose = "STMT CLOSE"
)

const (
	compressHeaderLen              = 7
	compressHeaderUncompressOffset = 4
	headerLen                      = 4
	protoVersionLen                = 1
	commandLen                     = 1
	responseCodeLen                = 1
	statementIDLen                 = 4
	errorCodeLen                   = 2
	sqlStateLen                    = 6
	intBaseLen                     = 1
	numberOffset                   = 3
	responseCodeOffset             = 0
	protoVersionOffset             = 0
	serverVersionOffset            = 1
	commandOffset                  = 0
	statementIDOffset              = 1
	errorCodeOffset                = 1
	sqlStateOffset                 = 3
)

const (
	serverStatusCodeMin uint16 = 1000
	clientStatusCodeMax uint16 = 2999
	clientStatusCodeMin uint16 = 2000
)

const (
	mysqlResponseCodeOK  byte = 0x00
	mysqlResponseCodeERR byte = 0xff
	mysqlResponseCodeEOF byte = 0xfe
	serverVersionEOF     byte = 0x00
	protoVersionFix      byte = 0x0a
	sqlStateMarker       byte = '#'
	intFlags2            byte = 0xfc
	intFlags3            byte = 0xfd
	intFlags8            byte = 0xfe
)

type mysqlPacketType int

const (
	packetUnknown  mysqlPacketType = 0
	packetRequest  mysqlPacketType = 1
	packetResponse mysqlPacketType = 2
	packetGreeting mysqlPacketType = 3
	packetSession  mysqlPacketType = 4
)

var sqlCommand = map[string]struct{}{
	"SELECT":    {},
	"PREPARE":   {},
	"INSERT":    {},
	"UPDATE":    {},
	"DELETE":    {},
	"SHOW":      {},
	"CREATE":    {},
	"DROP":      {},
	"ALTER":     {},
	"EXPLAIN":   {},
	"GRANT":     {},
	"SET":       {},
	"SAVEPOINT": {},
	"RELEASE":   {},
	"DECLARE":   {},
	"CALL":      {},
	"FETCH":     {},
	"IMPORT":    {},
	"REVOKE":    {},
}

var mysqlStart = map[string]struct{}{
	"XA":       {},
	"FLUSH":    {},
	"SHOW":     {},
	"USE":      {},
	"LOCK":     {},
	"UNLOCK":   {},
	"STOP":     {},
	"START":    {},
	"LOAD":     {},
	"ANALYZE":  {},
	"BEGIN":    {},
	"COMMIT":   {},
	"ROLLBACK": {},
	"DESC":     {},
}

var errContinue = errors.New("malformed packet, should continue")

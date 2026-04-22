//go:build linux
// +build linux

package l4log

import (
	"strconv"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
)

const (
	// sysctl -o net.ipv4.tcp_keepalive_time
	// defaultTCPKeepAlive = time.Hour * 2 // 7200s.

	twoMSL              = time.Minute     // buf default is 120s
	defaultTCPKeepAlive = time.Minute * 2 // >5min 直接结束，不使用内核设置，避免堆积

	tcpPktLimitPerChunk = 256
)

type L7ProtoEventAndMetric interface {
	Handle(txrx int8, cnt []byte, cntSize int64, ln *PktTCPHdr,
		k *PMeta, pktState int8) error
	DetectProto(cnt []byte) bool
}

type tcpStatus int8

const (
	TCPUnknownStatus tcpStatus = iota

	TCPSYNSend     // syn ->
	TCPSYNRcvd     // <- syn ;;  syn|ack ->
	TCPEstablished // ack ->

	// the party that actively closes.
	TCPFINWait1 // fin ->
	TCPFINWait2 // <- ack

	// not use.
	TCPTimeWait  // <- fin ;; ack ->
	TCPCloseWait // ack ->

	TCPLastAck // fin -> ;; <- ack

	TCPClose
)

type TCPMetrics struct {
	BytesRead    int `json:"bytes_read"`
	BytesWritten int `json:"bytes_written"`

	RTT    int64 `json:"rtt"`     // us
	RTTVar int64 `json:"rtt_var"` // us

	Retransmits int `json:"retransmits"`

	recEstab bool
	recClose [2]bool
}

const (
	chunkKindSYN uint8 = 1 << iota
	chunkKindFINRST
)

// func isSYNORFINChunk(k uint8) bool {
// 	return k != 0
// }

func isSYNChunk(k uint8) bool {
	return k&0b1 == 0b1
}

func isFINChunk(k uint8) bool {
	return k&0b10 == 0b10
}

type PktChunk struct {
	chunkKind uint8 // 0b01: syn, 0b10: fin/rst

	ChunkID int64 `json:"chunk_id"`

	// rtt float64

	txSeq [2]uint32
	rxSeq [2]uint32

	setFlag  [2]bool
	TxSeqPos uint32 `json:"tx_seq_pos"`
	RxSeqPos uint32 `json:"rx_seq_pos"`
	TimePos  int64  `json:"time_pos"`

	macCount   int
	macEntries [2]pktChunkMACEntry
	extraMAC   map[string]string
	MACMap     map[string]string `json:"mac_map"`
	TCPColName []string          `json:"tcp_series_col_name"`
	TCPSreries []PktTCPHdr       `json:"tcp_series"`

	RxBytes int `json:"rx_bytes"`
	TxBytes int `json:"tx_bytes"`

	RXPacket int64 `json:"rx_packets"`
	TXPacket int64 `json:"tx_packets"`

	// SPacket, DPacket int
	RetransmitsTx int `json:"tx_retrans"`
	RetransmitsRx int `json:"rx_retrans"`

	RSTTx int `json:"tx_rst"`
	RSTRx int `json:"rx_rst"`

	messageCache string
	messageDirty bool
}

type pktChunkMACEntry struct {
	mac string
	id  string
}

func (chunk *PktChunk) appendJSON(buf []byte) []byte {
	buf = append(buf, `{"chunk_id":`...)
	buf = strconv.AppendInt(buf, chunk.ChunkID, 10)
	buf = append(buf, `,"tx_seq_pos":`...)
	buf = strconv.AppendUint(buf, uint64(chunk.TxSeqPos), 10)
	buf = append(buf, `,"rx_seq_pos":`...)
	buf = strconv.AppendUint(buf, uint64(chunk.RxSeqPos), 10)
	buf = append(buf, `,"time_pos":`...)
	buf = strconv.AppendInt(buf, chunk.TimePos, 10)
	buf = append(buf, `,"mac_map":`...)
	buf = chunk.appendJSONMACMap(buf)
	buf = append(buf, `,"tcp_series_col_name":["txrx","src_mac","dst_mac","flags","seq","ack_seq","payload_size","win","ts"]`...)
	buf = append(buf, `,"tcp_series":[`...)
	for i, item := range chunk.TCPSreries {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = appendPktTCPHdrJSON(buf, item)
	}
	buf = append(buf, `],"rx_bytes":`...)
	buf = strconv.AppendInt(buf, int64(chunk.RxBytes), 10)
	buf = append(buf, `,"tx_bytes":`...)
	buf = strconv.AppendInt(buf, int64(chunk.TxBytes), 10)
	buf = append(buf, `,"rx_packets":`...)
	buf = strconv.AppendInt(buf, chunk.RXPacket, 10)
	buf = append(buf, `,"tx_packets":`...)
	buf = strconv.AppendInt(buf, chunk.TXPacket, 10)
	buf = append(buf, `,"tx_retrans":`...)
	buf = strconv.AppendInt(buf, int64(chunk.RetransmitsTx), 10)
	buf = append(buf, `,"rx_retrans":`...)
	buf = strconv.AppendInt(buf, int64(chunk.RetransmitsRx), 10)
	buf = append(buf, `,"tx_rst":`...)
	buf = strconv.AppendInt(buf, int64(chunk.RSTTx), 10)
	buf = append(buf, `,"rx_rst":`...)
	buf = strconv.AppendInt(buf, int64(chunk.RSTRx), 10)
	buf = append(buf, '}')
	return buf
}

func (chunk *PktChunk) appendJSONMACMap(buf []byte) []byte {
	if chunk.macCount == 0 && len(chunk.extraMAC) == 0 {
		return append(buf, "null"...)
	}

	buf = append(buf, '{')
	wrote := 0
	for i := 0; i < chunk.macCount; i++ {
		if wrote > 0 {
			buf = append(buf, ',')
		}
		buf = strconv.AppendQuote(buf, chunk.macEntries[i].mac)
		buf = append(buf, ':')
		buf = strconv.AppendQuote(buf, chunk.macEntries[i].id)
		wrote++
	}
	for mac, id := range chunk.extraMAC {
		if wrote > 0 {
			buf = append(buf, ',')
		}
		buf = strconv.AppendQuote(buf, mac)
		buf = append(buf, ':')
		buf = strconv.AppendQuote(buf, id)
		wrote++
	}
	buf = append(buf, '}')
	return buf
}

func (chunk *PktChunk) GetMacID(mac string) string {
	for i := 0; i < chunk.macCount; i++ {
		if chunk.macEntries[i].mac == mac {
			return chunk.macEntries[i].id
		}
	}

	if chunk.extraMAC != nil {
		if v, ok := chunk.extraMAC[mac]; ok {
			return v
		}
	}

	id := strconv.Itoa(chunk.macCount + len(chunk.extraMAC) + 1)
	if chunk.macCount < len(chunk.macEntries) {
		chunk.markMessageDirty()
		chunk.macEntries[chunk.macCount] = pktChunkMACEntry{
			mac: mac,
			id:  id,
		}
		chunk.macCount++
		return id
	}

	if chunk.extraMAC == nil {
		chunk.extraMAC = make(map[string]string, 2)
	}
	chunk.markMessageDirty()
	chunk.extraMAC[mac] = id
	return id
}

func (chunk *PktChunk) markMessageDirty() {
	chunk.messageDirty = true
}

func (chunk *PktChunk) recSeqRange(seq, ack uint32, tx bool, tcpflag TCPFlag) {
	chunk.markMessageDirty()
	var noAck, noSeq bool
	if tx {
		if tcpflag == TCPSYN || tcpflag == TCPRST {
			// ack 为 0，通过设置 flag 避免干扰回绕的处理
			noAck = true
		}
	} else {
		seq, ack = ack, seq
		if tcpflag == TCPSYN || tcpflag == TCPRST {
			// 翻转后 seq 为 0
			noSeq = true
		}
	}

	// 翻转后视为 tx 处理

	if !noSeq {
		if chunk.txSeq[0] == 0 || seq < chunk.txSeq[0] {
			chunk.txSeq[0] = seq
		}
	}

	if seq > chunk.txSeq[1] {
		chunk.txSeq[1] = seq
	}

	if !noAck {
		if chunk.rxSeq[0] == 0 || ack < chunk.rxSeq[0] {
			chunk.rxSeq[0] = ack
		}
	}

	if ack > chunk.rxSeq[1] {
		chunk.rxSeq[1] = ack
	}
}

type conndirection int8

const (
	directionUnknown conndirection = iota
	directionIncoming
	directionOutgoing
)

func (c conndirection) String() string {
	switch c {
	case directionIncoming:
		return netflow.DirectionIncoming
	case directionOutgoing:
		return netflow.DirectionOutgoing
	case directionUnknown:
		return netflow.DirectionUnknown
	default:
		return netflow.DirectionUnknown
	}
}

type RTT struct {
	// 当前一次不带载荷则不计算
	nextSeq uint32
	ack     bool

	ts     int64
	prvRtt int64
}

func (rtt *RTT) cal(txRx int8, seq uint32, ts int64) {
	if rtt.ack || seq == 0 {
		return
	}

	switch txRx {
	case directionTX:
		// 暂不考虑回绕，这需要记录 next seq
		if seq >= rtt.nextSeq {
			rtt.nextSeq = seq
			rtt.ts = ts
		}
	case directionRX:
		if seq == rtt.nextSeq {
			rtt.ts = ts - rtt.ts
			rtt.ack = true
		}
	}
}

func (rtt *RTT) toNext() {
	if rtt.ack {
		rtt.prvRtt = rtt.ts
	}

	rtt.ack = false
	rtt.nextSeq = 0
	rtt.ts = 0
}

func (rtt *RTT) getRTT() int64 {
	var rttVal int64
	if rtt.ack {
		rttVal = rtt.ts
	} else {
		rttVal = rtt.prvRtt
	}

	if rttVal < 0 {
		rttVal = 0
	}

	return rttVal
}

type TCPLog struct {
	direction conndirection
	reuseConn bool
	rstPkt    bool

	rtt RTT

	flowTracker *tcpFlowTracker

	tcpState tcpStatus

	// common info
	//

	synfinTS [4]int64

	synSeq, synAckSeq uint32

	l7proto        L7Proto
	RetransmitsSYN int16

	// win scale
	txWinScale int // https://www.rfc-editor.org/rfc/rfc7323.html#section-2.1
	rxWinScale int

	metric TCPMetrics
	//

	chunkID int64
	chunk   []*PktChunk
}

func (tcpl *TCPLog) classifyPacket(txRx int8, cnt []byte, ln *PktTCPHdr) streamPushResult {
	if tcpl.flowTracker == nil {
		tcpl.flowTracker = newTCPSeqTracker(8, 64*1024)
	}

	// Pure ACK packets don't contribute payload bytes to the stream reassembler.
	// We still keep TCP state transitions below, but they are not treated as
	// retransmissions on the hot path here.
	if len(cnt) == 0 && !ln.Flags.HasFlag(TCPSYN) && !ln.Flags.HasFlag(TCPFIN) && !ln.Flags.HasFlag(TCPRST) {
		return streamPushResult{}
	}

	return tcpl.flowTracker.Push(txRx, ln.Seq, ln.Flags, cnt, ln.TS)
}

func (tcpl *TCPLog) GetPktChunk(nxt bool, forceNew bool) *PktChunk {
	if len(tcpl.chunk) == 0 {
		tcpl.chunkID++
		c := &PktChunk{
			ChunkID:      tcpl.chunkID,
			messageDirty: true,
		}
		tcpl.chunk = append(tcpl.chunk, c)
		tcpl.rtt.toNext()
		return c
	}

	c := tcpl.chunk[len(tcpl.chunk)-1]
	diff := len(c.TCPSreries) - tcpPktLimitPerChunk
	if forceNew || (nxt && diff >= 0 &&
		(!isFINChunk(c.chunkKind) || diff > 64)) {
		tcpl.chunkID++
		c = &PktChunk{
			ChunkID:      tcpl.chunkID,
			messageDirty: true,
		}
		tcpl.chunk = append(tcpl.chunk, c)
		tcpl.rtt.toNext()
	}

	return c
}

func calSeqOffset(seq, ack, seqPos, ackPos uint32, seqPosFlag, ackPosFlag bool, tcpFlag TCPFlag) (
	uint32, uint32, uint32, uint32, bool, bool,
) {
	// 注意，RST 包上可能带有 ack seq，不能根据 ACK flag 更新 ack seq offset
	// 我们认为对于未 ack 对等端发来的数据包所造成当前 RST 包含有 ack seq 时
	// 而发生（溢出）回绕问题，不进行特殊处理可以减少逻辑复杂性
	//
	// 避开 SYN 包对 ack 的更新
	if !ackPosFlag && tcpFlag != TCPSYN {
		ackPosFlag = true
		ackPos = ack - 1
	}

	if !seqPosFlag {
		seqPosFlag = true
		seqPos = seq - 1
	}

	seq -= seqPos
	if ackPosFlag {
		ack -= ackPos
	}

	return seq, ack, seqPos, ackPos, seqPosFlag, ackPosFlag
}

func (tcpl *TCPLog) Handle(txRx int8, cnt []byte, cntLen int64, ln *PktTCPHdr, k *PMeta, scale int) (pktState int8) {
	trackerRes := tcpl.classifyPacket(txRx, cnt, ln)
	chunk := tcpl.GetPktChunk(true, trackerRes.Gap)
	chunk.markMessageDirty()
	if enableNetlog {
		lnCpy := *ln
		switch txRx {
		case directionRX:
			lnCpy.DstMAC = chunk.GetMacID(lnCpy.DstMAC)
			lnCpy.SrcMAC = chunk.GetMacID(lnCpy.SrcMAC)

			lnCpy.Seq, lnCpy.AckSeq, chunk.RxSeqPos, chunk.TxSeqPos,
				chunk.setFlag[1], chunk.setFlag[0] = calSeqOffset(
				lnCpy.Seq, lnCpy.AckSeq, chunk.RxSeqPos, chunk.TxSeqPos,
				chunk.setFlag[1], chunk.setFlag[0], lnCpy.Flags)
		case directionTX:
			lnCpy.SrcMAC = chunk.GetMacID(lnCpy.SrcMAC)
			lnCpy.DstMAC = chunk.GetMacID(lnCpy.DstMAC)

			lnCpy.Seq, lnCpy.AckSeq, chunk.TxSeqPos, chunk.RxSeqPos,
				chunk.setFlag[0], chunk.setFlag[1] = calSeqOffset(
				lnCpy.Seq, lnCpy.AckSeq, chunk.TxSeqPos, chunk.RxSeqPos,
				chunk.setFlag[0], chunk.setFlag[1], lnCpy.Flags)
		}
		if chunk.TimePos == 0 {
			// chunk.TimePos = lnCpy.TS - 1
			chunk.TimePos = lnCpy.TS
		}
		lnCpy.TS -= chunk.TimePos
		chunk.TCPSreries = append(chunk.TCPSreries, lnCpy)
	}

	if trackerRes.Retransmit {
		pktState = 1
	}

	if pktState == 1 {
		if txRx == directionRX {
			chunk.RetransmitsRx++
			tcpl.metric.Retransmits++
		} else if txRx == directionTX {
			chunk.RetransmitsTx++
			tcpl.metric.Retransmits++

			// 不含 payload 的不计入重传
			if cntLen != 0 {
				tcpl.metric.Retransmits++
			}
		}
	} else {
		if txRx == directionRX {
			tcpl.metric.BytesRead += int(cntLen)
		} else if txRx == directionTX {
			tcpl.metric.BytesWritten += int(cntLen)
		}
	}

	switch txRx {
	case directionTX:
		if ln.Flags.HasFlag(TCPSYN | TCPFIN) {
			tcpl.rtt.cal(txRx, ln.Seq+1+uint32(cntLen), ln.TS)
		} else {
			tcpl.rtt.cal(txRx, ln.Seq+uint32(cntLen), ln.TS)
		}

		chunk.recSeqRange(ln.Seq, ln.AckSeq, true, ln.Flags)
		chunk.TXPacket++
		if cntLen > 0 {
			chunk.TxBytes += int(cntLen)
		}
	case directionRX:
		tcpl.rtt.cal(txRx, ln.AckSeq, ln.TS)
		chunk.recSeqRange(ln.Seq, ln.AckSeq, false, ln.Flags)
		chunk.RXPacket++
		if cntLen > 0 {
			chunk.RxBytes += int(cntLen)
		}
	}

	if ln.Flags.HasFlag(TCPSYN) {
		chunk.chunkKind |= chunkKindSYN

		if scale > 0 {
			switch txRx {
			case directionTX:
				tcpl.txWinScale = scale
			case directionRX:
				tcpl.rxWinScale = scale
			}
		}

		if ln.Flags.HasFlag(TCPACK) {
			tcpl.synAckSeq = ln.Seq
			tcpl.synfinTS[1] = ln.TS
			if tcpl.synfinTS[0] == 0 {
				tcpl.synfinTS[0] = ln.TS
			}

			if txRx == directionTX && tcpl.tcpState == TCPSYNRcvd {
				tcpl.RetransmitsSYN++
			}

			tcpl.tcpState = TCPSYNRcvd
		} else {
			tcpl.synSeq = ln.Seq
			tcpl.synfinTS[0] = ln.TS
			if txRx == directionTX && tcpl.tcpState == TCPSYNSend {
				tcpl.RetransmitsSYN++
			}
			tcpl.tcpState = TCPSYNSend
		}
		return
	}

	// scale win
	switch txRx {
	case directionRX:
		if tcpl.rxWinScale > 0 {
			ln.Win <<= tcpl.rxWinScale
		}
	case directionTX:
		if tcpl.txWinScale > 0 {
			ln.Win <<= tcpl.txWinScale
		}
	}

	if tcpl.tcpState == TCPSYNRcvd {
		if ln.Flags.HasFlag(TCPACK) {
			tcpl.synfinTS[1] = ln.TS
			tcpl.tcpState = TCPEstablished
		}
		return
	}

	if ln.Flags.HasFlag(TCPRST) { // maybe after 4whs
		chunk.chunkKind |= chunkKindFINRST
		tcpl.rstPkt = true

		if tcpl.synfinTS[3] == 0 {
			tcpl.synfinTS[3] = ln.TS
		}

		switch txRx {
		case directionRX:
			chunk.RSTRx++
		case directionTX:
			chunk.RSTTx++
		}

		tcpl.tcpState = TCPClose
		return
	}

	// maybe start after tcp 3whs
	if tcpl.tcpState == TCPUnknownStatus {
		tcpl.tcpState = TCPEstablished
	}

	tcpstatus := tcpl.tcpState

	if ln.Flags.HasFlag(TCPFIN) {
		chunk.chunkKind |= chunkKindFINRST

		if tcpstatus == TCPEstablished {
			tcpl.synfinTS[2] = ln.TS
			tcpl.tcpState = TCPFINWait1
			goto fin
		}

		if tcpstatus == TCPFINWait1 {
			tcpl.synfinTS[3] = ln.TS
			tcpl.tcpState = TCPLastAck
			goto fin
		}

		if tcpstatus == TCPFINWait2 || tcpstatus == TCPCloseWait { // fin; ack; fin ^; ack
			tcpl.synfinTS[3] = ln.TS
			tcpl.tcpState = TCPLastAck
			goto fin
		}

	fin:
		return
	}

	if ln.Flags.HasFlag(TCPACK) {
		switch tcpstatus { //nolint:exhaustive
		case TCPFINWait1:
			tcpl.synfinTS[3] = ln.TS
			tcpl.tcpState = TCPFINWait2 // or close_wait
		case TCPLastAck:
			tcpl.synfinTS[3] = ln.TS
			tcpl.tcpState = TCPTimeWait
		}
	}
	return pktState
}

func (tcpl *TCPLog) Closed() bool {
	return tcpl.tcpState == TCPClose || tcpl.tcpState == TCPTimeWait
}

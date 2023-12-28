//go:build linux
// +build linux

package l4log

import (
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

type tcpSortElem struct {
	overflow bool
	idx      int64

	txRx int8

	seq uint32
	len uint32

	ackSeq uint32

	// todo: control flag
	// ack int8
	// fin int8
	// syn int8
	// rst int8
}

// const (
// 	_tcpelemLimit = 32
// 	_tcpelemCap   = _tcpelemLimit * 2
// )

type tcpRetransAndReorder struct {
	// todo: resort

	tcpStatus tcpStatus

	txPkts []*tcpSortElem // 记录少量数据包

	keepalive bool

	rxPkts []*tcpSortElem // 记录少量数据包
}

const (
	maxPktRecForRetransAndResort = 128
)

func (tcpr *tcpRetransAndReorder) _insert(elem *tcpSortElem, idx int) {
	var txrxPkts []*tcpSortElem

	switch elem.txRx {
	case directionTX:
		txrxPkts = tcpr.txPkts
	case directionRX:
		txrxPkts = tcpr.rxPkts
	default:
		return
	}

	curIdx := len(txrxPkts) - 1
	if idx > curIdx || idx < 0 {
		txrxPkts = append(txrxPkts, elem)
	} else {
		txrxPkts = append(txrxPkts, nil)
		copy(txrxPkts[idx+1:], txrxPkts[idx:])
		txrxPkts[idx] = elem
	}
	if len(txrxPkts) >= maxPktRecForRetransAndResort {
		tmp := make([]*tcpSortElem, 0, maxPktRecForRetransAndResort)
		txrxPkts = append(tmp, txrxPkts[maxPktRecForRetransAndResort/2:]...)
	}

	switch elem.txRx {
	case directionTX:
		tcpr.txPkts = txrxPkts
	case directionRX:
		tcpr.rxPkts = txrxPkts
	default:
		return
	}
}

func (tcpr *tcpRetransAndReorder) insert(elem *tcpSortElem) (ret int8) {
	// ret 0: ok, 1: retrans, 2: keepalive ;; todo: 3: need resort, will cache data

	var txrxPkts []*tcpSortElem

	switch elem.txRx {
	case directionTX:
		txrxPkts = tcpr.txPkts
	case directionRX:
		txrxPkts = tcpr.rxPkts
	default:
		return 0
	}

	overflowIdx := -1
	for i, v := range txrxPkts {
		if v != nil {
			if v.overflow {
				overflowIdx = i
			}
		}
	}

	if elem.seq+elem.len < elem.seq {
		elem.overflow = true
		if overflowIdx < 0 { // 之前未发生 seq 回绕
			tcpr._insert(elem, len(txrxPkts))
			return
		}
	}

	for i := len(txrxPkts) - 1; i >= 0; i-- {
		cachedElem := txrxPkts[i]
		if cachedElem == nil {
			continue
		}

		if tcpr.keepalive && cachedElem.seq+cachedElem.len == elem.seq &&
			elem.ackSeq == cachedElem.ackSeq {
			tcpr.keepalive = false
			if elem.len == 0 {
				ret = 2
				return
			} else {
				tcpr._insert(elem, i+1)
				return 0
			}
		}

		if elem.seq == cachedElem.seq &&
			elem.ackSeq == cachedElem.ackSeq &&
			elem.len == cachedElem.len {
			// retrans
			tcpr._insert(elem, i+1) // 追加
			return 1
		}

		if cachedElem.seq+cachedElem.len == elem.seq+1 &&
			elem.ackSeq == cachedElem.ackSeq {
			tcpr.keepalive = true
			return 2
		}

		if overflowIdx >= 0 && i >= overflowIdx { // 发生了回绕的记录；乱序将产生干扰，如果未识别到回绕
			curSeq := elem.seq
			if elem.overflow {
				curSeq = elem.seq + elem.len
			}

			if i != overflowIdx {
				if curSeq >= cachedElem.seq+cachedElem.len && (curSeq < txrxPkts[0].seq || overflowIdx == 0) {
					if cachedElem.ackSeq <= elem.ackSeq {
						tcpr._insert(elem, i+1) // 找到上一个小的seq，往后追加
						return 0
					}
				}
			} else {
				if curSeq == cachedElem.seq+cachedElem.len {
					if cachedElem.ackSeq <= elem.ackSeq {
						tcpr._insert(elem, i+1) // 找到上一个小的seq，往后追加
					} else {
						tcpr._insert(elem, i)
					}
					return 0
				}
			}
		} else if cachedElem.seq+cachedElem.len <= elem.seq {
			if cachedElem.ackSeq <= elem.ackSeq {
				tcpr._insert(elem, i+1) // 找到上一个小的seq，往后追加
				return 0
			}
		}
	}

	tcpr._insert(elem, 0)
	return 0
}

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

	TCPColName []string     `json:"tcp_series_col_name"`
	TCPSreries []*PktTCPHdr `json:"tcp_series"`

	RxBytes int `json:"rx_bytes"`
	TxBytes int `json:"tx_bytes"`

	RXPacket int64 `json:"rx_packets"`
	TXPacket int64 `json:"tx_packets"`

	TxFirstByteTS int64 `json:"tx_first_byte_ts"`
	RxFirstByteTS int64 `json:"rx_first_byte_ts"`

	TxLastByteTS int64 `json:"tx_last_byte_ts"`
	RxLastByteTS int64 `json:"rx_last_byte_ts"`

	// SPacket, DPacket int
	RetransmitsTx int `json:"tx_retrans"`
	RetransmitsRx int `json:"rx_retrans"`

	RSTTx int `json:"tx_rst"`
	RSTRx int `json:"rx_rst"`

	RetransmitsSYN int `json:"tcp_syn_retrans"`
}

func (chunk *PktChunk) recSeqRange(seq, ack uint32, tx bool, tcpflag TCPFlag) {
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

type TCPLog struct {
	reuseConn bool
	rstPkt    bool

	// 只能通过重复的 tx 数据反推重传
	tcpStatusRec tcpRetransAndReorder // seq, ack

	// common info
	//

	direction   conndirection
	synfinTS    [4]int64
	rttTxseqAck [2]uint32
	rttTxTS     int64

	synSeq, synAckSeq uint32

	// l7proto L7Proto

	tcpRTT float64

	// win scale
	txWinScale int // https://www.rfc-editor.org/rfc/rfc7323.html#section-2.1
	rxWinScale int

	metric TCPMetrics
	//

	chunkID int64
	chunk   []*PktChunk
}

func (tcpl *TCPLog) GetPktChunk(nxt bool) *PktChunk {
	if len(tcpl.chunk) == 0 {
		tcpl.chunkID++
		c := &PktChunk{
			ChunkID: tcpl.chunkID,
		}
		tcpl.chunk = append(tcpl.chunk, c)
		return c
	}

	c := tcpl.chunk[len(tcpl.chunk)-1]
	diff := len(c.TCPSreries) - tcpPktLimitPerChunk
	if nxt && diff >= 0 &&
		(!isFINChunk(c.chunkKind) || diff > 64) {
		tcpl.chunkID++
		c = &PktChunk{
			ChunkID: tcpl.chunkID,
		}
		tcpl.chunk = append(tcpl.chunk, c)
	}

	return c
}

func (tcpl *TCPLog) Handle(txRx int8, cnt []byte, cntLen int64, ln *PktTCPHdr, k *PMeta, scale int) (pktState int8) {
	chunk := tcpl.GetPktChunk(true)
	if enableNetlog {
		chunk.TCPSreries = append(chunk.TCPSreries, ln)
	}

	elem := &tcpSortElem{
		seq:    ln.Seq,
		ackSeq: ln.AckSeq,
		txRx:   txRx,
	}

	elem.len = uint32(cntLen)

	pktState = tcpl.tcpStatusRec.insert(elem)

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
		if tcpl.tcpRTT == 0 && ln.Seq > tcpl.rttTxseqAck[0] {
			tcpl.rttTxTS = ln.TS
			tcpl.rttTxseqAck[0] = ln.Seq
			if ln.Flags.HasFlag(TCPFIN | TCPSYN) {
				tcpl.rttTxseqAck[1] = ln.Seq + 1 + uint32(cntLen)
			} else {
				tcpl.rttTxseqAck[1] = ln.Seq + uint32(cntLen)
			}
		}
		chunk.recSeqRange(ln.Seq, ln.AckSeq, true, ln.Flags)
		chunk.TXPacket++
		if cntLen > 0 {
			chunk.TxBytes += int(cntLen)
			if chunk.TxFirstByteTS == 0 {
				chunk.TxFirstByteTS = ln.TS
			}
			chunk.TxLastByteTS = ln.TS
		}
	case directionRX:
		if tcpl.tcpRTT == 0 && ln.AckSeq != 0 && ln.AckSeq == tcpl.rttTxseqAck[1] {
			tcpl.tcpRTT = float64(ln.TS-tcpl.rttTxTS) / float64(time.Millisecond)
			tcpl.metric.RTT = (ln.TS - tcpl.rttTxTS) / int64(time.Microsecond)
		}
		chunk.recSeqRange(ln.Seq, ln.AckSeq, false, ln.Flags)
		chunk.RXPacket++
		if cntLen > 0 {
			chunk.RxBytes += int(cntLen)
			if chunk.RxFirstByteTS == 0 {
				chunk.RxFirstByteTS = ln.TS
			}
			chunk.RxLastByteTS = ln.TS
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

			if txRx == directionTX && tcpl.tcpStatusRec.tcpStatus == TCPSYNRcvd {
				chunk.RetransmitsSYN++
			}

			tcpl.tcpStatusRec.tcpStatus = TCPSYNRcvd
		} else {
			tcpl.synSeq = ln.Seq
			tcpl.synfinTS[0] = ln.TS
			if txRx == directionTX && tcpl.tcpStatusRec.tcpStatus == TCPSYNSend {
				chunk.RetransmitsSYN++
			}
			tcpl.tcpStatusRec.tcpStatus = TCPSYNSend
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

	if tcpl.tcpStatusRec.tcpStatus == TCPSYNRcvd {
		if ln.Flags.HasFlag(TCPACK) {
			tcpl.synfinTS[1] = ln.TS
			tcpl.tcpStatusRec.tcpStatus = TCPEstablished
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

		tcpl.tcpStatusRec.tcpStatus = TCPClose
		return
	}

	// maybe start after tcp 3whs
	if tcpl.tcpStatusRec.tcpStatus == TCPUnknownStatus {
		tcpl.tcpStatusRec.tcpStatus = TCPEstablished
	}

	tcpstatus := tcpl.tcpStatusRec.tcpStatus

	if ln.Flags.HasFlag(TCPFIN) {
		chunk.chunkKind |= chunkKindFINRST

		if tcpstatus == TCPEstablished {
			tcpl.synfinTS[2] = ln.TS
			tcpl.tcpStatusRec.tcpStatus = TCPFINWait1
			goto fin
		}

		if tcpstatus == TCPFINWait1 {
			tcpl.synfinTS[3] = ln.TS
			tcpl.tcpStatusRec.tcpStatus = TCPLastAck
			goto fin
		}

		if tcpstatus == TCPFINWait2 || tcpstatus == TCPCloseWait { // fin; ack; fin ^; ack
			tcpl.synfinTS[3] = ln.TS
			tcpl.tcpStatusRec.tcpStatus = TCPLastAck
			goto fin
		}

	fin:
		return
	}

	if ln.Flags.HasFlag(TCPACK) {
		switch tcpstatus { //nolint:exhaustive
		case TCPFINWait1:
			tcpl.synfinTS[3] = ln.TS
			tcpl.tcpStatusRec.tcpStatus = TCPFINWait2 // or close_wait
		case TCPLastAck:
			tcpl.synfinTS[3] = ln.TS
			tcpl.tcpStatusRec.tcpStatus = TCPTimeWait
		}
	}
	return pktState
}

func (tcpl *TCPLog) Closed() bool {
	return tcpl.tcpStatusRec.tcpStatus == TCPClose || tcpl.tcpStatusRec.tcpStatus == TCPTimeWait
}

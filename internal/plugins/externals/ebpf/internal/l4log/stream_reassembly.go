//go:build linux
// +build linux

package l4log

type streamDelivery struct {
	Seq     uint32
	Flags   TCPFlag
	TS      int64
	Payload []byte
}

type streamPushResult struct {
	Deliveries []streamDelivery
	Buffered   bool
	Retransmit bool
	Gap        bool
}

type streamSegment struct {
	seq     uint32
	dataSeq uint32
	endSeq  uint32
	flags   TCPFlag
	ts      int64
	payload []byte
}

type streamReassembler struct {
	maxBufferedSegments int
	maxReorderWindow    uint32
	capturePayload      bool

	initialized bool
	nextSeq     uint32
	buffered    []streamSegment
}

func newStreamReassembler(maxBuffered int, maxWindow uint32, capturePayload bool) *streamReassembler {
	if maxBuffered <= 0 {
		maxBuffered = 8
	}
	if maxWindow == 0 {
		maxWindow = 64 * 1024
	}

	return &streamReassembler{
		maxBufferedSegments: maxBuffered,
		maxReorderWindow:    maxWindow,
		capturePayload:      capturePayload,
	}
}

func (sr *streamReassembler) resetTo(seg streamSegment) streamPushResult {
	sr.initialized = true
	sr.nextSeq = seg.endSeq
	sr.buffered = sr.buffered[:0]

	res := streamPushResult{Gap: true}
	if sr.capturePayload && len(seg.payload) > 0 {
		res.Deliveries = append(res.Deliveries, streamDelivery{
			Seq:     seg.dataSeq,
			Flags:   seg.flags,
			TS:      seg.ts,
			Payload: seg.payload,
		})
	}

	return res
}

func (sr *streamReassembler) Push(seq uint32, flags TCPFlag, payload []byte, ts int64) streamPushResult {
	seg := newStreamSegment(seq, flags, payload, ts)
	if seg.endSeq == seg.seq && len(seg.payload) == 0 {
		return streamPushResult{}
	}

	if !sr.initialized {
		sr.initialized = true
		sr.nextSeq = seg.endSeq

		var res streamPushResult
		if sr.capturePayload && len(seg.payload) > 0 {
			res.Deliveries = append(res.Deliveries, streamDelivery{
				Seq:     seg.dataSeq,
				Flags:   seg.flags,
				TS:      seg.ts,
				Payload: seg.payload,
			})
		}
		return res
	}

	if seg.seq > sr.nextSeq {
		if seg.seq-sr.nextSeq > sr.maxReorderWindow || len(sr.buffered) >= sr.maxBufferedSegments {
			return sr.resetTo(seg)
		}
		if inserted := sr.insertBuffered(bufferedStreamSegment(seq, flags, payload, ts, sr.capturePayload)); !inserted {
			return streamPushResult{Retransmit: true}
		}
		return streamPushResult{Buffered: true}
	}

	return sr.consume(seg)
}

func (sr *streamReassembler) consume(seg streamSegment) streamPushResult {
	var res streamPushResult

	if seg.endSeq <= sr.nextSeq {
		res.Retransmit = true
		return res
	}

	if sr.capturePayload && len(seg.payload) > 0 {
		if tail, ok := trimPayloadToNextSeq(seg, sr.nextSeq); ok {
			res.Deliveries = append(res.Deliveries, streamDelivery{
				Seq:     sr.nextSeq,
				Flags:   seg.flags,
				TS:      seg.ts,
				Payload: tail,
			})
		}
	}

	sr.nextSeq = seg.endSeq
	sr.drain(&res)
	return res
}

func (sr *streamReassembler) drain(res *streamPushResult) {
	for len(sr.buffered) > 0 {
		seg := sr.buffered[0]
		if seg.seq > sr.nextSeq {
			return
		}

		sr.buffered = sr.buffered[1:]
		if seg.endSeq <= sr.nextSeq {
			res.Retransmit = true
			continue
		}

		if sr.capturePayload && len(seg.payload) > 0 {
			if tail, ok := trimPayloadToNextSeq(seg, sr.nextSeq); ok {
				res.Deliveries = append(res.Deliveries, streamDelivery{
					Seq:     sr.nextSeq,
					Flags:   seg.flags,
					TS:      seg.ts,
					Payload: tail,
				})
			}
		}

		sr.nextSeq = seg.endSeq
	}
}

func (sr *streamReassembler) insertBuffered(seg streamSegment) bool {
	for i := range sr.buffered {
		cur := sr.buffered[i]
		if seg.seq == cur.seq && seg.endSeq == cur.endSeq && seg.dataSeq == cur.dataSeq {
			return false
		}
		if seg.seq < cur.seq {
			sr.buffered = append(sr.buffered, streamSegment{})
			copy(sr.buffered[i+1:], sr.buffered[i:])
			sr.buffered[i] = seg
			return true
		}
	}

	sr.buffered = append(sr.buffered, seg)
	return true
}

func newStreamSegment(seq uint32, flags TCPFlag, payload []byte, ts int64) streamSegment {
	dataSeq := seq
	if flags.HasFlag(TCPSYN) {
		dataSeq++
	}

	endSeq := dataSeq + uint32(len(payload))
	if flags.HasFlag(TCPFIN) {
		endSeq++
	}

	return streamSegment{
		seq:     seq,
		dataSeq: dataSeq,
		endSeq:  endSeq,
		flags:   flags,
		ts:      ts,
		payload: payload,
	}
}

func trimPayloadToNextSeq(seg streamSegment, nextSeq uint32) ([]byte, bool) {
	if len(seg.payload) == 0 {
		return nil, false
	}

	if nextSeq <= seg.dataSeq {
		return seg.payload, true
	}
	if nextSeq >= seg.dataSeq+uint32(len(seg.payload)) {
		return nil, false
	}

	offset := int(nextSeq - seg.dataSeq)
	return seg.payload[offset:], true
}

func clonePayload(payload []byte) []byte {
	if len(payload) == 0 {
		return nil
	}

	buf := make([]byte, len(payload))
	copy(buf, payload)
	return buf
}

func bufferedStreamSegment(seq uint32, flags TCPFlag, payload []byte, ts int64, capturePayload bool) streamSegment {
	seg := newStreamSegment(seq, flags, payload, ts)
	if capturePayload {
		seg.payload = clonePayload(payload)
	} else {
		seg.payload = nil
	}
	return seg
}

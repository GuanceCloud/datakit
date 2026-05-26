//go:build linux
// +build linux

package l4log

type tcpFlowTracker struct {
	tx tcpDirectionTracker
	rx tcpDirectionTracker
}

type tcpDirectionTracker struct {
	reasm *streamReassembler
}

func newTCPFlowTracker(maxBuffered int, maxWindow uint32) *tcpFlowTracker {
	return &tcpFlowTracker{
		tx: tcpDirectionTracker{reasm: newStreamReassembler(maxBuffered, maxWindow, true)},
		rx: tcpDirectionTracker{reasm: newStreamReassembler(maxBuffered, maxWindow, true)},
	}
}

func newTCPSeqTracker(maxBuffered int, maxWindow uint32) *tcpFlowTracker {
	return &tcpFlowTracker{
		tx: tcpDirectionTracker{reasm: newStreamReassembler(maxBuffered, maxWindow, false)},
		rx: tcpDirectionTracker{reasm: newStreamReassembler(maxBuffered, maxWindow, false)},
	}
}

func (ft *tcpFlowTracker) direction(txrx int8) *tcpDirectionTracker {
	switch txrx {
	case directionTX:
		return &ft.tx
	case directionRX:
		return &ft.rx
	default:
		return nil
	}
}

func (ft *tcpFlowTracker) Push(txrx int8, seq uint32, flags TCPFlag, payload []byte, ts int64) streamPushResult {
	dir := ft.direction(txrx)
	if dir == nil || dir.reasm == nil {
		return streamPushResult{}
	}

	return dir.reasm.Push(seq, flags, payload, ts)
}

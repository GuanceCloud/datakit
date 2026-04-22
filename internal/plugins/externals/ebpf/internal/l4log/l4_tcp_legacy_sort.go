//go:build linux
// +build linux

package l4log

type tcpSortElem struct {
	overflow bool
	idx      int64

	txRx int8

	seq uint32
	len uint32

	ackSeq uint32
}

type tcpRetransAndReorder struct {
	keepalive bool

	txPkts []tcpSortElem
	txOFI  int

	rxPkts []tcpSortElem
	rxOFI  int
}

const (
	maxPktRecForRetransAndResort = 128
	noOverflowIdx                = -1
)

func (tcpr *tcpRetransAndReorder) packets(txrx int8) ([]tcpSortElem, int, bool) {
	switch txrx {
	case directionTX:
		return tcpr.txPkts, normalizeOverflowIdx(tcpr.txPkts, tcpr.txOFI), true
	case directionRX:
		return tcpr.rxPkts, normalizeOverflowIdx(tcpr.rxPkts, tcpr.rxOFI), true
	default:
		return nil, noOverflowIdx, false
	}
}

func normalizeOverflowIdx(pkts []tcpSortElem, overflowIdx int) int {
	if overflowIdx < 0 || overflowIdx >= len(pkts) {
		return noOverflowIdx
	}
	if !pkts[overflowIdx].overflow {
		return noOverflowIdx
	}

	return overflowIdx
}

func (tcpr *tcpRetransAndReorder) setPackets(txrx int8, pkts []tcpSortElem, overflowIdx int) {
	switch txrx {
	case directionTX:
		tcpr.txPkts = pkts
		tcpr.txOFI = overflowIdx
	case directionRX:
		tcpr.rxPkts = pkts
		tcpr.rxOFI = overflowIdx
	}
}

func recomputeOverflowIdx(pkts []tcpSortElem) int {
	for i, v := range pkts {
		if v.overflow {
			return i
		}
	}

	return noOverflowIdx
}

func (tcpr *tcpRetransAndReorder) _insert(elem *tcpSortElem, idx int) {
	txrxPkts, overflowIdx, ok := tcpr.packets(elem.txRx)
	if !ok {
		return
	}
	if txrxPkts == nil {
		txrxPkts = make([]tcpSortElem, 0, maxPktRecForRetransAndResort)
	}

	curIdx := len(txrxPkts) - 1
	if idx > curIdx || idx < 0 {
		txrxPkts = append(txrxPkts, *elem)
		if elem.overflow && overflowIdx == noOverflowIdx {
			overflowIdx = len(txrxPkts) - 1
		}
	} else {
		txrxPkts = append(txrxPkts, tcpSortElem{})
		copy(txrxPkts[idx+1:], txrxPkts[idx:])
		txrxPkts[idx] = *elem
		if overflowIdx >= idx {
			overflowIdx++
		}
		if elem.overflow && overflowIdx == noOverflowIdx {
			overflowIdx = idx
		}
	}
	if len(txrxPkts) >= maxPktRecForRetransAndResort {
		tmp := make([]tcpSortElem, 0, maxPktRecForRetransAndResort)
		txrxPkts = append(tmp, txrxPkts[maxPktRecForRetransAndResort/2:]...)
		overflowIdx = recomputeOverflowIdx(txrxPkts)
	}

	tcpr.setPackets(elem.txRx, txrxPkts, overflowIdx)
}

func (tcpr *tcpRetransAndReorder) insert(elem *tcpSortElem) (ret int8) {
	txrxPkts, overflowIdx, ok := tcpr.packets(elem.txRx)
	if !ok {
		return 0
	}

	if elem.seq+elem.len < elem.seq {
		elem.overflow = true
		if overflowIdx == noOverflowIdx {
			tcpr._insert(elem, len(txrxPkts))
			return
		}
	}

	if !tcpr.keepalive && len(txrxPkts) > 0 {
		last := txrxPkts[len(txrxPkts)-1]
		if overflowIdx == noOverflowIdx {
			switch {
			case elem.seq == last.seq && elem.ackSeq == last.ackSeq && elem.len == last.len:
				tcpr._insert(elem, len(txrxPkts))
				return 1
			case last.seq+last.len == elem.seq+1 && elem.ackSeq == last.ackSeq:
				tcpr.keepalive = true
				return 2
			case last.seq+last.len <= elem.seq && last.ackSeq <= elem.ackSeq:
				tcpr._insert(elem, len(txrxPkts))
				return 0
			}
		}
	}

	for i := len(txrxPkts) - 1; i >= 0; i-- {
		cachedElem := txrxPkts[i]

		if tcpr.keepalive && cachedElem.seq+cachedElem.len == elem.seq &&
			elem.ackSeq == cachedElem.ackSeq {
			tcpr.keepalive = false
			if elem.len == 0 {
				ret = 2
				return
			}
			tcpr._insert(elem, i+1)
			return 0
		}

		if elem.seq == cachedElem.seq &&
			elem.ackSeq == cachedElem.ackSeq &&
			elem.len == cachedElem.len {
			tcpr._insert(elem, i+1)
			return 1
		}

		if cachedElem.seq+cachedElem.len == elem.seq+1 &&
			elem.ackSeq == cachedElem.ackSeq {
			tcpr.keepalive = true
			return 2
		}

		if overflowIdx != noOverflowIdx && i >= overflowIdx {
			curSeq := elem.seq
			if elem.overflow {
				curSeq = elem.seq + elem.len
			}

			if i != overflowIdx {
				if curSeq >= cachedElem.seq+cachedElem.len && (curSeq < txrxPkts[0].seq || overflowIdx == 0) {
					if cachedElem.ackSeq <= elem.ackSeq {
						tcpr._insert(elem, i+1)
						return 0
					}
				}
			} else {
				if curSeq == cachedElem.seq+cachedElem.len {
					if cachedElem.ackSeq <= elem.ackSeq {
						tcpr._insert(elem, i+1)
					} else {
						tcpr._insert(elem, i)
					}
					return 0
				}
			}
		} else if cachedElem.seq+cachedElem.len <= elem.seq {
			if cachedElem.ackSeq <= elem.ackSeq {
				tcpr._insert(elem, i+1)
				return 0
			}
		}
	}

	tcpr._insert(elem, 0)
	return 0
}

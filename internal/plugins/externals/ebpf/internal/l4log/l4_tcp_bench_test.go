//go:build linux
// +build linux

package l4log

import "testing"

type tcpRetransAndReorderLegacy struct {
	keepalive bool
	txPkts    []*tcpSortElem
	rxPkts    []*tcpSortElem
}

func (tcpr *tcpRetransAndReorderLegacy) insertSlice(elem *tcpSortElem, idx int) {
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
	}
}

func (tcpr *tcpRetransAndReorderLegacy) insert(elem *tcpSortElem) (ret int8) {
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
		if v != nil && v.overflow {
			overflowIdx = i
		}
	}

	if elem.seq+elem.len < elem.seq {
		elem.overflow = true
		if overflowIdx < 0 {
			tcpr.insertSlice(elem, len(txrxPkts))
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
			}
			tcpr.insertSlice(elem, i+1)
			return 0
		}

		if elem.seq == cachedElem.seq &&
			elem.ackSeq == cachedElem.ackSeq &&
			elem.len == cachedElem.len {
			tcpr.insertSlice(elem, i+1)
			return 1
		}

		if cachedElem.seq+cachedElem.len == elem.seq+1 &&
			elem.ackSeq == cachedElem.ackSeq {
			tcpr.keepalive = true
			return 2
		}

		if overflowIdx >= 0 && i >= overflowIdx {
			curSeq := elem.seq
			if elem.overflow {
				curSeq = elem.seq + elem.len
			}

			if i != overflowIdx {
				if curSeq >= cachedElem.seq+cachedElem.len && (curSeq < txrxPkts[0].seq || overflowIdx == 0) {
					if cachedElem.ackSeq <= elem.ackSeq {
						tcpr.insertSlice(elem, i+1)
						return 0
					}
				}
			} else {
				if curSeq == cachedElem.seq+cachedElem.len {
					if cachedElem.ackSeq <= elem.ackSeq {
						tcpr.insertSlice(elem, i+1)
					} else {
						tcpr.insertSlice(elem, i)
					}
					return 0
				}
			}
		} else if cachedElem.seq+cachedElem.len <= elem.seq {
			if cachedElem.ackSeq <= elem.ackSeq {
				tcpr.insertSlice(elem, i+1)
				return 0
			}
		}
	}

	tcpr.insertSlice(elem, 0)
	return 0
}

func benchmarkTCPRetransInsert(b *testing.B, makeRec func() interface{ insert(*tcpSortElem) int8 }, reorder bool) {
	b.Helper()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rec := makeRec()

		for j := 0; j < maxPktRecForRetransAndResort; j++ {
			seq := uint32(j * 100)
			if reorder {
				seq = uint32((maxPktRecForRetransAndResort - j) * 100)
			}
			rec.insert(&tcpSortElem{
				txRx:   directionTX,
				idx:    int64(j),
				seq:    seq,
				len:    100,
				ackSeq: seq,
			})
		}
	}
}

func BenchmarkTCPRetransInsertLegacyOrdered(b *testing.B) {
	benchmarkTCPRetransInsert(b, func() interface{ insert(*tcpSortElem) int8 } {
		return &tcpRetransAndReorderLegacy{}
	}, false)
}

func BenchmarkTCPRetransInsertOrdered(b *testing.B) {
	benchmarkTCPRetransInsert(b, func() interface{ insert(*tcpSortElem) int8 } {
		return &tcpRetransAndReorder{txOFI: noOverflowIdx, rxOFI: noOverflowIdx}
	}, false)
}

func BenchmarkTCPRetransInsertLegacyReorder(b *testing.B) {
	benchmarkTCPRetransInsert(b, func() interface{ insert(*tcpSortElem) int8 } {
		return &tcpRetransAndReorderLegacy{}
	}, true)
}

func BenchmarkTCPRetransInsertReorder(b *testing.B) {
	benchmarkTCPRetransInsert(b, func() interface{ insert(*tcpSortElem) int8 } {
		return &tcpRetransAndReorder{txOFI: noOverflowIdx, rxOFI: noOverflowIdx}
	}, true)
}

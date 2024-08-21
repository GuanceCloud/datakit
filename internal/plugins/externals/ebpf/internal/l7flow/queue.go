//go:build linux
// +build linux

package l7flow

import (
	"math"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

const (
	queueWindow   = 36
	queuePopCount = 18
	queueWindowX6 = queueWindow * 6
)

func idxLess(l, r uint64) bool {
	// 可能发生回绕现象，预留窗口应与 buffer 长度相近
	if l > math.MaxUint64-(queueWindowX6) && r <= queueWindowX6 {
		return true
	}

	return l < r
}

type dataQueue struct {
	li []*comm.NetwrkData
	// 从 1 开始索引，如果值为 0，视为发生翻转
	prvDataPos uint64
}

func (q *dataQueue) insertSort(data *comm.NetwrkData) {
	for i := 0; i < len(q.li); i++ {
		if idxLess(data.Index, q.li[i].Index) {
			q.li = append(q.li, nil)
			copy(q.li[i+1:], q.li[i:])
			q.li[i] = data
			return
		}
	}

	q.li = append(q.li, data)
}

func (q *dataQueue) Queue(data *comm.NetwrkData) []*comm.NetwrkData {
	var val []*comm.NetwrkData
	if data == nil {
		return val
	}

	q.insertSort(data)

	if len(q.li) >= queueWindow {
		x := queuePopCount
		val = append(val, q.li[:x]...)

		copy(q.li, q.li[x:])
		q.li = q.li[:len(q.li)-x]

		q.prvDataPos = val[x-1].Index
	}

	var nxt int
	for i := 0; i < len(q.li); i++ {
		if q.li[i].Index == q.prvDataPos+1 {
			val = append(val, q.li[i])
			q.prvDataPos = q.li[i].Index
			nxt = i + 1
		}
	}

	if nxt > 0 {
		if nxt == len(q.li) {
			q.li = q.li[:0]
		} else {
			copy(q.li, q.li[nxt:])
			q.li = q.li[:len(q.li)-nxt]
		}
	}
	return val
}

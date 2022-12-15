//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

package httpflow

import (
	"sync"

	"github.com/shirou/gopsutil/host"
)

var _reqCache = NewReqCache()

type ReqCache struct {
	pathMap   map[CPayloadId]string
	finReqMap map[CPayloadId]*HTTPReqFinishedInfo

	pathMutex sync.Mutex
	reqMutex  sync.Mutex
}

func NewReqCache() *ReqCache {
	return &ReqCache{
		pathMap:   map[CPayloadId]string{},
		finReqMap: map[CPayloadId]*HTTPReqFinishedInfo{},
	}
}

func (cache *ReqCache) AppendPayload(buf *CL7Buffer) {
	if buf == nil {
		return
	}
	bufLen := int(buf.len)
	// l.Info("len ", bufLen)
	// l.Info(string(buf.payload[:bufLen]))

	if bufLen > len(buf.payload) {
		bufLen = len(buf.payload)
	}

	fistSpace := true
	var start, end int = -1, bufLen
	for i := 0; i < bufLen; i++ {
		if buf.payload[i] == ' ' {
			if fistSpace {
				fistSpace = false
				start = i + 1
			} else {
				end = i
				break
			}
		}
		if buf.payload[i] == '?' {
			end = i
			break
		}
	}

	if !(start > 0 && end > start && end <= bufLen) {
		return
	}

	reqPath := string(buf.payload[start:end])

	cache.pathMutex.Lock()
	defer cache.pathMutex.Unlock()

	cache.pathMap[CPayloadId(buf.id)] = reqPath

}

const connExpirationInterval = 6 * 3600 // 6 * 3600s

func reqExpr(uptimeS, tsNs uint64) bool {
	tsS := tsNs / 1000000000
	if uptimeS > tsS {
		if uptimeS-tsS > connExpirationInterval {
			return true
		}
	}
	return false
}

func (cache *ReqCache) AppendFinReq(id CPayloadId, finReq *HTTPReqFinishedInfo) {
	if finReq == nil {
		return
	}

	cache.reqMutex.Lock()
	defer cache.reqMutex.Unlock()

	cache.finReqMap[id] = finReq
}

func (cache *ReqCache) MergeReq() []*HTTPReqFinishedInfo {
	uptime, err := host.Uptime() // seconds since boot
	if err != nil {
		l.Error(err)
	}

	cache.pathMutex.Lock()
	defer cache.pathMutex.Unlock()

	cache.reqMutex.Lock()
	defer cache.reqMutex.Unlock()

	var finReqList []*HTTPReqFinishedInfo

	for id, finReq := range cache.finReqMap {
		if path, ok := cache.pathMap[id]; ok {
			finReq.HTTPStats.Path = path
			finReqList = append(finReqList, finReq)
			delete(cache.pathMap, id)
			delete(cache.finReqMap, id)
			continue
		}
		if reqExpr(uptime, uint64(id.ktime)) {
			delete(cache.finReqMap, id)
		}
	}

	return finReqList
}

func (cache *ReqCache) CleanPathExpr() {
	uptime, err := host.Uptime() // seconds since boot
	if err != nil {
		l.Error(err)
		return
	}
	cache.pathMutex.Lock()
	defer cache.pathMutex.Unlock()

	cache.reqMutex.Lock()
	defer cache.reqMutex.Unlock()

	for id := range cache.finReqMap {
		if reqExpr(uptime, uint64(id.ktime)) {
			delete(cache.finReqMap, id)
		}
	}

	for id := range cache.pathMap {
		if reqExpr(uptime, uint64(id.ktime)) {
			delete(cache.finReqMap, id)
		}
	}
}

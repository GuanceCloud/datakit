//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

package httpflow

import (
	"sync"

	"github.com/shirou/gopsutil/host"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/tracing"
)

var _reqCache = NewReqCache()

type ReqCache struct {
	pathMap   map[CPayloadID]*tracing.TraceInfo
	finReqMap map[CPayloadID]*HTTPReqFinishedInfo

	mutex sync.Mutex
}

func NewReqCache() *ReqCache {
	return &ReqCache{
		pathMap:   map[CPayloadID]*tracing.TraceInfo{},
		finReqMap: map[CPayloadID]*HTTPReqFinishedInfo{},
	}
}

func (cache *ReqCache) AppendPayload(payloadID CPayloadID, info *tracing.TraceInfo) {
	if info == nil {
		return
	}
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.pathMap[payloadID] = info
}

const connExpirationInterval = 61 // 61s

func reqExpr(uptimeS, tsNs uint64) bool {
	tsS := tsNs / 1e9
	l.Debugf("uptime: %d, reqTime: %d, uptime-reqTime: %d", uptimeS, tsS, uptimeS-tsS)
	if uptimeS > tsS {
		if uptimeS-tsS > connExpirationInterval {
			return true
		}
	}
	return false
}

func (cache *ReqCache) AppendFinReq(id CPayloadID, finReq *HTTPReqFinishedInfo) {
	if finReq == nil {
		return
	}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.finReqMap[id] = finReq
}

func (cache *ReqCache) MergeReq() []*HTTPReqFinishedInfo {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	var finReqList []*HTTPReqFinishedInfo
	for id, finReq := range cache.finReqMap {
		if traceInfo, ok := cache.pathMap[id]; ok && traceInfo != nil {
			finReq.HTTPStats.Path = traceInfo.Path
			finReqList = append(finReqList, finReq)
			delete(cache.pathMap, id)
			delete(cache.finReqMap, id)
		} else {
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
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	for id := range cache.finReqMap {
		if reqExpr(uptime, uint64(id.ktime)) {
			delete(cache.finReqMap, id)
		}
	}

	for id := range cache.pathMap {
		if reqExpr(uptime, uint64(id.ktime)) {
			delete(cache.pathMap, id)
		}
	}
}

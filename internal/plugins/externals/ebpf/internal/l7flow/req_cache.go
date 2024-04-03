//go:build linux
// +build linux

package l7flow

import (
	"sync"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/shirou/gopsutil/host"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/tracing"
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

const connExpirationInterval = 60 * 3 // 180s

func reqExpr(uptimeS, tsNs uint64) bool {
	tsS := tsNs / 1e9
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

func (cache *ReqCache) MergeReq(gtags map[string]string, etrace bool, procFilter *tracing.ProcessFilter) ([]*HTTPReqFinishedInfo, []*point.Point) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	var finReqList []*HTTPReqFinishedInfo
	var traceList []*point.Point
	for id, finReq := range cache.finReqMap {
		if traceInfo, ok := cache.pathMap[id]; ok {
			if traceInfo != nil {
				finReq.HTTPStats.Path = traceInfo.Path
				finReq.ConnInfo.ProcessName = traceInfo.ProcessName
				finReqList = append(finReqList, finReq)
				if etrace && traceInfo.AllowTrace && traceInfo.ProcessName != "datakit" &&
					traceInfo.ProcessName != "datakit-ebpf" {
					if pt, err := CreateTracePoint(gtags, traceInfo, finReq); err != nil {
						l.Warn(err)
					} else {
						traceList = append(traceList, pt)
					}
				}
			}
			delete(cache.pathMap, id)
			delete(cache.finReqMap, id)
		} else {
			l.Warn("finReq not found", " req_seq ", finReq.HTTPStats.ReqSeq,
				" resp_seq ", finReq.HTTPStats.RespSeq, " pid ", finReq.ConnInfo.Pid)
			delete(cache.finReqMap, id)
		}
	}

	return finReqList, traceList
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

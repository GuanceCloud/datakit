package aggregate

import (
	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

var _ Calculator = &algoCountLast{}

type algoCountLast struct {
	MetricBase
	last     float64
	lastTime int64
	count    int64
}

func (a *algoCountLast) Add(x any) {
	if inst, ok := x.(*algoCountLast); ok {
		a.count++
		if inst.lastTime > a.lastTime {
			a.last = inst.last
			a.lastTime = inst.lastTime
		}
	}
}

func (a *algoCountLast) Aggr() ([]*point.Point, error) {
	var kvs point.KVs

	kvs = kvs.Add(a.key, a.last).
		Add(a.key+"_count", a.count)

	for _, kv := range a.aggrTags {
		kvs = kvs.SetTag(kv[0], kv[1])
	}

	return []*point.Point{
		point.NewPoint(a.name, kvs, point.WithTimestamp(a.lastTime)),
	}, nil
}

func (a *algoCountLast) Reset() {
	a.last = 0
	a.lastTime = 0
	a.count = 0
}

func (a *algoCountLast) Base() *MetricBase {
	return &a.MetricBase
}

func (a *algoCountLast) doHash(h1 uint64) {
	h := HashCombine(h1, xxhash.Sum64([]byte("last")))
	h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(a.key)))
	a.MetricBase.hash = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(a.name)))
}

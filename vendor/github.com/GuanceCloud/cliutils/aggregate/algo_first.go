package aggregate

import (
	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

var _ Calculator = &algoCountFirst{}

type algoCountFirst struct {
	MetricBase
	first     float64
	firstTime int64
	count     int64
}

func (a *algoCountFirst) Add(x any) {
	if inst, ok := x.(*algoCountFirst); ok {
		a.count++
		if a.firstTime == 0 || inst.firstTime < a.firstTime {
			a.first = inst.first
			a.firstTime = inst.firstTime
		}
	}
}

func (a *algoCountFirst) Aggr() ([]*point.Point, error) {
	var kvs point.KVs

	kvs = kvs.Add(a.key, a.first).
		Add(a.key+"_count", a.count)

	for _, kv := range a.aggrTags {
		kvs = kvs.SetTag(kv[0], kv[1])
	}

	return []*point.Point{
		point.NewPoint(a.name, kvs, point.WithTimestamp(a.firstTime)),
	}, nil
}

func (a *algoCountFirst) Reset() {
	a.first = 0
	a.firstTime = 0
	a.count = 0
}

func (a *algoCountFirst) Base() *MetricBase {
	return &a.MetricBase
}

func (a *algoCountFirst) doHash(h1 uint64) {
	h := HashCombine(h1, xxhash.Sum64([]byte("first")))
	h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(a.key)))
	a.MetricBase.hash = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(a.name)))
}

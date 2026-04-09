package aggregate

import (
	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

type algoMin struct {
	MetricBase
	maxTime,
	count int64
	min float64
}

var _ Calculator = &algoMin{}

func (a *algoMin) Add(x any) {
	if inst, ok := x.(*algoMin); ok {
		a.count++
		if inst.min < a.min {
			a.min = inst.min
		}

		if inst.maxTime > a.maxTime {
			a.maxTime = inst.maxTime
		}
	}
}

func (a *algoMin) Aggr() ([]*point.Point, error) {
	var kvs point.KVs

	kvs = kvs.Add(a.key, a.min).
		Add(a.key+"_count", a.count)

	for _, kv := range a.aggrTags {
		// NOTE: if same-name tag key exist, apply the last one.
		kvs = kvs.SetTag(kv[0], kv[1])
	}

	return []*point.Point{
		point.NewPoint(a.name, kvs, point.WithTimestamp(a.maxTime)),
	}, nil
}

func (a *algoMin) Reset() {
	a.min = 0
	a.count = 0
	a.maxTime = 0
}

func (a *algoMin) Base() *MetricBase {
	return &a.MetricBase
}

func (a *algoMin) doHash(h1 uint64) {
	h := HashCombine(h1, xxhash.Sum64([]byte("min")))
	h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(a.key)))
	a.MetricBase.hash = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(a.name)))
}

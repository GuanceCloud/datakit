package aggregate

import (
	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

type algoCount struct {
	MetricBase
	maxTime, count int64
}

// type assertions.
var _ Calculator = &algoCount{}

func (c *algoCount) Add(x any) {
	if inst, ok := x.(*algoCount); ok {
		c.count++
		if inst.maxTime > c.maxTime {
			c.maxTime = inst.maxTime
		}
	}
}

func (c *algoCount) Aggr() ([]*point.Point, error) {
	var kvs point.KVs

	kvs = kvs.Add(c.key, c.count).
		Add(c.key+"_count", c.count)

	for _, kv := range c.aggrTags {
		// NOTE: if same-name tag key exist, apply the last one.
		kvs = kvs.SetTag(kv[0], kv[1])
	}

	return []*point.Point{
		point.NewPoint(c.name, kvs, point.WithTimestamp(c.maxTime)),
	}, nil
}

func (c *algoCount) Reset() {
	c.maxTime = 0
	c.count = 0
}

func (c *algoCount) doHash(h1 uint64) {
	h := HashCombine(h1, xxhash.Sum64([]byte("count")))
	h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(c.key)))
	c.MetricBase.hash = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(c.name)))
}

func (c *algoCount) Base() *MetricBase {
	return &c.MetricBase
}

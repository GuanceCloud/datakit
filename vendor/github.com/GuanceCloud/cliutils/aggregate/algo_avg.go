package aggregate

import (
	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

type algoAvg struct {
	MetricBase
	delta          float64
	maxTime, count int64
}

// type assertions.
var _ Calculator = &algoAvg{}

func (c *algoAvg) Add(x any) {
	if inst, ok := x.(*algoAvg); ok {
		c.count++
		c.delta += inst.delta

		if inst.maxTime > c.maxTime {
			c.maxTime = inst.maxTime
		}
	}
}

func (c *algoAvg) Aggr() ([]*point.Point, error) {
	var kvs point.KVs
	avg := c.delta / float64(c.count)
	kvs = kvs.Add(c.key, avg).
		Add(c.key+"_count", c.count)

	for _, kv := range c.aggrTags {
		// NOTE: if same-name tag key exist, apply the last one.
		kvs = kvs.SetTag(kv[0], kv[1])
	}

	return []*point.Point{
		point.NewPoint(c.name, kvs, point.WithTimestamp(c.maxTime)),
	}, nil
}

func (c *algoAvg) Reset() {
	c.delta = 0
	c.maxTime = 0
	c.count = 0
}

func (c *algoAvg) doHash(h1 uint64) {
	h := HashCombine(h1, xxhash.Sum64([]byte("avg")))
	h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(c.key)))
	c.MetricBase.hash = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(c.name)))
}

func (c *algoAvg) Base() *MetricBase {
	return &c.MetricBase
}

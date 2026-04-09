package aggregate

import (
	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

type algoHistogram struct {
	MetricBase
	count int64
	val   float64
	// tag:"le" 固定tag,存储："0":0,"10":1,"20":5
	// 使用le和指标的值 一一对应。
	leBucket map[string]float64
	maxTime  int64
}

// type assertions.
var _ Calculator = &algoHistogram{}

func (c *algoHistogram) Add(x any) {
	if inst, ok := x.(*algoHistogram); ok {
		c.count++

		if inst.pt != nil {
			kvs := point.KVs(inst.pt.Fields)
			if le := kvs.GetTag("le"); le != "" {
				c.leBucket[le] += inst.val
			}
		}
		if inst.maxTime > c.maxTime {
			c.maxTime = inst.maxTime
		}
	}
}

func (c *algoHistogram) Aggr() ([]*point.Point, error) {
	var pts []*point.Point
	// bucket
	for le, f := range c.leBucket {
		var kvs point.KVs
		kvs = kvs.AddTag("le", le).Add(c.key, f)

		for _, kv := range c.aggrTags {
			// NOTE: if same-name tag key exist, apply the last one.
			kvs = kvs.SetTag(kv[0], kv[1])
		}

		pts = append(pts, point.NewPoint(c.name, kvs, point.WithTimestamp(c.maxTime)))
	}

	// count
	var kvs point.KVs
	kvs = kvs.Add(c.key+"_count", c.count)
	for _, kv := range c.aggrTags {
		// NOTE: if same-name tag key exist, apply the last one.
		kvs = kvs.SetTag(kv[0], kv[1])
	}
	pts = append(pts, point.NewPoint(c.name, kvs, point.WithTimestamp(c.maxTime)))
	return pts, nil
}

func (c *algoHistogram) Reset() {
	c.leBucket = map[string]float64{}
	c.maxTime = 0
	c.count = 0
}

func (c *algoHistogram) doHash(h1 uint64) {
	h := HashCombine(h1, xxhash.Sum64([]byte("histogram")))
	h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(c.key)))
	c.MetricBase.hash = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(c.name)))
}

func (c *algoHistogram) Base() *MetricBase {
	return &c.MetricBase
}

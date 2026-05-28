package aggregate

import (
	"fmt"
	"hash/fnv"
	"math"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

const (
	countDistinctExactLimit = 4096
	countDistinctSketchBits = 1 << 18
)

type algoCountDistinct struct {
	MetricBase
	maxTime int64
	// Keep exact hashes until cardinality grows too high, then switch to
	// a fixed-size bitmap sketch to bound memory.
	distinctValues map[uint64]struct{}
	sketch         []uint64
}

// type assertions.
var _ Calculator = &algoCountDistinct{}

func (c *algoCountDistinct) Add(x any) {
	if inst, ok := x.(*algoCountDistinct); ok {
		if inst.sketch != nil {
			c.ensureSketch()
			for i, word := range inst.sketch {
				c.sketch[i] |= word
			}
		} else {
			for val := range inst.distinctValues {
				c.addHash(val)
			}
		}
		if inst.maxTime > c.maxTime {
			c.maxTime = inst.maxTime
		}
	}
}

func (c *algoCountDistinct) Aggr() ([]*point.Point, error) {
	var kvs point.KVs

	// 统计不重复值的数量
	distinctCount := c.count()

	kvs = kvs.Add(c.key, distinctCount).
		Add(c.key+"_count", distinctCount)

	for _, kv := range c.aggrTags {
		// NOTE: if same-name tag key exist, apply the last one.
		kvs = kvs.SetTag(kv[0], kv[1])
	}

	return []*point.Point{
		point.NewPoint(c.name, kvs, point.WithTimestamp(c.maxTime)),
	}, nil
}

func (c *algoCountDistinct) Reset() {
	c.maxTime = 0
	// 清空不重复值集合
	c.distinctValues = make(map[uint64]struct{})
	c.sketch = nil
}

func (c *algoCountDistinct) doHash(h1 uint64) {
	h := HashCombine(h1, xxhash.Sum64([]byte("count_distinct")))
	h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(c.key)))
	c.MetricBase.hash = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(c.name)))
}

func (c *algoCountDistinct) Base() *MetricBase {
	return &c.MetricBase
}

// 初始化函数，确保 map 被正确创建.
func newAlgoCountDistinct(mb MetricBase, maxTime int64, value any) *algoCountDistinct {
	calc := &algoCountDistinct{
		MetricBase:     mb,
		maxTime:        maxTime,
		distinctValues: make(map[uint64]struct{}, 1),
	}
	calc.addValue(value)
	return calc
}

func (c *algoCountDistinct) addValue(value any) {
	c.addHash(hashDistinctValue(value))
}

func (c *algoCountDistinct) addHash(hash uint64) {
	if c.sketch != nil {
		c.addSketchHash(hash)
		return
	}

	if c.distinctValues == nil {
		c.distinctValues = make(map[uint64]struct{})
	}
	if _, ok := c.distinctValues[hash]; ok || len(c.distinctValues) < countDistinctExactLimit {
		c.distinctValues[hash] = struct{}{}
		return
	}

	c.ensureSketch()
	c.addSketchHash(hash)
}

func (c *algoCountDistinct) ensureSketch() {
	if c.sketch != nil {
		return
	}

	c.sketch = make([]uint64, countDistinctSketchBits/64)
	for hash := range c.distinctValues {
		c.addSketchHash(hash)
	}
	c.distinctValues = nil
}

func (c *algoCountDistinct) addSketchHash(hash uint64) {
	idx := hash % countDistinctSketchBits
	c.sketch[idx/64] |= uint64(1) << (idx % 64)
}

func (c *algoCountDistinct) count() int64 {
	if c.sketch == nil {
		return int64(len(c.distinctValues))
	}

	zeroBits := 0
	for _, word := range c.sketch {
		if word == ^uint64(0) {
			continue
		}
		for bit := 0; bit < 64; bit++ {
			if word&(uint64(1)<<bit) == 0 {
				zeroBits++
			}
		}
	}
	if zeroBits == 0 {
		return countDistinctSketchBits
	}

	m := float64(countDistinctSketchBits)
	estimate := -m * math.Log(float64(zeroBits)/m)
	return int64(estimate + 0.5)
}

func hashDistinctValue(value any) uint64 {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%T:%v", value, value)
	return h.Sum64()
}

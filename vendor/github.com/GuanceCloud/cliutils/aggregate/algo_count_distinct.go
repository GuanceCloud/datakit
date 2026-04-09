package aggregate

import (
	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

type algoCountDistinct struct {
	MetricBase
	maxTime int64
	// 使用 map 来存储不重复的值
	distinctValues map[any]struct{}
}

// type assertions.
var _ Calculator = &algoCountDistinct{}

func (c *algoCountDistinct) Add(x any) {
	if inst, ok := x.(*algoCountDistinct); ok {
		// 合并不重复的值
		for val := range inst.distinctValues {
			c.distinctValues[val] = struct{}{}
		}
		if inst.maxTime > c.maxTime {
			c.maxTime = inst.maxTime
		}
	}
}

func (c *algoCountDistinct) Aggr() ([]*point.Point, error) {
	var kvs point.KVs

	// 统计不重复值的数量
	distinctCount := int64(len(c.distinctValues))

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
	c.distinctValues = make(map[any]struct{})
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
	return &algoCountDistinct{
		MetricBase:     mb,
		maxTime:        maxTime,
		distinctValues: map[any]struct{}{value: {}},
	}
}

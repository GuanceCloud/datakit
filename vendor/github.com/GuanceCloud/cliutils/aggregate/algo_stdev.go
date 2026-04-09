package aggregate

//
// SampleStdDev 计算样本标准差（除以 N-1

import (
	"errors"
	"math"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

type algoStdev struct {
	MetricBase
	data    []float64
	maxTime int64
}

// type assertions.
var _ Calculator = &algoStdev{}

func (c *algoStdev) Add(x any) {
	if inst, ok := x.(*algoStdev); ok {
		// 合并数据点
		c.data = append(c.data, inst.data...)

		if inst.maxTime > c.maxTime {
			c.maxTime = inst.maxTime
		}
	}
}

func (c *algoStdev) Aggr() ([]*point.Point, error) {
	var kvs point.KVs

	// 计算标准差
	stdev, err := SampleStdDev(c.data)
	if err != nil {
		return nil, err
	}

	count := len(c.data)
	kvs = kvs.Add(c.key, stdev).
		Add(c.key+"_count", count)

	for _, kv := range c.aggrTags {
		// NOTE: if same-name tag key exist, apply the last one.
		kvs = kvs.SetTag(kv[0], kv[1])
	}

	return []*point.Point{
		point.NewPoint(c.name, kvs, point.WithTimestamp(c.maxTime)),
	}, nil
}

func (c *algoStdev) Reset() {
	c.data = nil
	c.maxTime = 0
}

func (c *algoStdev) doHash(h1 uint64) {
	h := HashCombine(h1, xxhash.Sum64([]byte("stdev")))
	h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(c.key)))
	c.MetricBase.hash = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(c.name)))
}

func (c *algoStdev) Base() *MetricBase {
	return &c.MetricBase
}

// SampleStdDev 计算样本标准差（除以 N-1）.
func SampleStdDev(data []float64) (float64, error) {
	n := len(data)
	if n < 2 {
		return 0, errors.New("the sample standard deviation requires at least two data points")
	}

	sum := 0.0
	for _, v := range data {
		sum += v
	}
	mean := sum / float64(n)

	var sqDiff float64
	for _, v := range data {
		diff := v - mean
		sqDiff += diff * diff
	}

	variance := sqDiff / float64(n-1)
	return math.Sqrt(variance), nil
}

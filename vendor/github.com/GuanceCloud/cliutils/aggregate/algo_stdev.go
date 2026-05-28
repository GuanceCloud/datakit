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
	count   int64
	mean    float64
	m2      float64
	maxTime int64
}

// type assertions.
var _ Calculator = &algoStdev{}

func (c *algoStdev) Add(x any) {
	if inst, ok := x.(*algoStdev); ok {
		if inst.count == 0 {
			return
		}

		if c.count == 0 {
			c.count = inst.count
			c.mean = inst.mean
			c.m2 = inst.m2
		} else {
			delta := inst.mean - c.mean
			total := c.count + inst.count
			c.mean += delta * float64(inst.count) / float64(total)
			c.m2 += inst.m2 + delta*delta*float64(c.count)*float64(inst.count)/float64(total)
			c.count = total
		}

		if inst.maxTime > c.maxTime {
			c.maxTime = inst.maxTime
		}
	}
}

func (c *algoStdev) Aggr() ([]*point.Point, error) {
	var kvs point.KVs

	// 计算标准差
	stdev, err := c.SampleStdDev()
	if err != nil {
		return nil, err
	}

	kvs = kvs.Add(c.key, stdev).
		Add(c.key+"_count", c.count)

	for _, kv := range c.aggrTags {
		// NOTE: if same-name tag key exist, apply the last one.
		kvs = kvs.SetTag(kv[0], kv[1])
	}

	return []*point.Point{
		point.NewPoint(c.name, kvs, point.WithTimestamp(c.maxTime)),
	}, nil
}

func (c *algoStdev) Reset() {
	c.count = 0
	c.mean = 0
	c.m2 = 0
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

func newAlgoStdev(mb MetricBase, maxTime int64, value float64) *algoStdev {
	c := &algoStdev{
		MetricBase: mb,
		maxTime:    maxTime,
	}
	c.addValue(value)
	return c
}

func (c *algoStdev) addValue(value float64) {
	c.count++
	delta := value - c.mean
	c.mean += delta / float64(c.count)
	delta2 := value - c.mean
	c.m2 += delta * delta2
}

func (c *algoStdev) SampleStdDev() (float64, error) {
	if c.count < 2 {
		return 0, errors.New("the sample standard deviation requires at least two data points")
	}

	variance := c.m2 / float64(c.count-1)
	return math.Sqrt(variance), nil
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

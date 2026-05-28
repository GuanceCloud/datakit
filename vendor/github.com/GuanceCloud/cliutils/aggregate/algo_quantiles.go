package aggregate

import (
	"fmt"
	"sort"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

const quantileSampleLimit = 8192

type algoQuantiles struct {
	maxTime int64
	MetricBase
	all       []float64
	count     int64
	quantiles []float64
}

func (a *algoQuantiles) Add(x any) {
	if inst, ok := x.(*algoQuantiles); ok {
		before := a.count
		for _, v := range inst.all {
			a.addValue(v)
		}
		if missing := inst.count - (a.count - before); missing > 0 {
			a.count += missing
		}

		if inst.maxTime > a.maxTime {
			a.maxTime = inst.maxTime
		}
	}
}

func newAlgoQuantiles(mb MetricBase, maxTime int64, value float64) *algoQuantiles {
	calc := &algoQuantiles{
		MetricBase: mb,
		maxTime:    maxTime,
	}
	calc.addValue(value)
	return calc
}

func (a *algoQuantiles) addValue(v float64) {
	a.count++
	if len(a.all) < quantileSampleLimit {
		a.all = append(a.all, v)
		return
	}

	j := int(HashCombine(Seed1, uint64(a.count)) % uint64(a.count))
	if j < quantileSampleLimit {
		a.all[j] = v
	}
}

// GetPercentile 是一个通用方法，用于获取第 p 个百分位数 (0-100).
func (a *algoQuantiles) GetPercentile(p float64) float64 {
	if n := len(a.all); n == 0 {
		return 0
	}

	sort.Float64s(a.all)
	return percentileFromSorted(a.all, p)
}

func percentileFromSorted(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}

	// 2. 将百分比转换为 0-1 之间的比例
	fraction := p / 100.0

	// 3. 计算索引位置
	pos := fraction * float64(n-1)
	index := int(pos)
	rest := pos - float64(index)

	// 4. 边界处理与线性插值水电费
	if index >= n-1 {
		return sorted[n-1]
	}
	return sorted[index] + rest*(sorted[index+1]-sorted[index])
}

func (a *algoQuantiles) Aggr() ([]*point.Point, error) {
	var kvs point.KVs

	kvs = kvs.Add(a.key+"_count", a.count)
	sort.Float64s(a.all)
	for _, quantile := range a.quantiles {
		key := fmt.Sprintf("%s_P%.0f", a.key, quantile*100) // %.0f: float to int.
		kvs = kvs.Add(key, percentileFromSorted(a.all, quantile*100))
	}
	for _, kv := range a.aggrTags {
		// NOTE: if same-name tag key exist, apply the last one.
		kvs = kvs.SetTag(kv[0], kv[1])
	}

	return []*point.Point{
		point.NewPoint(a.name, kvs, point.WithTimestamp(a.maxTime)),
	}, nil
}

func (a *algoQuantiles) Reset() {
	a.maxTime = 0
	a.all = nil
	a.count = 0
}

func (a *algoQuantiles) Base() *MetricBase {
	return &a.MetricBase
}

var _ Calculator = &algoQuantiles{}

func (a *algoQuantiles) doHash(h1 uint64) {
	h := HashCombine(h1, xxhash.Sum64([]byte("quantiles")))
	h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(a.key)))
	a.MetricBase.hash = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(a.name)))
}

func (a *algoQuantiles) addOpts(opts *AggregationAlgo_QuantileOpts) {
	if opts.QuantileOpts != nil && opts.QuantileOpts.Percentiles != nil {
		a.quantiles = opts.QuantileOpts.Percentiles
	}
}

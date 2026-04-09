package aggregate

import (
	"fmt"
	"sort"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

type algoQuantiles struct {
	maxTime int64
	MetricBase
	all       []float64
	count     int64
	quantiles []float64
}

func (a *algoQuantiles) Add(x any) {
	if inst, ok := x.(*algoQuantiles); ok {
		a.count += inst.count
		a.all = append(a.all, inst.all...)

		if inst.maxTime > a.maxTime {
			a.maxTime = inst.maxTime
		}
	}
}

// GetPercentile 是一个通用方法，用于获取第 p 个百分位数 (0-100).
func (a *algoQuantiles) GetPercentile(p float64) float64 {
	n := len(a.all)
	if n == 0 {
		return 0
	}

	sort.Float64s(a.all)

	// 2. 将百分比转换为 0-1 之间的比例
	fraction := p / 100.0

	// 3. 计算索引位置
	pos := fraction * float64(n-1)
	index := int(pos)
	rest := pos - float64(index)

	// 4. 边界处理与线性插值水电费
	if index >= n-1 {
		return a.all[n-1]
	}
	return a.all[index] + rest*(a.all[index+1]-a.all[index])
}

func (a *algoQuantiles) Aggr() ([]*point.Point, error) {
	var kvs point.KVs

	kvs = kvs.Add(a.key+"_count", a.count)
	for _, quantile := range a.quantiles {
		key := fmt.Sprintf("%s_P%.0f", a.key, quantile*100) // %.0f: float to int.
		kvs = kvs.Add(key, a.GetPercentile(quantile*100))
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
	a.all = []float64{}
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

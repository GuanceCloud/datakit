package aggregate

import (
	"fmt"
	"reflect"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

/*
// CaculatorCache cache all Calculators.
type CaculatorCache struct {
	cache map[uint64]Calculator
	heap  []Calculator
	mtx   sync.RWMutex
}

func NewCaculatorCache() *CaculatorCache {
	return &CaculatorCache{
		cache: map[uint64]Calculator{},
	}
}

func (cc *CaculatorCache) AddBatches(batches ...*AggregationBatch) {
	cc.mtx.Lock()
	defer cc.mtx.Unlock()

	for _, b := range batches {
		for _, c := range newCalculators(b) {
			calcHash := c.Base().hash

			if calc, ok := cc.cache[calcHash]; ok {
				calc.Add(c)
			} else {
				c.Base().build()

				cc.cache[calcHash] = c
				heap.Push(cc, c)

			}
		}
	}
}

func (cc *CaculatorCache) DelBatch(id uint64) bool {
	cc.mtx.Lock()
	defer cc.mtx.Unlock()

	if c, ok := cc.cache[id]; !ok {
		return false // not exist
	} else {
		heap.Remove(cc, c.Base().heapIdx)
		delete(cc.cache, id)
		return true
	}
}

func (cc *CaculatorCache) PeekNext() (Calculator, bool) {
	cc.mtx.RLock()
	defer cc.mtx.RUnlock()

	if len(cc.heap) == 0 {
		return nil, false
	}

	return cc.heap[0], true
}

func (cc *CaculatorCache) ScheduleJob(c Calculator) {
	cc.mtx.Lock()
	defer cc.mtx.Unlock()

	mb := c.Base()
	mb.nextWallTime = AlignNextWallTime(time.Now(), time.Duration(mb.window))
	mb.heapIdx = len(cc.heap)
	heap.Push(cc, c)
	cc.cache[mb.hash] = c
}

func (cc *CaculatorCache) Len() int {
	return len(cc.heap)
}

func (cc *CaculatorCache) Less(i, j int) bool {
	// smallest nextWallTime pop first.
	less := cc.heap[i].Base().nextWallTime < cc.heap[j].Base().nextWallTime

	return less
}

func (cc *CaculatorCache) Swap(i, j int) {
	if len(cc.heap) == 0 {
		return
	}

	cc.heap[i], cc.heap[j] = cc.heap[j], cc.heap[i]
	cc.heap[i].Base().heapIdx = i
	cc.heap[j].Base().heapIdx = j
}

func (cc *CaculatorCache) Push(x any) {
	c := x.(Calculator)
	c.Base().heapIdx = len(cc.heap)
	cc.heap = append(cc.heap, c)
}

func (cc *CaculatorCache) Pop() any {
	old := cc.heap
	n := len(old)
	if n == 0 {
		return nil
	}

	// pop out the last one
	c := old[n-1]
	c.Base().heapIdx = -1 // label removed
	cc.heap = old[0 : n-1]
	return c
}


*/

type Calculator interface {
	Add(any)
	Aggr() ([]*point.Point, error)
	Reset()
	Base() *MetricBase
	ToString() string
}

func AlignNextWallTime(t time.Time, align time.Duration) int64 {
	truncated := t.Truncate(align)

	if truncated.Equal(t) {
		return t.Unix()
	}

	return truncated.Add(align).Unix()
}

func newCalculators(batch *AggregationBatch) (res []Calculator) {
	var ptwrap *point.Point
	// now    = time.Now()

	for key, algo := range batch.AggregationOpts {
		var extraTags [][2]string
		// nextWallTime = AlignNextWallTime(now, time.Duration(algo.Window))

		if len(algo.AddTags) > 0 {
			// NOTE: we first add these extra-tags from algorithm configure. If origin
			// data got same-name tag key, we override the algorithm configured tags by
			// using kv.SetTag().
			for k, v := range algo.AddTags {
				extraTags = append(extraTags, [2]string{k, v})
			}
		}

		for _, pt := range batch.Points.Arr {
			var (
				keyName string
				val     any
				ptwrap  = point.WrapPB(ptwrap, pt)
			)

			if algo.SourceField != "" {
				if val = ptwrap.Get(algo.SourceField); val == nil {
					continue
				}

				keyName = key
			} else {
				if val = ptwrap.Get(key); val == nil {
					continue
				}

				keyName = key
			}

			if keyName == "" {
				continue
			}

			mb := MetricBase{
				pt:       pt,
				key:      keyName,
				name:     ptwrap.Name(),
				aggrTags: extraTags,
				// align to next wall-time
				// XXX: what if the point reach too late?
				nextWallTime: AlignNextWallTime(ptwrap.Time(), time.Duration(algo.Window)),
				window:       algo.Window,
			}
			method := NormalizeAlgoMethod(algo.Method)
			f64, ok := val.(float64)
			if !ok {
				if i64, ok := val.(int64); !ok {
					if method == COUNT_DISTINCT || method == COUNT {
						// 这两种类型可以不转换成 float64
					} else {
						l.Warnf("key %s non-numeric type(%s) for method %s, ignored", keyName, reflect.TypeOf(val), method)
						continue
					}
				} else {
					f64 = float64(i64)
				}
			}
			// we get the kv for current algorithm.
			switch method {
			case MAX:
				calc := &algoMax{
					count:      1,
					max:        f64,
					maxTime:    ptwrap.Time().UnixNano(),
					MetricBase: mb,
				}
				calc.doHash(batch.RoutingKey)
				res = append(res, calc)

			case SUM:
				calc := &algoSum{
					count:      1,
					delta:      f64,
					maxTime:    ptwrap.Time().UnixNano(),
					MetricBase: mb,
				}
				calc.doHash(batch.RoutingKey)
				res = append(res, calc)

			case AVG:
				calc := &algoAvg{
					count:      1,
					delta:      f64,
					maxTime:    ptwrap.Time().UnixNano(),
					MetricBase: mb,
				}
				calc.doHash(batch.RoutingKey)
				res = append(res, calc)

			case COUNT:
				calc := &algoCount{
					maxTime:    ptwrap.Time().UnixNano(),
					MetricBase: mb,
					count:      1,
				}

				calc.doHash(batch.RoutingKey)
				res = append(res, calc)

			case MIN:
				calc := &algoMin{
					count:      1,
					min:        f64,
					maxTime:    ptwrap.Time().UnixNano(),
					MetricBase: mb,
				}

				calc.doHash(batch.RoutingKey)
				res = append(res, calc)

			case HISTOGRAM:
				calc := &algoHistogram{
					count:      1,
					val:        f64,
					maxTime:    ptwrap.Time().UnixNano(),
					MetricBase: mb,
					leBucket:   map[string]float64{},
				}
				le, _ := ptwrap.GetS("le")
				calc.leBucket[le] = f64

				calc.doHash(batch.RoutingKey)
				res = append(res, calc)

			case QUANTILES:
				calc := &algoQuantiles{
					count:      1,
					all:        []float64{f64},
					maxTime:    ptwrap.Time().UnixNano(),
					MetricBase: mb,
				}
				if algo.Options != nil {
					switch opt := algo.Options.(type) {
					case *AggregationAlgo_QuantileOpts:
						calc.addOpts(opt)
					default: //nolint
					}
				}

				calc.doHash(batch.RoutingKey)
				res = append(res, calc)

			case STDEV:
				calc := &algoStdev{
					MetricBase: mb,
					data:       []float64{f64},
					maxTime:    ptwrap.Time().UnixNano(),
				}
				calc.doHash(batch.RoutingKey)
				res = append(res, calc)

			case COUNT_DISTINCT:
				calc := newAlgoCountDistinct(mb, ptwrap.Time().UnixNano(), val)
				calc.doHash(batch.RoutingKey)
				res = append(res, calc)

			case LAST:
				calc := &algoCountLast{
					MetricBase: mb,
					last:       f64,
					lastTime:   ptwrap.Time().UnixNano(),
					count:      1,
				}
				calc.doHash(batch.RoutingKey)
				res = append(res, calc)

			case FIRST:
				calc := &algoCountFirst{
					MetricBase: mb,
					first:      f64,
					firstTime:  ptwrap.Time().UnixNano(),
					count:      1,
				}
				calc.doHash(batch.RoutingKey)
				res = append(res, calc)

			case METHOD_UNSPECIFIED:
			case EXPO_HISTOGRAM: // TODO
			default: // pass
			}
		}
	}

	return res
}

func prettyBatch(ab *AggregationBatch) string {
	return fmt.Sprintf("routingKey: %d\nconfigHash: %d\npoints: %d",
		ab.RoutingKey, ab.ConfigHash, len(ab.Points.Arr))
}

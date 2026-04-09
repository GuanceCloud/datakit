package aggregate

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
)

type (
	Action string
)

var l = logger.DefaultSLogger("aggregator")

const (
	// actions.
	ActionPassThrough = "passthrough"
	ActionDrop        = "drop"
)

// PickPoints organizes points into batches based on aggregation rules and grouping keys.
// 多种 category:M,L,RUM,T 类型的数据都可以筛选，返回的一定是指标类型。
func (ac *AggregatorConfigure) PickPoints(category string, pts []*point.Point) map[uint64]*Batchs {
	batchs := make(map[uint64]*Batchs)
	abs := make([]*AggregationBatch, 0)
	for _, ar := range ac.AggregateRules {
		if ar.Selector.Category == category {
			sPts := ar.SelectPoints(pts)
			abs = append(abs, ar.GroupbyBatch(ac, sPts)...)
		}
	}

	for _, ab := range abs {
		if _, ok := batchs[ab.PickKey]; !ok {
			bs := &Batchs{
				PickKey: ab.PickKey,
				Batchs:  []*AggregationBatch{ab},
			}
			batchs[ab.PickKey] = bs
		} else {
			batchs[ab.PickKey].Batchs = append(batchs[ab.PickKey].Batchs, ab)
		}
	}
	return batchs
}

// SelectPoints filters points based on the rule's selector criteria.
func (ar *AggregateRule) SelectPoints(pts []*point.Point) []*point.Point {
	return ar.Selector.doSelect(ar.Groupby, pts)
}

// GroupbyPoints groups points by their hash value calculated from grouping keys.
func (ar *AggregateRule) GroupbyPoints(pts []*point.Point) map[uint64][]*point.Point {
	res := map[uint64][]*point.Point{}
	for _, pt := range pts {
		h := hash(pt, ar.Groupby)
		res[h] = append(res[h], pt)
	}

	return res
}

// GroupbyBatch creates aggregation batches from points based on grouping keys.
func (ar *AggregateRule) GroupbyBatch(ac *AggregatorConfigure, pts []*point.Point) (batches []*AggregationBatch) {
	for _, pt := range pts {
		h := hash(pt, ar.Groupby)
		pickKey := pickHash(pt, ar.Groupby)
		b := &AggregationBatch{
			RoutingKey:      h,
			ConfigHash:      ac.hash,
			PickKey:         pickKey,
			AggregationOpts: ar.aggregationOpts,
			Points:          &point.PBPoints{Arr: []*point.PBPoint{pt.PBPoint()}},
		}
		batches = append(batches, b)
	}

	return batches
}

const (
	GuanceRoutingKey = "Guance-Routing-Key"
	GuancePickKey    = "Guance-Pick-Key"
)

// batchRequest creates an HTTP request for sending aggregation batch data.
func batchRequest(ab *AggregationBatch, url string) (*http.Request, error) {
	body, err := ab.Marshal()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set(GuanceRoutingKey, strconv.FormatUint(ab.RoutingKey, 10))
	return req, nil
}

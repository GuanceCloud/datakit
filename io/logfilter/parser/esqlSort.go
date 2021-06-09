package parser

import "strings"

var (
	// SortMap sort的映射关系
	SortMap = map[OrderType]string{
		OrderDesc: "desc",
		OrderAsc:  "asc",
	}

	// SortMissingMap sort的缺省值映射关系
	SortMissingMap = map[OrderType]string{
		OrderDesc: "_last",  // 降序的时候，空值默认为最后
		OrderAsc:  "_first", // 升序的时候，空值默认为最新
	}

	// SortMissing sort缺省值
	SortMissing = map[string]string{
		"desc": "_last",
		"asc":  "_first",
	}
)

// 获取sort
func sortTransport(m *DFQuery, esPtr *EST) ([]interface{}, error) {

	var res = []interface{}{}

	if esPtr.dfState.aggs == true {
		return res, nil
	}

	if m.SearchAfter != nil {
		return searchAfterSortTransport(m, esPtr)
	}

	if m.OrderBy != nil {
		// (1) 存在order by
		return specialSortTransport(m, esPtr)
	}

	// (2) 没有排序, 默认加上时间逆序排序
	return defaultSortTransport(m, esPtr)

}

// searchAfterSortTransport search_after 排序
func searchAfterSortTransport(m *DFQuery, esPtr *EST) ([]interface{}, error) {
	if m.OrderBy != nil {
		// (1) 存在order by
		return specialSortTransport(m, esPtr)
	}
	// (2) 默认排序，即按照 [time:desc, __docid:desc]
	return defaultSearchAfterSortTransport(m, esPtr)

}

// 指定了order by
func specialSortTransport(m *DFQuery, esPtr *EST) ([]interface{}, error) {
	var res = []interface{}{}
	for _, item := range m.OrderBy.List {
		elem, ok := item.(*OrderByElem)
		if !ok {
			return res, nil
		}

		fieldName := elem.Column
		fieldName, err := esPtr.getDSLFieldName(fieldName, true)
		if err != nil {
			return res, err
		}
		// field name
		inner := map[string]string{
			"order":         SortMap[elem.Opt],
			"missing":       SortMissingMap[elem.Opt],
			"unmapped_type": "string",
		}
		outer := SIMAP{fieldName: inner}
		res = append(res, outer)

	}
	return res, nil
}

// 默认时间逆序
func defaultSortTransport(m *DFQuery, esPtr *EST) ([]interface{}, error) {
	var res = []interface{}{}
	inner := map[string]string{
		"order":         "desc",
		"missing":       "_first",
		"unmapped_type": "string",
	}
	outer := SIMAP{esPtr.esNamespace.timeField: inner}
	res = append(res, outer)
	return res, nil
}

//search after 默认时间逆序
func defaultSearchAfterSortTransport(m *DFQuery, esPtr *EST) ([]interface{}, error) {
	var res = []interface{}{}
	tInner := map[string]string{
		"order":         "desc",
		"missing":       "_first",
		"unmapped_type": "string",
	}
	tOuter := SIMAP{esPtr.esNamespace.timeField: tInner}
	dInner := map[string]string{
		"order":         "desc",
		"missing":       "_first",
		"unmapped_type": "string",
	}
	dOuter := SIMAP{DefaultDocID: dInner}
	res = append(res, tOuter)
	res = append(res, dOuter)
	return res, nil
}

// 桶聚合的排序
func (esPtr *EST) updateBucketOrder(nodeList []aggsNode) []aggsNode {

	if len(nodeList) != 2 {
		return nodeList
	}

	// 第1层桶聚合只能有一个group by字段
	if len(nodeList[0].items) != 1 {
		return nodeList
	}

	// 第2层聚合函数, 只能对于singleBucket的聚合函数的结果排序，即不能是桶聚合
	for _, item := range nodeList[1].items {
		if item.tName != "metric" {
			break
		}
		switch item.fName {
		case AggsMetricFuncs[PercentileFName]: // pecentiles会有多指标，需要选择一个
			for k, v := range esPtr.groupOrders {
				if strings.HasPrefix(k, item.alias) { // percent需要添加具体的指标,例如: percent.99
					nodeList[0].items[0].args["order"] = map[string]string{k: v}
					break
				}
			}
		default:
			for k, v := range esPtr.groupOrders {
				if k == item.alias { // 其他函数不需要指定具体的指标,例如: max
					nodeList[0].items[0].args["order"] = map[string]string{k: v}
					break
				}
			}
		}

	}

	return nodeList
}

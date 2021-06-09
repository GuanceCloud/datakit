package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/kodo/utils"
)

// 聚合基本单元
type aggsItem struct {
	tName    string      // typeName, 表示聚合类型（metric/bucket/pipeline)
	fName    string      // functionName如果是指标聚合，函数名称
	args     SIMAP       // 函数参数
	alias    string      // 聚合别名
	esStruct interface{} // es返回结构
}

// 聚合节点
type aggsNode struct {
	items []aggsItem // 对应一个聚合层包含多个子聚合，例如: topHits与avg同层
}

var (
	// 支持聚合函数列表
	TopFName           = "top"
	BottomFName        = "bottom"
	FirstFName         = "first"
	LastFName          = "last"
	AvgFName           = "avg"
	CountdistinctFName = "count_distinct"
	DistinctFName      = "distinct"
	TermsFName         = "terms"
	MaxFName           = "max"
	MinFName           = "min"
	PercentileFName    = "percentile"
	SumFName           = "sum"
	CountFName         = "count"
	HistogramFName     = "histogram" // 直方图聚合

	// TopFuncs 可以转化为tophits的函数
	TopFuncs = []string{TopFName, BottomFName, FirstFName, LastFName}

	// AggsMetricFuncs 指标聚合函数集合
	// 函数名不区分大小写，ast统一lower
	AggsMetricFuncs = map[string]string{
		AvgFName:           AvgFName,      // 平均值
		CountdistinctFName: "cardinality", // ES中的cardinality不同于influxdb，定义为新函数countdistinct
		DistinctFName:      TermsFName,    // distinct使用terms实现
		MaxFName:           MaxFName,
		MinFName:           MinFName,
		PercentileFName:    "percentiles", // 百分位
		SumFName:           SumFName,
		CountFName:         "value_count", // count
		// HistogramFName:     "histogram",
	}

	// AggsBucketFuncs 桶聚合相关函数集合
	AggsBucketFuncs = map[string]string{
		AggsTophits:    "top_hits",
		HistogramFName: "histogram",
		// AGGSGROUPSIZE: "size",
	}

	// 内置cast函数

	// IntFName int 函数, string -> int
	IntFName = "int"

	// FloatFName float 函数, string -> float
	FloatFName = "float"

	// NestAggFuncs 嵌套聚合函数
	// CastFuncs = []string{IntFName, FloatFName}
	NestAggFuncs = map[string][]string{
		AvgFName:    []string{IntFName, FloatFName},
		MaxFName:    []string{IntFName, FloatFName},
		MinFName:    []string{IntFName, FloatFName},
		SumFName:    []string{IntFName, FloatFName},
		ScriptFName: []string{AvgFName, MaxFName, MinFName, SumFName},
	}
)

// aggsTransport aggs(聚合)部分的转换
func aggsTransport(m *DFQuery, esPtr *EST) (interface{}, error) {
	var res = SIMAP{}
	if esPtr.dfState.aggs == false {
		return res, nil
	}
	// 生成聚合节点列表
	al, err := genAggsNodeList(m, esPtr)
	if err != nil {
		return res, err
	}
	// 解析聚合节点列表
	res, err = parseAggsNodeList(al)
	if err != nil {
		return res, err
	}
	return res[AggsIdentifier], nil
}

// genAggsNodeList 生成聚合列表
func genAggsNodeList(m *DFQuery, esPtr *EST) ([]aggsNode, error) {
	var nodeList = []aggsNode{}

	// 1）生成桶聚合
	if esPtr.dfState.bucket {
		nodePtr, err := m.genBucketNodes(esPtr)
		if err != nil {
			return nodeList, err
		}
		for _, v := range *nodePtr {
			nodeList = append(nodeList, v)
		}
		// histogram
		if esPtr.dfState.hasHistogram {
			nodePtr, err := m.genHistogramNodes(esPtr)
			if err != nil {
				return nil, err
			}
			for _, v := range *nodePtr {
				nodeList = append(nodeList, v)
			}
		}
	}
	// 2) 生成时间范围聚合
	if esPtr.dfState.dateHg {
		nodePtr, err := m.genDateHgNode(esPtr)
		if err != nil {
			return nodeList, err
		}
		nodeList = append(nodeList, *nodePtr)
	}
	// 3）生成指标聚合
	if esPtr.dfState.metric {
		nodePtr, err := m.genMetricNodes(esPtr)
		if err != nil {
			return nodeList, err
		}
		if len(nodePtr.items) > 0 {
			nodeList = append(nodeList, *nodePtr)
		}
		// 如果有时间聚合，time字段名称改变
		// if esPtr.dfState.dateHg {
		// 	for i, v := range esPtr.SortFields {
		// 		if v == esPtr.esNamespace.timeField {
		// 			esPtr.SortFields[i] = TIMEDEFAULT
		// 		}
		// 	}
		// }
	}
	// 4) 如果有bucket聚合但是没有指标聚合，需要添加tophits
	if esPtr.dfState.bucket && esPtr.dfState.metric == false && esPtr.dfState.hasHistogram == false {
		nodePtr, err := m.genTophitsNodes(esPtr, DefaultTophitLimit)
		if err != nil {
			return nodeList, err
		}
		nodeList = append(nodeList, *nodePtr)
	}
	// 5) 如果有order by 内层聚合结果
	nodeList = esPtr.updateBucketOrder(nodeList)
	return nodeList, nil
}

func parseAggsNodeList(al []aggsNode) (SIMAP, error) {
	var p = new(SIMAP)
	l := len(al)
	for i := 0; i < l; i++ {
		ri := l - 1 - i
		err := al[ri].parse(ri, p)
		if err != nil {
			return nil, err
		}
	}
	return *p, nil
}

// genBucketNodes 生成桶聚合
func (m *DFQuery) genBucketNodes(esPtr *EST) (*[]aggsNode, error) {
	var res []aggsNode
	for i, fieldName := range esPtr.groupFields {
		// 获取索引时候的fieldName，
		// 如果使用aliasName，需要转为实际fileName
		var aliasName = fieldName
		fieldName, err := esPtr.getDSLFieldName(fieldName, true)
		if err != nil {
			return nil, err
		}

		size := esPtr.getBucketSize(i)

		item := aggsItem{
			tName: BucketAggs,
			fName: TermsFName,
			args: SIMAP{
				"field": fieldName,
				"size":  size,
			},
			alias: formatAliasField(aliasName),
		}
		// 添加桶聚合排序
		if groupOrder, ok := esPtr.groupOrders[fieldName]; ok {
			item.args["order"] = map[string]string{"_key": groupOrder}
		}
		node := aggsNode{items: []aggsItem{item}}
		res = append(res, node)
	}

	return &res, nil
}

// genHistogramNode 生成直方图桶聚合
func (m *DFQuery) genHistogramNodes(esPtr *EST) (*[]aggsNode, error) {
	var res []aggsNode
	fieldName := esPtr.histogramInfo["fieldName"].(string)
	start := esPtr.histogramInfo["start"].(float64)
	end := esPtr.histogramInfo["end"].(float64)
	interval := esPtr.histogramInfo["interval"].(float64)

	// 获取索引时候的fieldName，
	fieldName, err := esPtr.getDSLFieldName(fieldName, true)

	if err != nil {
		return nil, err
	}

	// distogram
	hItem, err := esPtr.genHistogramItem(fieldName, start, end, interval)
	if err != nil {
		return nil, err
	}
	hNode := aggsNode{items: []aggsItem{hItem}}

	// value_count
	vItem, err := esPtr.genValueCountItem(fieldName)
	if err != nil {
		return nil, err
	}
	vNode := aggsNode{items: []aggsItem{vItem}}

	// res = append(res, fNode)
	res = append(res, hNode)
	res = append(res, vNode)
	esPtr.appendSortFields(esPtr.esNamespace.timeField)
	esPtr.appendSortFields(fieldName)
	return &res, nil
}

func (esPtr *EST) genFilterItem(fieldName string, start, end float64) (aggsItem, error) {
	vRange := map[string]float64{
		"gte": start,
		"lte": end,
	}

	fRange := SIMAP{
		fieldName: vRange,
	}

	filterItem := aggsItem{
		tName: BucketAggs,
		fName: "filter",
		args: SIMAP{
			"range": fRange,
		},
		alias: esPtr.esNamespace.timeField,
	}
	return filterItem, nil
}

func (esPtr *EST) genHistogramItem(fieldName string, start, end, interval float64) (aggsItem, error) {
	bounds := map[string]float64{
		"min": start,
		"max": end,
	}

	hItem := aggsItem{
		tName: BucketAggs,
		fName: "histogram",
		args: SIMAP{
			"field":           fieldName,
			"interval":        interval,
			"extended_bounds": bounds,
		},
		alias: esPtr.esNamespace.timeField,
	}
	if minDoc, ok := esPtr.histogramInfo["minDoc"]; ok {
		hItem.args["min_doc_count"] = minDoc.(float64)
	}
	return hItem, nil
}

func (esPtr *EST) genValueCountItem(fieldName string) (aggsItem, error) {
	vItem := aggsItem{
		fName: AggsMetricFuncs[CountFName],
		tName: "metric",
		args:  SIMAP{"field": fieldName},
		alias: fieldName,
	}
	return vItem, nil
}

// genDateHgNode 生成时间范围聚合
func (m *DFQuery) genDateHgNode(esPtr *EST) (*aggsNode, error) {
	var inner = SIMAP{}
	interval, err := m.getTimeStamp(esPtr)
	if err != nil {
		return nil, err
	}
	inner["field"] = esPtr.esNamespace.timeField
	inner["interval"] = interval
	var item = aggsItem{
		fName: "date_histogram",
		tName: "bucket",
		args:  inner,
		alias: utils.Times,
	}
	var items = []aggsItem{item}
	return &aggsNode{items: items}, nil
}

// genMetricNodes 生成指标聚合
func (m *DFQuery) genMetricNodes(esPtr *EST) (*aggsNode, error) {
	var res = new(aggsNode)
	var items = []aggsItem{}
	var mNames = []string{}
	for i, t := range m.Targets {
		itemPtr, talias, err := genMetricNode(esPtr, t, m)
		if err != nil {
			return res, err
		}
		if itemPtr != nil {
			items = append(items, *itemPtr)
		}
		// target 别名
		m.Targets[i].Talias = talias
		// 同一层的指标聚合，不能是相同字段相同函数
		for _, mN := range mNames {
			if talias == mN {
				return nil, fmt.Errorf("The same field with function name should only appear once")
			}
		}
		mNames = append(mNames, talias)
	}
	res.items = items
	return res, nil
}

// genTophitsNodes 生成tophits聚合
func (m *DFQuery) genTophitsNodes(esPtr *EST, limitSize int) (*aggsNode, error) {
	var res = new(aggsNode)

	args := SIMAP{}
	// （1）args _source
	args["_source"] = esPtr.SortFields
	// （2）args size
	args["size"] = limitSize
	// （3）args sort, 默认是时间逆序
	sort := []interface{}{}
	inner := map[string]string{
		"order":         "desc",
		"missing":       "_first",
		"unmapped_type": "string",
	}
	outer := SIMAP{esPtr.esNamespace.timeField: inner}
	sort = append(sort, outer)
	args["sort"] = sort

	var topAgg = aggsItem{
		fName: AggsBucketFuncs[AggsTophits],
		tName: "metric",
		args:  args,
	}
	topAgg.alias = AggsBucketFuncs[AggsTophits]
	res.items = []aggsItem{topAgg}
	return res, nil
}

// genMetricNode
func genMetricNode(esPtr *EST, t *Target, m *DFQuery) (*aggsItem, string, error) {
	var (
		talias  string
		itemPtr *aggsItem
		err     error
	)
	if fc, ok := (t.Col).(*FuncExpr); ok {

		// （1）ES原生支持的指标函数 (max, min等)
		if fName, find := AggsMetricFuncs[fc.Name]; find {

			// distinct函数使用terms实现
			switch fc.Name {
			case DistinctFName:
				itemPtr, talias, err = genDistinctItem(m, esPtr, fc, t.Alias)
				if err != nil {
					return nil, "", fmt.Errorf("distinct aggs error")
				}
			default:
				itemPtr, talias, err = genMetricItem(esPtr, fName, fc, t.Alias)
				if err != nil {
					return nil, "", fmt.Errorf("metric aggs error")
				}
			}
		}

		// (2) 通过组合查询语句，构造的函数(top, bottom, first, last)
		if find := findStringSlice(fc.Name, TopFuncs); find {
			itemPtr, talias, err = genTopHitsItem(esPtr, fc, t.Alias)
			if err != nil {
				return nil, "", fmt.Errorf("tophits aggs error")
			}
			return itemPtr, talias, nil
		}

	}
	return itemPtr, talias, nil
}

//genDistinctItem 生成terms聚合
func genDistinctItem(m *DFQuery, esPtr *EST, fc *FuncExpr, alias string) (*aggsItem, string, error) {
	fieldName := GetStringParam(fc.Param[0])

	// 是否是别名, 如果是text字段，桶聚合需要使用fName.keyword
	rFieldName, err := esPtr.getDSLFieldName(fieldName, true)
	if err != nil {
		return nil, "", err
	}
	// distinctSize := DefaultLimit
	// if esPtr.limitSize > 0 {
	// 	distinctSize = esPtr.limitSize // input size
	// }
	res := aggsItem{
		tName: BucketAggs,
		fName: TermsFName,
		args:  SIMAP{"field": rFieldName, "size": DefaultLimit},
	}
	esPtr.distinctField = rFieldName
	if alias != "" {
		res.alias = alias
	} else {
		aFieldName := formatAliasField(fieldName)
		res.alias = fc.Name + "_" + aFieldName // 指标聚合，列名为 输入函数名_字段名，非es中函数名
	}
	esPtr.appendSortFields(esPtr.esNamespace.timeField)
	esPtr.appendSortFields(res.alias)
	return &res, res.alias, nil
}

//genMetricItem 生成一个metric聚合
func genMetricItem(esPtr *EST, fName string, fc *FuncExpr, alias string) (*aggsItem, string, error) {

	// percent 函数
	if fc.Name == PercentileFName {
		return genPercentMerticItem(esPtr, fName, fc, alias)
	}

	// 如果有内层函数
	if _, ok := fc.Param[0].(*StaticCast); ok {
		return genNestMetricItem(esPtr, fName, fc, alias)
	}

	// 如果有内层script函数
	if _, ok := fc.Param[0].(*FuncExpr); ok {
		return genNestScriptMetricItem(esPtr, fName, fc, alias)
	}

	fieldName := GetStringParam(fc.Param[0])

	var args = SIMAP{}

	// 是否是别名, 如果是text字段，桶聚合需要使用fName.keyword
	fieldName, err := esPtr.getDSLFieldName(fieldName, true)
	if err != nil {
		return nil, "", err
	}

	args["field"] = fieldName
	var res = aggsItem{
		fName: fName,
		tName: "metric",
		args:  args,
	}

	if alias != "" {
		res.alias = alias
	} else {
		rk := formatAliasField(fieldName)
		res.alias = fc.Name + "_" + rk // 指标聚合，列名为 输入函数名_字段名，非es中函数名
	}
	esPtr.appendSortFields(esPtr.esNamespace.timeField)
	esPtr.appendSortFields(res.alias)
	return &res, res.alias, nil
}

//genPercentMerticItem 生成一个percent metric聚合
func genPercentMerticItem(esPtr *EST, fName string, fc *FuncExpr, alias string) (*aggsItem, string, error) {

	fieldName := GetStringParam(fc.Param[0])

	var args = SIMAP{}

	// 是否是别名, 如果是text字段，桶聚合需要使用fName.keyword
	fieldName, err := esPtr.getDSLFieldName(fieldName, true)
	if err != nil {
		return nil, "", err
	}

	args["field"] = fieldName
	strNumber := ""
	if nv, ok := fc.Param[1].(*NumberLiteral); ok { // 数值
		if nv.IsInt {
			args["percents"] = []int64{nv.Int}
			strNumber = fmt.Sprintf("%.1f", float64(nv.Int))
		} else {
			floatNumber := math.Floor(nv.Float*10) / 10
			args["percents"] = []float64{floatNumber}
			strNumber = fmt.Sprintf("%.1f", floatNumber)
		}
	}
	var res = aggsItem{
		fName: fName,
		tName: "metric",
		args:  args,
	}

	if alias != "" {
		res.alias = alias
	} else {
		rk := formatAliasField(fieldName)
		res.alias = strings.Join([]string{fc.Name, rk, strNumber}, "_") // 指标聚合，列名为 输入函数名_字段名，非es中函数名
	}
	esPtr.appendSortFields(esPtr.esNamespace.timeField)
	esPtr.appendSortFields(res.alias + "_" + strNumber)
	return &res, res.alias, nil
}

// getCastScript
func getCastScript(fName, fieldName string) interface{} {
	source := ""
	switch fName {
	case IntFName:
		source = fmt.Sprintf(
			`if (doc['%s'].size() > 0) {Integer.parseInt(doc['%s'].value)}`,
			fieldName,
			fieldName,
		)
	case FloatFName:
		source = fmt.Sprintf(
			`if (doc['%s'].size() > 0) {Float.parseFloat(doc['%s'].value)}`,
			fieldName,
			fieldName,
		)
	}
	return SIMAP{"source": source}
}

//genNestMetricItem 生成一个包含内置函数的metric聚合
func genNestMetricItem(esPtr *EST, fName string, fc *FuncExpr, alias string) (*aggsItem, string, error) {
	ifc, _ := fc.Param[0].(*StaticCast)
	// fieldName := GetStringParam(ifc.Param[0])
	fieldName := ifc.Val.String()
	var args = SIMAP{}

	// 是否是别名, 如果是text字段，桶聚合需要使用fName.keyword
	fieldName, err := esPtr.getDSLFieldName(fieldName, true)
	if err != nil {
		return nil, "", err
	}

	ifName := ""
	if ifc.IsInt {
		ifName = IntFName
	} else {
		if ifc.IsFloat {
			ifName = FloatFName
		}
	}

	castScript := getCastScript(ifName, fieldName)
	args["script"] = castScript
	var res = aggsItem{
		fName: fName,
		tName: "metric",
		args:  args,
	}

	if alias != "" {
		res.alias = alias
	} else {
		rk := formatAliasField(fieldName)
		res.alias = fc.Name + "_" + rk // 指标聚合，列名为 输入函数名_字段名，非es中函数名
	}
	esPtr.appendSortFields(esPtr.esNamespace.timeField)
	esPtr.appendSortFields(res.alias)
	return &res, res.alias, nil
}

//genNestScriptMetricItem 生成一个包含内置script函数的聚合
func genNestScriptMetricItem(esPtr *EST, fName string, fc *FuncExpr, alias string) (*aggsItem, string, error) {
	ifc, _ := fc.Param[0].(*FuncExpr)
	// fieldName := GetStringParam(ifc.Param[0])
	fieldName := ifc.Name
	var args = SIMAP{}

	args["script"] = map[string]string{
		"source": GetStringParam(ifc.Param[0]),
	}
	var res = aggsItem{
		fName: fName,
		tName: "metric",
		args:  args,
	}

	if alias != "" {
		res.alias = alias
	} else {
		rk := formatAliasField(fieldName)
		res.alias = fc.Name + "_" + rk // 指标聚合，列名为 输入函数名_字段名，非es中函数名
	}
	esPtr.appendSortFields(esPtr.esNamespace.timeField)
	esPtr.appendSortFields(res.alias)
	return &res, res.alias, nil
}

// genTopHitsArgs 获取tophits聚合内部属性
func genTopHitsArgs(esPtr *EST, fc *FuncExpr) (SIMAP, error) {
	res := SIMAP{}
	source := []string{}
	size := "1"
	sort := []interface{}{}
	outer := SIMAP{}
	fieldName := GetStringParam(fc.Param[0])
	// 真正的fieldName, column值可能是别名
	fieldName = esPtr.findAliasField(fieldName)

	// 如果是text字段，桶聚合需要使用fName.keyword
	ok, err := esPtr.isTextField(fieldName)
	if err != nil {
		return nil, err
	}
	kFieldName := fieldName
	if ok {
		kFieldName = fieldName + ".keyword"
	}

	// (1) 不同类型，对应不同的size和sort

	switch fc.Name {
	case TopFName:
		size = fc.Param[1].String()
		inner := map[string]string{
			"order":         "desc",
			"missing":       "_first",
			"unmapped_type": "string",
		}
		outer = SIMAP{kFieldName: inner}

	case BottomFName:
		size = fc.Param[1].String()
		inner := map[string]string{
			"order":         "asc",
			"missing":       "_first",
			"unmapped_type": "string",
		}
		outer = SIMAP{kFieldName: inner}

	case FirstFName:

		inner := map[string]string{
			"order":         "asc",
			"missing":       "_first",
			"unmapped_type": "string",
		}
		outer = SIMAP{esPtr.esNamespace.timeField: inner}

	case LastFName:

		inner := map[string]string{
			"order":         "desc",
			"missing":       "_last",
			"unmapped_type": "string",
		}
		outer = SIMAP{esPtr.esNamespace.timeField: inner}
	}

	sort = append(sort, outer)
	// first, last add sort param
	sort = updateSortList(fc, sort)
	source = append(source, esPtr.esNamespace.timeField)
	source = append(source, fieldName)
	esPtr.appendSortFields(esPtr.esNamespace.timeField) // 添加time字段
	esPtr.SortFields = append(esPtr.SortFields, fieldName)

	// (2) 字段列表
	res["_source"] = source
	res["size"] = size
	res["sort"] = sort
	return res, nil

}

// updateSortList 更新sort列表，例如: last(f1, sort=['f2:desc'])
func updateSortList(fc *FuncExpr, sort []interface{}) []interface{} {

	if fc.Name == FirstFName || fc.Name == LastFName {
		if len(fc.Param) > 1 {
			fArg := fc.Param[1].(*FuncArg)
			argVals := fArg.ArgVal.(FuncArgList)
			for _, item := range argVals {
				str := GetStringParam(item)
				split := strings.Split(str, ":")
				orderVal := split[len(split)-1]
				fieldName := strings.Join(split[0:len(split)-1], ":")
				inner := map[string]string{
					"order":         orderVal,
					"missing":       SortMissing[orderVal],
					"unmapped_type": "string",
				}
				outer := SIMAP{fieldName: inner}
				sort = append(sort, outer)
			}
		}
	}
	return sort
}

// genTophits 针对top,bottom,first,last使用tophits实现
func genTopHitsItem(esPtr *EST, fc *FuncExpr, alias string) (*aggsItem, string, error) {

	args, err := genTopHitsArgs(esPtr, fc)
	if err != nil {
		return nil, "", err
	}

	var res = aggsItem{
		fName: AggsBucketFuncs[AggsTophits],
		tName: "metric",
		args:  args,
	}

	fieldName := GetStringParam(fc.Param[0])

	if alias != "" {
		res.alias = alias
		// top函数，需要将field列名替换为函数别名
		esPtr.AliasSlice[0][fieldName] = alias
		esPtr.AliasSlice[1][alias] = fieldName
	} else {
		res.alias = fc.Name + "_" + fieldName
	}

	return &res, res.alias, nil
}

// genDistinctTophitsNodes
func genDistinctTophitsNodes(m *DFQuery, esPtr *EST) (*aggsNode, error) {
	var res = new(aggsNode)

	args := SIMAP{}
	// （1）args _source
	source := []string{esPtr.esNamespace.timeField}
	source = append(source, esPtr.distinctField)
	args["_source"] = source
	// （2）args size
	args["size"] = DefaultTophitLimit
	esPtr.SortFields = append(esPtr.SortFields, esPtr.esNamespace.timeField)
	esPtr.SortFields = append(esPtr.SortFields, esPtr.distinctField)

	var topAgg = aggsItem{
		fName: AggsBucketFuncs[AggsTophits],
		tName: "metric",
		args:  args,
	}
	topAgg.alias = AggsBucketFuncs[AggsTophits]
	res.items = []aggsItem{topAgg}

	return res, nil
}

// parse 解析聚合层
func (node *aggsNode) parse(level int, ptr *SIMAP) error {
	aggs := map[string]SIMAP{}
	strLevel := strconv.Itoa(level)

	for index, v := range node.items {
		aggsName := ""
		strIndex := strconv.Itoa(index)
		if v.alias != "" {
			aggsName = v.alias //请求中带有聚合别名
		} else {
			aggsName = strings.Join( //自动生成
				[]string{v.tName, v.fName, strLevel, strIndex},
				"-",
			)
		}
		inner := SIMAP{v.fName: v.args}
		aggs[aggsName] = inner
	}

	if *ptr != nil {
		for _, v := range aggs {
			v[AggsIdentifier] = (*ptr)[AggsIdentifier]
			break
		}
	}

	*ptr = SIMAP{
		AggsIdentifier: aggs,
	}

	return nil
}

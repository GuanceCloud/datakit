package parser

import (
	"fmt"
	"strings"
)

var (
	objectNamespace = ESNamespace{
		abbrName:   OABBRNAME,
		fullName:   OFULLNAME,
		classField: OCLASS,
		timeField:  OTIME,
	}

	loggingNamespace = ESNamespace{
		abbrName:   LABBRNAME,
		fullName:   LFULLNAME,
		classField: LCLASS,
		timeField:  LTIME,
	}

	eventNamespace = ESNamespace{
		abbrName:   EABBRNAME,
		fullName:   EFULLNAME,
		classField: ECLASS,
		timeField:  ETIME,
	}

	tracingNamespace = ESNamespace{
		abbrName:   TABBRNAME,
		fullName:   TFULLNAME,
		classField: TCLASS,
		timeField:  TTIME,
	}

	rumNamespace = ESNamespace{
		abbrName:   RABBRNAME,
		fullName:   RFULLNAME,
		classField: RCLASS,
		timeField:  RTIME,
	}

	securityNamespace = ESNamespace{
		abbrName:   SABBRNAME,
		fullName:   SFULLNAME,
		classField: SCLASS,
		timeField:  STIME,
	}

	ESNamespaces = []*ESNamespace{
		&objectNamespace,
		&loggingNamespace,
		&eventNamespace,
		&tracingNamespace,
		&rumNamespace,
		&securityNamespace,
	}
)

func (m *DFQuery) checkValid() (*EST, error) {
	var err error
	// (1) basic check
	err = basicCheck(m)
	if err != nil {
		return nil, err
	}
	// (2) 初始化EST
	esPtr, err := m.initEST()
	if err != nil {
		return nil, err
	}
	// (3) 更新EST
	err = checkAndSetEST(m, esPtr)
	if err != nil {
		return nil, err
	}
	return esPtr, nil
}

func (m *DFQuery) initEST() (*EST, error) {
	res := new(EST)

	res.dfState = new(DFState)
	res.esNamespace = new(ESNamespace)
	res.limitSize = ZeroLimit
	res.fromSize = DefaultOffset
	res.groupFields = []string{}
	res.histogramInfo = map[string]interface{}{} // 直方图
	res.distinctField = ""
	res.groupOrders = map[string]string{}

	res.AliasSlice = []map[string]string{
		map[string]string{}, // k:v = fieldName: aliasName
		map[string]string{}, // k:v = aliasName: fieldName
	}
	res.ClassNames = ""
	res.SortFields = []string{}
	res.HighlightFields = []string{}
	return res, nil
}

// 判断是否有时间范围查询
func checkTimeRangeQuery(m *DFQuery, esPtr *EST) error {
	// 如果是object数据，查询不需要加上时间范围
	// if m.TimeRange != nil && (esPtr.esNamespace.fullName != OFULLNAME) {
	if m.TimeRange != nil {
		esPtr.dfState.timeRangeQuery = true
		return nil
	}
	esPtr.dfState.timeRangeQuery = false
	return nil
}

// 判断是否有聚合
func checkDfstateAggs(m *DFQuery, esPtr *EST) error {
	dateHg := m.existDateHgAggs(esPtr)
	bucket := m.existBucketAggs(esPtr)
	metric := m.existMetricAggs(esPtr)
	res := dateHg || bucket || metric
	esPtr.dfState.aggs = res
	return nil
}

// 判断是否有时间聚合
func (m *DFQuery) existDateHgAggs(esPtr *EST) bool {
	// if m.TimeRange != nil && m.TimeRange.Resolution != nil && (esPtr.esNamespace.fullName != OFULLNAME) {
	if m.TimeRange != nil && m.TimeRange.Resolution != nil {
		esPtr.dfState.dateHg = true
		return true
	}
	esPtr.dfState.dateHg = false
	return false
}

// 判断是否有桶聚合
func (m *DFQuery) existBucketAggs(esPtr *EST) bool {
	// 判断是否存在histogram
	for _, i := range m.Targets {
		s := i.Col
		if fc, ok := s.(*FuncExpr); ok {
			if fc.Name == HistogramFName {
				esPtr.dfState.hasHistogram = true
				esPtr.dfState.bucket = true
				return true
			}
		}
	}
	// 查看group by list
	if m.GroupBy != nil {
		esPtr.dfState.bucket = true
		return true
	}
	esPtr.dfState.bucket = false
	return false
}

// 判断是否有指标聚合或者topFuncs
func (m *DFQuery) existMetricAggs(esPtr *EST) bool {
	for _, i := range m.Targets {
		s := i.Col
		if fc, ok := s.(*FuncExpr); ok {
			if _, found := AggsMetricFuncs[fc.Name]; found {
				esPtr.dfState.metric = true
				if fc.Name == DistinctFName {
					esPtr.dfState.hasDistinct = true
				}
				return true
			}
			if find := findStringSlice(fc.Name, TopFuncs); find {
				esPtr.dfState.metric = true
				esPtr.dfState.hasTopFuncs = true
				return true
			}
		}
	}
	esPtr.dfState.metric = false
	return false
}

// 检查namespace
func checkNamespace(m *DFQuery) error {
	// 如果namespace 是缩写，转换为全称
	find := false
	for _, nPtr := range ESNamespaces {
		if m.Namespace == nPtr.abbrName {
			m.Namespace = nPtr.fullName
			find = true
			break
		} else {
			if m.Namespace == nPtr.fullName {
				find = true
				break
			}
		}
	}
	if find == false {
		return fmt.Errorf("invalid namespace")
	}
	return nil
}

// 检查是否包含指标集
func checkMetricList(m *DFQuery) error {
	if m.Names == nil && m.RegexNames == nil {
		return fmt.Errorf("no metrics")
	}
	return nil
}

// 基本检查
func basicCheck(m *DFQuery) error {
	var err error
	checkFuncs := []func(*DFQuery) error{
		checkNamespace,  // (1) 判断namespace是否正确
		checkMetricList, // (2) 判断是否有指标集
	}
	for _, f := range checkFuncs {
		err = f(m)
		if err != nil {
			return err
		}
	}
	return nil
}

// checkNestAggFuncs 判断内置聚合函数
func checkNestAggFuncs(fName string, fc *StaticCast) error {
	// 判断是否支持内置函数
	if _, ok := NestAggFuncs[fName]; !ok {
		return fmt.Errorf("%s func unsupport nest func", fName)
	}
	// 判断是否是合法的内置函数
	ifName := ""
	if fc.IsInt {
		ifName = IntFName
	} else {
		if fc.IsFloat {
			ifName = FloatFName
		}
	}
	if find := findStringSlice(ifName, NestAggFuncs[fName]); !find {
		return fmt.Errorf("%s func unsupport nest func %s", fName, ifName)
	}
	// // 判断内层函数参数
	// if len(fc.Param) != 1 {
	// 	return fmt.Errorf("%s func should have and at most one field name", fc.Name)
	// }
	// if IsStringParam(fc.Param[0]) == false {
	// 	return fmt.Errorf("%s func first param should be field name", fc.Name)
	// }
	return nil
}

// checkScriptNestAggFuncs 判断聚合函数中的内置script函数
func checkScriptNestAggFuncs(fName string, fc *FuncExpr) error {
	// 判断是否支持script函数
	find := findStringSlice(fName, NestAggFuncs[ScriptFName])
	if !find {
		return fmt.Errorf("%s func unsupport nest func %s", fName, ScriptFName)
	}
	// 判断script函数参数
	if len(fc.Param) != 1 {
		return fmt.Errorf("%s func should have and at most one field name", fc.Name)
	}
	if IsStringParam(fc.Param[0]) == false {
		return fmt.Errorf("%s func first param should be field name", fc.Name)
	}
	return nil
}

// checkAggsFuncs 检查是否是合法的聚合函数
func checkAggsFuncs(m *DFQuery, esPtr *EST) error {

	for _, t := range m.Targets {
		if fc, ok := (t.Col).(*FuncExpr); ok {
			found := false // 针对每个函数判断
			// 1) 判断是否是指标聚合
			if _, ok := AggsMetricFuncs[fc.Name]; ok {
				found = true
				err := checkAggsESFuncs(fc)
				if err != nil {
					return err
				}
				continue // check valid, go to next check
			}
			// 2) 判断是否是top等函数
			if find := findStringSlice(fc.Name, TopFuncs); find {
				found = true
				err := checkAggsTopFuncs(esPtr, fc)
				if err != nil {
					return err
				}
				continue // check valid, go to next check
			}
			// 3) 判断是否是histogram函数
			if fc.Name == HistogramFName {
				found = true
				err := checkAggsHistogramFunc(esPtr, fc)
				if err != nil {
					return err
				}
				continue // check valid, go to next check
			}
			if !found {
				return fmt.Errorf("unsupport func %s", fc.Name)
			}
		}
	}
	return nil
}

// checkGroupByFields 检查groupby聚合字段
func checkGroupByFields(m *DFQuery, esPtr *EST) error {
	if esPtr.dfState.bucket {
		if esPtr.dfState.hasHistogram {
			// histogram 和 groupby 不能同时使用
			if m.GroupBy != nil && len(m.GroupBy.List) > 0 {
				return fmt.Errorf("can not use group by and histogram at once")
			}
			// histogram函数不能与其他函数共用
			if len(m.Targets) > 1 {
				return fmt.Errorf("can not use histogram with other func")
			}
		} else {
			fields := []string{}
			lenField := len(m.GroupBy.List)
			if lenField > BucketDepthSize { // 只支持3层桶聚合
				return fmt.Errorf("bucket aggs depth limit is 2")
			}
			for _, i := range m.GroupBy.List {
				if IsStringParam(i) == false {
					// 只支持对field桶聚合
					return fmt.Errorf("only support field with group by")
				}
				fieldName := GetStringParam(i)
				if find := findStringSlice(fieldName, fields); find {
					// 多层聚合，不能对相同的field
					return fmt.Errorf("each field can only be grouped once")
				}
				fields = append(fields, fieldName)
				esPtr.groupFields = fields
			}
		}

	}
	return nil
}

// setAggsOrders 设置桶聚合相关的order
func checkAggsOrders(m *DFQuery, esPtr *EST) error {
	// 如果存在order by
	if m.OrderBy != nil {
		for _, item := range m.OrderBy.List {
			elem, ok := item.(*OrderByElem)
			if !ok {
				return nil
			}
			fieldName := elem.Column
			esPtr.groupOrders[fieldName] = SortMap[elem.Opt]
		}
	}
	return nil
}

// checkLimitSize 检查limit size
func checkLimitSize(m *DFQuery, esPtr *EST) error {
	// (1）有聚合，size limit最大值为1000
	if esPtr.dfState.aggs {
		return esPtr.checkAggsLimitSize(m)

	}
	// (2) 没有聚合
	return esPtr.checkQueryLimitSize(m)
}

func (esPtr *EST) checkAggsLimitSize(m *DFQuery) error {
	if m.Limit == nil {
		if esPtr.dfState.hasDistinct {
			esPtr.limitSize = DefaultLimit
		} else {
			esPtr.limitSize = DefaultGroupLimit
		}

	} else {
		inputLimit := int(m.Limit.Limit)
		if inputLimit > MaxGroupLimit {
			// esPtr.limitSize = MaxGroupLimit // ast 会自动填充
			return fmt.Errorf("aggs max limit size is %d, but you set %d", MaxGroupLimit, inputLimit)
		}
		esPtr.limitSize = inputLimit
	}
	return nil
}

func (esPtr *EST) checkQueryLimitSize(m *DFQuery) error {
	if m.Limit == nil {
		esPtr.limitSize = DefaultLimit
	} else {
		inputLimit := int(m.Limit.Limit)
		if inputLimit > MaxLimit {
			return fmt.Errorf("query max limit size is %d, but you set %d", MaxLimit, inputLimit)
		}
		esPtr.limitSize = inputLimit
	}
	return nil
}

// checkFromSize 检查from size
func checkFromSize(m *DFQuery, esPtr *EST) error {
	if esPtr.dfState.aggs { // 有聚合，设置from 为0
		return esPtr.checkAggsFromSize(m)
	}
	return esPtr.checkQueryFromSize(m)
}

func (esPtr *EST) checkAggsFromSize(m *DFQuery) error {
	if m.Offset == nil {
		esPtr.fromSize = DefaultOffset
	} else {
		fromSize := int(m.Offset.Offset)
		// query offset + query limit <= 1000
		if fromSize+esPtr.limitSize > MaxGroupLimit {
			return fmt.Errorf("aggs offset: %d add limit: %d should small than %d",
				fromSize, esPtr.limitSize, MaxGroupLimit,
			)
		}
		esPtr.fromSize = fromSize
	}
	return nil
}

func (esPtr *EST) checkQueryFromSize(m *DFQuery) error {
	if m.Offset != nil {
		fromSize := int(m.Offset.Offset)
		// query offset + query limit <= 10000
		if fromSize+esPtr.limitSize > MaxOffset {
			return fmt.Errorf("query offset: %d add limit: %d should small than %d",
				fromSize, esPtr.limitSize, MaxOffset,
			)
		}
		esPtr.fromSize = fromSize
	}
	return nil
}

// check with EST
func checkAndSetEST(m *DFQuery, esPtr *EST) error {
	var err error
	// (1) 初始化状态后，检查
	checkFuncs := []func(*DFQuery, *EST) error{
		checkESNamespace,
		checkDfstateAggs,
		checkTimeRangeQuery,
		checkAggsFuncs,     // 检查是否是合法函数
		checkGroupByFields, // 检查桶聚合
		checkLimitSize,     // 检查并且设置limit size
		checkFromSize,      // 检查并且设置offset
		checkAggsOrders,    // 检查设置 groupOrders
		checkHighlight,     // 检查设置 highlight
	}
	for _, f := range checkFuncs {
		err = f(m, esPtr)
		if err != nil {
			return err
		}
	}
	return nil
}

// checkESNamespace 设置timeField, classField
func checkESNamespace(m *DFQuery, esPtr *EST) error {
	for _, nPtr := range ESNamespaces {
		if nPtr.fullName == m.Namespace {
			esPtr.esNamespace = nPtr
			break
		}
	}
	return nil
}

// checkHighlight 设置高亮
func checkHighlight(m *DFQuery, esPtr *EST) error {
	esPtr.IsHighlight = m.Highlight
	return nil
}

// checkAggsTopFuncs 检查top等函数
func checkAggsTopFuncs(esPtr *EST, fc *FuncExpr) error {
	lenParam := len(fc.Param)
	switch fc.Name {
	case TopFName, BottomFName:
		//必须含有size的值
		if lenParam != 2 {
			return fmt.Errorf("%s func must have two params, one field name, one size", fc.Name)
		}
		if IsStringParam(fc.Param[0]) == false {
			// 第一个参数是字符串
			return fmt.Errorf("%s func first param should be field name", fc.Name)
		}
		if GetStringParam(fc.Param[0]) == esPtr.esNamespace.timeField {
			// 第一个参数不能是time字段
			return fmt.Errorf("%s func first param can not be time field", fc.Name)
		}
		if _, ok := fc.Param[1].(*NumberLiteral); !ok {
			// 第二个参数是数值
			return fmt.Errorf("%s func second param should be size value", fc.Name)
		}
	case FirstFName, LastFName:
		// first(f1, sort=['f2:desc'])
		if lenParam > 2 {
			return fmt.Errorf("%s func should have and at most two field name", fc.Name)
		}
		if IsStringParam(fc.Param[0]) == false {
			// 第一个参数是字符串
			return fmt.Errorf("%s func first param should be field name", fc.Name)
		}
		if GetStringParam(fc.Param[0]) == esPtr.esNamespace.timeField {
			// 第一个参数不能是time字段
			return fmt.Errorf("%s func first param can not be time field", fc.Name)
		}
		if lenParam > 1 {
			fArg, ok := fc.Param[1].(*FuncArg)
			// check func arg name
			if !ok || (ok && fArg.ArgName != "sort") {
				return fmt.Errorf("%s func second param should be named param sort", fc.Name)
			}
			// check func arg vals
			argVals, ok := fArg.ArgVal.(FuncArgList)
			if !ok || (ok && len(argVals) == 0) {
				return fmt.Errorf("%s func second param invalid, should like sort=['f1:asc']", fc.Name)
			}
			for _, item := range argVals {
				if IsStringParam(item) == false {
					return fmt.Errorf("%s func second param invalid, should like sort=['f1:asc']", fc.Name)
				}
				str := GetStringParam(item)
				split := strings.Split(str, ":")
				if len(split) < 2 {
					return fmt.Errorf("%s func second param invalid, should like sort=['f1:asc']", fc.Name)
				}
				if _, ok := SortMissing[split[len(split)-1]]; !ok {
					return fmt.Errorf("%s func sort type must be asc or desc", fc.Name)
				}
			}

		}
		esPtr.flFuncCount = esPtr.flFuncCount + 1
	}
	return nil
}

// checkAggsESFuncs 检查es支持的原生函数
func checkAggsESFuncs(fc *FuncExpr) error {
	lenParam := len(fc.Param)

	// percent 函数有2个参数, fieldname 和 size
	if fc.Name == PercentileFName {
		if lenParam != 2 {
			return fmt.Errorf("%s func must have two params, one field name, one size, size range [0, 100]", fc.Name)
		}
		if IsStringParam(fc.Param[0]) == false {
			// 第一个参数是字符串
			return fmt.Errorf("%s func first param should be field name", fc.Name)
		}
		if _, ok := fc.Param[1].(*NumberLiteral); !ok {
			// 第二个参数是数值
			return fmt.Errorf("%s func second param should be size value, size range [0, 100]", fc.Name)
		}
		return nil
	}

	//必须含有size的值
	if lenParam != 1 {
		return fmt.Errorf("%s func should have and at most one field name", fc.Name)
	}
	if ifc, ok := fc.Param[0].(*StaticCast); ok {
		// 检查内置聚合函数
		return checkNestAggFuncs(fc.Name, ifc)
	}
	// 内置函数为script
	if ifc, ok := fc.Param[0].(*FuncExpr); ok {
		if ifc.Name != ScriptFName {
			return fmt.Errorf("%s func only support script nested func", fc.Name)
		}
		return checkScriptNestAggFuncs(fc.Name, ifc)
	}
	if IsStringParam(fc.Param[0]) == false {
		return fmt.Errorf("%s func first param should be field name", fc.Name)
	}
	return nil
}

// checkAggsHistogramFunc 检查histogram函数
func checkAggsHistogramFunc(esPtr *EST, fc *FuncExpr) error {
	lenParam := len(fc.Param)
	// histogram(fieldName, [startValue:endValue:interval])
	if lenParam < 4 {
		return fmt.Errorf("%s func must have two params, one field name, three value range", fc.Name)
	}
	if IsStringParam(fc.Param[0]) == false {
		// 第一个参数是字符串
		return fmt.Errorf("%s func first param should be field name", fc.Name)
	}
	// 剩下3个参数为开始值、结束值、间隔(可能有第4个缺省参数，指定minDoc值)
	for i, v := range fc.Param[1:] {
		if _, ok := v.(*NumberLiteral); !ok {
			return fmt.Errorf("%s func the %d param should be number value", fc.Name, i)
		}
	}

	start := GetFloatParam(fc.Param[1])
	end := GetFloatParam(fc.Param[2])
	interval := GetFloatParam(fc.Param[3])
	// 分桶不超过10000
	if interval <= 0 || ((end-start)/interval > float64(MaxBucket)) {
		return fmt.Errorf("%s func, interval must > 0 and (end-start)/interval < %d", fc.Name, MaxBucket)
	}

	esPtr.histogramInfo["fieldName"] = GetStringParam(fc.Param[0])
	esPtr.histogramInfo["start"] = start
	esPtr.histogramInfo["end"] = end
	esPtr.histogramInfo["interval"] = interval
	// 如果minDoc设置
	if lenParam == 5 {
		minDoc := GetFloatParam(fc.Param[4])
		esPtr.histogramInfo["minDoc"] = minDoc
	}
	return nil
}

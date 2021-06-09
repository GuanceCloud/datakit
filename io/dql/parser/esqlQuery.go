package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	// QueryFuncs 查询函数集合
	QueryFuncs = map[string]string{
		"=":           "regexp",
		"match":       "match",
		"script":      "script",
		"wildcard":    "wildcard",
		"querystring": "query_string",
		"exists":      "exists",
	}

	// DefaultLogicOperator query_string, default logic operator
	DefaultLogicOperator = "AND"

	// AnalyzeWildcard queryString 是否可以使用通配方式
	AnalyzeWildcard = true

	// ArithmeticMap 算术运算符的映射
	ArithmeticMap = map[string]string{
		">":  "gt",
		"<":  "lt",
		">=": "gte",
		"<=": "lte",
	}

	// LogicMap 逻辑运算符的映射
	LogicMap = map[string]string{
		"and": "must",
		"or":  "should",
	}

	// LogicNotMap 逻辑运算符not
	LogicNotMap = map[string]string{
		"!=": "must_not",
	}

	// TermMap 精确查询映射
	TermMap = map[string]string{
		"=": "term",
	}

	// InQFName in 函数
	InQFName = "in"

	// ReservedChars 正则特殊字符
	ReservedChars = map[rune]rune{
		'{':  ' ',
		'}':  ' ',
		'[':  ' ',
		']':  ' ',
		'(':  ' ',
		')':  ' ',
		'\\': ' ',
		'/':  ' ',
		':':  ' ',
		'"':  ' ',
		'.':  ' ',
		'?':  ' ',
		'+':  ' ',
		'*':  ' ',
		'|':  ' ',
	}
	// ObjectTimeInterval 对象采集周期
	ObjectTimeInterval = "5m"

	// HighlightFragmentSize fragmentSize 高亮的分段大小, 默认是100，如果有长字段会缺失
	HighlightFragmentSize = 1000000
)

type esQlFunc struct {
	fName string // 函数名称
	// aType string      // 函数参数类型(可变列表/kv对)
	args interface{} // 函数参数(形式归一为 kv对)
}

// queryTransport query(查询)部分的转换
func queryTransport(m *DFQuery, esPtr *EST) (SIMAP, error) {
	var queryPart = SIMAP{}
	var arrayRes []interface{}

	// 添加 class过滤器
	classRes, err := classTermQuery(m, esPtr)

	if err != nil {
		return queryPart, err
	}

	// 如果为re(`.*`), classRes为nil
	if classRes != nil {
		arrayRes = append(arrayRes, classRes)
	}

	// 对象时间范围特别处理，只能查询最近时间
	if esPtr.esNamespace.fullName == OFULLNAME {
		timeRes, err := objectTimeRangeQuery(m, esPtr)
		if err != nil {
			return queryPart, err
		}
		for _, v := range timeRes {
			arrayRes = append(arrayRes, v)
		}

	} else {
		// 添加时间范围过滤器
		if esPtr.dfState.timeRangeQuery {
			timeRes, err := timeRangeQuery(m, esPtr)
			if err != nil {
				return queryPart, err
			}
			for _, v := range timeRes {
				arrayRes = append(arrayRes, v)
			}
		}
	}

	// 添加 filter过滤器
	filters, err := filterQuery(m, esPtr)
	if err != nil {
		return queryPart, err
	}
	for _, i := range filters {
		arrayRes = append(arrayRes, i)
	}

	// 添加 exist 过滤器(针对top, bottom, first, last 函数)
	if esPtr.dfState.hasTopFuncs {
		filters, err := existQuery(m, esPtr)
		if err != nil {
			return queryPart, err
		}
		for _, i := range filters {
			arrayRes = append(arrayRes, i)
		}
	}

	// 添加 histogram 范围的过滤器
	if esPtr.dfState.hasHistogram {
		filter, err := histogramRangeQuery(m, esPtr)
		if err != nil {
			return queryPart, err
		}
		arrayRes = append(arrayRes, filter)
	}

	if len(arrayRes) > 0 {
		var mustQuery = SIMAP{
			"must": arrayRes,
		}
		var boolQuery = SIMAP{
			"bool": mustQuery,
		}
		return boolQuery, nil
	}
	return nil, nil
}

// filterQuery filter过滤器
func filterQuery(m *DFQuery, esPtr *EST) ([]interface{}, error) {
	var arrayRes []interface{}
	for _, node := range m.WhereCondition {
		if pNode, ok := node.(*ParenExpr); ok {
			s, err := pNode.esQL(esPtr)
			if err != nil {
				return arrayRes, err
			}
			arrayRes = append(arrayRes, s)
		} else {
			if bNode, ok := node.(*BinaryExpr); ok {
				s, err := bNode.esQL(esPtr)
				if err != nil {
					return arrayRes, err
				}
				arrayRes = append(arrayRes, s)
			}
		}
	}
	return arrayRes, nil
}

// existQuery exist过滤器
func existQuery(m *DFQuery, esPtr *EST) ([]interface{}, error) {
	var arrayRes []interface{}
	for _, t := range m.Targets {
		if fc, ok := (t.Col).(*FuncExpr); ok {
			switch fc.Name {
			case TopFName, BottomFName, FirstFName, LastFName:
				fieldName := GetStringParam(fc.Param[0])
				inner := map[string]string{
					"field": fieldName,
				}
				outer := map[string]interface{}{
					"exists": inner,
				}
				arrayRes = append(arrayRes, outer)
			}
		}
	}
	return arrayRes, nil
}

// histogramRangeQuery 直方图的x轴过滤
func histogramRangeQuery(m *DFQuery, esPtr *EST) (interface{}, error) {

	iinner := SIMAP{
		"gte": esPtr.histogramInfo["start"],
		"lte": esPtr.histogramInfo["end"],
	}
	fName, ok := esPtr.histogramInfo["fieldName"].(string)
	if !ok {
		return nil, fmt.Errorf("histogram param error")
	}
	inner := SIMAP{
		fName: iinner,
	}
	outer := SIMAP{
		"range": inner,
	}
	return outer, nil
}

// classTermQuery 类型的过滤器
func classTermQuery(m *DFQuery, esPtr *EST) (SIMAP, error) {
	var shoulds = []SIMAP{}
	namesSlice := []string{}
	className := esPtr.esNamespace.classField

	// re(`.*`),为了检索性能, 不需要使用regex过滤
	if m.RegexNames != nil && len(m.RegexNames) == 1 {
		r, err := m.RegexNames[0].ESQL()
		if err != nil {
			return nil, err
		}
		str, err := voidToString(r)
		if err != nil {
			return nil, err
		}
		if str == ".*" {
			return nil, nil
		}
	}

	// 非 re(`.*`)
	if m.Names != nil {
		for _, v := range m.Names {
			s, err := esPtr.basicTermQuery("term", className, v, QueryValue)
			if err != nil {
				return nil, err
			}
			item, _ := s.(SIMAP)
			namesSlice = append(namesSlice, v)
			shoulds = append(shoulds, item)
		}
	}
	if m.RegexNames != nil {

		for _, v := range m.RegexNames {
			r, err := v.ESQL()
			if err != nil {
				return nil, err
			}
			str, err := voidToString(r)
			if err != nil {
				return nil, err
			}
			s, err := esPtr.basicTermQuery("regexp", className, str, QueryValue)
			if err != nil {
				return nil, err
			}
			if item, ok := s.(SIMAP); ok {
				namesSlice = append(namesSlice, str)
				shoulds = append(shoulds, item)
			} else {
				return nil, fmt.Errorf("can not get class query")
			}
		}

	}
	// ClassNames, 分类名称
	esPtr.ClassNames = strings.Join(namesSlice, ", ")

	// bool should
	var inner = SIMAP{"should": shoulds}
	return SIMAP{"bool": inner}, nil

}

// timeRangeQuery 时间范围的过滤器
func timeRangeQuery(m *DFQuery, esPtr *EST) ([]interface{}, error) {
	var start, end interface{}
	if m.TimeRange.Start == nil {
		start = time.Now().UnixNano() / int64(time.Millisecond)
	} else {
		start, _ = m.TimeRange.Start.ESQL() // 单位:毫秒
	}
	if m.TimeRange.End == nil {
		end = time.Now().UnixNano() / int64(time.Millisecond)
	} else {
		end, _ = m.TimeRange.End.ESQL() // 单位:毫秒
	}
	esPtr.StartTime = start.(int64)
	esPtr.EndTime = end.(int64)
	s, _ := rangeTermQuery("gte", esPtr.esNamespace.timeField, strconv.FormatInt(start.(int64), 10))
	e, _ := rangeTermQuery("lte", esPtr.esNamespace.timeField, strconv.FormatInt(end.(int64), 10))
	return []interface{}{s, e}, nil
}

// objectTimeRangeQuery 对象时间范围的过滤器
func objectTimeRangeQuery(m *DFQuery, esPtr *EST) ([]interface{}, error) {
	var start, end int64

	if m.TimeRange != nil {
		now := time.Now()
		if m.TimeRange.Start == nil {
			start = now.UnixNano() / int64(time.Millisecond)
		} else {
			start = m.TimeRange.Start.Time.UnixNano() / int64(time.Millisecond)
		}
		if m.TimeRange.End == nil {
			end = now.UnixNano() / int64(time.Millisecond)
		} else {
			end = m.TimeRange.End.Time.UnixNano() / int64(time.Millisecond)
		}

	} else {
		te := time.Now()
		interval, _ := time.ParseDuration(ObjectTimeInterval)
		ts := te.Add(-1 * interval)
		start = ts.UnixNano() / int64(time.Millisecond)
		end = te.UnixNano() / int64(time.Millisecond)
	}

	esPtr.StartTime = start
	esPtr.EndTime = end
	s, _ := rangeTermQuery("gte", esPtr.esNamespace.timeField, strconv.FormatInt(start, 10))
	e, _ := rangeTermQuery("lte", esPtr.esNamespace.timeField, strconv.FormatInt(end, 10))
	return []interface{}{s, e}, nil
}

// boolCompoundQuery, 复合bool查询
func boolCompoundQuery(v string, l interface{}, r interface{}) (interface{}, error) {
	iinter := []interface{}{
		l, r,
	}
	inter := map[string][]interface{}{
		v: iinter,
	}
	outer := SIMAP{
		"bool": inter,
	}
	return outer, nil
}

// boolNotQuery, must_not查询
func (esPtr *EST) boolNotQuery(v string, l interface{}, r interface{}) (interface{}, error) {
	s, err := esPtr.basicTermQuery("term", l, r, QueryValue)
	if err != nil {
		return nil, err
	}
	iinter := []interface{}{s}
	inter := map[string][]interface{}{
		v: iinter,
	}
	outer := SIMAP{
		"bool": inter,
	}
	return outer, nil
}

// 基本的term精确值匹配
func (esPtr *EST) basicTermQuery(fName string, lName, rValue interface{}, rName string) (interface{}, error) {
	strLName, err := voidToString(lName)
	if err != nil {
		return nil, err
	}
	strRValue, err := voidToString(rValue)
	if err != nil {
		return nil, err
	}
	// match 查询, 如果没有聚合并且需要高亮，添加高亮字段
	if fName == "match" {
		if (!esPtr.dfState.aggs) && esPtr.IsHighlight {
			esPtr.HighlightFields = append(esPtr.HighlightFields, strLName)
		}
	}
	// 正则匹配字符串长度限制, 截断处理
	if fName == "regexp" {
		if len(strRValue) >= RegexLimit {
			runes := []rune(strRValue[:RegexLimit])
			strRValue = string(runes[:len(runes)-1])
		}
	}
	// 非match查询, 如果是text字段，需要使用keyword
	if rName != "query" {
		ok, err := esPtr.isTextField(strLName)
		if err != nil {
			return nil, err
		}
		if ok {
			strLName = strLName + ".keyword"
		}
	}
	iinter := SIMAP{
		rName: strRValue,
	}
	inter := SIMAP{
		strLName: iinter,
	}
	outer := SIMAP{
		fName: inter,
	}
	return outer, nil
}

func rangeTermQuery(v string, l interface{}, r interface{}) (interface{}, error) {
	strL, err := voidToString(l)
	if err != nil {
		return nil, err
	}
	strR, err := voidToString(r)
	if err != nil {
		return nil, err
	}

	iinter := SIMAP{
		v: strR,
	}
	inter := SIMAP{
		strL: iinter,
	}
	outer := SIMAP{
		"range": inter,
	}
	return outer, nil
}

// query string
func (esPtr *EST) queryStringQuery(v string, l, r interface{}) (interface{}, error) {

	var err error

	strLValue, err := voidToString(l)
	if err != nil {
		return nil, err
	}

	strRValue, err := voidToString(r)
	if err != nil {
		return nil, err
	}

	// update regexp
	strRValue = updateRegexp(strRValue)

	inter := SIMAP{
		"default_field":    strLValue,
		"query":            strRValue,
		"default_operator": DefaultLogicOperator,
		"analyze_wildcard": AnalyzeWildcard,
	}
	outer := SIMAP{
		"query_string": inter,
	}

	if (!esPtr.dfState.aggs) && esPtr.IsHighlight {
		esPtr.HighlightFields = append(esPtr.HighlightFields, strLValue)
	}
	return outer, nil
}

// updateRegexp 更新查询的字符串
func updateRegexp(s string) string {
	s = strings.ToLower(s)
	var rs []rune
	// 1) 只能有1000个字符长度
	if len(s) >= RegexLimit {
		rs = []rune(s[:RegexLimit])
		rs = rs[:len(rs)-1]
	} else {
		rs = []rune(s)
	}

	nRs := []rune{}
	// 2) 特殊字符替换为空
	for _, r := range rs {
		if nr, ok := ReservedChars[r]; ok {
			nRs = append(nRs, nr)
		} else {
			nRs = append(nRs, r)
		}
	}
	return string(nRs)
}

// handleInFunc
func (x *BinaryExpr) handleInFunc(esPtr *EST) (interface{}, error) {
	fieldName := ""
	// 左值为fieldName
	if IsStringParam(x.LHS) {
		fieldName = GetStringParam(x.LHS)
		ok, err := esPtr.isTextField(fieldName) // 如果是text类型，使用fieldName.keyword
		if err != nil {
			return nil, err
		}
		if ok {
			fieldName = fieldName + ".keyword"
		}
	} else {
		return nil, fmt.Errorf("query with in should have field name")
	}
	// 右值为字符串或者数值
	if _, ok := (x.RHS).(NodeList); !ok {
		return nil, fmt.Errorf("query with in should have field value")
	}
	return inFuncQuery(fieldName, (x.RHS).(NodeList))
}

// in 查询
// f1 in [v1, v2, v3]
func inFuncQuery(fieldName string, r NodeList) (interface{}, error) {
	termsQuery := []interface{}{}
	for _, v := range r {
		iinner := SIMAP{}
		if ok := IsStringParam(v); ok { // 字符串
			iinner = map[string]interface{}{
				"value": GetStringParam(v),
			}
		} else {
			if nv, ok := v.(*NumberLiteral); ok { // 数值
				if nv.IsInt {
					iinner = map[string]interface{}{
						"value": nv.Int,
					}
				} else {
					iinner = map[string]interface{}{
						"value": nv.Float,
					}
				}
			} else { // 其他
				return nil, fmt.Errorf("query with in, just support number or string item")
			}
		}

		inner := SIMAP{fieldName: iinner}

		outer := SIMAP{"term": inner}
		termsQuery = append(termsQuery, outer)
	}
	shouldQuery := SIMAP{"should": termsQuery}
	return SIMAP{"bool": shouldQuery}, nil
}

// ESQL 对BinaryExpr结构的解析
func (x *BinaryExpr) esQL(esPtr *EST) (interface{}, error) {
	var (
		isRegex bool
		v       string
		ok      bool
		err     error
		l, r    interface{}
	)
	// 前序遍历
	op := x.Op.String()

	// in查询
	if op == InQFName {
		return x.handleInFunc(esPtr)
	}

	if xRes, ok := (x.LHS).(*BinaryExpr); ok {
		l, err = xRes.esQL(esPtr)
	} else {
		if xRes, ok := (x.LHS).(*ParenExpr); ok {
			l, err = xRes.esQL(esPtr)
		} else {
			l, err = (x.LHS).ESQL()
		}
	}

	if _, regexOk := (x.RHS).(*Regex); regexOk {
		isRegex = true
	}
	if xRes, ok := (x.RHS).(*BinaryExpr); ok {
		r, err = xRes.esQL(esPtr)
	} else {
		if xRes, ok := (x.RHS).(*ParenExpr); ok {
			r, err = xRes.esQL(esPtr)
		} else {
			r, err = (x.RHS).ESQL()
		}
	}

	if err != nil {
		return "", err
	}

	// 获取索引时候的fieldName，
	// 如果使用aliasName，需要转为实际fileName
	if strL, ok := l.(string); ok {
		if k, ok := esPtr.AliasSlice[1][strL]; ok {
			l = k
		}
	}

	// 正则查询
	if isRegex {
		return esPtr.basicTermQuery(QueryFuncs[op], l, r, QueryValue)
	}
	// match,script query
	if esFunc, ok := r.(*esQlFunc); ok {
		return esFunc.handleFunc(esPtr, l)
	}
	if v, ok = QueryFuncs[op]; ok && v == "match" {

		return esPtr.basicTermQuery(v, l, r, "query")
	}

	// 逻辑查询(and, or), 即ES DSL中的bool复合查询
	if v, ok = LogicMap[op]; ok {
		return boolCompoundQuery(v, l, r)
	}

	// 逻辑查询(not)，即ES DSL中的must_not查询
	if v, ok = LogicNotMap[op]; ok {
		return esPtr.boolNotQuery(v, l, r)
	}

	// 范围查询, 即ES DSL中的range查询
	if v, ok = ArithmeticMap[op]; ok {
		return rangeTermQuery(v, l, r)
	}

	// 精确值查询, 即ES DSL中的term查询
	if v, ok = TermMap[op]; ok {
		return esPtr.basicTermQuery(v, l, r, QueryValue)
	}

	return "", nil
}

// filterFunc filter is func
func (esFunc *esQlFunc) handleFunc(esPtr *EST, l interface{}) (interface{}, error) {
	switch esFunc.fName {
	case "match":
		if args, ok := (esFunc.args).([]string); ok && len(args) > 0 {
			return esPtr.basicTermQuery(esFunc.fName, l, args[0], "query")
		}
		return nil, fmt.Errorf("match func params should be field name")
	case "wildcard":
		if args, ok := (esFunc.args).([]string); ok && len(args) > 0 {
			return esPtr.basicTermQuery(esFunc.fName, l, args[0], QueryValue)
		}
		return nil, fmt.Errorf("wildcard func params should have one param")
	case "query_string":
		if args, ok := (esFunc.args).([]string); ok && len(args) > 0 {
			return esPtr.queryStringQuery(esFunc.fName, l, args[0])
		}
		return nil, fmt.Errorf("queryString func params should have one param")
	case "script":
		l, err := voidToString(l)
		if err != nil {
			return nil, err
		}
		if l != "__script" {
			return nil, fmt.Errorf("script func format shoule be __script=script()")
		}
		if args, ok := (esFunc.args).([]string); ok && len(args) > 0 {
			return esPtr.scriptQuery(args[0])
		}
	case "exists":
		return esPtr.existsQuery(l)
	}
	return nil, fmt.Errorf("unsupport query func: %s", esFunc.fName)
}

// script filter
func (esPtr *EST) scriptQuery(rValue string) (interface{}, error) {
	inter := SIMAP{
		"script": rValue,
	}
	outer := SIMAP{
		"script": inter,
	}
	return outer, nil
}

// exists query
func (esPtr *EST) existsQuery(l interface{}) (interface{}, error) {
	sl, err := voidToString(l)
	if err != nil {
		return nil, err
	}
	inter := SIMAP{
		"field": sl,
	}
	outer := SIMAP{
		"exists": inter,
	}
	return outer, nil
}

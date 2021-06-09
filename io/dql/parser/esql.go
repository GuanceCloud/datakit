package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//================== const ===================
//"object", "O", "logging", "L", "event", "E", "tracing", "T", "rum", "R"

const (
	// object 相关

	// OABBRNAME 缩写
	OABBRNAME = "O"
	// OFULLNAME 全称
	OFULLNAME = "object"
	// OCLASS 对象类型 分类字段
	OCLASS = "class"
	// OTIME 对象类型 时间字段
	OTIME = "last_update_time"

	// logging 相关

	// LABBRNAME 缩写
	LABBRNAME = "L"
	// LFULLNAME 全称
	LFULLNAME = "logging"
	// LCLASS 对象类型 分类字段
	LCLASS = "source"

	// LTIME 对象类型 时间字段
	LTIME = "date"

	// keyevent 相关

	// EABBRNAME 缩写
	EABBRNAME = "E"
	// EFULLNAME 全称
	EFULLNAME = "event"
	// ECLASS 对象类型 分类字段
	ECLASS = "source"

	// ETIME 对象类型 时间字段
	ETIME = "date"

	// tracing 相关

	// TABBRNAME 缩写
	TABBRNAME = "T"
	// TFULLNAME 全称
	TFULLNAME = "tracing"
	// TCLASS 对象类型 分类字段
	TCLASS = "service"

	// TTIME 对象类型 时间字段
	TTIME = "date"

	// rum 相关

	// RABBRNAME 缩写
	RABBRNAME = "R"
	// RFULLNAME 全称
	RFULLNAME = "rum"
	// RCLASS 对象类型 分类字段
	RCLASS = "source"

	// RTIME 对象类型 时间字段
	RTIME = "date"

	// SABBRNAME security 缩写
	SABBRNAME = "S"
	// SFULLNAME security 全称
	SFULLNAME = "security"
	// SCLASS security 分类字段
	SCLASS = "category"
	// STIME security 时间字段
	STIME = "date"

	// TIMEDEFAULT 默认time字段名称
	TIMEDEFAULT = "time" // 默认的时间字段别名

	// DefaultDocID docid
	DefaultDocID = "__docid" // 文档唯一标识字段
)

// ESNamespace 包含 timeField, classField信息
type ESNamespace struct {
	abbrName   string
	fullName   string
	classField string
	timeField  string
}

// DFState 状态描述
type DFState struct {
	timeRangeQuery bool // 是否包含时间范围查询
	aggs           bool // 是否有聚合
	dateHg         bool // 是否有日期聚合
	bucket         bool // 是否有桶聚合
	metric         bool // 是否有指标聚合
	hasTopFuncs    bool // 非聚合，是否有top，bottom，first，last的查询
	hasDistinct    bool // 是否有distinct函数
	hasHistogram   bool // 是否有histogram函数
}

// EST ESTransInfo（对应ast）
type EST struct {
	dfState     *DFState     // DFState 状态描述
	esNamespace *ESNamespace // timeField, classField
	limitSize   int          // size
	fromSize    int          // offset

	// 函数
	groupFields   []string               // 桶聚合字段列表
	groupOrders   map[string]string      // 桶聚合的顺序
	distinctField string                 // distinct字段
	flFuncCount   int                    // first,last函数个数
	histogramInfo map[string]interface{} // histogram信息

	// 请求结果解析
	AliasSlice      []map[string]string // AliasSlice 别名信息
	ClassNames      string              // ClassNames 指标集名称
	SortFields      []string            // SortFields 选择字段有序列表
	AggsStruct      interface{}         // aggs结构，用于反序列化es返回结果
	StartTime       int64               // 时间范围
	EndTime         int64               // 时间范围
	IsHighlight     bool                // 是否高亮
	HighlightFields []string            // 高亮字段
}

//==================全局变量，只读===================
var (

	// script
	ScriptFName = "script"

	// text 字段匹配规则
	ObjectTextFields        = []string{"message"}
	LoggingTextFields       = []string{"message"}
	BackupLoggingTextFields = []string{"message"}
	TracingTextFields       = []string{"message"}
	EventTextFields         = []string{"message", "title"}
	RumRegExp               = ".*(message|stack)$"
	SecurityTextFields      = []string{"message", "title"}

	//
	AggsIdentifier = "aggs"
	BucketAggs     = "bucket"  // 桶聚合类型名称
	QueryValue     = "value"   // query field 关键字
	AggsTophits    = "tophits" // tophits
	SelectAll      = "*"       // 查询返回所有字段

)

//==================unit func===================

func voidToString(s interface{}) (string, error) {
	if res, ok := s.(string); ok {
		return res, nil
	}
	return "", fmt.Errorf("%v type is %T, can not convert to string", s, s)
}

// isTextField 判断是否为text类型
func (esPtr *EST) isTextField(fieldName string) (bool, error) {
	switch esPtr.esNamespace.fullName {
	case RFULLNAME: // rum
		match, err := regexp.MatchString(RumRegExp, fieldName)
		if err != nil {
			return false, err
		}
		return match, nil
	case OFULLNAME: // object
		return findStringSlice(fieldName, ObjectTextFields), nil
	case LFULLNAME: // logging
		return findStringSlice(fieldName, LoggingTextFields), nil
	case EFULLNAME: // event
		return findStringSlice(fieldName, EventTextFields), nil
	case TFULLNAME: // tracing
		return findStringSlice(fieldName, TracingTextFields), nil
	case SFULLNAME: // security
		return findStringSlice(fieldName, SecurityTextFields), nil
	}
	return false, nil
}

func (esPtr *EST) findAliasField(fieldName string) string {
	if k, ok := esPtr.AliasSlice[1][fieldName]; ok {
		return k //如果存在别名，需要转化
	}
	return fieldName
}

//  获取查询语句中应该使用的fieldname
func (esPtr *EST) getDSLFieldName(fieldName string, checkText bool) (string, error) {
	// (1) 如果有别名，需要转为实际字段
	fieldName = esPtr.findAliasField(fieldName)

	if checkText {
		// (2) 如果是text字段，桶聚合需要使用fName.keyword
		ok, err := esPtr.isTextField(fieldName)
		if err != nil {
			return fieldName, err
		}
		if ok {
			fieldName = fieldName + ".keyword"
		}
	}
	return fieldName, nil
}

// findStringSlice 查看string是否在slice string中
func findStringSlice(s string, items []string) bool {
	for _, item := range items {
		if item == s {
			return true
		}
	}
	return false
}

// appendSortFields sortField添加新的field
func (esPtr *EST) appendSortFields(fieldName string) error {
	if fieldName == esPtr.esNamespace.timeField {
		// 判断是否已经有time字段
		if find := findStringSlice(fieldName, esPtr.SortFields); find {
			return nil
		}
	}
	esPtr.SortFields = append(esPtr.SortFields, fieldName)
	return nil
}

// IsStringParam 函数参数，类型判断
func IsStringParam(ptr interface{}) bool {
	// StringLiteral, Identifier
	if _, ok := ptr.(*Identifier); ok {
		return true
	}
	if _, ok := ptr.(*StringLiteral); ok {
		return true
	}
	return false
}

// GetStringParam 获取string param
func GetStringParam(ptr interface{}) string {
	// StringLiteral, Identifier
	if rPtr, ok := ptr.(*Identifier); ok {
		return rPtr.Name
	}
	if rPtr, ok := ptr.(*StringLiteral); ok {
		return rPtr.Val
	}
	return ""
}

// GetFloatParam 获取float
func GetFloatParam(ptr interface{}) float64 {
	// StringLiteral, Identifier
	if rPtr, ok := ptr.(*NumberLiteral); ok {
		if rPtr.IsInt {
			return float64(rPtr.Int)
		}
		return rPtr.Float
	}
	return float64(0)
}

//==================customize struct===================

// SIMAP 自定义类型
type SIMAP map[string]interface{}

// ESMetric ES查询语句的抽象结构
// 实例化——>序列化后为ES的查询字符串
type ESMetric struct {
	Aggs      interface{}   `json:"aggs"`                       //  聚合部分
	Fields    []string      `json:"_source,omitempty"`          //  返回列列表
	Query     SIMAP         `json:"query,omitempty"`            //  查询部分
	Size      int           `json:"size"`                       //  返回数量
	From      int           `json:"from,omitempty"`             //  offset
	Sort      []interface{} `json:"sort,omitempty"`             //  排序,可缺失
	TotalHits bool          `json:"track_total_hits,omitempty"` //  查询添加totalHits
	Highlight interface{}   `json:"highlight,omitempty"`        //  高亮查询结果
}

// getTimeStamp 获取时间聚合信息
func (m *DFQuery) getTimeStamp(esPtr *EST) (string, error) {
	var interval int64
	interval = int64(m.TimeRange.Resolution.Duration / time.Millisecond) // 单位:毫秒
	if interval < 1 {
		return "", fmt.Errorf("time interval should large than 1ms")
	}
	return strconv.FormatInt(interval, 10) + "ms", nil

}

// 字段别名format
func formatAliasField(fName string) string {
	res := ""
	s := strings.Split(fName, ".")
	l := len(s)
	if l > 1 {
		if s[l-1] == "keyword" {
			res = s[l-2]
		} else {
			res = s[l-1]
		}
	} else {
		res = s[0]
	}
	return res
}

// ESQL 用于ElasticSearch实例化
// 返回值为elasticsearch的查询语句
func (m *DFQuery) ESQL() (interface{}, error) {

	// var res = map[string]string{}
	var esPtr *EST

	// step1: 基本合法性判断
	esPtr, err := m.checkValid()
	if err != nil {
		return "", err
	}

	// step2: 通过metric信息，实例化ESMetric
	var em ESMetric
	em, err = transport(m, esPtr)

	if err != nil {
		return "", fmt.Errorf("metric transport error, %s", err)
	}

	qRes, err := json.Marshal(em)

	if err != nil {
		return "", fmt.Errorf("json marshal error, %s", err)
	}

	// res["dql"] = string(qRes)
	esAlias := map[string]string{}
	for k, v := range esPtr.AliasSlice[0] {
		rk := formatAliasField(k)
		esAlias[rk] = v
	}

	// 如果是对象数据，需要另外添加time字段的别名
	if esPtr.esNamespace.fullName == OFULLNAME {
		esAlias[esPtr.esNamespace.timeField] = TIMEDEFAULT
	}

	// 将解析信息添加到AST上，用于结果解析
	estResPtr := &ESTRes{
		Alias:        esAlias,
		ClassNames:   esPtr.ClassNames,
		SortFields:   esPtr.SortFields,
		Show:         false,
		FLFuncCount:  esPtr.flFuncCount,
		AggsFromSize: esPtr.getAggsFromSize(),
		StartTime:    esPtr.StartTime,
		EndTime:      esPtr.EndTime,
	}
	// 如果有时间聚合，第一列列名为time
	if esPtr.dfState.dateHg {
		if len(estResPtr.SortFields) > 0 {
			estResPtr.SortFields[0] = TIMEDEFAULT
		}
	}
	// 如果需要高亮查询结果
	if esPtr.IsHighlight && len(esPtr.HighlightFields) > 0 {
		estResPtr.HighlightFields = esPtr.HighlightFields
	}
	helper := &Helper{
		ESTResPtr: estResPtr,
	}
	m.Helper = helper
	return string(qRes), nil
}

// transport 转换函数
func transport(m *DFQuery, esPtr *EST) (ESMetric, error) {
	var em ESMetric
	// 1）fields
	fp, err := fieldsTransport(m, esPtr)
	if err != nil {
		return em, err
	}
	// 2) 聚合
	ap, err := aggsTransport(m, esPtr)
	if err != nil {
		return em, err
	}
	// 3）查询
	qp, err := queryTransport(m, esPtr)
	if err != nil {
		return em, err
	}
	// 4) size
	sp, err := sizeTransport(m, esPtr)
	if err != nil {
		return em, err
	}
	// 5) from
	fromSize, err := fromTransport(m, esPtr)
	if err != nil {
		return em, err
	}
	// 6) sort
	stp, err := sortTransport(m, esPtr)
	if err != nil {
		return em, err
	}

	// 7) totalHits
	hits, err := hitsTransport(m, esPtr)
	if err != nil {
		return em, err
	}

	// 8) highlight
	highlights, err := highlightTransport(m, esPtr)
	if err != nil {
		return em, err
	}
	em.Aggs = ap
	em.Query = qp
	em.Fields = fp
	em.Size = sp
	em.From = fromSize
	em.Sort = stp
	em.TotalHits = hits
	em.Highlight = highlights

	return em, nil
}

// ESQL groupby
func (x *GroupBy) ESQL() (interface{}, error) {
	var res = []string{}
	if x == nil {
		return res, nil
	}
	for _, i := range x.List {
		str := i.String() // Identifier.ESQL()
		// 多层聚合，不能对相同的field
		for _, item := range res {
			if item == str {
				return res, fmt.Errorf("each field can only be grouped once")
			}
		}
		res = append(res, str)
	}
	return res, nil
}

// 获取totalHits
func hitsTransport(m *DFQuery, esPtr *EST) (bool, error) {
	if esPtr.dfState.aggs { // 存在聚合,没有 totalHits
		return false, nil
	}
	return true, nil
}

// rangeTermQuery 范围查询
// 示例:
// {
// 	"range": {
// 		"age": {
// 			"lt": "30"
// 		}
// 	}
// }

// ESQL BinaryExpr
func (x *BinaryExpr) ESQL() (interface{}, error) {
	return nil, nil
}

// ESQL 对FuncExpr结构的解析
func (x *FuncExpr) ESQL() (interface{}, error) {
	var res = new(esQlFunc)

	// match, script
	if v, ok := QueryFuncs[x.Name]; ok {
		res.fName = v
		var args = []string{}
		for _, i := range x.Param {
			arg, err := i.ESQL()
			if err != nil {
				return nil, err
			}
			strArg, err := voidToString(arg)
			if err != nil {
				return nil, fmt.Errorf("invalid namespace")
			}
			args = append(args, strArg)
		}
		res.args = args
	}

	return res, nil
}

// ESQL target 返回列值
func (x *Target) ESQL() (interface{}, error) {
	if v, ok := x.Col.(*Identifier); ok {
		s, err := v.ESQL()
		if err != nil {
			return "", err
		}
		return s, nil
	}
	return nil, nil

}

// ESQL column
func (x *Identifier) ESQL() (interface{}, error) {
	res := x.String()
	return res, nil
}

// ESQL stringliteral
func (x *StringLiteral) ESQL() (interface{}, error) {
	return x.Val, nil
}

// ESQL orderby
func (x *OrderBy) ESQL() (interface{}, error) {
	return x.String(), nil
}

// ESQL orderbyElem
func (x *OrderByElem) ESQL() (interface{}, error) {
	return x.String(), nil
}

// ESQL timeexpr
// 返回int64类型的时间戳，单位为毫秒
func (x *TimeExpr) ESQL() (interface{}, error) {
	return int64(x.Time.UnixNano() / int64(time.Millisecond)), nil
}

// ESQL limit
func (x *Limit) ESQL() (interface{}, error) {
	return x.Limit, nil
}

// ESQL slimit
func (x *SLimit) ESQL() (interface{}, error) {
	return x.String(), nil
}

// ESQL offset
func (x *Offset) ESQL() (interface{}, error) {
	return x.String(), nil
}

// ESQL soffset
func (x *SOffset) ESQL() (interface{}, error) {
	return x.String(), nil
}

// ESQL niliteral
func (x *NilLiteral) ESQL() (interface{}, error) {
	return "", fmt.Errorf("nil no support")
}

// ESQL bool
func (x *BoolLiteral) ESQL() (interface{}, error) {
	return fmt.Sprintf("%v", x.Val), nil
}

// ESQL regex
func (x *Regex) ESQL() (interface{}, error) {
	return x.Regex, nil
}

// ESQL numberliteral
func (x *NumberLiteral) ESQL() (interface{}, error) {
	return x.String(), nil
}

// ESQL funcarg
func (x *FuncArg) ESQL() (interface{}, error) {
	return "", fmt.Errorf("not impl")
}

// ESQL fill
func (x *Fill) ESQL() (interface{}, error) {
	return "", fmt.Errorf("not impl")
}

// ESQL nodelist
func (x NodeList) ESQL() (interface{}, error) {
	return "", nil
}

// ESQL timezone
func (x *TimeZone) ESQL() (interface{}, error) {
	return "", nil
}

// ESQL timerange
func (x *TimeRange) ESQL() (interface{}, error) {
	return "", nil
}

// esQL parenexpr
func (x *ParenExpr) esQL(esPtr *EST) (interface{}, error) {
	if xRes, ok := (x.Param).(*BinaryExpr); ok {
		return xRes.esQL(esPtr)
	}
	return x.Param.ESQL()
}

// ESQL parenexpr
func (x *ParenExpr) ESQL() (interface{}, error) {
	return nil, nil
}

// ESQL stmts
func (x Stmts) ESQL() (interface{}, error) {
	return "", nil
}

// ESQL FuncArgList
func (x FuncArgList) ESQL() (interface{}, error) {
	return "", fmt.Errorf("not impl")
}

// ESQL Star
func (x *Star) ESQL() (interface{}, error) {
	return "", fmt.Errorf("not impl")
}

// ESQL AttrExpr
func (x *AttrExpr) ESQL() (interface{}, error) {
	return "", fmt.Errorf("not impl")
}

// ESQL CascadeFunctions
func (x *CascadeFunctions) ESQL() (interface{}, error) {
	return "", fmt.Errorf("not impl")
}

// ESQL TimeResolution
func (x *TimeResolution) ESQL() (interface{}, error) {
	return "", fmt.Errorf("not impl")
}

// ESQL Lambda
func (x *Lambda) ESQL() (interface{}, error) {
	return "", fmt.Errorf("not impl")
}

// ESQL StaticCast
func (x *StaticCast) ESQL() (interface{}, error) {
	return "", fmt.Errorf("not impl")
}

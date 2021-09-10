package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	Nil = "nil"
)

type Value interface {
	Type() ValueType
	String() string
}

type ValueType string

///////////////////////////////////////
// Node
///////////////////////////////////////
type Node interface {
	String() string
	Pos() *PositionRange
}

type CascadeFunctions struct {
	Funcs []*FuncExpr `json:"function_list,omitempty"`
}

func (n *CascadeFunctions) Pos() *PositionRange { return nil }
func (n *CascadeFunctions) String() string {
	arr := []string{}
	for _, f := range n.Funcs {
		arr = append(arr, f.String())
	}

	return strings.Join(arr, ".")
}

type Anonymous struct{}

func (n *Anonymous) Pos() *PositionRange { return nil }
func (n *Anonymous) String() string {
	return ""
}

type AttrExpr struct {
	Obj  Node `json:"object,omitempty"`
	Attr Node `json:"attr,omitempty"`
}

func (n *AttrExpr) Pos() *PositionRange { return nil }
func (n *AttrExpr) String() string {
	return fmt.Sprintf("%s.%s", n.Obj.String(), n.Attr.String())
}

type Star struct{}

func (n *Star) MarshalJSON() ([]byte, error) {
	return []byte(`"*"`), nil
}

func (n *Star) Pos() *PositionRange { return nil }
func (n *Star) String() string {
	return "*"
}

type Stmts []Node

func (n Stmts) Pos() *PositionRange { return nil }
func (n Stmts) String() string {
	arr := []string{}
	for _, n := range n {
		arr = append(arr, n.String())
	}

	return strings.Join(arr, "; ")
}

type BoolLiteral struct {
	Val bool `json:"val,omitempty"`
}

func (n *BoolLiteral) Pos() *PositionRange { return nil }
func (n *BoolLiteral) String() string {
	return fmt.Sprintf("%v", n.Val)
}

type NilLiteral struct{}

func (n *NilLiteral) Pos() *PositionRange { return nil }
func (n *NilLiteral) String() string {
	return Nil
}

type Limit struct {
	Limit int64 `json:"val,omitempty"`
}

func (n *Limit) Pos() *PositionRange { return nil }
func (n *Limit) String() string {
	return fmt.Sprintf("LIMIT %d", n.Limit)
}

type SLimit struct {
	SLimit int64 `json:"val,omitempty"`
}

func (n *SLimit) Pos() *PositionRange { return nil }
func (n *SLimit) String() string {
	return fmt.Sprintf("SLIMIT %d", n.SLimit)
}

type Offset struct {
	Offset int64 `json:"val,omitempty"`
}

func (n *Offset) Pos() *PositionRange { return nil }
func (n *Offset) String() string {
	return fmt.Sprintf("OFFSET %d", n.Offset)
}

type SOffset struct {
	SOffset int64 `json:"val,omitempty"`
}

func (n *SOffset) Pos() *PositionRange { return nil }
func (n *SOffset) String() string {
	return fmt.Sprintf("SOFFSET %d", n.SOffset)
}

type OrderType int

const (
	OrderAsc OrderType = iota
	OrderDesc
)

type OrderByElem struct {
	Column string    `json:"-"`
	Opt    OrderType `json:"-"`
}

func (n *OrderByElem) Pos() *PositionRange { return nil }
func (n *OrderByElem) String() string {
	if n.Opt == OrderDesc {
		return fmt.Sprintf("%s DESC", n.Column)
	}
	return fmt.Sprintf("%s ASC", n.Column)
}

func (n *OrderByElem) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(n.Column)
	if err != nil {
		return nil, err
	}

	const output = `{"column":%s, "opt":"%s"}`

	switch n.Opt {
	case OrderAsc:
		return []byte(fmt.Sprintf(output, b, "asc")), nil
	case OrderDesc:
		return []byte(fmt.Sprintf(output, b, "desc")), nil
	}
	return []byte(`""`), nil
}

type OrderBy struct {
	List NodeList
}

func (n *OrderBy) Pos() *PositionRange { return nil }
func (n *OrderBy) String() string {
	var arr []string
	for _, elem := range n.List {
		arr = append(arr, elem.String())
	}
	return fmt.Sprintf("ORDER BY %s", strings.Join(arr, ", "))
}

type GroupBy struct {
	List NodeList `json:"val_list,omitempty"`
}

func (n *GroupBy) Pos() *PositionRange { return nil }
func (n *GroupBy) String() string {
	return fmt.Sprintf("BY %s", n.List.String())
}

func (n *GroupBy) ColumnList() []string {
	var res []string
	for _, node := range n.List {
		res = append(res, node.String())
	}
	return res
}

type TimeZone struct {
	Input    string `json:"input,omitempty"`
	TimeZone string `json:"-"`
}

func (n *TimeZone) Pos() *PositionRange { return nil }
func (n *TimeZone) String() string {
	return fmt.Sprintf(`tz("%s")`, n.Input)
}

type NodeList []Node

func (n NodeList) Pos() *PositionRange { return nil }
func (n NodeList) String() string {
	arr := []string{}
	for _, arg := range n {
		arr = append(arr, arg.String())
	}
	return "[" + strings.Join(arr, ", ") + "]"
}

type FuncArg struct {
	ArgName string `json:"name,omitempty"`
	ArgVal  Node   `json:"val,omitempty"`
}

func (n *FuncArg) Pos() *PositionRange { return nil }
func (n *FuncArg) String() string {
	if n.ArgVal != nil {
		return fmt.Sprintf("%s=%s", n.ArgName, n.ArgVal.String())
	} else {
		return n.ArgName
	}
}

type FuncArgList []Node

func (n FuncArgList) Pos() *PositionRange { return nil }
func (n FuncArgList) String() string {
	arr := []string{}
	for _, x := range n {
		arr = append(arr, x.String())
	}
	return "[" + strings.Join(arr, ", ") + "]"
}

func getFuncArgList(nl NodeList) FuncArgList {
	var res FuncArgList
	for _, x := range nl {
		res = append(res, x)
	}
	return res
}

// SearchAfter 深度分页
type SearchAfter struct {
	Vals []interface{} `json:"vals,omitempty"`
}

// Pos pos
func (sa *SearchAfter) Pos() *PositionRange { return nil }

// String string
func (sa *SearchAfter) String() string {
	return fmt.Sprintf("%v", sa.Vals)
}

const (
	FillNil int = iota + 1
	FillInt
	FillFloat
	FillStr
	FillLinear
	FillPrevious
)

type Fill struct {
	FillType int     `json:"-"`
	Str      string  `json:"-"`
	Float    float64 `json:"-"`
	Int      int64   `json:"-"`
	Boolean  bool    `json:"-"`
}

func (n *Fill) MarshalJSON() ([]byte, error) {
	const output1 = `{"type":"%s"}`
	const output2 = `{"type":"%s", "%s":%v}`

	switch n.FillType {
	case FillNil:
		return []byte(fmt.Sprintf(output1, Nil)), nil

	case FillInt:
		return []byte(fmt.Sprintf(output2, "integer", "integer_val", n.Int)), nil

	case FillFloat:
		return []byte(fmt.Sprintf(output2, "float", "float_val", n.Float)), nil

	case FillStr:
		return []byte(fmt.Sprintf(output2, "str", "str_val", fmt.Sprintf(`"%s"`, n.Str))), nil

	case FillLinear:
		return []byte(fmt.Sprintf(output1, "linear")), nil

	case FillPrevious:
		return []byte(fmt.Sprintf(output1, "previous")), nil
	}

	return []byte(`""`), nil
}

func (n *Fill) String() string {
	switch n.FillType {
	case FillNil:
		return "<nil>"
	case FillInt:
		return fmt.Sprintf("<%d>", n.Int)
	case FillFloat:
		return fmt.Sprintf("<%f>", n.Float)
	case FillStr:
		return fmt.Sprintf("<%s>", n.Str)
	case FillLinear:
		return "<linear>"
	case FillPrevious:
		return "<previous>"
	}

	log.Warn("invalid ifill: %+#v", n)
	return ""
}

func (n *Fill) StringInfluxql() string {
	switch n.FillType {
	case FillNil:
		return Nil
	case FillInt:
		return fmt.Sprintf("%d", n.Int)
	case FillFloat:
		return fmt.Sprintf("%f", n.Float)
	case FillStr:
		return n.Str
	case FillLinear:
		return "linear"
	case FillPrevious:
		return "previous"
	}

	log.Warn("invalid fill: %+#v", n)
	return ""
}

func (n *Fill) Pos() *PositionRange { return nil } // TODO

type FuncExpr struct {
	Name  string `json:"name,omitempty"`
	Param []Node `json:"param,omitempty"`
	//Pos   *PositionRange
}

const (
	fillFuncArgs = 2
)

func (n *FuncExpr) SplitFill() (val Node, fill *Fill, err error) {

	switch strings.ToLower(n.Name) {
	case "fill":
		const typ = "(Nil,NumberLiteral,StringLiteral,PREVIOUS,LINEAR)"

		if len(n.Param) != fillFuncArgs {
			err = fmt.Errorf("fill function only accept 2 parameter, left value and %s", typ)
			return
		}

		paramErr := fmt.Errorf("unknown fill function parameter, only accept %s", typ)

		switch v := n.Param[1].(type) {
		case *Identifier:
			switch strings.ToLower(v.Name) {
			case "previous":
				fill = &Fill{FillType: FillPrevious}
			case "linear":
				fill = &Fill{FillType: FillLinear}
			default:
				err = paramErr
				return
			}

		case *NilLiteral:
			fill = &Fill{FillType: FillNil}

		case *NumberLiteral:
			if v.IsInt {
				fill = &Fill{FillType: FillInt, Int: v.Int}
			} else {
				fill = &Fill{FillType: FillFloat, Float: v.Float}
			}

		case *StringLiteral:
			fill = &Fill{FillType: FillStr, Str: v.Val}

		default:
			err = paramErr
			return
		}

		val = n.Param[0]

	default:
		val = n
	}

	return
}

func (n *FuncExpr) String() string {
	arr := []string{}
	for _, n := range n.Param {
		arr = append(arr, n.String())
	}
	return fmt.Sprintf("%s(%s)", strings.ToLower(n.Name), strings.Join(arr, ", "))
}

func (n *FuncExpr) Pos() *PositionRange { return nil } // TODO

// stmt
type ESTRes struct {
	Alias           map[string]string // 别名信息
	SortFields      []string          // 返回字段有序列表
	ClassNames      string            // metric 类型名称
	Show            bool              // 是否是show查询
	ShowFields      bool              // 是否是show fields查询
	TimeField       string            // time字段
	FLFuncCount     int               // first,last函数个数
	AggsFromSize    int               // 聚合的分页
	StartTime       int64             // 开始时间
	EndTime         int64             // 结束时间
	HighlightFields []string          // 高亮字段
}

// Helper tranlate中间结果
type Helper struct {
	ESTResPtr *ESTRes // es translate，当结果转为influxdb结构使用
}

type DFQuery struct { // impl Node
	Namespace string `json:"namespace,omitempty"`

	// data source
	Names      []string `json:"names,omitempty"`
	RegexNames []*Regex `json:"regex_names,omitempty"`

	Anonymous bool `json:"-"`

	Subquery *DFQuery `json:"subquery,omitempty"`

	Targets        []*Target    `json:"targets,omitempty"`
	WhereCondition []Node       `json:"where_condition,omitempty"`
	GroupBy        *GroupBy     `json:"groupby,omitempty"`
	OrderBy        *OrderBy     `json:"orderby,omitempty"`
	Limit          *Limit       `json:"limit,omitempty"`
	Offset         *Offset      `json:"offset,omitempty"`
	SLimit         *SLimit      `json:"slimit,omitempty"`
	SOffset        *SOffset     `json:"soffset,omitempty"`
	TimeZone       *TimeZone    `json:"timezone,omitempty"`
	SearchAfter    *SearchAfter `json:"search_after,omitempty"` // search_after
	Highlight      bool         `json:"highlight,omitempty"`

	Helper *Helper `json:"-"`
}

func (m *DFQuery) JSON() ([]byte, error) {
	// json.Marshal escaping < and >
	// https://stackoverflow.com/questions/28595664/how-to-stop-json-marshal-from-escaping-and
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(m)

	// Encode() followed by a newline character
	return buffer.Bytes(), err
}

// IsAllTargets, 未指定 target 或为手动填写 "*"，即 ALL，like SELECT * FROM XX
func (m *DFQuery) IsAllTargets() bool {
	return m.IsMatchTargetsNum(0)
}

func (m *DFQuery) IsMatchTargetsNum(num int) bool {
	if len(m.Targets) == 0 || m.Targets[0].Col.String() == "*" {
		return num == 0
	}
	return len(m.Targets) == num
}

func (m *DFQuery) GroupByList() []string {
	if m.GroupBy == nil {
		return nil
	}
	return m.GroupBy.ColumnList()
}

func (m *DFQuery) String() string {
	parts := []string{m.From()}

	if m.Targets != nil {
		arr := []string{}
		for _, t := range m.Targets {
			arr = append(arr, t.String())
		}

		parts = append(parts, ":("+strings.Join(arr, ", ")+")")
	}

	if m.WhereCondition != nil {
		arr := []string{}
		for _, f := range m.WhereCondition {
			arr = append(arr, f.String())
		}
		parts = append(parts, "{"+strings.Join(arr, ", ")+"}")
	}

	if m.GroupBy != nil {
		parts = append(parts, " "+m.GroupBy.String())
	}

	if m.OrderBy != nil {
		parts = append(parts, " "+m.OrderBy.String())
	}

	if m.Limit != nil {
		parts = append(parts, " "+m.Limit.String())
	}

	if m.Offset != nil {
		parts = append(parts, " "+m.Offset.String())
	}

	if m.SLimit != nil {
		parts = append(parts, " "+m.SLimit.String())
	}

	if m.SOffset != nil {
		parts = append(parts, " "+m.SOffset.String())
	}

	if m.TimeZone != nil {
		parts = append(parts, " "+m.TimeZone.String())
	}

	return strings.Join(parts, "")
}

func (m *DFQuery) From() string {
	arr := []string{}

	arr = append(arr, m.Names...)

	for _, x := range m.RegexNames {
		arr = append(arr, x.String())
	}

	if m.Subquery != nil {
		arr = append(arr, "("+m.Subquery.String()+")")
	}

	return m.Namespace + "::" + strings.Join(arr, ",")
}

func (m *DFQuery) Pos() *PositionRange { return nil } // TODO

type Target struct { // impl Node
	Col    Node   `json:"col,omitempty"`
	Alias  string `json:"alias,omitempty"`
	Fill   *Fill  `json:"fill,omitempty"`
	Talias string `json:"tailas,omitempty"` // dql翻译别名
}

func (n *Target) String() string {
	if n.Col == nil {
		panic("unreachable: Target col is nil")
	}

	if n.Fill != nil {
		fillFn := &FuncExpr{
			// FIXME: support upper and lower
			Name:  "fill",
			Param: []Node{n.Col, n.Fill},
		}
		return fillFn.String()
	}

	var res []string

	if c := n.Col.String(); c != "" {
		res = append(res, c)
	}

	if n.Alias != "" {
		res = append(res, fmt.Sprintf("AS %s", n.Alias))
	}

	if n.Fill != nil && n.Fill.String() != "" {
		res = append(res, n.Fill.String())
	}

	return strings.Join(res, " ")
}

func (n *Target) String2() string {
	if n.Alias != "" {
		return n.Alias
	}
	return n.String()
}

func (n *Target) Pos() *PositionRange { return nil /* TODO */ }

type TimeResolution struct {
	Duration time.Duration
	Auto     bool
	PointNum *NumberLiteral
}

func (n *TimeResolution) String() string {
	if n.Auto {
		return "auto"
	}
	if n.PointNum != nil {
		return fmt.Sprintf("auto(%s)", n.PointNum)
	}
	return fmt.Sprintf("%v", n.Duration)
}

func (n *TimeResolution) Pos() *PositionRange { return nil /* TODO */ }

///////////////////////////////////////
// Expr
///////////////////////////////////////
type Expr interface {
	Node
	Type() ValueType

	DQLExpr()
}

type Regex struct {
	Regex string `json:"regex,omitempty"`
}

func (e *Regex) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, e.String())), nil
}

func (e *Regex) Type() ValueType     { return "" /* TODO */ }
func (e *Regex) String() string      { return fmt.Sprintf("re('%s')", e.Regex) }
func (e *Regex) Pos() *PositionRange { return nil } // TODO
func (e *Regex) DQLExpr()            {}             // not used

type StringLiteral struct {
	Val string `json:"val,omitempty"`
}

func (e *StringLiteral) Type() ValueType     { return "" /* TODO */ }
func (e *StringLiteral) DQLExpr()            { /* not used */ }
func (e *StringLiteral) String() string      { return fmt.Sprintf("'%s'", e.Val) }
func (e *StringLiteral) Pos() *PositionRange { return nil /* TODO */ }

type BinaryExpr struct { // impl Expr & Node
	Op         ItemType `json:"operator,omitempty"`
	LHS        Node     `json:"left,omitempty"`
	RHS        Node     `json:"right,omitempty"`
	ReturnBool bool     `json:"-"`
}

func (e *BinaryExpr) Type() ValueType     { return "" }  // TODO
func (e *BinaryExpr) Pos() *PositionRange { return nil } // TODO
func (e *BinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s",
		e.LHS.String(),
		e.Op.String(),
		e.RHS.String())
}

func (e *BinaryExpr) DQLExpr() {} // not used

type ParenExpr struct {
	Param Node `json:"paren"`
}

func (e *ParenExpr) Type() ValueType     { return "" }  // TODO
func (e *ParenExpr) Pos() *PositionRange { return nil } // TODO
func (e *ParenExpr) String() string {
	return fmt.Sprintf("(%s)", e.Param.String())
}

func (e *ParenExpr) DQLExpr() {} // not used

type NumberLiteral struct {
	IsInt bool
	Float float64
	Int   int64
}

func (e *NumberLiteral) IsPositiveInteger() bool {
	return e.IsInt && e.Int > 0
}

func (e *NumberLiteral) MarshalJSON() ([]byte, error) {
	return []byte(e.String()), nil
}

func (e *NumberLiteral) Type() ValueType     { return "" }
func (e *NumberLiteral) DQLExpr()            {}
func (e *NumberLiteral) Pos() *PositionRange { return nil } // not used
func (e *NumberLiteral) Reverse() {
	if e.IsInt {
		e.Int = -e.Int
	} else {
		e.Float = -e.Float
	}
}

func (e *NumberLiteral) String() string {
	if e.IsInt {
		return fmt.Sprintf("%d", e.Int)
	} else {
		return fmt.Sprintf("%f", e.Float)
	}
}

type Identifier struct { // impl Expr
	Name string `json:"val,omitempty"`
}

func (e *Identifier) String() string      { return e.Name }
func (e *Identifier) Pos() *PositionRange { return nil } // TODO
func (e *Identifier) DQLExpr()            {}             // not used
func (e *Identifier) Type() ValueType     { return "" }

type StaticCast struct {
	IsInt   bool
	IsFloat bool
	Val     *Identifier
}

func (e *StaticCast) MarshalJSON() ([]byte, error) {
	const res = `{"%s":"%s"}`
	if e.IsInt {
		return []byte(fmt.Sprintf(res, "int", e.Val.String())), nil
	}
	if e.IsFloat {
		return []byte(fmt.Sprintf(res, "float", e.Val.String())), nil
	}
	return nil, fmt.Errorf("unreachable")
}

func (e *StaticCast) String() string {
	if e.IsInt {
		return fmt.Sprintf("int(%s)", e.Val.Name)
	}
	if e.IsFloat {
		return fmt.Sprintf("float(%s)", e.Val.Name)
	}
	return ""
}
func (e *StaticCast) Pos() *PositionRange { return nil } // TODO
func (e *StaticCast) DQLExpr()            {}             // not used
func (e *StaticCast) Type() ValueType     { return "" }

///////////////////////////////////////
// stmt
///////////////////////////////////////
type Statement interface {
	Node
	DQLStmt() // not used
}

// OuterFunc outerFunc
type OuterFunc struct {
	Func         *FuncExpr     `json:"func,omitempty"`
	FuncArgVals  []interface{} `json:"func_arg_vals,omitempty"`
	FuncArgTypes []string      `json:"func_arg_types,omitempty"`
	FuncArgNames []string      `json:"func_arg_names,omitempty"`
}

type OuterFuncs struct {
	Funcs []*OuterFunc `json:"funcs,omitempty"`
}

func (ofuncs *OuterFuncs) JSON() ([]byte, error) {
	// json.Marshal escaping < and >
	// https://stackoverflow.com/questions/28595664/how-to-stop-json-marshal-from-escaping-and
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(ofuncs)

	// Encode() followed by a newline character
	return buffer.Bytes(), err
}

func (ofuncs *OuterFunc) String() string {
	return "outer func"
}

func (ofuncs *OuterFunc) Pos() *PositionRange {
	return nil
}

func (ofuncs *OuterFuncs) String() string {
	// TODO
	return "outer func"
}

func (ofuncs *OuterFuncs) Pos() *PositionRange {
	return nil
}

// DeleteFunc delete info
type DeleteFunc struct {
	// (1) dql语句;
	// (2) es indexName, 索引名称不包含wsid，例如: rum, log等;
	// (3) influxdb measurement name, 例如: cpu,mem等
	StrDql            string
	Func              *FuncExpr
	DeleteIndex       bool // 是否删除整个ES索引
	DeleteMeasurement bool // 是否删除整个Influxdb measurement
}

func (d *DeleteFunc) String() string {
	return "outer delete func"
}

func (d *DeleteFunc) Pos() *PositionRange {
	return nil
}

type Evaluable interface {
	Eval(source string, tags map[string]string, fields map[string]interface{}) bool
}

type WhereCondition struct {
	conditions []Node
}

func (x *WhereCondition) Eval(source string, tags map[string]string, fields map[string]interface{}) bool {
	for _, c := range x.conditions {
		switch expr := c.(type) {
		case *BinaryExpr:
			if !expr.Eval(source, tags, fields) {
				return false
			}

		default:
			log.Errorf("Eval only accept BinaryExpr")
			return false
		}
	}

	return true
}

func (x *WhereCondition) String() string {
	arr := []string{}
	for _, f := range x.conditions {
		arr = append(arr, f.String())
	}

	return "{" + strings.Join(arr, " and ") + "}"
}

func (x *WhereCondition) Pos() *PositionRange {
	return nil
}

type WhereConditions []Node

func (x WhereConditions) Pos() *PositionRange { return nil }

func (x WhereConditions) String() string {
	arr := []string{}
	for _, c := range x {
		if c == nil {
			continue
		}

		arr = append(arr, c.String())
	}
	return strings.Join(arr, "; ")
}

func (x WhereConditions) Eval(source string,
	tags map[string]string,
	fields map[string]interface{}) bool {

	for _, item := range x {
		switch c := item.(type) {
		case *WhereCondition:
			if c.Eval(source, tags, fields) {
				return true
			}

		default:
			log.Warnf("invalid where condition: %s", c)
			return false
		}
	}

	return false
}

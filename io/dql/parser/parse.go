package parser

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/prometheus/util/strutil"
)

var parserPool = sync.Pool{
	New: func() interface{} {
		return &parser{}
	},
}

type parser struct {
	lex      Lexer
	yyParser yyParserImpl

	parseResult interface{}
	lastClosing Pos
	errs        ParseErrors

	inject    ItemType
	injecting bool

	curQuery   *DFQuery
	ExtraParam *ExtraParam

	context interface{}
}

func newParser(input string) *parser {
	return newParserWithParam(input, nil)
}

func newParserWithParam(input string, param *ExtraParam) *parser {
	p := parserPool.Get().(*parser)

	// reset parser fields
	p.injecting = false
	p.errs = nil
	p.parseResult = nil
	p.curQuery = nil
	p.lex = Lexer{
		input: input,
		state: lexStatements,
	}

	if param != nil {
		p.ExtraParam = param
	}

	return p
}

var errUnexpected = errors.New("unexpected error")

func (p *parser) unexpected(context string, expected string) {
	var errMsg strings.Builder

	if p.yyParser.lval.item.Typ == ERROR { // do not report lex error twice
		return
	}

	errMsg.WriteString("unexpected: ")
	errMsg.WriteString(p.yyParser.lval.item.desc())

	if context != "" {
		errMsg.WriteString(" in: ")
		errMsg.WriteString(context)
	}

	if expected != "" {
		errMsg.WriteString(", expected: ")
		errMsg.WriteString(expected)
	}

	p.addParseErr(p.yyParser.lval.item.PositionRange(), errors.New(errMsg.String()))
}

func (p *parser) recover(errp *error) {
	e := recover()
	if _, ok := e.(runtime.Error); ok {
		buf := make([]byte, 64<<10) // 64k
		buf = buf[:runtime.Stack(buf, false)]
		fmt.Fprintf(os.Stderr, "parser panic: %v\n%s", e, buf)
		*errp = errUnexpected
	} else if e != nil {
		*errp = e.(error)
	}
}

func (p *parser) addParseErr(pr *PositionRange, err error) {
	p.errs = append(p.errs, ParseErr{
		Pos:   pr,
		Err:   err,
		Query: p.lex.input,
	})
}

func (p *parser) addParseErrf(pr *PositionRange, format string, args ...interface{}) {
	p.addParseErr(pr, fmt.Errorf(format, args...))
}

// impl Lex interface
func (p *parser) Lex(lval *yySymType) int {
	var typ ItemType

	if p.injecting {
		p.injecting = false
		return int(p.inject)
	}

	for { // skip comment
		p.lex.NextItem(&lval.item)
		typ = lval.item.Typ
		if typ != COMMENT {
			break
		}
	}

	switch typ {
	case ERROR:
		pos := PositionRange{
			Start: p.lex.start,
			End:   Pos(len(p.lex.input)),
		}

		p.addParseErr(&pos, errors.New(p.yyParser.lval.item.Val))
		return 0 // tell yacc it's the end of input

	case EOF:
		lval.item.Typ = EOF
		p.InjectItem(0)
	case RIGHT_BRACE, RIGHT_PAREN, RIGHT_BRACKET, DURATION:
		p.lastClosing = lval.item.Pos + Pos(len(lval.item.Val))
	}
	return int(typ)
}

func (p *parser) Error(e string) {}

func (p *parser) unquoteString(s string) string {
	unq, err := strutil.Unquote(s)
	if err != nil {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			"error unquoting string %q: %s", s, err)
	}
	return unq
}

func (p *parser) doParse() {
	p.InjectItem(START_STMTS)
	p.yyParser.Parse(p)
	if len(p.errs) == 0 {
		p.checkingSemantic()
	}
	p.addDefaultLimit()
	p.addDefaultSLimit()
	p.addSearchAfter()
	p.addHighlight()
}

func (p *parser) checkingSemantic() {
	if p.parseResult == nil {
		return
	}

	stmts, ok := p.parseResult.(Stmts)
	if !ok {
		p.addParseErr(p.yyParser.lval.item.PositionRange(),
			fmt.Errorf("unknown parser result: %v", p.parseResult))
		return
	}

	for _, node := range stmts {
		switch v := node.(type) {
		case *DFQuery:
			errs := v.checkingSemantic()
			for _, err := range errs {
				if err.Err == nil {
					continue
				}

				pos := p.yyParser.lval.item.PositionRange()
				if err.Pos != nil {
					pos = err.Pos
				}
				p.addParseErr(pos, err.Err)
			}
		default:
			// pass
		}
	}
}

func (p *parser) InjectItem(typ ItemType) {
	if p.injecting {
		log.Warnf("current inject is %v, new inject is %v", p.inject, typ)
		panic("cannot inject multiple Items into the token stream")
	}

	if typ != 0 && (typ <= startSymbolsStart || typ >= startSymbolsEnd) {
		log.Warnf("current inject is %v", typ)
		panic("cannot inject symbol that isn't start symbol")
	}
	p.inject = typ
	p.injecting = true
}

func (p *parser) number(v string) *NumberLiteral {
	nl := &NumberLiteral{}

	n, err := strconv.ParseInt(v, 0, 64)
	if err != nil {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			p.addParseErrf(p.yyParser.lval.item.PositionRange(),
				"error parsing number: %s", err)
		}
		nl.Float = f
	} else {
		nl.IsInt = true
		nl.Int = n
	}

	return nl
}

func (p *parser) parseDuration(v string) (time.Duration, error) {
	du, err := parseDuration(v)
	if err != nil {
		return -1, err
	}
	return du, nil
}

func (p *parser) checkAST(node Node) (typ ValueType) {
	// TODO
	return ""
}

//////////////////////////////////////
// yylex.(*parser).newXXXX
//////////////////////////////////////
func (p *parser) newQuery(x interface{}) (*DFQuery, error) {
	m := &DFQuery{}

	switch v := x.(type) {
	case *Regex:
		m.RegexNames = append(m.RegexNames, v)
	case Item:
		m.Names = append(m.Names, v.Val)
		if x.(Item).Val == "_" {
			m.Anonymous = true
		}
	case *StringLiteral:
		m.Names = append(m.Names, v.Val)
	case *Anonymous:
	default:
		p.addParseErr(p.yyParser.lval.item.PositionRange(),
			fmt.Errorf("error parsing metric name, should not been here"))
		return nil, fmt.Errorf("invalid query from source: %+#v", x)
	}

	p.curQuery = m
	p.context = p.curQuery
	return m, nil
}

func (p *parser) newShow(f *FuncExpr) (interface{}, error) {
	if f == nil {
		panic("function should not be nil")
	}
	show := &Show{}
	switch strings.ToLower(f.Name) {
	case "show_object_class", "show_object_source", "show_object_field":
		show.Namespace = "object"

	case "show_event_source", "show_event_field":
		show.Namespace = "event"

	case "show_logging_source", "show_logging_field":
		show.Namespace = "logging"

	case "show_tracing_service", "show_tracing_source", "show_tracing_field":
		show.Namespace = "tracing"

	case "show_rum_type", "show_rum_source", "show_rum_field":
		show.Namespace = "rum"

	case "show_measurement", "show_field_key", "show_tag_key", "show_tag_value":
		show.Namespace = "metric"

	case "show_security_category", "show_security_source", "show_security_field":
		show.Namespace = "security"

	default:
		return nil, fmt.Errorf("unknown show function")
	}
	show.Func = f
	return show, nil
}

func (p *parser) newLambda(left *DFQuery, op Item, where NodeList) *Lambda {
	ld := &Lambda{Left: left, WhereCondition: where}

	switch op.Typ {
	case FILTER:
		ld.Opt = LambdaFilter
	case LINK:
		ld.Opt = LambdaLink
	}

	return ld
}

func (p *parser) newTarget(n Node, alias string) (*Target, error) {
	t := Target{Col: n, Alias: alias}

	switch v := n.(type) {
	case *FuncExpr:
		val, fill, err := v.SplitFill()
		if err != nil {
			return nil, err
		}

		if fill != nil {
			t.Col = val
			t.Fill = fill
		}
	}

	return &t, nil
}

func (p *parser) newSubquery(m *DFQuery) *DFQuery {
	ret := &DFQuery{
		Subquery: m,
	}

	p.curQuery = ret // cur-metric points to father
	return ret
}

func (p *parser) newBinExpr(l, r Node, op Item) *BinaryExpr {
	switch op.Typ {
	case DIV, MOD:
		rightNumber, ok := r.(*NumberLiteral)
		if ok {
			if rightNumber.IsInt && rightNumber.Int == 0 ||
				!rightNumber.IsInt && rightNumber.Float == 0 {
				p.addParseErrf(p.yyParser.lval.item.PositionRange(), "division or modulo by zero")
				return nil
			}
		}
	}

	return &BinaryExpr{
		RHS: r,
		LHS: l,
		Op:  op.Typ,
	}
}

func (p *parser) newFunc(fname string, args []Node) *FuncExpr {
	agg := &FuncExpr{
		Name:  strings.ToLower(fname),
		Param: args,
	}
	return agg
}

func (p *parser) newOrderByElem(column string, op Item) *OrderByElem {
	order := &OrderByElem{Column: column}

	switch op.Typ {
	case DESC:
		order.Opt = OrderDesc
	case ASC:
		order.Opt = OrderAsc
	}

	return order
}

func (p *parser) newTimeExpr(t *TimeExpr) *TimeExpr {
	if t.IsDuration {
		t.Time = time.Now().Add(-1 * t.Duration)
	}
	return t
}

func (p *parser) newTimeZone(t *StringLiteral) *TimeZone {
	_, err := time.LoadLocation(t.Val)
	if err == nil {
		return &TimeZone{Input: t.Val, TimeZone: t.Val}
	}

	if tz, ok := timezoneList[t.Val]; ok {
		return &TimeZone{Input: t.Val, TimeZone: tz}
	}

	p.addParseErrf(p.yyParser.lval.item.PositionRange(), "invalid time zone or UTC offset")
	return nil
}

// end of yylex.(*parser).newXXXX

type ParseErrors []ParseErr

type ParseErr struct {
	Pos        *PositionRange
	Err        error
	Query      string
	LineOffset int
}

func (e *ParseErr) Error() string {

	if e.Pos == nil {
		return fmt.Sprintf("%s", e.Err)
	}

	pos := int(e.Pos.Start)
	lastLineBrk := -1
	ln := e.LineOffset + 1
	var posStr string

	if pos < 0 || pos > len(e.Query) {
		posStr = "invalid position:"
	} else {
		for i, c := range e.Query[:pos] {
			if c == '\n' {
				lastLineBrk = i
				ln++
			}
		}

		col := pos - lastLineBrk
		posStr = fmt.Sprintf("%d:%d", ln, col)
	}

	return fmt.Sprintf("%s parse error: %s", posStr, e.Err)
}

// impl Error() interface
func (errs ParseErrors) Error() string {
	var errArray []string
	for _, err := range errs {
		errStr := err.Error()
		if errStr != "" {
			errArray = append(errArray, errStr)
		}
	}

	return strings.Join(errArray, "\n")
}

type PositionRange struct {
	Start, End Pos
}

func newShow(f *FuncExpr) (*Show, error) {
	if f == nil {
		return nil, fmt.Errorf("empty show function, should not been here")
	}

	show := &Show{}
	switch strings.ToLower(f.Name) {
	case "show_object_class":
		show.Namespace = "object"

	case "show_event_source":
		show.Namespace = "event"

	case "show_logging_source":
		show.Namespace = "logging"

	case "show_tracing_service":
		show.Namespace = "tracing"

	case "show_rum_type":
		show.Namespace = "rum"

	case "show_measurement", "show_field_key", "show_tag_key", "show_tag_value":
		show.Namespace = "metric"

	default:
		return nil, fmt.Errorf("unknown show function `%s'", f.Name)
	}
	return show, nil
}

var durationRE = regexp.MustCompile("^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?(([0-9]+)us)?(([0-9]+)ns)?$")

func parseDuration(s string) (time.Duration, error) {
	switch s {
	case "0":
		return 0, nil
	case "":
		return 0, fmt.Errorf("empty duration string")
	}

	m := durationRE.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("invalid duration string: %q", s)
	}

	var du time.Duration
	f := func(pos int, mult time.Duration) {
		if m[pos] == "" {
			return
		}

		n, _ := strconv.Atoi(m[pos])
		d := time.Duration(n)
		du += d * mult
	}

	f(2, 60*60*24*365*time.Second) // y
	f(4, 60*60*24*7*time.Second)   // w
	f(6, 60*60*24*time.Second)     // d
	f(8, 60*60*time.Second)        // h
	f(10, 60*time.Second)          // m
	f(12, time.Second)             // s
	f(14, time.Millisecond)        // ms
	f(16, time.Microsecond)        // us
	f(18, time.Nanosecond)         // ns
	return time.Duration(du), nil
}

// ChainFuncsInfo dql support outer funcs list
// 支持外层函数列表
var ChainFuncsInfo = map[string]map[string]string{
	"abs": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"avg": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"cumsum": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"derivative": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"difference": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"first": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"last": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"log10": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"log2": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"max": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"min": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"moving_average": map[string]string{
		"lenArg": "2", "arg0Name": "dql", "arg0Type": "string",
		"arg1Name": "size", "arg1Type": "int",
	},
	"non_negative_derivative": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"non_negative_difference": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"sum": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"count": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
	"count_distinct": map[string]string{
		"lenArg": "1", "arg0Name": "dql", "arg0Type": "string",
	},
}

// (1) 判断是否是支持的 outer func
func checkOuterFuncName(f *FuncExpr, oFunc *OuterFunc, notFirst bool) error {
	fName := strings.ToLower(f.Name)
	for k := range ChainFuncsInfo {
		if k == fName {
			if fName != f.Name {
				f.Name = fName
			}
			oFunc.Func = f
			return nil
		}
	}
	return fmt.Errorf("unsupport outer func: %s", fName)
}

// 获取外层函数的命名参数类型
func getArgType(argVal interface{}) (string, interface{}) {
	if IsStringParam(argVal) {
		return "string", GetStringParam(argVal)
	}
	if argVal, ok := argVal.(*NumberLiteral); ok {
		if argVal.IsInt {
			return "int", int64(argVal.Int)
		}
	}
	return "unknown", nil
}

// (2) 检查outer func的参数是否合法
func checkOuterFuncParam(f *FuncExpr, oFunc *OuterFunc, notFirst bool) error {
	// 第一个外层函数
	if !notFirst {
		return checkFirstOuterFuncParam(f, oFunc)
	}
	// 其他的链接函数
	fInfo := ChainFuncsInfo[f.Name]
	expectLen, err := strconv.Atoi(fInfo["lenArg"])
	if err != nil {
		return err
	}

	// 链式函数，除了第一个函数，其他不需要dql参数
	lenParam := len(f.Param) + 1
	i := 1

	if lenParam != expectLen {
		return fmt.Errorf("outer func %s : must have %d params, but give %d params", f.Name, expectLen, lenParam)
	}

	for i < lenParam {

		exceptName := fInfo["arg"+strconv.Itoa(i)+"Name"]
		exceptType := fInfo["arg"+strconv.Itoa(i)+"Type"]

		var inputArgType string
		var inputArgVal interface{}

		switch f.Param[i-1].(type) {
		case *FuncArg: // 命名参数
			fArg := f.Param[i-1].(*FuncArg)
			inputArgType, inputArgVal = getArgType(fArg.ArgVal)
			if fArg.ArgName != exceptName || inputArgType != exceptType {
				return fmt.Errorf(
					"outer func %s : arg name:type should be %s:%s, but input arg is %s:%s",
					f.Name, exceptName, exceptType, fArg.ArgName, inputArgType,
				)
			}
		case *NumberLiteral: // 非命名参数
			fNumber := f.Param[i-1].(*NumberLiteral)
			if fNumber.IsInt {
				inputArgType = "int"
				inputArgVal = fNumber.Int
			} else {
				inputArgType = "float"
				inputArgVal = fNumber.Float
			}
			if inputArgType != exceptType {
				return fmt.Errorf(
					"outer func %s : the arg %s type should be %s, but input type is %s",
					f.Name, exceptName, exceptType, inputArgType,
				)
			}
		default:
			return fmt.Errorf("outer func %s args error", f.Name)
		}

		oFunc.FuncArgTypes = append(oFunc.FuncArgTypes, inputArgType)
		oFunc.FuncArgVals = append(oFunc.FuncArgVals, inputArgVal)
		oFunc.FuncArgNames = append(oFunc.FuncArgNames, exceptName)
		i = i + 1
	}
	return nil
}

// (2) 检查outer func的参数是否合法
func checkFirstOuterFuncParam(f *FuncExpr, oFunc *OuterFunc) error {
	fInfo := ChainFuncsInfo[f.Name]
	expectLen, err := strconv.Atoi(fInfo["lenArg"])
	if err != nil {
		return err
	}
	lenParam := len(f.Param)
	i := 0
	if lenParam != expectLen {
		return fmt.Errorf("outer func %s : must have %d params, but give %d params", f.Name, expectLen, lenParam)
	}

	for i < lenParam {
		exceptName := fInfo["arg"+strconv.Itoa(i)+"Name"]
		exceptType := fInfo["arg"+strconv.Itoa(i)+"Type"]

		var inputArgType string
		var inputArgVal interface{}

		switch f.Param[i].(type) {
		case *FuncArg: // 命名参数
			fArg := f.Param[i].(*FuncArg)
			inputArgType, inputArgVal = getArgType(fArg.ArgVal)
			if fArg.ArgName != exceptName || inputArgType != exceptType {
				return fmt.Errorf(
					"outer func %s : arg name:type should be %s:%s, but input arg is %s:%s",
					f.Name, exceptName, exceptType, fArg.ArgName, inputArgType,
				)
			}
		case *NumberLiteral: // 非命名参数, 数值类型（外层函数的首个函数第一个参数必须是dql字符串）
			fNumber := f.Param[i].(*NumberLiteral)
			if fNumber.IsInt {
				inputArgType = "int"
				inputArgVal = fNumber.Int
			} else {
				inputArgType = "float"
				inputArgVal = fNumber.Float
			}
			if inputArgType != exceptType {
				return fmt.Errorf(
					"outer func %s : the arg %s type should be %s, but input type is %s",
					f.Name, exceptName, exceptType, inputArgType,
				)
			}
		case *Identifier, *StringLiteral: // 非命名参数，字符串类型
			inputArgVal = GetStringParam(f.Param[i])
			inputArgType = "string"
			if inputArgType != exceptType {
				return fmt.Errorf(
					"outer func %s : the arg %s type should be %s, but input type is %s",
					f.Name, exceptName, exceptType, inputArgType,
				)
			}
		default:
			return fmt.Errorf("outer func %s args error", f.Name)
		}
		oFunc.FuncArgTypes = append(oFunc.FuncArgTypes, inputArgType)
		oFunc.FuncArgVals = append(oFunc.FuncArgVals, inputArgVal)
		oFunc.FuncArgNames = append(oFunc.FuncArgNames, exceptName)
		i = i + 1
	}
	return nil
}

func (f *FuncExpr) lowerFuncName() string {
	return strings.ToLower(f.Name)
}

func (f *FuncExpr) isShowFunc() bool {
	fName := f.lowerFuncName()
	return len(fName) > 5 && fName[:5] == "show_"
}

func (f *FuncExpr) isDeleteFunc() bool {
	fName := f.lowerFuncName()
	return fName == "delete"
}

// check chain funcs
func (oFuncs *OuterFuncs) checkFuncsSort(f *FuncExpr) error {
	if len(oFuncs.Funcs) == 1 {
		preFunc := oFuncs.Funcs[0].Func
		if preFunc.isShowFunc() || preFunc.isDeleteFunc() { // show 和 delete函数不能链式调用
			return fmt.Errorf("you can not call the func chain: %s func, %s func", preFunc.lowerFuncName(), f.lowerFuncName())
		}
	}
	return nil
}

// newOuterFunc
func (p *parser) newOuterFunc(chainFuncs *OuterFuncs, f *FuncExpr) (interface{}, error) {
	if f == nil {
		return nil, fmt.Errorf("empty outer function")
	}
	// 1) 可能为show函数
	if f.isShowFunc() {
		return p.newShow(f)
	}
	// 2) 可能为delete函数
	if f.isDeleteFunc() {
		return p.newDeleteFunc(f)
	}
	// 3) 外层链接调用函数
	if chainFuncs == nil || len(chainFuncs.Funcs) == 0 {
		chainFuncs = &OuterFuncs{
			Funcs: []*OuterFunc{},
		}
	}

	var err error

	oFunc := &OuterFunc{}

	checkFuncs := []func(*FuncExpr, *OuterFunc, bool) error{
		checkOuterFuncName,  // (1) 检查是否是支持的 outer func
		checkOuterFuncParam, // (2) 检查参数
		// checkOuterFuncSort 检查多个外层函数的嵌套是否有意义, e.g. max(dql='').top(3)
	}
	for _, checkFunc := range checkFuncs {
		err = checkFunc(f, oFunc, len(chainFuncs.Funcs) >= 1)
		if err != nil {
			return nil, err
		}
	}
	oFunc.Func = f
	err = chainFuncs.checkFuncsSort(f)
	if err != nil {
		return nil, err
	}
	chainFuncs.Funcs = append(chainFuncs.Funcs, oFunc)
	return chainFuncs, nil

}

// 外层函数，删除查询出来的结果
func (p *parser) newDeleteFunc(f *FuncExpr) (interface{}, error) {
	if f == nil {
		return nil, fmt.Errorf("empty outer delete function")
	}
	fName := strings.ToLower(f.Name)
	if fName != "delete" {
		return nil, fmt.Errorf("unsupport func")
	}

	lenParam := len(f.Param)
	if lenParam != 1 {
		return nil, fmt.Errorf("outer delete func %s : must have 1 params, but give %d params", f.Name, lenParam)
	}

	var (
		inputArgVal       string
		inputArgType      string
		deleteIndex       bool
		deleteMeasurement bool
	)

	switch f.Param[0].(type) {
	case *FuncArg: // 命名参数, 必须为dql或者index,measurement
		fArg := f.Param[0].(*FuncArg)
		if IsStringParam(fArg.ArgVal) {
			inputArgType, inputArgVal = "string", GetStringParam(fArg.ArgVal)
		}
		if (fArg.ArgName != "index" && fArg.ArgName != "dql" && fArg.ArgName != "measurement") || inputArgType != "string" {
			return nil, fmt.Errorf(
				"outer delete func: arg name:type should be dql:string or index:string measurement:string, but input arg is %s:%s",
				fArg.ArgName, inputArgType,
			)
		}
		if fArg.ArgName == "index" {
			deleteIndex = true
		} else if fArg.ArgName == "measurement" {
			deleteMeasurement = true
		}
	case *Identifier, *StringLiteral: // 非命名参数，字符串类型
		inputArgVal = GetStringParam(f.Param[0])
	default:
		return nil, fmt.Errorf("outer func delete args error, should be string dql")
	}

	deleteFunc := DeleteFunc{
		Func:              f,
		StrDql:            inputArgVal,
		DeleteIndex:       deleteIndex,
		DeleteMeasurement: deleteMeasurement,
	}
	return &deleteFunc, nil

}

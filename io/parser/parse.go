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

	context interface{}
}

func GetConds(input string) WhereConditions {
	log.Debug(input)

	var err error
	p := newParser(input)
	defer parserPool.Put(p)
	defer p.recover(&err)

	p.doParse()

	if len(p.errs) > 0 {
		log.Error(p.errs.Error())

		return nil
	}

	return p.parseResult.(WhereConditions)
}

func newParser(input string) *parser {
	p := parserPool.Get().(*parser)

	// reset parser fields
	p.injecting = false
	p.errs = nil
	p.parseResult = nil
	p.lex = Lexer{
		input: input,
		state: lexStatements,
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
	p.InjectItem(START_WHERE_CONDITION)
	p.yyParser.Parse(p)
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

	p.context = p.parseResult
	return m, nil
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

func (p *parser) newWhereConditions(conditions []Node) *WhereCondition {
	return &WhereCondition{
		conditions: conditions,
	}
}

package parser

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/prometheus/util/strutil"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var log = logger.DefaultSLogger("parser")

type Node interface {
	String() string
	Pos() *PositionRange
}

type Ast struct {
	Functions []*FuncExpr
}

func (e *Ast) String() string {
	arr := []string{}
	for _, f := range e.Functions {
		arr = append(arr, f.String())
	}
	return strings.Join(arr, "\n")
}

func (e *Ast) Pos() *PositionRange { return nil } // TODO

type IndexExpr struct {
	Obj   *Identifier
	Index []int64
}

func (e *IndexExpr) String() string {
	x := ""
	if e.Obj != nil {
		x = e.Obj.String()
	}
	for i := range e.Index {
		x += fmt.Sprintf("[%d]", e.Index[i])
	}

	return x
}
func (e *IndexExpr) Pos() *PositionRange { return nil } // TODO

type AttrExpr struct {
	Obj  Node
	Attr Node
}

func (e *AttrExpr) String() string {
	if e.Attr != nil {
		if e.Obj == nil {
			return e.Attr.String()
		}
		return e.Obj.String() + "." + e.Attr.String()
	} else {
		return e.Obj.String()
	}
}

func (e *AttrExpr) Pos() *PositionRange { return nil } // TODO

type Identifier struct { // impl Expr
	Name string
}

func (e *Identifier) String() string      { return e.Name }
func (e *Identifier) Pos() *PositionRange { return nil } // TODO

type NumberLiteral struct {
	IsInt bool
	Float float64
	Int   int64
}

func (e *NumberLiteral) Pos() *PositionRange { return nil } // not used
func (e *NumberLiteral) String() string {
	if e.IsInt {
		return fmt.Sprintf("%d", e.Int)
	} else {
		return fmt.Sprintf("%f", e.Float)
	}
}

type StringLiteral struct {
	Val string
}

func (e *StringLiteral) Pos() *PositionRange { return nil /* TODO */ }
func (e *StringLiteral) String() string      { return fmt.Sprintf("'%s'", e.Val) }

type BoolLiteral struct {
	Val bool
}

func (n *BoolLiteral) Pos() *PositionRange { return nil }
func (n *BoolLiteral) String() string {
	return fmt.Sprintf("%v", n.Val)
}

type NilLiteral struct{}

func (n *NilLiteral) Pos() *PositionRange { return nil }
func (n *NilLiteral) String() string {
	return "nil"
}

type Regex struct {
	Regex string
}

func (e *Regex) String() string      { return fmt.Sprintf("re('%s')", e.Regex) }
func (e *Regex) Pos() *PositionRange { return nil } // TODO

type Jspath struct {
	Jspath string
}

func (j *Jspath) String() string      { return fmt.Sprintf("jp('%s')", j.Jspath) }
func (j *Jspath) Pos() *PositionRange { return nil } // TODO

type ParenExpr struct {
	Param Node
}

func (e *ParenExpr) Pos() *PositionRange { return nil } // TODO
func (e *ParenExpr) String() string {
	return fmt.Sprintf("(%s)", e.Param.String())
}

type BinaryExpr struct { // impl Expr & Node
	Op         ItemType
	LHS, RHS   Node
	ReturnBool bool
}

func (e *BinaryExpr) Pos() *PositionRange { return nil } // TODO
func (e *BinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.LHS.String(), e.Op.String(), e.RHS.String())
}

type FuncExpr struct {
	Name  string
	RunOk bool
	Param []Node
}

func (n *FuncExpr) String() string {
	arr := []string{}
	for _, n := range n.Param {
		arr = append(arr, n.String())
	}
	return fmt.Sprintf("%s(%s)", strings.ToLower(n.Name), strings.Join(arr, ", "))
}

func (n *FuncExpr) Pos() *PositionRange { return nil } // TODO

type Funcs []Node

func (n Funcs) Pos() *PositionRange { return nil }
func (n Funcs) String() string {
	arr := []string{}
	for _, n := range n {
		arr = append(arr, n.String())
	}

	return strings.Join(arr, "; ")
}

type NodeList []Node

func (n NodeList) Pos() *PositionRange { return nil }
func (n NodeList) String() string {
	arr := []string{}
	for _, arg := range n {
		arr = append(arr, arg.String())
	}
	return strings.Join(arr, ", ")
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

var parserPool = sync.Pool{
	New: func() interface{} {
		return &parser{}
	},
}

type parser struct {
	lex      Lexer
	yyParser yyParserImpl

	parseResult Node
	lastClosing Pos
	errs        ParseErrors

	inject    ItemType
	injecting bool
	context   interface{}
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
	p.errs = append(p.errs, ParseError{
		Pos:   pr,
		Err:   err,
		Query: p.lex.input,
	})
}

func (p *parser) addParseErrf(pr *PositionRange, format string, args ...interface{}) {
	p.addParseErr(pr, fmt.Errorf(format, args...))
}

func (p *parser) unquoteString(s string) string {
	unq, err := strutil.Unquote(s)
	if err != nil {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			"error unquoting string %q: %s", s, err)
	}
	return unq
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

// impl Lex interface.
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
	case RIGHT_PAREN:
		p.lastClosing = lval.item.Pos + Pos(len(lval.item.Val))
	}
	return int(typ)
}

func (p *parser) Error(e string) {}

func newParser(input string) *parser {
	p := parserPool.Get().(*parser)

	p.injecting = false
	p.errs = nil
	p.parseResult = nil
	p.lex = Lexer{
		input: input,
		state: lexStatements,
	}
	return p
}

// end of yylex.(*parser).newXXXX

type ParseErrors []ParseError

type ParseError struct {
	Pos        *PositionRange
	Err        error
	Query      string
	LineOffset int
}

func (e *ParseError) Error() string {
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

// impl Error() interface.
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

func ParsePipeline(input string) (res Node, err error) {
	p := newParser(input)
	defer parserPool.Put(p)
	defer p.recover(&err)

	p.InjectItem(START_PIPELINE)
	p.yyParser.Parse(p)

	if p.parseResult != nil {
		res = p.parseResult
	}

	if len(p.errs) != 0 {
		err = p.errs
	}

	return res, err
}

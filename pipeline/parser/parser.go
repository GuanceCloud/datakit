package parser

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/prometheus/util/strutil"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var log = logger.DefaultSLogger("parser")

func InitLog() {
	log = logger.SLogger("parser")
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
	e := recover() //nolint: ifshort
	if _, ok := e.(runtime.Error); ok {
		buf := make([]byte, 64<<10) // 64k
		buf = buf[:runtime.Stack(buf, false)]
		fmt.Fprintf(os.Stderr, "parser panic: %v\n%s", e, buf)
		*errp = errUnexpected
	} else if e != nil {
		if x, ok := e.(error); ok {
			*errp = x
		}
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

func (p *parser) newIfelseStmt(ifExpr Node, elifList []Node, elseExpr Node) *IfelseStmt {
	if ifExpr == nil {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(), "invalid if expression is empty")
		return nil
	}

	ie, ok := ifExpr.(*IfExpr)
	if !ok {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			fmt.Sprintf("invalid if expression type %s", reflect.TypeOf(ifExpr)))
		return nil
	}
	ifList := IfList{ie}

	for _, e := range elifList {
		if i, ok := e.(*IfExpr); ok {
			ifList = append(ifList, i)
		} else {
			p.addParseErrf(p.yyParser.lval.item.PositionRange(),
				fmt.Sprintf("invalid elif expression type %s", reflect.TypeOf(e)))
			return nil
		}
	}

	expr := &IfelseStmt{
		IfList: ifList,
	}

	if v, ok := elseExpr.(Stmts); ok {
		expr.Else = v
	}

	return expr
}

func (p *parser) newIfExpr(condition Node, stmts Node) *IfExpr {
	if condition == nil {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(), "invalid if/elif condition")
		return nil
	}

	ifexpr := &IfExpr{Condition: condition}

	if stmts != nil {
		if s, ok := stmts.(Stmts); ok {
			ifexpr.Stmts = s
		} else {
			p.addParseErrf(p.yyParser.lval.item.PositionRange(),
				fmt.Sprintf("invalid if/elif statements type %s", reflect.TypeOf(stmts)))
			return nil
		}
	}

	return ifexpr
}

func (p *parser) newAssignmentStmt(l, r Node) *AssignmentStmt {
	return &AssignmentStmt{
		LHS: l,
		RHS: r,
	}
}

func (p *parser) newConditionalExpr(l, r Node, op Item) *ConditionalExpr {
	return &ConditionalExpr{
		RHS: r,
		LHS: l,
		Op:  op.Typ,
	}
}

/*
func (p *parser) newComputationExpr(l, r Node, op Item) *ComputationExpr {
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

	return &ComputationExpr{
		RHS: r,
		LHS: l,
		Op:  op.Typ,
	}
}
*/

func (p *parser) newIndexExpr(obj Node, index Node) Node {
	if index == nil {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(), "invalid array index is emepty")
		return nil
	}

	var idx int64

	switch v := index.(type) {
	case *NumberLiteral:
		if !v.IsInt {
			p.addParseErrf(p.yyParser.lval.item.PositionRange(),
				fmt.Sprintf("array index should be int, got `%f'", v.Float))
			return nil
		} else {
			idx = v.Int
		}
	default:
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			fmt.Sprintf("invalid indexExpr type %s", reflect.TypeOf(index)))
		return nil
	}

	if obj == nil {
		return &IndexExpr{Index: []int64{idx}}
	}

	switch v := obj.(type) {
	case *Identifier:
		return &IndexExpr{Obj: v, Index: []int64{idx}}
	case *IndexExpr:
		v.Index = append(v.Index, idx)
		return v
	default:
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			fmt.Sprintf("invalid indexExpr object type %s", reflect.TypeOf(obj)))
	}
	return nil
}

func (p *parser) newFuncStmt(fname string, args []Node) (*FuncStmt, error) {
	f := &FuncStmt{
		Name: strings.ToLower(fname),
	}

	for _, arg := range args {
		switch v := arg.(type) {
		/*
			switch lv := v.LHS.(type) {
			case *Identifier, *AttrExpr:
				f.KwParam[lv.String()] = v.RHS
			default:
				return nil, fmt.Errorf("invalid arg name: %s", lv.String())
			}
		*/

		case *Identifier, *AttrExpr, *StringLiteral,
			*NumberLiteral, *Regex, *FuncStmt,
			FuncArgList, *AssignmentStmt,
			*NilLiteral, *BoolLiteral, *ComputationExpr:
			f.Param = append(f.Param, v)

		default:
			return nil, fmt.Errorf("unknown arg type %s: %s", reflect.TypeOf(v).String(), v.String())
		}
	}

	if len(f.Param) > 0 && len(f.KwParam) > 0 {
		return nil, fmt.Errorf("naming args can't passing with anonymous args")
	}

	return f, nil
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
	p, ok := parserPool.Get().(*parser)
	if !ok {
		return nil
	}

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

func ParsePipeline(input string) (res Node, err error) {
	p := newParser(input)
	defer parserPool.Put(p)
	defer p.recover(&err)

	p.InjectItem(START_STMTS)
	p.yyParser.Parse(p)

	if p.parseResult != nil {
		res = p.parseResult
	}

	if len(p.errs) != 0 {
		err = p.errs
	}

	return res, err
}

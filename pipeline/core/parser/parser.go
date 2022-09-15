// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package parser

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
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

	parseResult ast.Stmts
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

func (p *parser) number(v string) *ast.NumberLiteral {
	nl := &ast.NumberLiteral{}

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
	unq, err := Unquote(s)
	if err != nil {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			"error unquoting string %q: %s", s, err)
	}
	return unq
}

func (p *parser) unquoteMultilineString(s string) string {
	unq, err := UnquoteMultiline(s)
	if err != nil {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			"error unquoting multiline string %q: %s", s, err)
	}
	return unq
}

func (p *parser) newBreakStmt() *ast.Node {
	return ast.WrapBreakStmt(&ast.BreakStmt{})
}

func (p *parser) newContinueStmt() *ast.Node {
	return ast.WrapContinueStmt(&ast.ContinueStmt{})
}

func (p *parser) newForStmt(initExpr *ast.Node, condExpr *ast.Node, loopExpr *ast.Node, body ast.Stmts) *ast.Node {
	return ast.WrapForStmt(&ast.ForStmt{
		Init: initExpr,
		Loop: loopExpr,
		Cond: condExpr,
		Body: body,
	})
}

func (p *parser) newForInStmt(varb string, iter *ast.Node, body ast.Stmts) *ast.Node {
	switch iter.NodeType { //nolint:exhaustive
	case ast.TypeBoolLiteral, ast.TypeNilLiteral,
		ast.TypeNumberLiteral:
		p.addParseErrf(p.yyParser.lval.item.PositionRange(), "%s object is not iterable", iter.NodeType)
	}
	return ast.WrapForInStmt(&ast.ForInStmt{
		Varb: ast.WrapIdentifier(&ast.Identifier{Name: varb}),
		Iter: iter,
		Body: body,
	})
}

func (p *parser) newIfelseStmt(ifElseStmt *ast.Node, ifExpr *ast.IfStmtElem,
	elifExpr *ast.IfStmtElem, elseExpr ast.Stmts,
) *ast.Node {
	if ifElseStmt == nil {
		if ifExpr == nil {
			p.addParseErrf(p.yyParser.lval.item.PositionRange(), "invalid if expression is empty")
			return nil
		} else { // 创建 if
			return ast.WrapIfelseStmt(&ast.IfelseStmt{
				IfList: ast.IfList{ifExpr},
			})
		}
	}

	if ifElseStmt.NodeType != ast.TypeIfelseStmt {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			fmt.Sprintf("invalid if expression type %s", ifElseStmt.NodeType))
		return nil
	}

	if elifExpr != nil {
		ifElseStmt.IfelseStmt.IfList = append(ifElseStmt.IfelseStmt.IfList, elifExpr)
	}

	if elseExpr != nil {
		ifElseStmt.IfelseStmt.Else = elseExpr
	}

	return ifElseStmt
}

func (p *parser) newIfExpr(condition *ast.Node, stmts ast.Stmts) *ast.IfStmtElem {
	if condition == nil {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(), "invalid if/elif condition")
		return nil
	}

	ifexpr := &ast.IfStmtElem{
		Condition: condition,
		Stmts:     stmts,
	}

	return ifexpr
}

func (p *parser) newAssignmentExpr(l, r *ast.Node) *ast.Node {
	return ast.WrapAssignmentExpr(&ast.AssignmentExpr{
		LHS: l,
		RHS: r,
	})
}

func (p *parser) newConditionalExpr(l, r *ast.Node, op Item) *ast.Node {
	return ast.WrapConditionExpr(&ast.ConditionalExpr{
		RHS: r,
		LHS: l,
		Op:  AstOp(op.Typ),
	})
}

func (p *parser) newArithmeticExpr(l, r *ast.Node, op Item) *ast.Node {
	switch op.Typ {
	case DIV, MOD:
		if r.NodeType == ast.TypeNumberLiteral {
			if r.NumberLiteral.IsInt && r.NumberLiteral.Int == 0 ||
				!r.NumberLiteral.IsInt && r.NumberLiteral.Float == 0 {
				p.addParseErrf(p.yyParser.lval.item.PositionRange(), "division or modulo by zero")
				return nil
			}
		}
	}

	return ast.WrapArithmeticExpr(
		&ast.ArithmeticExpr{
			RHS: r,
			LHS: l,
			Op:  AstOp(op.Typ),
		},
	)
}

func (p *parser) newAttrExpr(obj, attr *ast.Node) *ast.Node {
	return ast.WrapAttrExpr(&ast.AttrExpr{
		Obj:  obj,
		Attr: attr,
	})
}

func (p *parser) newIndexExpr(obj, index *ast.Node) *ast.Node {
	if index == nil {
		p.addParseErrf(p.yyParser.lval.item.PositionRange(), "invalid array index is emepty")
		return nil
	}

	if obj == nil {
		// .[idx]
		return ast.WrapIndexExpr(&ast.IndexExpr{
			Index: []*ast.Node{index},
		})
	}

	switch obj.NodeType { //nolint:exhaustive
	case ast.TypeIdentifier:
		return ast.WrapIndexExpr(&ast.IndexExpr{
			Obj: obj.Identifier, Index: []*ast.Node{index},
		})
	case ast.TypeIndexExpr:
		obj.IndexExpr.Index = append(obj.IndexExpr.Index, index)
		return obj
	default:
		p.addParseErrf(p.yyParser.lval.item.PositionRange(),
			fmt.Sprintf("invalid indexExpr object type %s", obj.NodeType))
	}
	return nil
}

func (p *parser) newCallExpr(fname string, args []*ast.Node) (*ast.Node, error) {
	f := &ast.CallExpr{
		Name: strings.ToLower(fname),
	}

	// TODO: key-value param support
	f.Param = append(f.Param, args...)

	return ast.WrapCallExpr(f), nil
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

func ParsePipeline(input string) (res ast.Stmts, err error) {
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

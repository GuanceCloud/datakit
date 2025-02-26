// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ast

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/GuanceCloud/grok"
	"github.com/GuanceCloud/platypus/pkg/token"
)

type Op string

const (
	ADD Op = "+"
	SUB Op = "-"
	MUL Op = "*"
	DIV Op = "/"
	MOD Op = "%"

	// XOR Op = "^"
	// ~~~ POW Op = "^" ~~~.

	EQEQ Op = "=="
	NEQ  Op = "!="
	LTE  Op = "<="
	LT   Op = "<"
	GTE  Op = ">="
	GT   Op = ">"

	AND Op = "&&"
	OR  Op = "||"
	NOT Op = "!"

	EQ    Op = "="
	ADDEQ Op = "+="
	SUBEQ Op = "-="
	MULEQ Op = "*="
	DIVEQ Op = "/="
	MODEQ Op = "%="
)

type Identifier struct {
	Name  string
	Start token.LnColPos
}

func (e *Identifier) IsExpr() bool {
	return true
}

func (e *Identifier) String() string {
	return e.Name
}

type StringLiteral struct {
	Val   string
	Start token.LnColPos
}

func (e *StringLiteral) IsExpr() bool {
	return true
}

func (e *StringLiteral) String() string {
	return fmt.Sprintf("'%s'", e.Val)
}

// type NumberLiteral struct {
// 	IsInt bool
// 	Float float64
// 	Int   int64
// 	Start token.FilePos
// 	End   token.FilePos
// }

type IntegerLiteral struct {
	Val   int64
	Start token.LnColPos
}

func (e *IntegerLiteral) String() string {
	return fmt.Sprintf("%d", e.Val)
}

func (e *IntegerLiteral) IsExpr() bool {
	return true
}

type FloatLiteral struct {
	Val   float64
	Start token.LnColPos
}

func (e *FloatLiteral) String() string {
	return fmt.Sprintf("%f", e.Val)
}

func (e *FloatLiteral) IsExpr() bool {
	return true
}

type BoolLiteral struct {
	Val   bool
	Start token.LnColPos
}

func (e *BoolLiteral) String() string {
	return fmt.Sprintf("%v", e.Val)
}

type NilLiteral struct {
	Start token.LnColPos
}

func (e *NilLiteral) IsExpr() bool {
	return true
}

func (e *NilLiteral) String() string {
	return "nil"
}

type MapLiteral struct {
	KeyValeList [][2]*Node // key,value list
	LBrace      token.LnColPos
	RBrace      token.LnColPos
}

func (e *MapLiteral) IsExpr() bool {
	return true
}

func (e *MapLiteral) String() string {
	v := "{"
	for i, item := range e.KeyValeList {
		if i == 0 {
			v += item[0].String() + ": " + item[1].String()
		} else {
			v += ", " + item[0].String() + ": " + item[1].String()
		}
	}
	return v + "}"
}

type ListLiteral struct {
	List     []*Node
	LBracket token.LnColPos
	RBracket token.LnColPos
}

func (e *ListLiteral) IsExpr() bool {
	return true
}

func (e *ListLiteral) String() string {
	arr := []string{}
	for _, x := range e.List {
		arr = append(arr, x.String())
	}
	return "[" + strings.Join(arr, ", ") + "]"
}

type UnaryExpr struct {
	Op    Op
	RHS   *Node
	OpPos token.LnColPos
}

func (e *UnaryExpr) IsExpr() bool {
	return true
}

func (e *UnaryExpr) String() string {
	return fmt.Sprintf("%s%s", e.Op, e.RHS.String())
}

type ConditionalExpr struct {
	Op       Op
	LHS, RHS *Node
	OpPos    token.LnColPos
}

func (e *ConditionalExpr) IsExpr() bool {
	return true
}

func (e *ConditionalExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.LHS.String(), e.Op, e.RHS.String())
}

type InExpr struct {
	Op       Op
	LHS, RHS *Node
	OpPos    token.LnColPos
}

func (e *InExpr) IsExpr() bool {
	return true
}

func (e *InExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.LHS.String(), e.Op, e.RHS.String())
}

type ArithmeticExpr struct {
	Op       Op
	LHS, RHS *Node
	OpPos    token.LnColPos
}

func (e *ArithmeticExpr) IsExpr() bool {
	return true
}

func (e *ArithmeticExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.LHS.String(), e.Op, e.RHS.String())
}

type AttrExpr struct {
	Obj   *Node
	Attr  *Node
	Start token.LnColPos
}

func (e *AttrExpr) IsExpr() bool {
	return true
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

type IndexExpr struct {
	Obj      *Identifier
	Index    []*Node // int float string bool
	LBracket []token.LnColPos
	RBracket []token.LnColPos
}

func (e *IndexExpr) IsExpr() bool {
	return true
}

func (e *IndexExpr) String() string {
	x := ""
	if e.Obj != nil {
		x = e.Obj.String()
	}
	for _, i := range e.Index {
		x += fmt.Sprintf("[%s]", i)
	}

	return x
}

type ParenExpr struct {
	Param  *Node
	LParen token.LnColPos
	RParen token.LnColPos
}

func (e *ParenExpr) IsExpr() bool {
	return true
}

func (e *ParenExpr) String() string {
	return fmt.Sprintf("(%s)", e.Param.String())
}

type CallExpr struct {
	// TODO
	// Name *ast.Node

	// temporary record function name location
	NamePos token.LnColPos // as 'Start' (token.FilePos)

	LParen token.LnColPos
	RParen token.LnColPos

	Name string

	Param           []*Node
	ParamNormalized []*Node

	PrivateData interface{}

	// ParamIndex []int

	Grok *grok.GrokRegexp
	Re   *regexp.Regexp
}

func (e *CallExpr) IsExpr() bool {
	return true
}

func (e *CallExpr) String() string {
	arr := []string{}
	for _, n := range e.Param {
		arr = append(arr, n.String())
	}
	return fmt.Sprintf("%s(%s)", strings.ToLower(e.Name), strings.Join(arr, ", "))
}

type SliceExpr struct {
	Obj      *Node
	Start    *Node
	End      *Node
	Step     *Node
	Colon2   bool
	LBracket token.LnColPos
	RBracket token.LnColPos
}

func (e *SliceExpr) IsExpr() bool {
	return true
}

func (e *SliceExpr) String() string {
	startStr := ""
	if e.Start != nil {
		startStr = e.Start.String()
	}
	endStr := ""
	if e.End != nil {
		endStr = e.End.String()
	}
	stepStr := ""
	if e.Step != nil {
		stepStr = ":" + e.Step.String()
	} else if e.Colon2 {
		stepStr = ":"
	}
	return fmt.Sprintf("%s[%s:%s%s]", e.Obj.String(), startStr, endStr, stepStr)
}

type AssignmentExpr struct {
	LHS, RHS *Node
	Op       Op
	OpPos    token.LnColPos
}

func (e *AssignmentExpr) IsExpr() bool {
	return true
}

func (e *AssignmentExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.LHS, e.Op, e.RHS)
}

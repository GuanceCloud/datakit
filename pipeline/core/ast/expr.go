// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ast

import (
	"fmt"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/grok"
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

	EQ Op = "="
)

type Identifier struct {
	Name string
}

func (e *Identifier) IsExpr() bool {
	return true
}

func (e *Identifier) String() string {
	return e.Name
}

type StringLiteral struct{ Val string }

func (e *StringLiteral) IsExpr() bool {
	return true
}

func (e *StringLiteral) String() string {
	return fmt.Sprintf("'%s'", e.Val)
}

type NumberLiteral struct {
	IsInt bool
	Float float64
	Int   int64
}

func (e *NumberLiteral) IsExpr() bool {
	return true
}

func (e *NumberLiteral) String() string {
	if e.IsInt {
		return fmt.Sprintf("%d", e.Int)
	} else {
		return fmt.Sprintf("%f", e.Float)
	}
}

func (e *NumberLiteral) Value() any {
	if e.IsInt {
		return e.Int
	} else {
		return e.Float
	}
}

type BoolLiteral struct{ Val bool }

func (e *BoolLiteral) String() string {
	return fmt.Sprintf("%v", e.Val)
}

type NilLiteral struct{}

func (e *NilLiteral) IsExpr() bool {
	return true
}

func (e *NilLiteral) String() string {
	return "nil"
}

type MapInitExpr struct {
	KeyValeList [][2]*Node // key,value list
}

func (e *MapInitExpr) IsExpr() bool {
	return true
}

func (e *MapInitExpr) String() string {
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

type ListInitExpr struct {
	List []*Node
}

func (e *ListInitExpr) IsExpr() bool {
	return true
}

func (e *ListInitExpr) String() string {
	arr := []string{}
	for _, x := range e.List {
		arr = append(arr, x.String())
	}
	return "[" + strings.Join(arr, ", ") + "]"
}

type ConditionalExpr struct {
	Op       Op
	LHS, RHS *Node
}

func (e *ConditionalExpr) IsExpr() bool {
	return true
}

func (e *ConditionalExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.LHS.String(), e.Op, e.RHS.String())
}

type ArithmeticExpr struct {
	Op       Op
	LHS, RHS *Node
}

func (e *ArithmeticExpr) IsExpr() bool {
	return true
}

func (e *ArithmeticExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.LHS.String(), e.Op, e.RHS.String())
}

type AttrExpr struct {
	Obj  *Node
	Attr *Node
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
	Obj   *Identifier
	Index []*Node // int float string bool
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
	Param *Node
}

func (e *ParenExpr) IsExpr() bool {
	return true
}

func (e *ParenExpr) String() string {
	return fmt.Sprintf("(%s)", e.Param.String())
}

type CallExpr struct {
	Name string

	Param []*Node

	// ParamIndex []int

	Grok *grok.GrokRegexp
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

type AssignmentExpr struct {
	LHS, RHS *Node
}

func (e *AssignmentExpr) IsExpr() bool {
	return true
}

func (e *AssignmentExpr) String() string {
	return fmt.Sprintf("%s = %s", e.LHS, e.RHS)
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ast pipeline ast node
package ast

import (
	"fmt"
	"sort"
	"strings"
)

type NodeType uint

func (NodeType) String() string {
	return ""
}

const (
	// expr.
	TypeInvaild NodeType = iota

	TypeIdentifier
	TypeStringLiteral
	TypeNumberLiteral
	TypeBoolLiteral
	TypeNilLiteral

	TypeListInitExpr
	TypeMapInitExpr

	TypeParenExpr

	TypeAttrExpr
	TypeIndexExpr

	TypeArithmeticExpr
	TypeConditionalExpr
	TypeAssignmentExpr

	TypeCallExpr

	// stmt.
	TypeIfelseStmt
	TypeForStmt
	TypeForInStmt
	TypeContinueStmt
	TypeBreakStmt
)

type Stmts []*Node

type KwArgs map[string]*Node

type FuncArgList []*Node

type Regex struct{ Regex string }

type Jspath struct{ Jspath string }

func (e KwArgs) String() string {
	keys := []string{}
	for k := range e {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	arr := []string{}
	for _, key := range keys {
		arr = append(arr, fmt.Sprintf("%s = %s", key, e[key]))
	}
	return strings.Join(arr, ", ")
}

func (e Stmts) String() string {
	arr := []string{}
	for _, x := range e {
		arr = append(arr, x.String())
	}
	return strings.Join(arr, "\n")
}

func (e *Regex) String() string {
	return fmt.Sprintf("re('%s')", e.Regex)
}

func (e *Jspath) String() string {
	return fmt.Sprintf("jp('%s')", e.Jspath)
}

type Node struct {
	// node type
	NodeType NodeType

	// expr
	Identifier *Identifier

	StringLiteral *StringLiteral
	NumberLiteral *NumberLiteral
	BoolLiteral   *BoolLiteral
	NilLiteral    *NilLiteral

	ListInitExpr *ListInitExpr
	MapInitExpr  *MapInitExpr

	ParenExpr *ParenExpr

	AttrExpr  *AttrExpr
	IndexExpr *IndexExpr

	ArithmeticExpr  *ArithmeticExpr
	ConditionalExpr *ConditionalExpr
	AssignmentExpr  *AssignmentExpr

	CallExpr *CallExpr

	// stmt
	IfelseStmt *IfelseStmt

	ForStmt      *ForStmt
	ForInStmt    *ForInStmt
	ContinueStmt *ContinueStmt
	BreakStmt    *BreakStmt
}

func (node *Node) String() string {
	switch node.NodeType { //nolint:exhaustive
	case TypeIdentifier:
		return node.Identifier.String()
	case TypeStringLiteral:
		return node.StringLiteral.String()
	case TypeNumberLiteral:
		return node.NumberLiteral.String()
	case TypeBoolLiteral:
		return node.BoolLiteral.String()
	case TypeNilLiteral:
		return node.NilLiteral.String()
	case TypeListInitExpr:
		return node.ListInitExpr.String()
	case TypeMapInitExpr:
		return node.MapInitExpr.String()
	case TypeParenExpr:
		return node.ParenExpr.String()
	case TypeAttrExpr:
		return node.AttrExpr.String()
	case TypeIndexExpr:
		return node.IndexExpr.String()
	case TypeArithmeticExpr:
		return node.ArithmeticExpr.String()
	case TypeConditionalExpr:
		return node.ConditionalExpr.String()
	case TypeAssignmentExpr:
		return node.AssignmentExpr.String()
	case TypeCallExpr:
		return node.CallExpr.String()
	case TypeIfelseStmt:
		return node.IfelseStmt.String()
	case TypeForStmt:
		return node.ForStmt.String()
	case TypeForInStmt:
		return node.ForInStmt.String()
	case TypeContinueStmt:
		return node.ContinueStmt.String()
	case TypeBreakStmt:
		return node.BreakStmt.String()
	}
	return "node conv to string failed"
}

func WrapIdentifier(node *Identifier) *Node {
	return &Node{
		NodeType:   TypeIdentifier,
		Identifier: node,
	}
}

func WrapStringLiteral(node *StringLiteral) *Node {
	return &Node{
		NodeType:      TypeStringLiteral,
		StringLiteral: node,
	}
}

func WrapNumberLiteral(node *NumberLiteral) *Node {
	return &Node{
		NodeType:      TypeNumberLiteral,
		NumberLiteral: node,
	}
}

func WrapBoolLiteral(node *BoolLiteral) *Node {
	return &Node{
		NodeType:    TypeBoolLiteral,
		BoolLiteral: node,
	}
}

func WrapNilLiteral(node *NilLiteral) *Node {
	return &Node{
		NodeType:   TypeNilLiteral,
		NilLiteral: node,
	}
}

func WrapListInitExpr(node *ListInitExpr) *Node {
	return &Node{
		NodeType:     TypeListInitExpr,
		ListInitExpr: node,
	}
}

func WrapMapInitExpr(node *MapInitExpr) *Node {
	return &Node{
		NodeType:    TypeMapInitExpr,
		MapInitExpr: node,
	}
}

func WrapParenExpr(node *ParenExpr) *Node {
	return &Node{
		NodeType:  TypeParenExpr,
		ParenExpr: node,
	}
}

func WrapAttrExpr(node *AttrExpr) *Node {
	return &Node{
		NodeType: TypeAttrExpr,
		AttrExpr: node,
	}
}

func WrapIndexExpr(node *IndexExpr) *Node {
	return &Node{
		NodeType:  TypeIndexExpr,
		IndexExpr: node,
	}
}

func WrapArithmeticExpr(node *ArithmeticExpr) *Node {
	return &Node{
		NodeType:       TypeArithmeticExpr,
		ArithmeticExpr: node,
	}
}

func WrapConditionExpr(node *ConditionalExpr) *Node {
	return &Node{
		NodeType:        TypeConditionalExpr,
		ConditionalExpr: node,
	}
}

func WrapAssignmentExpr(node *AssignmentExpr) *Node {
	return &Node{
		NodeType:       TypeAssignmentExpr,
		AssignmentExpr: node,
	}
}

func WrapCallExpr(node *CallExpr) *Node {
	return &Node{
		NodeType: TypeCallExpr,
		CallExpr: node,
	}
}

func WrapIfelseStmt(node *IfelseStmt) *Node {
	return &Node{
		NodeType:   TypeIfelseStmt,
		IfelseStmt: node,
	}
}

func WrapForStmt(node *ForStmt) *Node {
	return &Node{
		NodeType: TypeForStmt,
		ForStmt:  node,
	}
}

func WrapForInStmt(node *ForInStmt) *Node {
	return &Node{
		NodeType:  TypeForInStmt,
		ForInStmt: node,
	}
}

func WrapContinueStmt(node *ContinueStmt) *Node {
	return &Node{
		NodeType:     TypeContinueStmt,
		ContinueStmt: node,
	}
}

func WrapBreakStmt(node *BreakStmt) *Node {
	return &Node{
		NodeType:  TypeBreakStmt,
		BreakStmt: node,
	}
}

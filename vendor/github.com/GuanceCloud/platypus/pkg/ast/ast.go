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

	"github.com/GuanceCloud/platypus/pkg/token"
)

type NodeType uint

const (
	// expr.
	TypeInvalid NodeType = iota

	TypeIdentifier
	TypeStringLiteral
	TypeIntegerLiteral
	TypeFloatLiteral
	TypeBoolLiteral
	TypeNilLiteral

	TypeListLiteral
	TypeMapLiteral

	TypeInExpr

	TypeParenExpr

	TypeAttrExpr
	TypeIndexExpr

	TypeUnaryExpr
	TypeArithmeticExpr
	TypeConditionalExpr
	TypeAssignmentExpr

	TypeCallExpr
	TypeSliceExpr

	// stmt.
	TypeBlockStmt
	TypeIfelseStmt
	TypeForStmt
	TypeForInStmt
	TypeContinueStmt
	TypeBreakStmt
)

func (t NodeType) String() string {
	switch t {
	case TypeInvalid:
		return "Invalid"
	case TypeIdentifier:
		return "Identifier"
	case TypeInExpr:
		return "InExpr"
	case TypeStringLiteral:
		return "StringLiteral"
	case TypeIntegerLiteral:
		return "IntLiteral"
	case TypeFloatLiteral:
		return "FloatLiteral"
	case TypeBoolLiteral:
		return "BoolLiteral"
	case TypeNilLiteral:
		return "NilLiteral"
	case TypeListLiteral:
		return "ListLiteral"
	case TypeMapLiteral:
		return "MapLiteral"
	case TypeParenExpr:
		return "ParenExpr"
	case TypeAttrExpr:
		return "AttrExpr"
	case TypeIndexExpr:
		return "IndexExpr"
	case TypeUnaryExpr:
		return "UnaryExpr"
	case TypeArithmeticExpr:
		return "ArithmeticExpr"
	case TypeConditionalExpr:
		return "ConditionalExpr"
	case TypeAssignmentExpr:
		return "AssignmentExpr"
	case TypeCallExpr:
		return "CallExpr"
	case TypeSliceExpr:
		return "SliceExpr"
	case TypeBlockStmt:
		return "BlockStmt"
	case TypeIfelseStmt:
		return "IfelseStmt"
	case TypeForStmt:
		return "ForStmt"
	case TypeForInStmt:
		return "ForInStmt"
	case TypeContinueStmt:
		return "ContinueStmt"
	case TypeBreakStmt:
		return "BreakStmt"
	}
	return "Undefined"
}

type Stmts []*Node

type KwArgs map[string]*Node

type FuncArgList []*Node

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

type AstNode interface {
	String() string
}

type Node struct {
	// node type
	NodeType NodeType
	elem     AstNode
}

func (n *Node) String() string {
	if n.elem != nil {
		return n.elem.String()
	}
	return n.NodeType.String()
}

func (n *Node) Identifier() *Identifier {
	return n.elem.(*Identifier)
}
func (n *Node) StringLiteral() *StringLiteral {
	return n.elem.(*StringLiteral)
}
func (n *Node) IntegerLiteral() *IntegerLiteral {
	return n.elem.(*IntegerLiteral)
}
func (n *Node) FloatLiteral() *FloatLiteral {
	return n.elem.(*FloatLiteral)
}
func (n *Node) BoolLiteral() *BoolLiteral {
	return n.elem.(*BoolLiteral)
}
func (n *Node) NilLiteral() *NilLiteral {
	return n.elem.(*NilLiteral)
}
func (n *Node) ListLiteral() *ListLiteral {
	return n.elem.(*ListLiteral)
}
func (n *Node) MapLiteral() *MapLiteral {
	return n.elem.(*MapLiteral)
}
func (n *Node) ParenExpr() *ParenExpr {
	return n.elem.(*ParenExpr)
}
func (n *Node) AttrExpr() *AttrExpr {
	return n.elem.(*AttrExpr)
}
func (n *Node) IndexExpr() *IndexExpr {
	return n.elem.(*IndexExpr)
}
func (n *Node) InExpr() *InExpr {
	return n.elem.(*InExpr)
}
func (n *Node) UnaryExpr() *UnaryExpr {
	return n.elem.(*UnaryExpr)
}
func (n *Node) ArithmeticExpr() *ArithmeticExpr {
	return n.elem.(*ArithmeticExpr)
}
func (n *Node) ConditionalExpr() *ConditionalExpr {
	return n.elem.(*ConditionalExpr)
}
func (n *Node) AssignmentExpr() *AssignmentExpr {
	return n.elem.(*AssignmentExpr)
}
func (n *Node) CallExpr() *CallExpr {
	return n.elem.(*CallExpr)
}
func (n *Node) SliceExpr() *SliceExpr {
	return n.elem.(*SliceExpr)
}
func (n *Node) BlockStmt() *BlockStmt {
	return n.elem.(*BlockStmt)
}
func (n *Node) IfelseStmt() *IfelseStmt {
	return n.elem.(*IfelseStmt)
}
func (n *Node) ForStmt() *ForStmt {
	return n.elem.(*ForStmt)
}
func (n *Node) ForInStmt() *ForInStmt {
	return n.elem.(*ForInStmt)
}
func (n *Node) ContinueStmt() *ContinueStmt {
	return n.elem.(*ContinueStmt)
}
func (n *Node) BreakStmt() *BreakStmt {
	return n.elem.(*BreakStmt)
}

func (n *Node) StartPos() token.LnColPos {
	return NodeStartPos(n)
}

func WrapIdentifier(node *Identifier) *Node {
	return &Node{
		NodeType: TypeIdentifier,
		elem:     node,
	}
}

func WrapStringLiteral(node *StringLiteral) *Node {
	return &Node{
		NodeType: TypeStringLiteral,
		elem:     node,
	}
}

func WrapIntegerLiteral(node *IntegerLiteral) *Node {
	return &Node{
		NodeType: TypeIntegerLiteral,
		elem:     node,
	}
}

func WrapFloatLiteral(node *FloatLiteral) *Node {
	return &Node{
		NodeType: TypeFloatLiteral,
		elem:     node,
	}
}

func WrapBoolLiteral(node *BoolLiteral) *Node {
	return &Node{
		NodeType: TypeBoolLiteral,
		elem:     node,
	}
}

func WrapNilLiteral(node *NilLiteral) *Node {
	return &Node{
		NodeType: TypeNilLiteral,
		elem:     node,
	}
}

func WrapListInitExpr(node *ListLiteral) *Node {
	return &Node{
		NodeType: TypeListLiteral,
		elem:     node,
	}
}

func WrapMapLiteral(node *MapLiteral) *Node {
	return &Node{
		NodeType: TypeMapLiteral,
		elem:     node,
	}
}

func WrapParenExpr(node *ParenExpr) *Node {
	return &Node{
		NodeType: TypeParenExpr,
		elem:     node,
	}
}

func WrapAttrExpr(node *AttrExpr) *Node {
	return &Node{
		NodeType: TypeAttrExpr,
		elem:     node,
	}
}

func WrapIndexExpr(node *IndexExpr) *Node {
	return &Node{
		NodeType: TypeIndexExpr,
		elem:     node,
	}
}

func WrapArithmeticExpr(node *ArithmeticExpr) *Node {
	return &Node{
		NodeType: TypeArithmeticExpr,
		elem:     node,
	}
}

func WrapConditionExpr(node *ConditionalExpr) *Node {
	return &Node{
		NodeType: TypeConditionalExpr,
		elem:     node,
	}
}

func WrapInExpr(node *InExpr) *Node {
	return &Node{
		NodeType: TypeInExpr,
		elem:     node,
	}
}

func WrapUnaryExpr(node *UnaryExpr) *Node {
	return &Node{
		NodeType: TypeUnaryExpr,
		elem:     node,
	}
}

func WrapAssignmentStmt(node *AssignmentExpr) *Node {
	return &Node{
		NodeType: TypeAssignmentExpr,
		elem:     node,
	}
}

func WrapCallExpr(node *CallExpr) *Node {
	return &Node{
		NodeType: TypeCallExpr,
		elem:     node,
	}
}

func WrapSliceExpr(node *SliceExpr) *Node {
	return &Node{
		NodeType: TypeSliceExpr,
		elem:     node,
	}
}
func WrapIfelseStmt(node *IfelseStmt) *Node {
	return &Node{
		NodeType: TypeIfelseStmt,
		elem:     node,
	}
}

func WrapForStmt(node *ForStmt) *Node {
	return &Node{
		NodeType: TypeForStmt,
		elem:     node,
	}
}

func WrapForInStmt(node *ForInStmt) *Node {
	return &Node{
		NodeType: TypeForInStmt,
		elem:     node,
	}
}

func WrapContinueStmt(node *ContinueStmt) *Node {
	return &Node{
		NodeType: TypeContinueStmt,
		elem:     node,
	}
}

func WrapBreakStmt(node *BreakStmt) *Node {
	return &Node{
		NodeType: TypeBreakStmt,
		elem:     node,
	}
}

func WrapeBlockStmt(node *BlockStmt) *Node {
	return &Node{
		NodeType: TypeBlockStmt,
		elem:     node,
	}
}

func NodeStartPos(node *Node) token.LnColPos {
	if node == nil || node.elem == nil {
		return token.InvalidLnColPos
	}
	switch node.NodeType {
	case TypeInvalid:
		return token.InvalidLnColPos
	case TypeIdentifier:
		return node.Identifier().Start
	case TypeStringLiteral:
		return node.StringLiteral().Start
	case TypeIntegerLiteral:
		return node.IntegerLiteral().Start
	case TypeFloatLiteral:
		return node.FloatLiteral().Start
	case TypeBoolLiteral:
		return node.BoolLiteral().Start
	case TypeNilLiteral:
		return node.NilLiteral().Start

	case TypeListLiteral:
		return node.ListLiteral().LBracket
	case TypeMapLiteral:
		return node.MapLiteral().LBrace

	case TypeParenExpr:
		return node.ParenExpr().LParen

	case TypeAttrExpr:
		return node.AttrExpr().Start

	case TypeIndexExpr:
		return node.IndexExpr().Obj.Start

	case TypeUnaryExpr:
		return node.UnaryExpr().OpPos
	case TypeArithmeticExpr:
		return node.ArithmeticExpr().LHS.StartPos()
	case TypeConditionalExpr:
		return node.ConditionalExpr().LHS.StartPos()
	case TypeAssignmentExpr:
		return node.AssignmentExpr().LHS[0].StartPos()

	case TypeCallExpr:
		return node.CallExpr().NamePos

	case TypeSliceExpr:
		return node.SliceExpr().LBracket
	case TypeBlockStmt:
		return node.BlockStmt().LBracePos

	case TypeIfelseStmt:
		if len(node.IfelseStmt().IfList) > 0 {
			return node.IfelseStmt().IfList[0].Start
		} else {
			return token.InvalidLnColPos
		}

	case TypeForStmt:
		return node.ForStmt().ForPos
	case TypeForInStmt:
		return node.ForInStmt().ForPos
	case TypeContinueStmt:
		return node.ContinueStmt().Start
	case TypeBreakStmt:
		return node.BreakStmt().Start
	}
	return token.InvalidLnColPos
}

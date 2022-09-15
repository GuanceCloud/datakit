// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ast

import (
	"strings"
)

type IfelseStmt struct {
	IfList IfList
	Else   Stmts
}

func (e *IfelseStmt) IsExpr() bool {
	return false
}

func (e *IfelseStmt) String() string {
	arr := []string{e.IfList.String()}
	if len(e.Else) != 0 {
		arr = append(arr, "else", "{", e.Else.String(), "}")
	}
	return strings.Join(arr, " ")
}

// IfList index [0] is IF, [1..end] is ELIF.
type IfList []*IfStmtElem

func (e IfList) String() string {
	if len(e) == 0 {
		return ""
	}
	arr := []string{"if", e[0].String()}
	for i := 1; i < len(e); i++ {
		arr = append(arr, "elif", e[i].String())
	}
	return strings.Join(arr, " ")
}

type IfStmtElem struct {
	Condition *Node
	Stmts     Stmts
}

func (e *IfStmtElem) String() string {
	arr := []string{e.Condition.String(), "{", e.Stmts.String(), "}"}
	return strings.Join(arr, " ")
}

type BreakStmt struct{}

func (e *BreakStmt) String() string {
	return "break"
}

type ContinueStmt struct{}

func (e *ContinueStmt) String() string {
	return "continue"
}

type ForInStmt struct {
	Varb *Node
	Iter *Node
	Body Stmts
}

func (e *ForInStmt) String() string {
	return "for in stmt"
}

type ForStmt struct {
	// init
	Init *Node

	// step1: -> step2 or break
	Cond *Node

	// step3: -> step1
	Loop *Node

	// step2: -> step3
	Body Stmts
}

func (e *ForStmt) String() string {
	return "for stmt"
}

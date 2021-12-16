package parser

import (
	"fmt"
	"sort"
	"strings"
)

type Node interface {
	String() string
}

type Stmts []Node

type AssignmentStmt struct {
	LHS, RHS Node
}

type KwArgs map[string]Node

type FuncStmt struct {
	Name    string
	RunOk   bool
	Param   []Node
	KwParam KwArgs
}

type IfelseStmt struct {
	IfList IfList
	Else   Stmts
}

// IfList index [0] is IF, [1..end] is ELIF.
type IfList []*IfExpr

type IfExpr struct {
	Condition Node
	Stmts     Stmts
}

type ConditionalExpr struct {
	Op       ItemType
	LHS, RHS Node
}

type ComputationExpr struct {
	Op       ItemType
	LHS, RHS Node
}

type AttrExpr struct {
	Obj  Node
	Attr Node
}

type IndexExpr struct {
	Obj   *Identifier
	Index []int64
}

type ParenExpr struct{ Param Node }

type FuncArgList []Node

type NodeList []Node

type BoolLiteral struct{ Val bool }

type NilLiteral struct{}

type Identifier struct{ Name string }

type Regex struct{ Regex string }

type Jspath struct{ Jspath string }

type StringLiteral struct{ Val string }

type NumberLiteral struct {
	IsInt bool
	Float float64
	Int   int64
}

func (e *AssignmentStmt) String() string { return fmt.Sprintf("%s = %s", e.LHS, e.RHS) }
func (e *Identifier) String() string     { return e.Name }
func (e *StringLiteral) String() string  { return fmt.Sprintf("'%s'", e.Val) }
func (e *BoolLiteral) String() string    { return fmt.Sprintf("%v", e.Val) }
func (e *NilLiteral) String() string     { return "nil" }
func (e *Regex) String() string          { return fmt.Sprintf("re('%s')", e.Regex) }
func (e *Jspath) String() string         { return fmt.Sprintf("jp('%s')", e.Jspath) }
func (e *ParenExpr) String() string      { return fmt.Sprintf("(%s)", e.Param.String()) }

func (e *ComputationExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.LHS.String(), e.Op.String(), e.RHS.String())
}

func (e *ConditionalExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.LHS.String(), e.Op.String(), e.RHS.String())
}

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

func (e *NumberLiteral) String() string {
	if e.IsInt {
		return fmt.Sprintf("%d", e.Int)
	} else {
		return fmt.Sprintf("%f", e.Float)
	}
}

func (e *FuncStmt) String() string {
	arr := []string{}
	for _, n := range e.Param {
		arr = append(arr, n.String())
	}
	if len(e.KwParam) != 0 {
		arr = append(arr, e.KwParam.String())
	}
	return fmt.Sprintf("%s(%s)", strings.ToLower(e.Name), strings.Join(arr, ", "))
}

func (e *IfExpr) String() string {
	arr := []string{e.Condition.String(), "{", e.Stmts.String(), "}"}
	return strings.Join(arr, " ")
}

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

func (e *IfelseStmt) String() string {
	arr := []string{e.IfList.String()}
	if len(e.Else) != 0 {
		arr = append(arr, "else", "{", e.Else.String(), "}")
	}
	return strings.Join(arr, " ")
}

func (e NodeList) String() string {
	arr := []string{}
	for _, x := range e {
		arr = append(arr, x.String())
	}
	return strings.Join(arr, ", ")
}

func (e FuncArgList) String() string {
	arr := []string{}
	for _, x := range e {
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

type PositionRange struct {
	Start, End Pos
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
)

type ContextCheck struct {
	forstmt []bool

	breakstmt    bool
	continuestmt bool
}

func RunStmtsCheck(ctx *Context, ctxCheck *ContextCheck, nodes ast.Stmts) error {
	for _, node := range nodes {
		if err := RunStmtCheck(ctx, ctxCheck, node); err != nil {
			return err
		}
	}
	return nil
}

func RunStmtCheck(ctx *Context, ctxCheck *ContextCheck, node *ast.Node) error {
	if node == nil {
		return nil
	}
	switch node.NodeType { //nolint:exhaustive
	case ast.TypeInvaild:
		// skip
	case ast.TypeIdentifier:
		// skip
	case ast.TypeStringLiteral:
		// skip
	case ast.TypeNumberLiteral:
		// skip
	case ast.TypeBoolLiteral:
		// skip
	case ast.TypeNilLiteral:
		// skip
	case ast.TypeListInitExpr:
		return RunListInitExprCheck(ctx, ctxCheck, node.ListInitExpr)
	case ast.TypeMapInitExpr:
		return RunMapInitExprCheck(ctx, ctxCheck, node.MapInitExpr)

	case ast.TypeParenExpr:
		return RunParenExprCheck(ctx, ctxCheck, node.ParenExpr)

	case ast.TypeAttrExpr:
		return RunAttrExprCheck(ctx, ctxCheck, node.AttrExpr)
	case ast.TypeIndexExpr:
		return RunIndexExprGetCheck(ctx, ctxCheck, node.IndexExpr)

	case ast.TypeArithmeticExpr:
		return RunArithmeticExprCheck(ctx, ctxCheck, node.ArithmeticExpr)
	case ast.TypeConditionalExpr:
		return RunConditionExprCheck(ctx, ctxCheck, node.ConditionalExpr)

	case ast.TypeAssignmentExpr:
		return RunAssignmentExprCheck(ctx, ctxCheck, node.AssignmentExpr)

	case ast.TypeCallExpr:
		return RunCallExprCheck(ctx, ctxCheck, node.CallExpr)

	case ast.TypeIfelseStmt:
		return RunIfElseStmtCheck(ctx, ctxCheck, node.IfelseStmt)
	case ast.TypeForStmt:
		return RunForStmtCheck(ctx, ctxCheck, node.ForStmt)
	case ast.TypeForInStmt:
		return RunForInStmtCheck(ctx, ctxCheck, node.ForInStmt)
	case ast.TypeContinueStmt:
		return RunContinueStmtCheck(ctx, ctxCheck, node.ContinueStmt)
	case ast.TypeBreakStmt:
		return RunBreakStmtCheck(ctx, ctxCheck, node.BreakStmt)
	}

	return nil
}

func RunListInitExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.ListInitExpr) error {
	for _, v := range expr.List {
		if err := RunStmtCheck(ctx, ctxCheck, v); err != nil {
			return err
		}
	}
	return nil
}

func RunMapInitExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.MapInitExpr) error {
	for _, v := range expr.KeyValeList {
		switch v[0].NodeType { //nolint:exhaustive
		case ast.TypeNumberLiteral, ast.TypeBoolLiteral, ast.TypeNilLiteral,
			ast.TypeListInitExpr, ast.TypeMapInitExpr:
			return fmt.Errorf("map key expect string")
		default:
		}
		if err := RunStmtCheck(ctx, ctxCheck, v[0]); err != nil {
			return err
		}
		if err := RunStmtCheck(ctx, ctxCheck, v[1]); err != nil {
			return err
		}
	}
	return nil
}

func RunParenExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.ParenExpr) error {
	return RunStmtCheck(ctx, ctxCheck, expr.Param)
}

func RunAttrExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.AttrExpr) error {
	if err := RunStmtCheck(ctx, ctxCheck, expr.Obj); err != nil {
		return err
	}

	if err := RunStmtCheck(ctx, ctxCheck, expr.Attr); err != nil {
		return err
	}
	return nil
}

func RunArithmeticExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.ArithmeticExpr) error {
	if err := RunStmtCheck(ctx, ctxCheck, expr.LHS); err != nil {
		return err
	}
	if err := RunStmtCheck(ctx, ctxCheck, expr.RHS); err != nil {
		return err
	}
	return nil
}

func RunConditionExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.ConditionalExpr) error {
	if err := RunStmtCheck(ctx, ctxCheck, expr.LHS); err != nil {
		return err
	}
	if err := RunStmtCheck(ctx, ctxCheck, expr.RHS); err != nil {
		return err
	}
	return nil
}

func RunIndexExprGetCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.IndexExpr) error {
	for _, v := range expr.Index {
		if err := RunStmtCheck(ctx, ctxCheck, v); err != nil {
			return err
		}
	}
	return nil
}

func RunCallExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.CallExpr) error {
	_, ok := ctx.GetFuncCall(expr.Name)
	if !ok {
		return fmt.Errorf("unsupported func: `%v'", expr.Name)
	}
	funcCheck, ok := ctx.GetFuncCheck(expr.Name)
	if !ok {
		return fmt.Errorf("not found check for func: `%v'", expr.Name)
	}

	return funcCheck(ctx, expr)
}

func RunAssignmentExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.AssignmentExpr) error {
	if err := RunStmtCheck(ctx, ctxCheck, expr.RHS); err != nil {
		return err
	}
	if err := RunStmtCheck(ctx, ctxCheck, expr.LHS); err != nil {
		return err
	}
	return nil
}

func RunIfElseStmtCheck(ctx *Context, ctxCheck *ContextCheck, stmt *ast.IfelseStmt) error {
	ctx.StackEnterNew()
	defer ctx.StackExitCur()

	for _, ifelem := range stmt.IfList {
		if err := RunStmtCheck(ctx, ctxCheck, ifelem.Condition); err != nil {
			return err
		}

		ctx.StackEnterNew()
		if err := RunStmtsCheck(ctx, ctxCheck, ifelem.Stmts); err != nil {
			return err
		}
		ctx.StackExitCur()
	}

	ctx.StackEnterNew()
	if err := RunStmtsCheck(ctx, ctxCheck, stmt.Else); err != nil {
		return err
	}
	ctx.StackExitCur()

	return nil
}

func RunForStmtCheck(ctx *Context, ctxCheck *ContextCheck, stmt *ast.ForStmt) error {
	ctx.StackEnterNew()
	defer ctx.StackExitCur()

	// check init
	if err := RunStmtCheck(ctx, ctxCheck, stmt.Init); err != nil {
		return err
	}

	// check cond
	if err := RunStmtCheck(ctx, ctxCheck, stmt.Cond); err != nil {
		return err
	}

	ctxCheck.forstmt = append(ctxCheck.forstmt, true)

	// check body
	ctx.StackEnterNew()
	if err := RunStmtsCheck(ctx, ctxCheck, stmt.Body); err != nil {
		ctx.StackExitCur()
		return err
	}
	ctx.StackExitCur()

	// check loop
	if err := RunStmtCheck(ctx, ctxCheck, stmt.Loop); err != nil {
		return err
	}

	ctxCheck.forstmt = ctxCheck.forstmt[0 : len(ctxCheck.forstmt)-1]
	ctxCheck.breakstmt = false
	ctxCheck.continuestmt = false
	return nil
}

func RunForInStmtCheck(ctx *Context, ctxCheck *ContextCheck, stmt *ast.ForInStmt) error {
	ctx.StackEnterNew()
	defer ctx.StackExitCur()

	// check varb
	switch stmt.Varb.NodeType { //nolint:exhaustive
	case ast.TypeIdentifier:
	default:
		return fmt.Errorf("node type expect identifier, but %s", stmt.Varb.NodeType)
	}

	// check iter
	if err := RunStmtCheck(ctx, ctxCheck, stmt.Iter); err != nil {
		return err
	}

	ctxCheck.forstmt = append(ctxCheck.forstmt, true)

	// check body
	ctx.StackEnterNew()
	if err := RunStmtsCheck(ctx, ctxCheck, stmt.Body); err != nil {
		return err
	}
	ctx.StackExitCur()

	ctxCheck.forstmt = ctxCheck.forstmt[0 : len(ctxCheck.forstmt)-1]
	ctxCheck.breakstmt = false
	ctxCheck.continuestmt = false
	return nil
}

func RunBreakStmtCheck(ctx *Context, ctxCheck *ContextCheck, stmt *ast.BreakStmt) error {
	if len(ctxCheck.forstmt) == 0 {
		return fmt.Errorf("break not in loop")
	}
	ctxCheck.breakstmt = true
	return nil
}

func RunContinueStmtCheck(ctx *Context, ctxCheck *ContextCheck, stmt *ast.ContinueStmt) error {
	if len(ctxCheck.forstmt) == 0 {
		return fmt.Errorf("continue not in loop")
	}
	ctxCheck.continuestmt = true
	return nil
}

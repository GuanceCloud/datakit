// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"fmt"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

type ContextCheck struct {
	forstmt []bool

	breakstmt    bool
	continuestmt bool
}

func RunStmtsCheck(ctx *Context, ctxCheck *ContextCheck, nodes ast.Stmts) *errchain.PlError {
	for _, node := range nodes {
		if err := RunStmtCheck(ctx, ctxCheck, node); err != nil {
			return err
		}
	}
	return nil
}

func RunStmtCheck(ctx *Context, ctxCheck *ContextCheck, node *ast.Node) *errchain.PlError {
	if node == nil {
		return nil
	}
	switch node.NodeType {
	case ast.TypeInvalid:
		// skip
	case ast.TypeIdentifier:
		// skip
	case ast.TypeStringLiteral:
		// skip
	case ast.TypeFloatLiteral:
		// skip
	case ast.TypeIntegerLiteral:
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
	case ast.TypeUnaryExpr:
		return RunUnaryExprCheck(ctx, ctxCheck, node.UnaryExpr)
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

func RunListInitExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.ListInitExpr) *errchain.PlError {
	for _, v := range expr.List {
		if err := RunStmtCheck(ctx, ctxCheck, v); err != nil {
			return err.ChainAppend(ctx.name, expr.LBracket)
		}
	}
	return nil
}

func RunMapInitExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.MapInitExpr) *errchain.PlError {
	for _, v := range expr.KeyValeList {
		switch v[0].NodeType { //nolint:exhaustive
		case ast.TypeFloatLiteral, ast.TypeIntegerLiteral,
			ast.TypeBoolLiteral, ast.TypeNilLiteral,
			ast.TypeListInitExpr, ast.TypeMapInitExpr:
			return NewRunError(ctx, "map key expect string",
				ast.NodeStartPos(v[0]))
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

func RunParenExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.ParenExpr) *errchain.PlError {
	return RunStmtCheck(ctx, ctxCheck, expr.Param)
}

func RunAttrExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.AttrExpr) *errchain.PlError {
	if err := RunStmtCheck(ctx, ctxCheck, expr.Obj); err != nil {
		return err
	}

	if err := RunStmtCheck(ctx, ctxCheck, expr.Attr); err != nil {
		return err
	}
	return nil
}

func RunArithmeticExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.ArithmeticExpr) *errchain.PlError {
	if err := RunStmtCheck(ctx, ctxCheck, expr.LHS); err != nil {
		return err
	}
	if err := RunStmtCheck(ctx, ctxCheck, expr.RHS); err != nil {
		return err
	}

	// TODO
	// check op

	return nil
}

func RunConditionExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.ConditionalExpr) *errchain.PlError {
	if err := RunStmtCheck(ctx, ctxCheck, expr.LHS); err != nil {
		return err
	}
	if err := RunStmtCheck(ctx, ctxCheck, expr.RHS); err != nil {
		return err
	}
	return nil
}

func RunUnaryExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.UnaryExpr) *errchain.PlError {
	if err := RunStmtCheck(ctx, ctxCheck, expr.RHS); err != nil {
		return err
	}
	return nil
}

func RunIndexExprGetCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.IndexExpr) *errchain.PlError {
	for _, v := range expr.Index {
		if err := RunStmtCheck(ctx, ctxCheck, v); err != nil {
			return err
		}
	}
	return nil
}

func RunCallExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.CallExpr) *errchain.PlError {
	_, ok := ctx.GetFuncCall(expr.Name)
	if !ok {
		return NewRunError(ctx, fmt.Sprintf(
			"unsupported func: `%v`", expr.Name), expr.NamePos)
	}

	if err := RunStmtsCheck(ctx, ctxCheck, expr.Param); err != nil {
		return err.ChainAppend(ctx.name, expr.NamePos)
	}

	funcCheck, ok := ctx.GetFuncCheck(expr.Name)
	if !ok {
		return NewRunError(ctx, fmt.Sprintf(
			"not found check for func: `%v`", expr.Name), expr.NamePos)
	}

	return funcCheck(ctx, expr)
}

func RunAssignmentExprCheck(ctx *Context, ctxCheck *ContextCheck, expr *ast.AssignmentExpr) *errchain.PlError {
	if err := RunStmtCheck(ctx, ctxCheck, expr.RHS); err != nil {
		return err
	}
	if err := RunStmtCheck(ctx, ctxCheck, expr.LHS); err != nil {
		return err
	}
	return nil
}

func RunIfElseStmtCheck(ctx *Context, ctxCheck *ContextCheck, stmt *ast.IfelseStmt) *errchain.PlError {
	ctx.StackEnterNew()
	defer ctx.StackExitCur()

	for _, ifelem := range stmt.IfList {
		if err := RunStmtCheck(ctx, ctxCheck, ifelem.Condition); err != nil {
			return err
		}

		ctx.StackEnterNew()
		if ifelem.Block != nil {
			if err := RunStmtsCheck(ctx, ctxCheck, ifelem.Block.Stmts); err != nil {
				return err
			}
		}
		ctx.StackExitCur()
	}

	ctx.StackEnterNew()
	if stmt.Else != nil {
		if err := RunStmtsCheck(ctx, ctxCheck, stmt.Else.Stmts); err != nil {
			return err
		}
	}
	ctx.StackExitCur()

	return nil
}

func RunForStmtCheck(ctx *Context, ctxCheck *ContextCheck, stmt *ast.ForStmt) *errchain.PlError {
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
	if stmt.Body != nil {
		if err := RunStmtsCheck(ctx, ctxCheck, stmt.Body.Stmts); err != nil {
			ctx.StackExitCur()
			return err
		}
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

func RunForInStmtCheck(ctx *Context, ctxCheck *ContextCheck, stmt *ast.ForInStmt) *errchain.PlError {
	ctx.StackEnterNew()
	defer ctx.StackExitCur()

	// check varb
	switch stmt.Varb.NodeType { //nolint:exhaustive
	case ast.TypeIdentifier:
	default:
		return NewRunError(ctx, fmt.Sprintf("varb node type expect identifier, but %s",
			stmt.Varb.NodeType), stmt.ForPos)
	}

	// check iter
	if err := RunStmtCheck(ctx, ctxCheck, stmt.Iter); err != nil {
		return err
	}

	ctxCheck.forstmt = append(ctxCheck.forstmt, true)

	// check body
	ctx.StackEnterNew()
	if stmt.Body != nil {
		if err := RunStmtsCheck(ctx, ctxCheck, stmt.Body.Stmts); err != nil {
			return err
		}
	}

	ctx.StackExitCur()

	ctxCheck.forstmt = ctxCheck.forstmt[0 : len(ctxCheck.forstmt)-1]
	ctxCheck.breakstmt = false
	ctxCheck.continuestmt = false
	return nil
}

func RunBreakStmtCheck(ctx *Context, ctxCheck *ContextCheck, stmt *ast.BreakStmt) *errchain.PlError {
	if len(ctxCheck.forstmt) == 0 {
		return NewRunError(ctx, "break not in loop", stmt.Start)
	}
	ctxCheck.breakstmt = true
	return nil
}

func RunContinueStmtCheck(ctx *Context, ctxCheck *ContextCheck, stmt *ast.ContinueStmt) *errchain.PlError {
	if len(ctxCheck.forstmt) == 0 {
		return NewRunError(ctx, "continue not in loop", stmt.Start)
	}
	ctxCheck.continuestmt = true
	return nil
}
